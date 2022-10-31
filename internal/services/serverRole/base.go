package serverRole

import (
	"context"
	"fmt"
	common2 "github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/PGSSoft/terraform-provider-mssql/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
)

var attributes = map[string]tfsdk.Attribute{
	"id": {
		MarkdownDescription: "Role principal ID. Can be retrieved using `SELECT SUSER_SID('<login_name>')`.",
		Type:                types.StringType,
	},
	"name": {
		MarkdownDescription: fmt.Sprintf("Role name. %s and cannot be longer than 128 chars.", common2.RegularIdentifiersDoc),
		Type:                types.StringType,
		Validators:          validators.UserNameValidators,
	},
	"owner_id": {
		MarkdownDescription: "ID of another server role or login owning this role. Can be retrieved using `mssql_server_role` or `mssql_sql_login`.",
		Type:                types.StringType,
	},
}

type resourceData struct {
	Id      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	OwnerId types.String `tfsdk:"owner_id"`
}

func (d resourceData) withSettings(settings sql.ServerRoleSettings) resourceData {
	d.Name = types.StringValue(settings.Name)
	d.OwnerId = types.StringValue(fmt.Sprint(settings.OwnerId))

	return d
}

func (d resourceData) toSettings(ctx context.Context) sql.ServerRoleSettings {
	settings := sql.ServerRoleSettings{
		Name:    d.Name.ValueString(),
		OwnerId: sql.EmptyServerPrincipalId,
	}

	if common2.IsAttrSet(d.OwnerId) {
		id, err := strconv.Atoi(d.OwnerId.ValueString())
		utils.AddError(ctx, "Failed to parse owner ID", err)
		settings.OwnerId = sql.GenericServerPrincipalId(id)
	}

	return settings
}
