package acctest

import (
	"context"
	"database/sql"
	"fmt"
	sql2 "github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/kofalt/go-memoize"
	"github.com/microsoft/go-mssqldb/azuread"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

const DefaultDbName = "acc-test-db"

func NewContext(t *testing.T, providerFactory func(connection sql2.Connection) provider.Provider) *TestContext {
	testCtx := TestContext{
		Require: require.New(t),
		Assert:  assert.New(t),

		IsAzureTest:      os.Getenv("TF_MSSQL_EDITION") == "azure",
		IsAcceptanceTest: os.Getenv("TF_ACC") == "1",

		AzureADTestGroup: struct {
			Id   string
			Name string
		}{Id: os.Getenv("TF_AZURE_AD_TEST_GROUP_ID"), Name: os.Getenv("TF_AZURE_AD_TEST_GROUP_NAME")},

		AzureTestMSI: struct {
			Name     string
			ObjectId string
			ClientId string
		}{Name: os.Getenv("TF_MSSQL_MSI_NAME"), ObjectId: os.Getenv("TF_MSSQL_MSI_OBJECT_ID"), ClientId: os.Getenv("TF_MSSQL_MSI_CLIENT_ID")},

		t: t,

		sqlElasticPoolName: os.Getenv("TF_MSSQL_ELASTIC_POOL_NAME"),

		connCache:       memoize.NewMemoizer(2*time.Hour, time.Hour),
		createdDBsCache: cache.New(2*time.Hour, time.Hour),

		providerFactory: providerFactory,
	}

	if testCtx.AzureADTestGroup.Id == "" {
		testCtx.AzureADTestGroup.Id = "2d242970-dcf6-4a1d-8abb-e0b167de4e29"
	}

	if testCtx.AzureADTestGroup.Name == "" {
		testCtx.AzureADTestGroup.Name = "terraform-provider-test-group"
	}

	testCtx.connCache.Storage.OnEvicted(func(_ string, conn interface{}) {
		conn.(*sql.DB).Close()
	})

	testCtx.createdDBsCache.OnEvicted(func(_ string, finalizer interface{}) {
		finalizer.(func())()
	})

	if testCtx.IsAcceptanceTest {
		testCtx.setSqlConnectionDetails()
		testCtx.DefaultDBId = panicOnError(testCtx.tryCreateDB(DefaultDbName, false))
		testCtx.openProviderConnection()
	}

	return &testCtx
}

type TestContext struct {
	Require *require.Assertions
	Assert  *assert.Assertions

	IsAzureTest      bool
	IsAcceptanceTest bool
	DefaultDBId      int

	AzureADTestGroup struct {
		Id   string
		Name string
	}

	AzureTestMSI struct {
		Name     string
		ObjectId string
		ClientId string
	}

	t *testing.T

	sqlDriverName      string
	sqlConnString      url.URL
	sqlElasticPoolName string

	providerConnection sql2.Connection

	connCache       *memoize.Memoizer
	createdDBsCache *cache.Cache

	providerFactory func(connection sql2.Connection) provider.Provider
}

func (t *TestContext) Cleanup() {
	for key := range t.connCache.Storage.Items() {
		t.connCache.Storage.Delete(key)
	}

	t.connCache.Storage.Flush()
}

func (t *TestContext) GetDBConnection(dbName string) *sql.DB {
	return panicOnError(t.tryGetDBConnection(dbName))
}

func (t *TestContext) GetDefaultDBConnection() *sql.DB {
	return t.GetDBConnection(DefaultDbName)
}

func (t *TestContext) GetMasterDBConnection() *sql.DB {
	return t.GetDBConnection("master")
}

func (t *TestContext) SqlCheck(dbName string, check func(conn *sql.DB) error) resource.TestCheckFunc {
	return func(*terraform.State) error {
		return check(t.GetDBConnection(dbName))
	}
}

func (t *TestContext) SqlCheckDefaultDB(check func(conn *sql.DB) error) resource.TestCheckFunc {
	return t.SqlCheck(DefaultDbName, check)
}

func (t *TestContext) SqlCheckMaster(check func(conn *sql.DB) error) resource.TestCheckFunc {
	return t.SqlCheck("master", check)
}

func (t *TestContext) CreateDB(dbName string) int {
	dbId, err := t.tryCreateDB(dbName, false)
	t.Require.NoError(err, "Failed to create DB %q", dbName)
	return dbId
}

func (t *TestContext) ExecDB(dbName string, statFmt string, args ...any) {
	err := t.execDB(dbName, statFmt, args...)
	t.Require.NoError(err, "Failed to execute SQL")
}

func (t *TestContext) ExecDefaultDB(statFmt string, args ...any) {
	t.ExecDB(DefaultDbName, statFmt, args...)
}

func (t *TestContext) ExecMasterDB(statFmt string, args ...any) {
	t.ExecDB("master", statFmt, args...)
}

func (t *TestContext) NewProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	prov := t.providerFactory(t.providerConnection)

	return map[string]func() (tfprotov6.ProviderServer, error){
		"mssql": providerserver.NewProtocol6WithError(prov),
	}
}

