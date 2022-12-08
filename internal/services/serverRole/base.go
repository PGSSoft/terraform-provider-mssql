package serverRole

import (
	"context"
	"fmt"
	common2 "github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
)

var attrDescriptions = map[string]string{
	"id":       "Role principal ID.",
	"name":     fmt.Sprintf("Role name. %s and cannot be longer than 128 chars.", common2.RegularIdentifiersDoc),
	"owner_id": "ID of another server role or login owning this role. Can be retrieved using `mssql_server_role` or `mssql_sql_login`.",
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

func parseId(ctx context.Context, id types.String) sql.ServerRoleId {
	intId, err := strconv.Atoi(id.ValueString())
	utils.AddError(ctx, "Failed to parse ID", err)
	return sql.ServerRoleId(intId)
}
