package sql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
)

type SchemaPermission struct {
	Name            string
	WithGrantOption bool
}

type SchemaPermissions map[string]SchemaPermission

type Schema interface {
	GetDb(context.Context) Database
	GetId(context.Context) SchemaId
	GetName(context.Context) string
	GetOwnerId(context.Context) GenericDatabasePrincipalId
	ChangeOwner(_ context.Context, ownerId GenericDatabasePrincipalId)
	Drop(context.Context)
	GetPermissions(ctx context.Context, id GenericDatabasePrincipalId) SchemaPermissions
	GrantPermission(ctx context.Context, id GenericDatabasePrincipalId, permission SchemaPermission)
	UpdatePermission(ctx context.Context, id GenericDatabasePrincipalId, permission SchemaPermission)
	RevokePermission(ctx context.Context, id GenericDatabasePrincipalId, permission string)
}

func GetSchema(_ context.Context, db Database, id SchemaId) Schema {
	return schema{id: id, db: db}
}

func GetSchemaByName(ctx context.Context, db Database, name string) Schema {
	conn := db.connect(ctx)
	var id sql.NullInt32

	utils.StopOnError(ctx).
		Then(func() {
			err := QueryRowContextWithRetry(ctx, conn, "SELECT SCHEMA_ID(@p1)", name).Scan(&id)
			utils.AddError(ctx, "Failed to fetch schema ID", err)
		}).
		Then(func() {
			if !id.Valid {
				utils.AddError(ctx, "Schema does not exist", fmt.Errorf("did not find schema with %q", name))
			}
		})

	return GetSchema(ctx, db, SchemaId(id.Int32))
}

func GetSchemas(ctx context.Context, db Database) map[SchemaId]Schema {
	conn := db.connect(ctx)
	schemas := map[SchemaId]Schema{}

	utils.StopOnError(ctx).Then(func() {
		res, err := conn.QueryContext(ctx, "SELECT [schema_id] FROM sys.schemas")

		switch err {
		case sql.ErrNoRows:
			return
		case nil:
			for res.Next() {
				var id SchemaId
				rowErr := res.Scan(&id)
				utils.AddError(ctx, "Failed to parse schemas dataset", rowErr)
				schemas[id] = GetSchema(ctx, db, id)
			}
		default:
			utils.AddError(ctx, "Failed to fetch DB schemas", err)
		}
	})

	return schemas
}

func CreateSchema[T DatabasePrincipalId](ctx context.Context, db Database, name string, ownerId T) Schema {
	conn := db.connect(ctx)
	ownerName := db.getUserName(ctx, GenericDatabasePrincipalId(ownerId))

	utils.StopOnError(ctx).Then(func() {
		_, err := conn.ExecContext(ctx, fmt.Sprintf("CREATE SCHEMA [%s] AUTHORIZATION [%s]", name, ownerName))
		utils.AddError(ctx, "Failed to create schema", err)
	})

	return GetSchemaByName(ctx, db, name)
}

type schema struct {
	id SchemaId
	db Database
}

func (s schema) GetDb(ctx context.Context) Database {
	return s.db
}

func (s schema) GetId(ctx context.Context) SchemaId {
	return s.id
}

func (s schema) GetName(ctx context.Context) string {
	var name string
	conn := s.db.connect(ctx)

	utils.StopOnError(ctx).Then(func() {
		err := QueryRowContextWithRetry(ctx, conn, "SELECT SCHEMA_NAME(@p1)", s.id).Scan(&name)
		utils.AddError(ctx, "Failed to fetch schema name", err)
	})

	return name
}

func (s schema) GetOwnerId(ctx context.Context) GenericDatabasePrincipalId {
	var (
		ownerId GenericDatabasePrincipalId
		conn    *sql.DB
	)

	utils.StopOnError(ctx).
		Then(func() { conn = s.db.connect(ctx) }).
		Then(func() {
			err := QueryRowContextWithRetry(ctx, conn, "SELECT [principal_id] FROM sys.schemas WHERE [schema_id] = @p1", s.id).Scan(&ownerId)
			utils.AddError(ctx, "Failed to fetch owner ID", err)
		})

	return ownerId
}

