package provider

import (
	"bufio"
	"context"
	sql2 "database/sql"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/denisenkom/go-mssqldb/azuread"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/glendc/go-external-ip"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	a "github.com/microsoft/kiota-authentication-azure-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/stretchr/testify/require"
	"io"
	"math/rand"
	"net/url"
	"os"
	"regexp"
	"testing"
	"time"
)

const (
	mssqlSaPassword = "Terraform-acc-test-sa123"
	containerName   = "terraform-mssql-acc-test"
	mappedPort      = 11433
)

var docker *client.Client
var sqlHost = fmt.Sprintf("localhost:%d", mappedPort)
var azureSubscription = os.Getenv("TF_AZURE_SUBSCRIPTION_ID")
var azureResourceGroup = os.Getenv("TF_AZURE_RESOURCE_GROUP")
var imgTag = os.Getenv("TF_MSSQL_IMG_TAG")
var isAzureTest = imgTag == "azure-sql"
var azureServerName string

func init() {
	rand.Seed(time.Now().UnixNano())
	azureServerName = fmt.Sprintf("tfmssqltest%d", rand.Intn(1000))
	d, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	docker = d
}

type testRunner struct {
	m *testing.M
}

func (r *testRunner) Run() int {
	if imgTag == "" {
		imgTag = "2019-latest"
	}

	if os.Getenv("TF_ACC") == "1" {
		if isAzureTest {
			createAzureSQL()
			defer destroyAzureSQL()
		} else {
			startMSSQL(imgTag)
			defer stopMSSQL()
		}
	}

	return r.m.Run()
}

func TestMain(m *testing.M) {
	resource.TestMain(&testRunner{m: m})
}

func panicOnError[T any](result T, err error) T {
	if err != nil {
		panic(err)
	}
	return result
}

func createAzureSQL() {
	fmt.Fprintln(os.Stdout, "Creating Azure SQL instance..")

	ctx := context.Background()
	token := panicOnError(azidentity.NewDefaultAzureCredential(nil))

	clientId := os.Getenv("AZURE_CLIENT_ID")
	if clientId == "" {
		auth := panicOnError(a.NewAzureIdentityAuthenticationProvider(token))
		graphAdapter := panicOnError(msgraphsdk.NewGraphRequestAdapter(auth))
		graphClient := msgraphsdk.NewGraphServiceClient(graphAdapter)
		me := panicOnError(graphClient.Me().Get())
		clientId = *me.GetId()
	}

	serverClient := panicOnError(armsql.NewServersClient(azureSubscription, token, nil))

	request := panicOnError(serverClient.BeginCreateOrUpdate(ctx, azureResourceGroup, azureServerName, armsql.Server{
		Location: to.Ptr("WestEurope"),
		Properties: &armsql.ServerProperties{
			Administrators: &armsql.ServerExternalAdministrator{
				AzureADOnlyAuthentication: to.Ptr(true),
				Sid:                       &clientId,
				Login:                     &clientId,
			},
		},
	}, nil))
	response := panicOnError(request.PollUntilDone(ctx, nil))

	poolClient := panicOnError(armsql.NewElasticPoolsClient(azureSubscription, token, nil))
	poolRequest := panicOnError(poolClient.BeginCreateOrUpdate(ctx, azureResourceGroup, azureServerName, azureServerName, armsql.ElasticPool{
		Location: response.Location,
	}, nil))
	panicOnError(poolRequest.PollUntilDone(ctx, nil))

	externalIp := panicOnError(externalip.DefaultConsensus(nil, nil).ExternalIP())
	networkRulesClient := panicOnError(armsql.NewFirewallRulesClient(azureSubscription, token, nil))
	panicOnError(networkRulesClient.CreateOrUpdate(ctx, azureResourceGroup, azureServerName, "test", armsql.FirewallRule{
		Properties: &armsql.ServerFirewallRuleProperties{
			StartIPAddress: to.Ptr(externalIp.String()),
			EndIPAddress:   to.Ptr(externalIp.String()),
		},
	}, nil))
	sqlHost = *response.Server.Properties.FullyQualifiedDomainName
	fmt.Fprintln(os.Stdout, "Azure SQL instance created!")
}

func destroyAzureSQL() {
	ctx := context.Background()
	token := panicOnError(azidentity.NewDefaultAzureCredential(nil))
	client := panicOnError(armsql.NewServersClient(azureSubscription, token, nil))
	panicOnError(client.BeginDelete(ctx, azureResourceGroup, azureServerName, nil))
}

