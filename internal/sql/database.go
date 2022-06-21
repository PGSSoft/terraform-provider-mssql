package sql

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"strings"
)

const NullDatabaseId = DatabaseId(-1)

type DatabaseSettings struct {
	Name      string
	Collation string
}

type DatabaseId int

type Database interface {
	GetConnection(context.Context) Connection
	GetId(context.Context) DatabaseId
	Exists(context.Context) bool
	GetSettings(context.Context) DatabaseSettings
	Rename(_ context.Context, name string)
	SetCollation(_ context.Context, collation string)
	Drop(context.Context)
	connect(context.Context) *sql.DB
}

type database struct {
	conn connection
	id   DatabaseId
}

func (c connection) CreateDatabase(ctx context.Context, settings DatabaseSettings) Database {
	var query strings.Builder
	query.WriteString(fmt.Sprintf("CREATE DATABASE [%s]", settings.Name))

	if settings.Collation != "" {
		query.WriteString(fmt.Sprintf(" COLLATE %s", settings.Collation))
	}

	c.exec(ctx, query.String())

	if utils.HasError(ctx) {
		return nil
	}

	return c.GetDatabaseByName(ctx, settings.Name)
}

func (c connection) GetDatabase(_ context.Context, id DatabaseId) Database {
	return &database{conn: c, id: id}
}

func (c connection) GetDatabaseByName(ctx context.Context, name string) Database {
	id := DatabaseId(0)

	if err := c.conn.QueryRowContext(ctx, "SELECT DB_ID(@p1)", name).Scan(&id); err != nil {
		utils.AddError(ctx, fmt.Sprintf("Failed to retrieve DB ID for name '%s'", name), err)
		return nil
	}

	return c.GetDatabase(ctx, id)
}

func (c connection) GetDatabases(ctx context.Context) map[DatabaseId]Database {
	const errorSummary = "Failed to retrieve list of DBs"
	result := map[DatabaseId]Database{}

	switch rows, err := c.conn.QueryContext(ctx, "SELECT [database_id] FROM sys.databases"); err {
	case sql.ErrNoRows: // ignore
	case nil:
		for rows.Next() {
			var db = database{conn: c}
			err = rows.Scan(&db.id)
			if err != nil {
				utils.AddError(ctx, errorSummary, err)
			}
			result[db.id] = &db
		}
	default:
		utils.AddError(ctx, errorSummary, err)
	}

	return result
}

func (db database) GetConnection(context.Context) Connection {
	return db.conn
}

func (db database) GetId(context.Context) DatabaseId {
	return db.id
}

func (db database) Exists(ctx context.Context) bool {
	switch _, err := db.getSettingsRaw(ctx); err {
	case sql.ErrNoRows:
		return false
	case nil:
		return true
	default:
		utils.AddError(ctx, "Could not retrieve DB info", err)
		return false
	}
}

func (db database) GetSettings(ctx context.Context) DatabaseSettings {
	settings, err := db.getSettingsRaw(ctx)

	if err != nil {
		utils.AddError(ctx, "Could not retrieve DB info", err)
	}

	return settings
}

func (db *database) Rename(ctx context.Context, name string) {
	settings := db.GetSettings(ctx)
	db.conn.exec(ctx, fmt.Sprintf("ALTER DATABASE [%s] MODIFY NAME = %s", settings.Name, name))
}

func (db database) SetCollation(ctx context.Context, collation string) {
	settings := db.GetSettings(ctx)
	db.conn.exec(ctx, fmt.Sprintf("ALTER DATABASE [%s] COLLATE %s", settings.Name, collation))
}

func (db database) Drop(ctx context.Context) {
	settings := db.GetSettings(ctx)
	db.conn.exec(ctx, fmt.Sprintf("DROP DATABASE [%s]", settings.Name))
}

func (db database) getSettingsRaw(ctx context.Context) (DatabaseSettings, error) {
	var settings DatabaseSettings
	err := db.conn.conn.QueryRowContext(ctx, "SELECT [name], collation_name FROM sys.databases WHERE [database_id] = @p1", db.id).Scan(&settings.Name, &settings.Collation)
	return settings, err
}

func (db database) connect(ctx context.Context) *sql.DB {
	settings := db.GetSettings(ctx)
	if utils.HasError(ctx) {
		return nil
	}

	connDetails := db.conn.connDetails
	connDetails.Database = settings.Name

	connStr, diags := connDetails.getConnectionString(ctx)
	utils.AppendDiagnostics(ctx, diags...)
	if utils.HasError(ctx) {
		return nil
	}

	conn, err := sql.Open(connDetails.Auth.getDriverName(), connStr)

	if err != nil {
		utils.AddError(ctx, "Failed to open DB connection", err)
		return nil
	}

	return conn
}
