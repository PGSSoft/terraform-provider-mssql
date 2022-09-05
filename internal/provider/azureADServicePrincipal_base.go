package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/PGSSoft/terraform-provider-mssql/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var azureADServicePrincipalAttributes = map[string]tfsdk.Attribute{
	"id": {
		MarkdownDescription: "`<database_id>/<user_id>`. User ID can be retrieved using `sys.database_principals` view.",
		Type:                types.StringType,
	},
	"name": {
		MarkdownDescription: "User name. Cannot be longer than 128 chars.",
		Type:                types.StringType,
		Validators:          validators.UserNameValidators,
	},
	"database_id": databaseIdAttribute,
	"client_id": {
		MarkdownDescription: "Azure AD client_id of the Service Principal. This can be either regular Service Principal or Managed Service Identity.",
		Type:                types.StringType,
	},
}

type azureADServicePrincipalResourceData struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	DatabaseId types.String `tfsdk:"database_id"`
	ClientId   types.String `tfsdk:"client_id"`
}

func (d azureADServicePrincipalResourceData) toSettings() sql.UserSettings {
	return sql.UserSettings{
		Name:        d.Name.Value,
		AADObjectId: sql.AADObjectId(d.ClientId.Value),
		Type:        sql.USER_TYPE_AZUREAD,
	}
}

func (d azureADServicePrincipalResourceData) withSettings(ctx context.Context, settings sql.UserSettings) azureADServicePrincipalResourceData {
	if settings.Type != sql.USER_TYPE_AZUREAD {
		utils.AddError(ctx, "Invalid user type", fmt.Errorf("expected user type %d, but got %d", sql.USER_TYPE_AZUREAD, settings.Type))
		return d
	}

	d.Name = types.String{Value: settings.Name}
	d.ClientId = types.String{Value: strings.ToUpper(fmt.Sprint(settings.AADObjectId))}
	return d
}
