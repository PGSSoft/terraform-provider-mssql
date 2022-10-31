package serverRole

import (
	"context"
	"errors"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/datasource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithValidation[resourceData] = dataSource{}

type dataSource struct{}

func (d dataSource) GetName() string {
	return "server_role"
}

func (d dataSource) GetSchema(context.Context) tfsdk.Schema {
	const requiredNote = " Either `name` or `id` must be provided."

	return tfsdk.Schema{
		Description: "Obtains information about single server role.",
		Attributes: map[string]tfsdk.Attribute{
			"id": func() tfsdk.Attribute {
				attr := attributes["id"]
				attr.Optional = true
				attr.Computed = true
				attr.MarkdownDescription += requiredNote

				return attr
			}(),
			"name": func() tfsdk.Attribute {
				attr := attributes["name"]
				attr.Optional = true
				attr.Computed = true
				attr.MarkdownDescription += requiredNote

				return attr
			}(),
			"owner_id": func() tfsdk.Attribute {
				attr := attributes["owner_id"]
				attr.Computed = true

				return attr
			}(),
		},
	}
}

func (d dataSource) Read(ctx context.Context, req datasource.ReadRequest[resourceData], resp *datasource.ReadResponse[resourceData]) {
	id := sql.ServerRoleId(0)

	if common.IsAttrSet(req.Config.Id) {
		id = parseId(ctx, req.Config)
	}

	var role sql.ServerRole
	var settings sql.ServerRoleSettings

	req.
		Then(func() {
			if common.IsAttrSet(req.Config.Id) {
				role = sql.GetServerRole(ctx, req.Conn, id)
			} else {
				role = sql.GetServerRoleByName(ctx, req.Conn, req.Config.Name.ValueString())
			}
		}).
		Then(func() { settings = role.GetSettings(ctx) }).
		Then(func() {
			state := req.Config.withSettings(settings)
			state.Id = types.StringValue(fmt.Sprint(role.GetId(ctx)))
			resp.SetState(state)
		})
}

func (d dataSource) Validate(ctx context.Context, req datasource.ValidateRequest[resourceData], resp *datasource.ValidateResponse[resourceData]) {
	if !common.IsAttrSet(req.Config.Id) && !common.IsAttrSet(req.Config.Name) {
		utils.AddError(ctx, "Either name or id must be provided", errors.New("both name and id are empty values"))
	}
}
