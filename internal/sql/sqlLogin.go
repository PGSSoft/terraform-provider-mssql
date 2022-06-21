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

func (s SqlLoginSettings) toSqlOptions(ctx context.Context, conn connection) string {
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
		err := conn.conn.QueryRowContext(ctx, "SELECT DB_NAME(@p1)", s.DefaultDatabaseId).Scan(&defaultDatabaseName)
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
	conn connection
}

func (l sqlLogin) GetId(context.Context) LoginId {
	return l.id
}

func (l sqlLogin) Exists(ctx context.Context) bool {
	const query = "SELECT [name] FROM sys.sql_logins WHERE CONVERT(VARCHAR(85), [sid], 1) = @p1"

	switch err := l.conn.conn.QueryRowContext(ctx, query, l.id).Err(); err {
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
	err := l.conn.conn.QueryRowContext(ctx, `
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
	err := l.conn.conn.QueryRowContext(ctx, "SELECT SUSER_SNAME(CONVERT(VARBINARY(85), @p1, 1))", l.id).Scan(&name)
	if err != nil {
		utils.AddError(ctx, "Failed to retrieve current login name", err)
	}

	return name
}

func (c connection) getLoginId(ctx context.Context, loginName string) LoginId {
	var id sql.NullString
	err := c.conn.QueryRowContext(ctx, "SELECT CONVERT(VARCHAR(85), SUSER_SID(@p1), 1)", loginName).Scan(&id)
	if err != nil {
		utils.AddError(ctx, "Failed to retrieve login ID", err)
	}

	if id.Valid {
		return LoginId(id.String)
	} else {
		return NullLoginId
	}
}

func (c connection) CreateSqlLogin(ctx context.Context, settings SqlLoginSettings) SqlLogin {
	sqlOptions := settings.toSqlOptions(ctx, c)
	if utils.HasError(ctx) {
		return nil
	}
	c.exec(ctx, fmt.Sprintf("CREATE LOGIN [%s] WITH %s", settings.Name, sqlOptions))
	if utils.HasError(ctx) {
		return nil
	}

	return c.GetSqlLoginByName(ctx, settings.Name)
}

func (c connection) GetSqlLogin(_ context.Context, id LoginId) SqlLogin {
	return sqlLogin{conn: c, id: id}
}

func (c connection) GetSqlLoginByName(ctx context.Context, name string) SqlLogin {
	id := c.getLoginId(ctx, name)
	if utils.HasError(ctx) || id == NullLoginId {
		return nil
	}

	return sqlLogin{conn: c, id: id}
}

func (c connection) GetSqlLogins(ctx context.Context) map[LoginId]SqlLogin {
	const errorSummary = "Failed to retrieve list of SQL logins"
	result := map[LoginId]SqlLogin{}

	switch rows, err := c.conn.QueryContext(ctx, "SELECT CONVERT(VARCHAR(85), [sid], 1) FROM sys.sql_logins"); err {
	case sql.ErrNoRows: // ignore
	case nil:
		for rows.Next() {
			var login = sqlLogin{conn: c}
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
