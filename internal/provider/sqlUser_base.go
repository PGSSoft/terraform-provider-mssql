package provider

import (
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var sqlUserAttributes = map[string]tfsdk.Attribute{
	"id": {
		MarkdownDescription: "`<database_id>/<user_id>`. User ID can be retrieved using `SELECT DATABASE_PRINCIPAL_ID('<user_name>')`.",
		Type:                types.StringType,
	},
	"name": {
		MarkdownDescription: "User name. Cannot be longer than 128 chars.",
		Type:                types.StringType,
		Validators:          validators.UserNameValidators,
	},
	"database_id": databaseIdAttribute,
	"login_id": {
		MarkdownDescription: "SID of SQL login. Can be retrieved using `mssql_sql_login` or `SELECT SUSER_SID('<login_name>')`.",
		Type:                types.StringType,
	},
}

type sqlUserResourceData struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	DatabaseId types.String `tfsdk:"database_id"`
	LoginId    types.String `tfsdk:"login_id"`
}

func (d sqlUserResourceData) toSettings() sql.UserSettings {
	return sql.UserSettings{
		Name:    d.Name.Value,
		LoginId: sql.LoginId(d.LoginId.Value),
		Type:    sql.USER_TYPE_SQL,
	}
}

func (d sqlUserResourceData) withSettings(settings sql.UserSettings) sqlUserResourceData {
	return sqlUserResourceData{
		Id:         d.Id,
		DatabaseId: d.DatabaseId,
		Name:       types.String{Value: settings.Name},
		LoginId:    types.String{Value: fmt.Sprint(settings.LoginId)},
	}
}

func (d sqlUserResourceData) withIds(dbId sql.DatabaseId, userId sql.UserId) sqlUserResourceData {
	return sqlUserResourceData{
		Id:         types.String{Value: fmt.Sprintf("%v/%v", dbId, userId)},
		DatabaseId: types.String{Value: fmt.Sprint(dbId)},
		Name:       d.Name,
		LoginId:    d.LoginId,
	}
}

type sqlUserResourceBase struct {
	Resource
}
