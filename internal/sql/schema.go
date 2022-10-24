package sql

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
)

type Schema interface {
	GetDb(context.Context) Database
	GetId(context.Context) SchemaId
	GetOwnerId(context.Context) GenericDatabasePrincipalId
	ChangeOwner(_ context.Context, ownerId GenericDatabasePrincipalId)
	Drop(context.Context)
}

func GetSchema(_ context.Context, db Database, id SchemaId) Schema {
	return schema{id: id, db: db}
}

func GetSchemaByName(ctx context.Context, db Database, name string) Schema {
	conn := db.connect(ctx)
	var id SchemaId

	utils.StopOnError(ctx).Then(func() {
		err := conn.QueryRowContext(ctx, "SELECT SCHEMA_ID(@p1)", name).Scan(&id)
		utils.AddError(ctx, "Failed to fetch schema ID", err)
	})

	return GetSchema(ctx, db, id)
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

func (s schema) GetOwnerId(ctx context.Context) GenericDatabasePrincipalId {
	var (
		ownerId GenericDatabasePrincipalId
		conn    *sql.DB
	)

	utils.StopOnError(ctx).
		Then(func() { conn = s.db.connect(ctx) }).
		Then(func() {
			err := conn.QueryRowContext(ctx, "SELECT [principal_id] FROM sys.schemas WHERE [schema_id] = @p1", s.id).Scan(&ownerId)
			utils.AddError(ctx, "Failed to fetch owner ID", err)
		})

	return ownerId
}

func (s schema) ChangeOwner(ctx context.Context, ownerId GenericDatabasePrincipalId) {
	schemaName := s.getName(ctx)
	ownerName := s.db.getUserName(ctx, ownerId)
	conn := s.db.connect(ctx)

	utils.StopOnError(ctx).Then(func() {
		_, err := conn.ExecContext(ctx, fmt.Sprintf("ALTER AUTHORIZATION ON schema::[%s] TO [%s]", schemaName, ownerName))
		utils.AddError(ctx, "Failed to change owner", err)
	})
}

func (s schema) Drop(ctx context.Context) {
	schemaName := s.getName(ctx)
	conn := s.db.connect(ctx)

	utils.StopOnError(ctx).Then(func() {
		_, err := conn.ExecContext(ctx, fmt.Sprintf("DROP SCHEMA [%s]", schemaName))
		utils.AddError(ctx, "Failed to drop schema", err)
	})
}

func (s schema) getName(ctx context.Context) string {
	var name string
	conn := s.db.connect(ctx)

	utils.StopOnError(ctx).Then(func() {
		err := conn.QueryRowContext(ctx, "SELECT SCHEMA_NAME(@p1)", s.id).Scan(&name)
		utils.AddError(ctx, "Failed to fetch schema name", err)
	})

	return name
}
