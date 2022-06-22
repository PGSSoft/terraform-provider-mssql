package sql

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"strings"
)

type LoginId string

const NullLoginId LoginId = "<null>"

type SqlLoginSettings struct {
	Name                    string
	Password                string
	MustChangePassword      bool
	DefaultDatabaseId       DatabaseId
	DefaultLanguage         string
	CheckPasswordExpiration bool
	CheckPasswordPolicy     bool
}

func (s SqlLoginSettings) toSqlOptions(ctx context.Context, conn Connection) string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("PASSWORD='%s'", s.Password))

	if s.MustChangePassword {
		builder.WriteString(" MUST_CHANGE")
	}

	var addOption = func(name string, value string) {
		if value != "" {
			builder.WriteString(fmt.Sprintf(", %s=[%s]", name, value))
		}
	}

	var addOptionFlag = func(name string, value bool) {
		builder.WriteString(fmt.Sprintf(", %s=", name))
		if value {
			builder.WriteString("ON")
		} else {
			builder.WriteString("OFF")
		}
	}

	var defaultDatabaseName string
	if s.DefaultDatabaseId != DatabaseId(0) {
		err := conn.getSqlConnection(ctx).QueryRowContext(ctx, "SELECT DB_NAME(@p1)", s.DefaultDatabaseId).Scan(&defaultDatabaseName)
		if err != nil {
			utils.AddError(ctx, "Failed to retrieve DB name for given ID", err)
			return ""
		}
	}

	addOption("DEFAULT_DATABASE", defaultDatabaseName)
	addOption("DEFAULT_LANGUAGE", s.DefaultLanguage)

	addOptionFlag("CHECK_EXPIRATION", s.CheckPasswordExpiration)
	addOptionFlag("CHECK_POLICY", s.CheckPasswordPolicy)

	return builder.String()
}

func getLoginId(ctx context.Context, conn Connection, loginName string) LoginId {
	var id sql.NullString
	err := conn.getSqlConnection(ctx).QueryRowContext(ctx, "SELECT CONVERT(VARCHAR(85), SUSER_SID(@p1), 1)", loginName).Scan(&id)
	if err != nil {
		utils.AddError(ctx, "Failed to retrieve login ID", err)
	}

	if id.Valid {
		return LoginId(id.String)
	} else {
		return NullLoginId
	}
}

type SqlLogin interface {
	GetId(context.Context) LoginId
	Exists(context.Context) bool
	GetSettings(context.Context) SqlLoginSettings
	UpdateSettings(ctx context.Context, settings SqlLoginSettings)
	Drop(ctx context.Context)
	getName(ctx context.Context) string
}

type sqlLogin struct {
	id   LoginId
	conn Connection
}

func GetSqlLogin(_ context.Context, conn Connection, id LoginId) SqlLogin {
	return sqlLogin{conn: conn, id: id}
}

func GetSqlLoginByName(ctx context.Context, conn Connection, name string) SqlLogin {
	id := getLoginId(ctx, conn, name)
	if utils.HasError(ctx) || id == NullLoginId {
		return nil
	}

	return sqlLogin{conn: conn, id: id}
}

func GetSqlLogins(ctx context.Context, conn Connection) map[LoginId]SqlLogin {
	const errorSummary = "Failed to retrieve list of SQL logins"
	result := map[LoginId]SqlLogin{}

	switch rows, err := conn.getSqlConnection(ctx).QueryContext(ctx, "SELECT CONVERT(VARCHAR(85), [sid], 1) FROM sys.sql_logins"); err {
	case sql.ErrNoRows: // ignore
	case nil:
		for rows.Next() {
			var login = sqlLogin{conn: conn}
			err := rows.Scan(&login.id)
			if err != nil {
				utils.AddError(ctx, errorSummary, err)
			}
			result[login.id] = login
		}
	default:
		utils.AddError(ctx, errorSummary, err)
	}

	return result
}

func CreateSqlLogin(ctx context.Context, conn Connection, settings SqlLoginSettings) SqlLogin {
	sqlOptions := settings.toSqlOptions(ctx, conn)
	if utils.HasError(ctx) {
		return nil
	}
	conn.exec(ctx, fmt.Sprintf("CREATE LOGIN [%s] WITH %s", settings.Name, sqlOptions))
	if utils.HasError(ctx) {
		return nil
	}

	return GetSqlLoginByName(ctx, conn, settings.Name)
}

func (l sqlLogin) GetId(context.Context) LoginId {
	return l.id
}

func (l sqlLogin) Exists(ctx context.Context) bool {
	const query = "SELECT [name] FROM sys.sql_logins WHERE CONVERT(VARCHAR(85), [sid], 1) = @p1"

	switch err := l.conn.getSqlConnection(ctx).QueryRowContext(ctx, query, l.id).Err(); err {
	case sql.ErrNoRows:
		return false
	case nil:
		return true
	default:
		utils.AddError(ctx, "Failed to check if login exists", err)
		return false
	}
}

func (l sqlLogin) GetSettings(ctx context.Context) SqlLoginSettings {
	var settings SqlLoginSettings
	err := l.conn.getSqlConnection(ctx).QueryRowContext(ctx, `
SELECT 
    [name], 
    [password_hash], 
    LOGINPROPERTY([name], 'IsMustChange') AS is_must_change, 
    DB_ID([default_database_name]) AS default_database_id, 
    [default_language_name], 
    [is_expiration_checked], 
    [is_policy_checked] 
FROM sys.sql_logins 
WHERE CONVERT(VARCHAR(85), [sid], 1) = @p1`, l.id).
		Scan(
			&settings.Name,
			&settings.Password,
			&settings.MustChangePassword,
			&settings.DefaultDatabaseId,
			&settings.DefaultLanguage,
			&settings.CheckPasswordExpiration,
			&settings.CheckPasswordPolicy)

	if err != nil {
		utils.AddError(ctx, "Failed to retrieve DB settings", err)
	}

	return settings
}

func (l sqlLogin) UpdateSettings(ctx context.Context, settings SqlLoginSettings) {
	sqlOptions := settings.toSqlOptions(ctx, l.conn)
	if utils.HasError(ctx) {
		return
	}

	currentName := l.getName(ctx)
	if utils.HasError(ctx) {
		return
	}

	l.conn.exec(ctx, fmt.Sprintf("ALTER LOGIN [%s] WITH %s, NAME=[%s]", currentName, sqlOptions, settings.Name))
}

func (l sqlLogin) Drop(ctx context.Context) {
	currentName := l.getName(ctx)
	if utils.HasError(ctx) {
		return
	}

	l.conn.exec(ctx, fmt.Sprintf("DROP LOGIN [%s]", currentName))
}

func (l sqlLogin) getName(ctx context.Context) string {
	var name string
	err := l.conn.getSqlConnection(ctx).QueryRowContext(ctx, "SELECT SUSER_SNAME(CONVERT(VARBINARY(85), @p1, 1))", l.id).Scan(&name)
	if err != nil {
		utils.AddError(ctx, "Failed to retrieve current login name", err)
	}

	return name
}
