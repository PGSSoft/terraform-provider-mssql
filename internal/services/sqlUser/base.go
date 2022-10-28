package sqlUser

import (
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var attributes = map[string]tfsdk.Attribute{
	"id": {
		MarkdownDescription: "`<database_id>/<user_id>`. User ID can be retrieved using `SELECT DATABASE_PRINCIPAL_ID('<user_name>')`.",
		Type:                types.StringType,
	},
	"name": {
		MarkdownDescription: "User name. Cannot be longer than 128 chars.",
		Type:                types.StringType,
		Validators:          validators.UserNameValidators,
	},
	"database_id": common.DatabaseIdAttribute,
	"login_id": {
		MarkdownDescription: "SID of SQL login. Can be retrieved using `mssql_sql_login` or `SELECT SUSER_SID('<login_name>')`.",
		Type:                types.StringType,
	},
}

type resourceData struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	DatabaseId types.String `tfsdk:"database_id"`
	LoginId    types.String `tfsdk:"login_id"`
}

func (d resourceData) toSettings() sql.UserSettings {
	return sql.UserSettings{
		Name:    d.Name.ValueString(),
		LoginId: sql.LoginId(d.LoginId.ValueString()),
		Type:    sql.USER_TYPE_SQL,
	}
}

func (d resourceData) withSettings(settings sql.UserSettings) resourceData {
	return resourceData{
		Id:         d.Id,
		DatabaseId: d.DatabaseId,
		Name:       types.StringValue(settings.Name),
		LoginId:    types.StringValue(fmt.Sprint(settings.LoginId)),
	}
}

func (d resourceData) withIds(dbId sql.DatabaseId, userId sql.UserId) resourceData {
	return resourceData{
		Id:         types.StringValue(fmt.Sprintf("%v/%v", dbId, userId)),
		DatabaseId: types.StringValue(fmt.Sprint(dbId)),
		Name:       d.Name,
		LoginId:    d.LoginId,
	}
}
