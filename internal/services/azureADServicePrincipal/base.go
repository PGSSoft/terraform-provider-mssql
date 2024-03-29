package azureADServicePrincipal

import (
	"context"
	"fmt"
	"strings"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var attrDescriptions = map[string]string{
	"id":        "`<database_id>/<user_id>`. User ID can be retrieved using `sys.database_principals` view.",
	"name":      "User name. Cannot be longer than 128 chars.",
	"client_id": "Azure AD client_id of the Service Principal. This can be either regular Service Principal or Managed Service Identity.",
}

type resourceData struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	DatabaseId types.String `tfsdk:"database_id"`
	ClientId   types.String `tfsdk:"client_id"`
}

func (d resourceData) toSettings() sql.UserSettings {
	return sql.UserSettings{
		Name:        d.Name.ValueString(),
		AADObjectId: sql.AADObjectId(d.ClientId.ValueString()),
		Type:        sql.USER_TYPE_AZUREAD,
	}
}

func (d resourceData) withSettings(ctx context.Context, settings sql.UserSettings) resourceData {
	if settings.Type != sql.USER_TYPE_AZUREAD {
		utils.AddError(ctx, "Invalid user type", fmt.Errorf("expected user type %d, but got %d", sql.USER_TYPE_AZUREAD, settings.Type))
		return d
	}

	d.Name = types.StringValue(settings.Name)
	d.ClientId = types.StringValue(strings.ToUpper(fmt.Sprint(settings.AADObjectId)))
	return d
}
