package database

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/resource"
	common2 "github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type res struct{}

func (r *res) GetName() string {
	return "database"
}

func (r *res) GetSchema(context.Context) tfsdk.Schema {
	return tfsdk.Schema{
		Description: "Manages single database.",
		Attributes: map[string]tfsdk.Attribute{
			"id":   common2.ToResourceId(attributes["id"]),
			"name": common2.ToRequired(attributes["name"]),
			"collation": func() tfsdk.Attribute {
				attr := attributes["collation"]
				attr.Optional = true
				attr.Computed = true
				attr.Description += " Defaults to SQL Server instance's default collation."
				return attr
			}(),
		},
	}
}

func (r *res) Create(ctx context.Context, req resource.CreateRequest[resourceData], resp *resource.CreateResponse[resourceData]) {
	var db sql.Database

	req.
		Then(func() { db = sql.CreateDatabase(ctx, req.Conn, req.Plan.toSettings()) }).
		Then(func() { resp.State = req.Plan.withSettings(db.GetSettings(ctx)) }).
		Then(func() { resp.State.Id = types.String{Value: fmt.Sprint(db.GetId(ctx))} })
}

func (r *res) Read(ctx context.Context, req resource.ReadRequest[resourceData], resp *resource.ReadResponse[resourceData]) {
	var db sql.Database
	var dbExists bool
	var settings sql.DatabaseSettings

	req.
		Then(func() { db = sql.GetDatabase(ctx, req.Conn, req.State.getDbId(ctx)) }).
		Then(func() {
			dbExists = db.Exists(ctx)
			settings = db.GetSettings(ctx)
		}).
		Then(func() {
			if dbExists {
				resp.SetState(req.State.withSettings(settings))
			}
		})
}

func (r *res) Update(ctx context.Context, req resource.UpdateRequest[resourceData], resp *resource.UpdateResponse[resourceData]) {
	var db sql.Database

	req.
		Then(func() { db = sql.GetDatabase(ctx, req.Conn, req.Plan.getDbId(ctx)) }).
		Then(func() {
			if req.State.Name.Value != req.Plan.Name.Value {
				db.Rename(ctx, req.Plan.Name.Value)
			}
		}).
		Then(func() {
			if req.State.Collation.Value != req.Plan.Collation.Value {
				db.SetCollation(ctx, req.Plan.Collation.Value)
			}
		}).
		Then(func() {
			resp.State = req.Plan.withSettings(db.GetSettings(ctx))
		})
}

func (r *res) Delete(ctx context.Context, req resource.DeleteRequest[resourceData], _ *resource.DeleteResponse[resourceData]) {
	var dbId sql.DatabaseId
	var db sql.Database

	req.
		Then(func() { dbId = req.State.getDbId(ctx) }).
		Then(func() { db = sql.GetDatabase(ctx, req.Conn, dbId) }).
		Then(func() { db.Drop(ctx) })
}