func (t *TestContext) Run(name string, test func(testCtx *TestContext)) *TestContext {
	ctx := *t

	t.t.Run(name, func(t *testing.T) {
		ctx.t = t
		ctx.Require = require.New(t)
		ctx.Assert = assert.New(t)
		test(&ctx)
	})

	return t
}

func (t *TestContext) Test(testCase resource.TestCase) {
	testCase.ProtoV6ProviderFactories = t.NewProviderFactories()
	resource.Test(t.t, testCase)
}

func (t *TestContext) FormatId(idParts ...any) string {
	str := make([]string, len(idParts))
	for i, id := range idParts {
		str[i] = fmt.Sprint(id)
	}

	return strings.Join(str, "/")
}

func (t *TestContext) DefaultDbId(idParts ...any) string {
	return t.FormatId(append([]any{t.DefaultDBId}, idParts...)...)
}

func (t *TestContext) setSqlConnectionDetails() {
	t.sqlDriverName = "sqlserver"
	t.sqlConnString = url.URL{
		Scheme: "sqlserver",
		Host:   os.Getenv("TF_MSSQL_HOST"),
		User:   url.UserPassword("sa", os.Getenv("TF_MSSQL_PASSWORD")),
	}

	if t.IsAzureTest {
		t.sqlDriverName = azuread.DriverName
		t.sqlConnString.User = nil

		q := t.sqlConnString.Query()
		q.Set("fedauth", "ActiveDirectoryDefault")
		t.sqlConnString.RawQuery = q.Encode()
	}
}

func (t *TestContext) tryGetDBConnection(dbName string) (*sql.DB, error) {
	u := t.sqlConnString
	q := u.Query()
	q.Set("database", dbName)
	u.RawQuery = q.Encode()

	db, err, _ := t.connCache.Memoize(u.String(), func() (interface{}, error) {
		return sql.Open(t.sqlDriverName, u.String())
	})

	return db.(*sql.DB), err
}

func (t *TestContext) tryCreateDB(dbName string, recreate bool) (int, error) {
	dropDB := func() error {
		return t.execDB("master", "DROP DATABASE [%s]", dbName)
	}

	if id, err := t.tryGetDBId(dbName); err == nil {
		if !recreate {
			return id, err
		}

		if execErr := dropDB(); execErr != nil {
			return id, execErr
		}
	} else if err != sql.ErrNoRows {
		return id, err
	}

	dbOptions := ""
	if t.IsAzureTest {
		dbOptions = fmt.Sprintf("( SERVICE_OBJECTIVE = ELASTIC_POOL ( name = %s ) )", t.sqlElasticPoolName)
	}

	if err := t.execDB("master", "CREATE DATABASE [%[1]s] %[2]s", dbName, dbOptions); err != nil {
		return 0, err
	}

	t.createdDBsCache.SetDefault(dbName, func() { dropDB() })

	return t.tryGetDBId(dbName)
}

func (t *TestContext) tryGetDBId(dbName string) (int, error) {
	var id int
	err := t.GetMasterDBConnection().QueryRow("SELECT database_id FROM sys.databases WHERE [name] = @p1", dbName).Scan(&id)
	return id, err
}

func (t *TestContext) execDB(dbName string, statFmt string, args ...any) error {
	if !t.IsAcceptanceTest {
		return nil
	}

	stat := fmt.Sprintf(statFmt, args...)
	_, err := t.GetDBConnection(dbName).Exec(stat)
	return err
}

func (t *TestContext) openProviderConnection() {
	connDetails := sql2.ConnectionDetails{
		Host: t.sqlConnString.Host,
		Auth: sql2.ConnectionAuthAzure{},
	}

	if !t.IsAzureTest {
		auth := sql2.ConnectionAuthSql{Username: t.sqlConnString.User.Username()}
		auth.Password, _ = t.sqlConnString.User.Password()
		connDetails.Auth = auth
	}

	conn, diags := connDetails.Open(context.Background())

	for _, d := range diags {
		if d.Severity() == diag.SeverityError {
			panic(d)
		}
	}

	t.providerConnection = conn
}

func panicOnError[T any](result T, err error) T {
	if err != nil {
		panic(err)
	}
	return result
}
