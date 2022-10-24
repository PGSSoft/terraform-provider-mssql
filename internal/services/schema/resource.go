package schema

import (
	"context"
	"errors"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/resource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type res struct{}

func (r res) GetName() string {
	return "schema"
}

func (r res) GetSchema(ctx context.Context) tfsdk.Schema {
	return tfsdk.Schema{
		MarkdownDescription: "Manages single DB schema.",
		Attributes: map[string]tfsdk.Attribute{
			"id":          common.ToResourceId(attributes["id"]),
			"name":        common.ToRequired(attributes["name"]),
			"database_id": common.DatabaseIdResourceAttribute,
			"owner_id": func() tfsdk.Attribute {
				attr := attributes["owner_id"]
				attr.Optional = true
				attr.Computed = true

				return attr
			}(),
		},
	}
}

func (r res) Read(ctx context.Context, req resource.ReadRequest[resourceData], resp *resource.ReadResponse[resourceData]) {
	var (
		db       sql.Database
		schemaId common.DbObjectId[sql.SchemaId]
		schema   sql.Schema
	)

	req.
		Then(func() { schemaId = common.ParseDbObjectId[sql.SchemaId](ctx, req.State.Id.Value) }).
		Then(func() { db = sql.GetDatabase(ctx, req.Conn, schemaId.DbId) }).
		Then(func() { schema = sql.GetSchema(ctx, db, schemaId.ObjectId) }).
		Then(func() { resp.SetState(req.State.withSchemaData(ctx, schema)) })
}

func (r res) Create(ctx context.Context, req resource.CreateRequest[resourceData], resp *resource.CreateResponse[resourceData]) {
	var (
		db      sql.Database
		schema  sql.Schema
		ownerId common.DbObjectId[sql.GenericDatabasePrincipalId]
	)

	req.
		Then(func() { db = common.GetResourceDb(ctx, req.Conn, req.Plan.DatabaseId.Value) }).
		Then(func() { ownerId = r.getOwnerId(ctx, req.Plan, db) }).
		Then(func() {
			owner := sql.EmptyDatabasePrincipalId

			if !ownerId.IsEmpty {
				owner = ownerId.ObjectId
			}

			schema = sql.CreateSchema(ctx, db, req.Plan.Name.Value, owner)
		}).
		Then(func() { resp.State = req.Plan.withSchemaData(ctx, schema) })
}

func (r res) Update(ctx context.Context, req resource.UpdateRequest[resourceData], resp *resource.UpdateResponse[resourceData]) {
	var (
		db       sql.Database
		schemaId common.DbObjectId[sql.SchemaId]
		schema   sql.Schema
		ownerId  common.DbObjectId[sql.GenericDatabasePrincipalId]
	)

	req.
		Then(func() { schemaId = common.ParseDbObjectId[sql.SchemaId](ctx, req.Plan.Id.Value) }).
		Then(func() { db = common.GetResourceDb(ctx, req.Conn, req.Plan.DatabaseId.Value) }).
		Then(func() { ownerId = r.getOwnerId(ctx, req.Plan, db) }).
		Then(func() { schema = sql.GetSchema(ctx, db, schemaId.ObjectId) }).
		Then(func() {
			owner := sql.EmptyDatabasePrincipalId

			if !ownerId.IsEmpty {
				owner = ownerId.ObjectId
			}

			schema.ChangeOwner(ctx, owner)
		}).
		Then(func() { resp.State = req.Plan.withSchemaData(ctx, schema) })
}

func (r res) Delete(ctx context.Context, req resource.DeleteRequest[resourceData], resp *resource.DeleteResponse[resourceData]) {
	var (
		db       sql.Database
		schemaId common.DbObjectId[sql.SchemaId]
		schema   sql.Schema
	)

	req.
		Then(func() { schemaId = common.ParseDbObjectId[sql.SchemaId](ctx, req.State.Id.Value) }).
		Then(func() { db = common.GetResourceDb(ctx, req.Conn, req.State.DatabaseId.Value) }).
		Then(func() { schema = sql.GetSchema(ctx, db, schemaId.ObjectId) }).
		Then(func() { schema.Drop(ctx) })
}

func (r res) getOwnerId(ctx context.Context, data resourceData, db sql.Database) common.DbObjectId[sql.GenericDatabasePrincipalId] {
	if !common.IsAttrSet(data.OwnerId) {
		return common.DbObjectId[sql.GenericDatabasePrincipalId]{IsEmpty: true}
	}

	var (
		ownerId common.DbObjectId[sql.GenericDatabasePrincipalId]
		dbId    sql.DatabaseId
	)

	utils.StopOnError(ctx).
		Then(func() { ownerId = common.ParseDbObjectId[sql.GenericDatabasePrincipalId](ctx, data.OwnerId.Value) }).
		Then(func() { dbId = db.GetId(ctx) }).
		Then(func() {
			if ownerId.DbId != dbId {
				utils.AddError(ctx, "Schema owner must be principal defined in the same DB as the schema", errors.New("owner and schema DBs are different"))
			}
		})

	return ownerId
}
