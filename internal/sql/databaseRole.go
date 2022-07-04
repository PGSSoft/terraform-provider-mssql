package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
)

type DatabaseRole interface {
	GetId(context.Context) DatabaseRoleId
	GetOwnerId(context.Context) GenericDatabasePrincipalId
	GetDb(context.Context) Database
	GetName(context.Context) string
	Drop(context.Context)
	Rename(_ context.Context, name string)
	ChangeOwner(_ context.Context, ownerId GenericDatabasePrincipalId)
}

type databaseRole struct {
	id DatabaseRoleId
	db Database
}

func CreateDatabaseRole[T DatabasePrincipalId](ctx context.Context, db Database, name string, ownerId T) DatabaseRole {
	return WithConnection(ctx, db.connect, func(conn *sql.DB) DatabaseRole {
		stat := fmt.Sprintf("CREATE ROLE [%s]", name)

		if int(ownerId) != int(EmptyDatabasePrincipalId) {
			ownerName := getPrincipalName(ctx, conn, ownerId)
			stat += fmt.Sprintf(" AUTHORIZATION [%s]", ownerName)
		}

		if _, err := conn.ExecContext(ctx, stat); err != nil {
			utils.AddError(ctx, "Failed to create role", err)
			return nil
		}

		return GetDatabaseRoleByName(ctx, db, name)
	})
}

func GetDatabaseRole(_ context.Context, db Database, id DatabaseRoleId) DatabaseRole {
	return databaseRole{db: db, id: id}
}

func GetDatabaseRoleByName(ctx context.Context, db Database, name string) DatabaseRole {
	return WithConnection(ctx, db.connect, func(conn *sql.DB) DatabaseRole {
		res := databaseRole{db: db}
		id := sql.NullInt64{}

		if err := conn.QueryRowContext(ctx, "SELECT DATABASE_PRINCIPAL_ID(@p1)", name).Scan(&id); err != nil {
			utils.AddError(ctx, "Failed to resolve role ID", err)
			return nil
		}

		if !id.Valid {
			utils.AddError(ctx, "Role does not exist", errors.New("role does not exist"))
			return nil
		}

		res.id = DatabaseRoleId(id.Int64)
		return res
	})
}

func GetDatabaseRoles(ctx context.Context, db Database) map[DatabaseRoleId]DatabaseRole {
	const errorSummary = "Failed to retrieve list of database roles"

	return WithConnection(ctx, db.connect, func(conn *sql.DB) map[DatabaseRoleId]DatabaseRole {
		res := map[DatabaseRoleId]DatabaseRole{}

		switch queryRes, err := conn.QueryContext(ctx, "SELECT [principal_id] FROM sys.database_principals WHERE [type] = 'R'"); err {
		case sql.ErrNoRows: //ignore
		case nil:
			for queryRes.Next() {
				role := databaseRole{db: db}

				if err := queryRes.Scan(&role.id); err != nil {
					utils.AddError(ctx, errorSummary, err)
					return nil
				}
				res[role.id] = role
			}
		default:
			utils.AddError(ctx, errorSummary, err)
		}

		return res
	})
}

func (d databaseRole) GetId(context.Context) DatabaseRoleId {
	return d.id
}

func (d databaseRole) GetOwnerId(ctx context.Context) GenericDatabasePrincipalId {
	return WithConnection(ctx, d.db.connect, func(conn *sql.DB) GenericDatabasePrincipalId {
		var res GenericDatabasePrincipalId

		if err := conn.QueryRowContext(ctx, "SELECT owning_principal_id FROM sys.database_principals WHERE principal_id=@p1", d.id).Scan(&res); err != nil {
			utils.AddError(ctx, "Failed to retrieve owner ID", err)
		}

		return res
	})
}

func (d databaseRole) GetDb(context.Context) Database {
	return d.db
}

func (d databaseRole) GetName(ctx context.Context) string {
	return WithConnection(ctx, d.db.connect, func(conn *sql.DB) string {
		return getPrincipalName(ctx, conn, d.id)
	})
}

func (d databaseRole) Drop(ctx context.Context) {
	WithConnection(ctx, d.db.connect, func(conn *sql.DB) any {
		name := getPrincipalName(ctx, conn, d.id)
		if utils.HasError(ctx) {
			return nil
		}

		if _, err := conn.ExecContext(ctx, fmt.Sprintf("DROP ROLE [%s]", name)); err != nil {
			utils.AddError(ctx, "Failed to drop role", err)
		}

		return nil
	})
}

func (d databaseRole) Rename(ctx context.Context, name string) {
	WithConnection(ctx, d.db.connect, func(conn *sql.DB) any {
		oldName := getPrincipalName(ctx, conn, d.id)
		if utils.HasError(ctx) {
			return nil
		}

		if _, err := conn.ExecContext(ctx, fmt.Sprintf("ALTER ROLE [%s] WITH NAME = [%s]", oldName, name)); err != nil {
			utils.AddError(ctx, "Failed to rename role", err)
		}

		return nil
	})
}

func (d databaseRole) ChangeOwner(ctx context.Context, ownerId GenericDatabasePrincipalId) {
	WithConnection(ctx, d.db.connect, func(conn *sql.DB) any {
		roleName := getPrincipalName(ctx, conn, d.id)
		var ownerName string
		if ownerId == EmptyDatabasePrincipalId {
			ownerName = getCurrentUserName(ctx, conn)
		} else {
			ownerName = getPrincipalName(ctx, conn, ownerId)
		}

		if utils.HasError(ctx) {
			return nil
		}

		if _, err := conn.ExecContext(ctx, fmt.Sprintf("ALTER AUTHORIZATION ON ROLE::[%s] TO [%s]", roleName, ownerName)); err != nil {
			utils.AddError(ctx, "Failed to change role ownership", err)
		}

		return nil
	})
}
