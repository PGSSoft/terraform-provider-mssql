package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"strings"
)

type UserSettings struct {
	Name    string
	LoginId LoginId
}

type User interface {
	GetId(context.Context) UserId
	GetDatabaseId(context.Context) DatabaseId
	GetSettings(context.Context) UserSettings
	Drop(context.Context)
	UpdateSettings(context.Context, UserSettings)
}

func CreateUser(ctx context.Context, db Database, settings UserSettings) User {
	return WithConnection(ctx, db.connect, func(conn *sql.DB) User {
		sqlStat := strings.Builder{}
		sqlStat.WriteString(fmt.Sprintf("CREATE USER [%s]", settings.Name))

		loginName := GetSqlLogin(ctx, db.GetConnection(ctx), settings.LoginId).getName(ctx)
		if utils.HasError(ctx) {
			return nil
		}

		sqlStat.WriteString(fmt.Sprintf(" FOR LOGIN [%s]", loginName))

		if _, err := conn.ExecContext(ctx, sqlStat.String()); err != nil {
			utils.AddError(ctx, "Failed to create user", err)
			return nil
		}

		return GetUserByName(ctx, db, settings.Name)
	})
}

func GetUser(_ context.Context, db Database, id UserId) User {
	return user{db: db, id: id}
}

func GetUserByName(ctx context.Context, db Database, name string) User {
	return WithConnection(ctx, db.connect, func(conn *sql.DB) User {
		user := user{db: db}
		id := sql.NullInt32{}
		err := conn.QueryRowContext(ctx, "SELECT USER_ID(@p1)", name).Scan(&id)
		if err != nil {
			utils.AddError(ctx, "Failed to resolve user ID", err)
			return nil
		}

		if !id.Valid {
			utils.AddError(ctx, "User does not exist", errors.New("user does not exist"))
			return nil
		}

		user.id = UserId(id.Int32)
		return user
	})
}

func GetUsers(ctx context.Context, db Database) map[UserId]User {
	const errorSummary = "Failed to retrieve list of SQL users"

	return WithConnection(ctx, db.connect, func(conn *sql.DB) map[UserId]User {
		result := map[UserId]User{}

		switch res, err := conn.QueryContext(ctx, "SELECT [principal_id] FROM sys.database_principals WHERE [type] = 'S' AND [sid] IS NOT NULL"); err {
		case sql.ErrNoRows: //ignore
		case nil:
			for res.Next() {
				user := user{db: db}
				err := res.Scan(&user.id)
				if err != nil {
					utils.AddError(ctx, errorSummary, err)
				}
				result[user.id] = user
			}
		default:
			utils.AddError(ctx, errorSummary, err)
		}

		return result
	})
}

type user struct {
	db Database
	id UserId
}

func (u user) GetId(context.Context) UserId {
	return u.id
}

func (u user) GetDatabaseId(ctx context.Context) DatabaseId {
	return u.db.GetId(ctx)
}

func (u user) GetSettings(ctx context.Context) UserSettings {
	var settings UserSettings
	return WithConnection(ctx, u.db.connect, func(conn *sql.DB) UserSettings {
		err := conn.QueryRowContext(ctx, "SELECT [name], CONVERT(VARCHAR(85), [sid], 1) FROM sys.database_principals WHERE [principal_id]=@p1", u.id).
			Scan(&settings.Name, &settings.LoginId)
		if err != nil {
			utils.AddError(ctx, "Failed to retrieve user settings", err)
		}

		return settings
	})
}

func (u user) Drop(ctx context.Context) {
	WithConnection(ctx, u.db.connect, func(conn *sql.DB) any {
		name := u.getName(ctx, conn)
		if utils.HasError(ctx) {
			return nil
		}

		_, err := conn.ExecContext(ctx, fmt.Sprintf("DROP USER [%s]", name))
		if err != nil {
			utils.AddError(ctx, "Failed to drop user", err)
		}

		return nil
	})
}

func (u user) UpdateSettings(ctx context.Context, settings UserSettings) {
	WithConnection(ctx, u.db.connect, func(conn *sql.DB) any {
		name := u.getName(ctx, conn)
		if utils.HasError(ctx) {
			return nil
		}

		loginName := GetSqlLogin(ctx, u.db.GetConnection(ctx), settings.LoginId).getName(ctx)
		if utils.HasError(ctx) {
			return nil
		}

		_, err := conn.ExecContext(ctx, fmt.Sprintf("ALTER USER [%s] WITH NAME=[%s], LOGIN=[%s]", name, settings.Name, loginName))
		if err != nil {
			utils.AddError(ctx, "Failed to update user", err)
		}

		return nil
	})
}

func (u user) getName(ctx context.Context, conn *sql.DB) string {
	var name string
	err := conn.QueryRowContext(ctx, "SELECT USER_NAME(@p1)", u.id).Scan(&name)
	if err != nil {
		utils.AddError(ctx, "Failed to resolve user name", err)
	}

	return name
}
