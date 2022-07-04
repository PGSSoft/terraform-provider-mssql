package provider

import (
	"bufio"
	"context"
	sql2 "database/sql"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
	"io"
	"net/url"
	"os"
	"regexp"
	"testing"
)

const (
	mssqlSaPassword = "Terraform-acc-test-sa123"
	containerName   = "terraform-mssql-acc-test"
	mappedPort      = 11433
)

var docker *client.Client
var sqlHost = fmt.Sprintf("localhost:%d", mappedPort)

func init() {
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
	startMSSQL()
	defer stopMSSQL()
	return r.m.Run()
}

func TestMain(m *testing.M) {
	resource.TestMain(&testRunner{m: m})
}

func startMSSQL() {
	const (
		imgName = "mcr.microsoft.com/mssql/server"
	)

	imgTag := os.Getenv("TF_MSSQL_IMG_TAG")
	if imgTag == "" {
		imgTag = "2019-latest"
	}

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

func openDBConnection() *sql2.DB {
	u := url.URL{
		Scheme: "sqlserver",
		Host:   sqlHost,
		User:   url.UserPassword("sa", mssqlSaPassword),
	}

	conn, err := sql2.Open("sqlserver", u.String())
	if err != nil {
		panic(err)
	}

	return conn
}

func withDBConnection(f func(conn *sql2.DB)) {
	conn := openDBConnection()
	defer conn.Close()
	f(conn)
}

func sqlCheck(check func(db *sql2.DB) error) resource.TestCheckFunc {
	return func(*terraform.State) error {
		db := openDBConnection()
		defer db.Close()
		return check(db)
	}
}

func createDB(t *testing.T, name string) {
	withDBConnection(func(conn *sql2.DB) {
		_, err := conn.Exec(fmt.Sprintf("CREATE DATABASE [%s]", name))
		require.NoError(t, err, "creating DB")
	})
}
