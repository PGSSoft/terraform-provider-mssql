package provider

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/provider/resource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	sdkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type databaseResource struct {
	BaseResource
}

func (p mssqlProvider) NewDatabaseResource() func() sdkresource.Resource {
	return func() sdkresource.Resource {
		return resource.WrapResource[databaseResourceData](&databaseResource{})
	}
}

func (r *databaseResource) GetName() string {
	return "database"
}

func (r *databaseResource) GetSchema(context.Context) tfsdk.Schema {
	return tfsdk.Schema{
		Description: "Manages single database.",
		Attributes: map[string]tfsdk.Attribute{
			"id":   toResourceId(databaseAttributes["id"]),
			"name": toRequired(databaseAttributes["name"]),
			"collation": func() tfsdk.Attribute {
				attr := databaseAttributes["collation"]
				attr.Optional = true
				attr.Computed = true
				attr.Description += " Defaults to SQL Server instance's default collation."
				return attr
			}(),
		},
	}
}

func (r *databaseResource) Create(ctx context.Context, req resource.CreateRequest[databaseResourceData], resp *resource.CreateResponse[databaseResourceData]) {
	var db sql.Database

	req.
		Then(func() { db = sql.CreateDatabase(ctx, r.conn, req.Plan.toSettings()) }).
		Then(func() { resp.State = req.Plan.withSettings(db.GetSettings(ctx)) }).
		Then(func() { resp.State.Id = types.String{Value: fmt.Sprint(db.GetId(ctx))} })
}

func (r *databaseResource) Read(ctx context.Context, req resource.ReadRequest[databaseResourceData], resp *resource.ReadResponse[databaseResourceData]) {
	var db sql.Database
	var dbExists bool
	var settings sql.DatabaseSettings

	req.
		Then(func() { db = sql.GetDatabase(ctx, r.conn, req.State.getDbId(ctx)) }).
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

func (r *databaseResource) Update(ctx context.Context, req resource.UpdateRequest[databaseResourceData], resp *resource.UpdateResponse[databaseResourceData]) {
	var db sql.Database

	req.
		Then(func() { db = sql.GetDatabase(ctx, r.conn, req.Plan.getDbId(ctx)) }).
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

func (r *databaseResource) Delete(ctx context.Context, req resource.DeleteRequest[databaseResourceData], _ *resource.DeleteResponse[databaseResourceData]) {
	var dbId sql.DatabaseId
	var db sql.Database

	req.
		Then(func() { dbId = req.State.getDbId(ctx) }).
		Then(func() { db = sql.GetDatabase(ctx, r.conn, dbId) }).
		Then(func() { db.Drop(ctx) })
}
