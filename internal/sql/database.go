package sql

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"strings"
	"time"
)

const NullDatabaseId = DatabaseId(-1)

type DatabaseSettings struct {
	Name      string
	Collation string
}

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
	conn Connection
	id   DatabaseId
}

func CreateDatabase(ctx context.Context, conn Connection, settings DatabaseSettings) Database {
	var query strings.Builder
	query.WriteString(fmt.Sprintf("CREATE DATABASE [%s]", settings.Name))

	if settings.Collation != "" {
		query.WriteString(fmt.Sprintf(" COLLATE %s", settings.Collation))
	}

	conn.exec(ctx, query.String())

	if utils.HasError(ctx) {
		return nil
	}

	return GetDatabaseByName(ctx, conn, settings.Name)
}

func GetDatabase(_ context.Context, conn Connection, id DatabaseId) Database {
	return &database{conn: conn, id: id}
}

func GetDatabaseByName(ctx context.Context, conn Connection, name string) Database {
	id := DatabaseId(0)

	if err := conn.getSqlConnection(ctx).QueryRowContext(ctx, "SELECT database_id FROM sys.databases WHERE [name] = @p1", name).Scan(&id); err != nil {
		utils.AddError(ctx, fmt.Sprintf("Failed to retrieve DB ID for name '%s'", name), err)
		return nil
	}

	return GetDatabase(ctx, conn, id)
}

func GetDatabases(ctx context.Context, conn Connection) map[DatabaseId]Database {
	const errorSummary = "Failed to retrieve list of DBs"
	result := map[DatabaseId]Database{}

	switch rows, err := conn.getSqlConnection(ctx).QueryContext(ctx, "SELECT [database_id] FROM sys.databases"); err {
	case sql.ErrNoRows: // ignore
	case nil:
		for rows.Next() {
			var db = database{conn: conn}
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
	err := db.conn.getSqlConnection(ctx).
		QueryRowContext(ctx, "SELECT [name], collation_name FROM sys.databases WHERE [database_id] = @p1", db.id).
		Scan(&settings.Name, &settings.Collation)
	return settings, err
}

func (db database) connect(ctx context.Context) *sql.DB {
	settings := db.GetSettings(ctx)
	if utils.HasError(ctx) {
		return nil
	}

	connDetails := db.conn.getConnectionDetails(ctx)
	connDetails.Database = settings.Name

	connStr, diags := connDetails.getConnectionString(ctx)
	utils.AppendDiagnostics(ctx, diags...)
	if utils.HasError(ctx) {
		return nil
	}

	var err error
	var conn *sql.DB
	for i := time.Second; i <= 5*time.Second; i += time.Second {
		conn, err = sql.Open(connDetails.Auth.getDriverName(), connStr)

		if err == nil {
			return conn
		}

		time.Sleep(i)
	}

	utils.AddError(ctx, "Failed to open DB connection", err)
	return nil
}
