package script

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/resource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceData struct {
	Id           types.String `tfsdk:"id"`
	DatabaseId   types.String `tfsdk:"database_id"`
	ReadScript   types.String `tfsdk:"read_script"`
	CreateScript types.String `tfsdk:"create_script"`
	UpdateScript types.String `tfsdk:"update_script"`
	DeleteScript types.String `tfsdk:"delete_script"`

	State map[string]types.String `tfsdk:"state"`
}

type res struct{}

func (r *res) GetName() string {
	return "script"
}

func (r *res) GetSchema(context.Context) tfsdk.Schema {
	return tfsdk.Schema{
		MarkdownDescription: `Allows execution of arbitrary SQL scripts to check state and apply desired state. 

-> **Note** This resource is meant to be an escape hatch for all cases not supported by the provider's resources. Whenever possible, use dedicated resources, which offer better plan, validation and error reporting.  
`,
		Attributes: map[string]tfsdk.Attribute{
			"id": common.ToResourceId(tfsdk.Attribute{
				MarkdownDescription: "Used only internally by Terraform. Always set to `script`",
				Type:                types.StringType,
			}),
			"database_id": common.ToRequiredImmutable(common.DatabaseIdAttribute),
			"state": {
				MarkdownDescription: "Desired state of the DB. It is arbitrary map of string values that will be compared against the values returned by the `read_script`.",
				Type:                types.MapType{ElemType: types.StringType},
				Required:            true,
			},
			"read_script": {
				MarkdownDescription: "SQL script returning current state of the DB. It must return single-row result set where column names match the keys of `state` map and all values are strings that will be compared against `state` to determine if the resource state matches DB state.",
				Type:                types.StringType,
				Required:            true,
			},
			"create_script": {
				MarkdownDescription: "SQL script executed when the resource does not exist in Terraform state. When not provided, `update_script` will be used to create the resource.",
				Type:                types.StringType,
				Optional:            true,
			},
			"update_script": {
				MarkdownDescription: "SQL script executed when the desired state specified in `state` attribute does not match the state returned by `read_script`",
				Type:                types.StringType,
				Required:            true,
			},
			"delete_script": {
				MarkdownDescription: "SQL script executed when the resource is being destroyed. When not provided, no action will be taken during resource destruction.",
				Type:                types.StringType,
				Optional:            true,
			},
		},
	}
}

func (r *res) Read(ctx context.Context, req resource.ReadRequest[resourceData], resp *resource.ReadResponse[resourceData]) {
	req.
		Then(func() { req.State.State = r.queryState(ctx, req.Conn, req.State) }).
		Then(func() { resp.SetState(req.State) })
}

func (r *res) Create(ctx context.Context, req resource.CreateRequest[resourceData], resp *resource.CreateResponse[resourceData]) {
	req.
		Then(func() { r.queryState(ctx, req.Conn, req.Plan) }). // report error if planned read script produces and error
		Then(func() {
			script := req.Plan.UpdateScript.ValueString()

			if common.IsAttrSet(req.Plan.CreateScript) {
				script = req.Plan.CreateScript.ValueString()
			}

			r.execScript(ctx, req.Conn, script, req.Plan)
		}).
		Then(func() {
			resp.State = req.Plan
			resp.State.Id = types.StringValue("script")
		})
}

func (r *res) Update(ctx context.Context, req resource.UpdateRequest[resourceData], resp *resource.UpdateResponse[resourceData]) {
	req.
		Then(func() {
			if req.State.ReadScript != req.Plan.ReadScript {
				r.queryState(ctx, req.Conn, req.Plan) // report error if planned read script produces and error
			}
		}).
		Then(func() { r.execScript(ctx, req.Conn, req.Plan.UpdateScript.ValueString(), req.Plan) }).
		Then(func() { resp.State = req.Plan })
}

func (r *res) Delete(ctx context.Context, req resource.DeleteRequest[resourceData], _ *resource.DeleteResponse[resourceData]) {
	if common.IsAttrSet(req.State.DeleteScript) {
		req.Then(func() { r.execScript(ctx, req.Conn, req.State.DeleteScript.ValueString(), req.State) })
	}
}

func (r *res) execScript(ctx context.Context, conn sql.Connection, script string, data resourceData) {
	var db sql.Database

	utils.StopOnError(ctx).
		Then(func() { db = common.GetResourceDb(ctx, conn, data.DatabaseId.ValueString()) }).
		Then(func() { db.Exec(ctx, script) })
}

func (r *res) queryState(ctx context.Context, conn sql.Connection, data resourceData) map[string]types.String {
	var (
		db       sql.Database
		queryRes []map[string]string
	)

	state := map[string]types.String{}

	utils.StopOnError(ctx).
		Then(func() { db = common.GetResourceDb(ctx, conn, data.DatabaseId.ValueString()) }).
		Then(func() { queryRes = db.Query(ctx, data.ReadScript.ValueString()) }).
		Then(func() {
			if len(queryRes) != 1 {
				utils.AddError(ctx, "Invalid read_script result", fmt.Errorf("expected 1 row, got %d", len(queryRes)))
			}
		}).
		Then(func() {
			for name, val := range queryRes[0] {
				state[name] = types.StringValue(val)
			}
		})

	return state
}