func startMSSQL(imgTag string) {
	const (
		imgName = "mcr.microsoft.com/mssql/server"
	)

	img := fmt.Sprintf("%s:%s", imgName, imgTag)

	ctx := context.Background()

	log, err := docker.ImagePull(ctx, img, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, log)

	stopMSSQL()

	containerConfig := container.Config{
		Image:        img,
		Env:          []string{"ACCEPT_EULA=Y", fmt.Sprintf("MSSQL_SA_PASSWORD=%s", mssqlSaPassword)},
		ExposedPorts: nat.PortSet{"1433/tcp": struct{}{}},
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	}
	hostConfig := container.HostConfig{
		PortBindings: nat.PortMap{
			"1433/tcp": {{HostIP: "0.0.0.0", HostPort: fmt.Sprint(mappedPort)}},
		},
	}
	resp, err := docker.ContainerCreate(ctx, &containerConfig, &hostConfig, nil, nil, containerName)
	if err != nil {
		panic(err)
	}

	if err = docker.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	log, err = docker.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{Follow: true, ShowStdout: true, ShowStderr: true})
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(log)
	scanner.Split(bufio.ScanLines)
	defer log.Close()
	readyPattern := regexp.MustCompile("Recovery is complete")
	for scanner.Scan() {
		if readyPattern.Match(scanner.Bytes()) {
			return
		}
	}

	for i := time.Second; i <= 5*time.Second; i += time.Second {
		var conn *sql2.DB
		conn, err = tryOpenDBConnection("master")

		if err == nil {
			err = conn.QueryRow("SELECT 1;").Err()
			conn.Close()

			if err == nil {
				return
			}
		}

		time.Sleep(i)
	}
	panic(err)
}

func stopMSSQL() {
	ctx := context.Background()
	containers, err := docker.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		panic(err)
	}
	for _, c := range containers {
		for _, name := range c.Names {
			if name == fmt.Sprintf("/%s", containerName) {
				err = docker.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{Force: true})
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func newProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	connDetails := sql.ConnectionDetails{
		Host: sqlHost,
		Auth: sql.ConnectionAuthSql{Username: "sa", Password: mssqlSaPassword},
	}

	if isAzureTest {
		connDetails.Auth = sql.ConnectionAuthAzure{}
	}

	return map[string]func() (tfprotov6.ProviderServer, error){
		"mssql": func() (tfprotov6.ProviderServer, error) {
			connection, diagnostics := connDetails.Open(context.Background())

			for _, d := range diagnostics {
				if d.Severity() == diag.SeverityError {
					return nil, fmt.Errorf("%v", d)
				}
			}

			prov := provider{
				Version: VersionTest,
				Db:      connection,
			}

			return providerserver.NewProtocol6WithError(&prov)()
		},
	}
}

func tryOpenDBConnection(dbName string) (*sql2.DB, error) {
	driverName := "sqlserver"
	u := url.URL{
		Scheme: "sqlserver",
		Host:   sqlHost,
		User:   url.UserPassword("sa", mssqlSaPassword),
	}
	q := u.Query()
	q.Set("database", dbName)

	if isAzureTest {
		driverName = azuread.DriverName
		u.User = nil
		q.Set("fedauth", "ActiveDirectoryDefault")
	}

	u.RawQuery = q.Encode()
	return sql2.Open(driverName, u.String())
}

func openDBConnection(dbName string) *sql2.DB {
	return panicOnError(tryOpenDBConnection(dbName))
}

func withDBConnection(dbName string, f func(conn *sql2.DB)) {
	conn := openDBConnection(dbName)
	defer conn.Close()
	f(conn)
}

func sqlCheck(dbName string, check func(db *sql2.DB) error) resource.TestCheckFunc {
	return func(*terraform.State) error {
		db := openDBConnection(dbName)
		defer db.Close()
		return check(db)
	}
}

func createDB(t *testing.T, name string) int {
	masterConn := openDBConnection("master")
	defer masterConn.Close()

	var dbId int

	dbOptions := ""
	if isAzureTest {
		dbOptions = fmt.Sprintf("( SERVICE_OBJECTIVE = ELASTIC_POOL ( name = %s ) )", azureServerName)
	}

	err := masterConn.QueryRow(fmt.Sprintf(`CREATE DATABASE [%[1]s] %[2]s; SELECT database_id FROM sys.databases WHERE [name] = '%[1]s'`, name, dbOptions)).Scan(&dbId)
	require.NoError(t, err, "creating DB")

	return dbId
}
