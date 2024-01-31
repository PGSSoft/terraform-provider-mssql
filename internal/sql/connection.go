package sql

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"regexp"
	"time"

	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/kofalt/go-memoize"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	_ "github.com/microsoft/go-mssqldb"
	_ "github.com/microsoft/go-mssqldb/azuread"
)

var azureSQLEditionPattern = regexp.MustCompile("^SQL Azure.*")

type ServerPermission struct {
	Name            string
	WithGrantOption bool
}

type ServerPermissions map[string]ServerPermission

type ConnectionAuth interface {
	configure(context.Context, *url.URL) diag.Diagnostics
	getDriverName() string
}

type ConnectionDetails struct {
	Host     string
	Database string
	Auth     ConnectionAuth
}

type Connection interface {
	IsAzure(context.Context) bool
	GetPermissions(ctx context.Context, principalId GenericServerPrincipalId) ServerPermissions
	GrantPermission(ctx context.Context, principalId GenericServerPrincipalId, permission ServerPermission)
	RevokePermission(ctx context.Context, principalId GenericServerPrincipalId, permission string)
	exec(_ context.Context, query string, args ...any) sql.Result
	getConnectionDetails(context.Context) ConnectionDetails
	getSqlConnection(context.Context) *sql.DB
	getDBSqlConnection(_ context.Context, dbName string) *sql.DB
	lookupServerPrincipalName(ctx context.Context, id GenericServerPrincipalId) string
	lookupServerPrincipalId(ctx context.Context, name string) GenericServerPrincipalId
}

type connection struct {
	connDetails ConnectionDetails
	conn        *sql.DB
	dbConnCache *memoize.Memoizer
}

func (cd ConnectionDetails) Open(ctx context.Context) (Connection, diag.Diagnostics) {
	cs, diags := cd.getConnectionString(ctx)
	db, err := sql.Open(cd.Auth.getDriverName(), cs)

	if err != nil {
		diags.AddError("Could not connect to SQL endpoint", err.Error())
	}

	conn := connection{conn: db, connDetails: cd, dbConnCache: memoize.NewMemoizer(2*time.Hour, time.Hour)}

	conn.dbConnCache.Storage.OnEvicted(func(_ string, dbConn interface{}) {
		dbConn.(*sql.DB).Close()
	})

	return &conn, diags
}

func (c *connection) IsAzure(ctx context.Context) bool {
	var edition string
	if err := QueryRowContextWithRetry(ctx, c.conn, "SELECT SERVERPROPERTY('edition')").Scan(&edition); err != nil {
		utils.AddError(ctx, "Failed to determine server edition", err)
	}
	return azureSQLEditionPattern.MatchString(edition)
}

func (c *connection) GetPermissions(ctx context.Context, principalId GenericServerPrincipalId) ServerPermissions {
	res, err := QueryContextWithRetry(ctx, c.conn, "SELECT [permission_name], [state] FROM sys.server_permissions WHERE [class]=100 AND [grantee_principal_id]=@p1", principalId)
	perms := ServerPermissions{}

	switch err {
	case sql.ErrNoRows:
	case nil:
		for res.Next() {
			perm := ServerPermission{}
			var state string
			err := res.Scan(&perm.Name, &state)
			utils.AddError(ctx, "Failed to parse permissions result", err)
			perm.WithGrantOption = state == "W"
			perms[perm.Name] = perm
		}
	default:
		utils.AddError(ctx, "Failed to fetch server permissions", err)
	}

	return perms
}

func (c *connection) GrantPermission(ctx context.Context, principalId GenericServerPrincipalId, permission ServerPermission) {
	name := c.lookupServerPrincipalName(ctx, principalId)
	if utils.HasError(ctx) {
		return
	}

	stat := fmt.Sprintf("GRANT %s TO [%s]", permission.Name, name)
	if permission.WithGrantOption {
		stat += " WITH GRANT OPTION"
	}
	c.exec(ctx, stat)
}

func (c *connection) RevokePermission(ctx context.Context, principalId GenericServerPrincipalId, permission string) {
	name := c.lookupServerPrincipalName(ctx, principalId)
	if utils.HasError(ctx) {
		return
	}

	c.exec(ctx, fmt.Sprintf("REVOKE %s FROM [%s] CASCADE", permission, name))
}

func (cd ConnectionDetails) getConnectionString(ctx context.Context) (string, diag.Diagnostics) {
	query := url.Values{
		"app name": []string{"Terraform - mssql provider"},
	}

	if cd.Database != "" {
		query.Set("database", cd.Database)
	}

	u := url.URL{
		Scheme:   "sqlserver",
		Host:     cd.Host,
		RawQuery: query.Encode(),
	}

	diags := cd.Auth.configure(ctx, &u)

	return u.String(), diags
}

func (c *connection) exec(ctx context.Context, query string, args ...any) sql.Result {
	res, err := c.conn.ExecContext(ctx, query, args...)

	if err != nil {
		utils.AddError(ctx, "Could not execute SQL", err)
	}

	return res
}

func (c *connection) getConnectionDetails(context.Context) ConnectionDetails {
	return c.connDetails
}

func (c *connection) getSqlConnection(context.Context) *sql.DB {
	return c.conn
}

func (c *connection) getDBSqlConnection(ctx context.Context, dbName string) *sql.DB {
	connDetails := c.getConnectionDetails(ctx)
	connDetails.Database = dbName

	connStr, diags := connDetails.getConnectionString(ctx)
	utils.AppendDiagnostics(ctx, diags...)
	if utils.HasError(ctx) {
		return nil
	}

	driverName := connDetails.Auth.getDriverName()

	conn, err, _ := c.dbConnCache.Memoize(fmt.Sprintf("%s||%s", driverName, connStr), func() (interface{}, error) {
		var err error
		var conn *sql.DB
		for i := time.Second; i <= 5*time.Second; i += time.Second {
			conn, err = sql.Open(driverName, connStr)

			if err == nil {
				return conn, nil
			}

			time.Sleep(i)
		}

		return nil, err
	})

	if err != nil {
		utils.AddError(ctx, "Failed to open DB connection", err)
		return nil
	}

	return conn.(*sql.DB)
}

func (c *connection) lookupServerPrincipalName(ctx context.Context, id GenericServerPrincipalId) string {
	var name string
	err := QueryRowContextWithRetry(ctx, c.conn, "SELECT [name] FROM sys.server_principals WHERE [principal_id]=@p1", id).Scan(&name)
	utils.AddError(ctx, "Failed to lookup server principal name", err)
	return name
}

func (c *connection) lookupServerPrincipalId(ctx context.Context, name string) GenericServerPrincipalId {
	var id GenericServerPrincipalId
	err := QueryRowContextWithRetry(ctx, c.conn, "SELECT [principal_id] FROM sys.server_principals WHERE [name]=@p1", name).Scan(&id)
	utils.AddError(ctx, "Failed to lookup server principal ID", err)
	return id
}
