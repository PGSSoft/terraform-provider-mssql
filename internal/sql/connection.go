package sql

import (
	"context"
	"database/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"net/url"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	_ "github.com/microsoft/go-mssqldb"
	_ "github.com/microsoft/go-mssqldb/azuread"
)

var azureSQLEditionPattern = regexp.MustCompile("^SQL Azure.*")

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
	exec(_ context.Context, query string, args ...any) sql.Result
	getConnectionDetails(context.Context) ConnectionDetails
	getSqlConnection(context.Context) *sql.DB
}

type connection struct {
	connDetails ConnectionDetails
	conn        *sql.DB
}

func (cd ConnectionDetails) Open(ctx context.Context) (Connection, diag.Diagnostics) {
	cs, diags := cd.getConnectionString(ctx)
	db, err := sql.Open(cd.Auth.getDriverName(), cs)

	if err != nil {
		diags.AddError("Could not connect to SQL endpoint", err.Error())
	}

	return connection{conn: db, connDetails: cd}, diags
}

func (c connection) IsAzure(ctx context.Context) bool {
	var edition string
	if err := c.conn.QueryRowContext(ctx, "SELECT SERVERPROPERTY('edition')").Scan(&edition); err != nil {
		utils.AddError(ctx, "Failed to determine server edition", err)
	}
	return azureSQLEditionPattern.MatchString(edition)
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

func (c connection) exec(ctx context.Context, query string, args ...any) sql.Result {
	res, err := c.conn.ExecContext(ctx, query, args...)

	if err != nil {
		utils.AddError(ctx, "Could not execute SQL", err)
	}

	return res
}

func (c connection) getConnectionDetails(context.Context) ConnectionDetails {
	return c.connDetails
}

func (c connection) getSqlConnection(context.Context) *sql.DB {
	return c.conn
}
