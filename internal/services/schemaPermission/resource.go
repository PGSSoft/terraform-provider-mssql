package schemaPermission

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/resource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
)

type resourceData struct {
	Id              types.String `tfsdk:"id"`
	SchemaId        types.String `tfsdk:"schema_id"`
	PrincipalId     types.String `tfsdk:"principal_id"`
	Permission      types.String `tfsdk:"permission"`
	WithGrantOption types.Bool   `tfsdk:"with_grant_option"`
}

var _ resource.ResourceWithValidation[resourceData] = res{}

type res struct{}

func (r res) GetName() string {
	return "schema_permission"
}

func (r res) GetSchema(ctx context.Context) tfsdk.Schema {
	return tfsdk.Schema{
		MarkdownDescription: "Grants database-level permission.",
		Attributes: map[string]tfsdk.Attribute{
			"id":           common.ToResourceId(attributes["id"]),
			"schema_id":    common.ToRequiredImmutable(attributes["schema_id"]),
			"principal_id": common.ToRequiredImmutable(attributes["principal_id"]),
			"permission":   common.ToRequiredImmutable(attributes["permission"]),
			"with_grant_option": func() tfsdk.Attribute {
				attr := attributes["with_grant_option"]
				attr.MarkdownDescription += " Defaults to `false`."
				attr.Optional = true
				attr.Computed = true
				return attr
			}(),
		},
	}
}

func (r res) Read(ctx context.Context, req resource.ReadRequest[resourceData], resp *resource.ReadResponse[resourceData]) {
	schema, principalId, permission := r.parseInputs(ctx, req.Conn, req.State)
	var permissions sql.SchemaPermissions

	req.
		Then(func() { permissions = schema.GetPermissions(ctx, principalId.MemberId) }).
		Then(func() {
			state := resourceData{
				Id:          req.State.Id,
				SchemaId:    types.StringValue(principalId.DbObjectId.String()),
				PrincipalId: types.StringValue(principalId.GetMemberId().String()),
				Permission:  types.StringValue(permission),
			}

			if perm, ok := permissions[permission]; ok {
				state.WithGrantOption = types.BoolValue(perm.WithGrantOption)
				resp.SetState(state)
			}
		})
}

func (r res) Create(ctx context.Context, req resource.CreateRequest[resourceData], resp *resource.CreateResponse[resourceData]) {
	schemaId := common.ParseDbObjectId[sql.SchemaId](ctx, req.Plan.SchemaId.ValueString())
	principalId := common.ParseDbObjectId[sql.GenericDatabasePrincipalId](ctx, req.Plan.PrincipalId.ValueString())

	var (
		db     sql.Database
		schema sql.Schema
	)

	req.
		Then(func() { db = sql.GetDatabase(ctx, req.Conn, schemaId.DbId) }).
		Then(func() { schema = sql.GetSchema(ctx, db, schemaId.ObjectId) }).
		Then(func() {
			perm := sql.SchemaPermission{
				Name:            req.Plan.Permission.ValueString(),
				WithGrantOption: req.Plan.WithGrantOption.ValueBool(),
			}

			schema.GrantPermission(ctx, principalId.ObjectId, perm)
		}).
		Then(func() {
			resp.State = req.Plan
			resp.State.Id = types.StringValue(fmt.Sprintf("%s/%d/%s", schemaId, principalId.ObjectId, req.Plan.Permission.ValueString()))
			resp.State.WithGrantOption = types.BoolValue(req.Plan.WithGrantOption.ValueBool())
		})
}

func (r res) Update(ctx context.Context, req resource.UpdateRequest[resourceData], resp *resource.UpdateResponse[resourceData]) {
	schema, principalId, permission := r.parseInputs(ctx, req.Conn, req.Plan)

	req.
		Then(func() {
			perm := sql.SchemaPermission{
				Name:            permission,
				WithGrantOption: req.Plan.WithGrantOption.ValueBool(),
			}

			schema.UpdatePermission(ctx, principalId.MemberId, perm)
		}).
		Then(func() {
			resp.State = req.Plan
			resp.State.WithGrantOption = types.BoolValue(req.Plan.WithGrantOption.ValueBool())
		})
}

func (r res) Delete(ctx context.Context, req resource.DeleteRequest[resourceData], _ *resource.DeleteResponse[resourceData]) {
	schema, principalId, permission := r.parseInputs(ctx, req.Conn, req.State)

	req.Then(func() { schema.RevokePermission(ctx, principalId.MemberId, permission) })
}

func (r res) Validate(ctx context.Context, req resource.ValidateRequest[resourceData], _ *resource.ValidateResponse[resourceData]) {
	schemaId := common.ParseDbObjectId[sql.SchemaId](ctx, req.Config.SchemaId.ValueString())
	principalId := common.ParseDbObjectId[sql.GenericDatabasePrincipalId](ctx, req.Config.PrincipalId.ValueString())

	req.Then(func() {
		if schemaId.DbId != principalId.DbId {
			err := fmt.Errorf("schema_id points to DB with ID %d while principal_id points to DB with ID %d", schemaId.DbId, principalId.DbId)
			utils.AddError(ctx, "Schema and Principal must belong to the same DB", err)
		}
	})
}

func (r res) parseInputs(ctx context.Context, conn sql.Connection, data resourceData) (sql.Schema, common.DbObjectMemberId[sql.SchemaId, sql.GenericDatabasePrincipalId], string) {
	parts := strings.Split(data.Id.ValueString(), "/")
	permission := parts[len(parts)-1]
	principalId := common.ParseDbObjectMemberId[sql.SchemaId, sql.GenericDatabasePrincipalId](ctx, data.Id.ValueString()[:len(data.Id.ValueString())-len(permission)-1])

	var (
		db     sql.Database
		schema sql.Schema
	)

	utils.StopOnError(ctx).
		Then(func() { db = sql.GetDatabase(ctx, conn, principalId.DbId) }).
		Then(func() { schema = sql.GetSchema(ctx, db, principalId.ObjectId) })

	return schema, principalId, permission
}