func (s schema) ChangeOwner(ctx context.Context, ownerId GenericDatabasePrincipalId) {
	schemaName := s.GetName(ctx)
	ownerName := s.db.getUserName(ctx, ownerId)
	conn := s.db.connect(ctx)

	utils.StopOnError(ctx).Then(func() {
		_, err := conn.ExecContext(ctx, fmt.Sprintf("ALTER AUTHORIZATION ON schema::[%s] TO [%s]", schemaName, ownerName))
		utils.AddError(ctx, "Failed to change owner", err)
	})
}

func (s schema) Drop(ctx context.Context) {
	schemaName := s.GetName(ctx)
	conn := s.db.connect(ctx)

	utils.StopOnError(ctx).Then(func() {
		_, err := conn.ExecContext(ctx, fmt.Sprintf("DROP SCHEMA [%s]", schemaName))
		utils.AddError(ctx, "Failed to drop schema", err)
	})
}

func (s schema) GetPermissions(ctx context.Context, principalId GenericDatabasePrincipalId) SchemaPermissions {
	conn := s.db.connect(ctx)
	if utils.HasError(ctx) {
		return nil
	}

	res, err := conn.QueryContext(ctx, "SELECT [permission_name], [state] FROM sys.database_permissions WHERE [class]=3 AND [major_id]=@p1 AND [grantee_principal_id]=@p2", s.id, principalId)

	perms := SchemaPermissions{}

	switch err {
	case sql.ErrNoRows:
		return perms
	case nil:
		for res.Next() {
			var state string
			perm := SchemaPermission{}
			err := res.Scan(&perm.Name, &state)
			utils.AddError(ctx, "Failed to parse schema permissions", err)
			perm.WithGrantOption = state == "W"
			perms[perm.Name] = perm
		}
	default:
		utils.AddError(ctx, "Failed to fetch schema permissions", err)
		return nil
	}

	return perms
}

func (s schema) GrantPermission(ctx context.Context, principalId GenericDatabasePrincipalId, permission SchemaPermission) {
	schemaName := s.GetName(ctx)
	principalName := s.db.getUserName(ctx, principalId)
	var conn *sql.DB

	utils.StopOnError(ctx).
		Then(func() { conn = s.db.connect(ctx) }).
		Then(func() {
			stat := fmt.Sprintf("GRANT %s ON schema::[%s] TO [%s]", permission.Name, schemaName, principalName)
			if permission.WithGrantOption {
				stat += " WITH GRANT OPTION"
			}
			_, err := conn.ExecContext(ctx, stat)
			utils.AddError(ctx, "Failed to grant schema permission", err)
		})
}

func (s schema) UpdatePermission(ctx context.Context, principalId GenericDatabasePrincipalId, permission SchemaPermission) {
	if permission.WithGrantOption {
		s.GrantPermission(ctx, principalId, permission)
		return
	}

	schemaName := s.GetName(ctx)
	principalName := s.db.getUserName(ctx, principalId)
	var conn *sql.DB

	utils.StopOnError(ctx).
		Then(func() { conn = s.db.connect(ctx) }).
		Then(func() {
			_, err := conn.ExecContext(ctx, fmt.Sprintf("REVOKE GRANT OPTION FOR %s ON schema::[%s] FROM [%s] CASCADE", permission.Name, schemaName, principalName))
			utils.AddError(ctx, "Failed to revoke grant option", err)
		})
}

func (s schema) RevokePermission(ctx context.Context, principalId GenericDatabasePrincipalId, permission string) {
	schemaName := s.GetName(ctx)
	principalName := s.db.getUserName(ctx, principalId)
	var conn *sql.DB

	utils.StopOnError(ctx).
		Then(func() { conn = s.db.connect(ctx) }).
		Then(func() {
			_, err := conn.ExecContext(ctx, fmt.Sprintf("REVOKE %s ON schema::[%s] FROM [%s] CASCADE", permission, schemaName, principalName))
			utils.AddError(ctx, "Failed to revoke permission", err)
		})
}
