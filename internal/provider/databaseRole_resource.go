package provider

import (
	"context"
	"errors"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// To ensure resource types fully satisfy framework interfaces
var (
	_ resource.ResourceWithConfigure   = &databaseRoleResource{}
	_ resource.ResourceWithImportState = databaseRoleResource{}
)

type databaseRoleResource struct {
	Resource
}

func (p mssqlProvider) NewDatabaseRoleResource() func() resource.Resource {
	return func() resource.Resource {
		return &databaseRoleResource{}
	}
}

func (s databaseRoleResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "mssql_database_role"
}

func (r *databaseRoleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.Resource.Configure(ctx, req.ProviderData, &resp.Diagnostics)
}

func (d databaseRoleResource) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Manages database-level role.",
		Attributes: map[string]tfsdk.Attribute{
			"id":          toResourceId(databaseRoleAttributes["id"]),
			"name":        toRequired(databaseRoleAttributes["name"]),
			"database_id": databaseIdResourceAttribute,
			"owner_id": func() tfsdk.Attribute {
				attr := databaseRoleAttributes["owner_id"]
				attr.Optional = true
				attr.Computed = true
				attr.MarkdownDescription += " Defaults to ID of current user, used to authorize the Terraform provider."
				return attr
			}(),
		},
	}, nil
}

func (d databaseRoleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var (
		data databaseRoleResourceData
		db   sql.Database
		dbId sql.DatabaseId
		role sql.DatabaseRole
	)
	ownerId := dbObjectId[sql.GenericDatabasePrincipalId]{IsEmpty: true}

	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	utils.StopOnError(ctx).
		Then(func() { data = utils.GetData[databaseRoleResourceData](ctx, request.Config) }).
		Then(func() { db = getResourceDb(ctx, d.Db, data.DatabaseId.Value) }).
		Then(func() { dbId = db.GetId(ctx) }).
		Then(func() {
			if !data.OwnerId.Null && !data.OwnerId.Unknown {
				ownerId = parseDbObjectId[sql.GenericDatabasePrincipalId](ctx, data.OwnerId.Value)

				if ownerId.DbId != dbId {
					utils.AddError(ctx, "Role owner must be principal defined in the same DB as the role", errors.New("owner and principal DBs are different"))
				}
			}
		}).
		Then(func() {
			if ownerId.IsEmpty {
				role = sql.CreateDatabaseRole(ctx, db, data.Name.Value, sql.EmptyDatabasePrincipalId)
			} else {
				role = sql.CreateDatabaseRole(ctx, db, data.Name.Value, ownerId.ObjectId)
			}
		}).
		Then(func() { data = data.withRoleData(ctx, role) }).
		Then(func() { utils.SetData(ctx, &response.State, data) })
}

func (d databaseRoleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var (
		data   databaseRoleResourceData
		db     sql.Database
		roleId dbObjectId[sql.DatabaseRoleId]
		role   sql.DatabaseRole
	)

	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	utils.StopOnError(ctx).
		Then(func() { data = utils.GetData[databaseRoleResourceData](ctx, request.State) }).
		Then(func() { roleId = parseDbObjectId[sql.DatabaseRoleId](ctx, data.Id.Value) }).
		Then(func() { db = sql.GetDatabase(ctx, d.Db, roleId.DbId) }).
		Then(func() { role = sql.GetDatabaseRole(ctx, db, roleId.ObjectId) }).
		Then(func() { data = data.withRoleData(ctx, role) }).
		Then(func() { utils.SetData(ctx, &response.State, data) })
}

func (d databaseRoleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		data   databaseRoleResourceData
		dbId   sql.DatabaseId
		db     sql.Database
		roleId dbObjectId[sql.DatabaseRoleId]
		role   sql.DatabaseRole
	)
	ownerId := dbObjectId[sql.GenericDatabasePrincipalId]{IsEmpty: true}

	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	utils.StopOnError(ctx).
		Then(func() { data = utils.GetData[databaseRoleResourceData](ctx, request.Plan) }).
		Then(func() { db = getResourceDb(ctx, d.Db, data.DatabaseId.Value) }).
		Then(func() { dbId = db.GetId(ctx) }).
		Then(func() { roleId = parseDbObjectId[sql.DatabaseRoleId](ctx, data.Id.Value) }).
		Then(func() { role = sql.GetDatabaseRole(ctx, db, roleId.ObjectId) }).
		Then(func() {
			if !data.OwnerId.Null && !data.OwnerId.Unknown {
				ownerId = parseDbObjectId[sql.GenericDatabasePrincipalId](ctx, data.OwnerId.Value)

				if ownerId.DbId != dbId {
					utils.AddError(ctx, "Role owner must be principal defined in the same DB as the role", errors.New("owner and principal DBs are different"))
				}
			}
		}).
		Then(func() {
			if role.GetName(ctx) != data.Name.Value && !utils.HasError(ctx) {
				role.Rename(ctx, data.Name.Value)
			}
		}).
		Then(func() {
			if role.GetOwnerId(ctx) != ownerId.ObjectId && !utils.HasError(ctx) {
				if ownerId.IsEmpty {
					role.ChangeOwner(ctx, sql.EmptyDatabasePrincipalId)
				} else {
					role.ChangeOwner(ctx, ownerId.ObjectId)
				}
			}
		}).
		Then(func() { utils.SetData(ctx, &response.State, data) })
}

func (d databaseRoleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var (
		data   databaseRoleResourceData
		db     sql.Database
		roleId dbObjectId[sql.DatabaseRoleId]
		role   sql.DatabaseRole
	)

	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	utils.StopOnError(ctx).
		Then(func() { data = utils.GetData[databaseRoleResourceData](ctx, request.State) }).
		Then(func() { db = getResourceDb(ctx, d.Db, data.DatabaseId.Value) }).
		Then(func() { roleId = parseDbObjectId[sql.DatabaseRoleId](ctx, data.Id.Value) }).
		Then(func() { role = sql.GetDatabaseRole(ctx, db, roleId.ObjectId) }).
		Then(func() { role.Drop(ctx) }).
		Then(func() { response.State.RemoveResource(ctx) })
}

func (d databaseRoleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}
