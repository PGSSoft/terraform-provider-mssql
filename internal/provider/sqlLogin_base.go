package provider

import (
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var sqlLoginAttributes = map[string]tfsdk.Attribute{
	"id": {
		MarkdownDescription: "Login SID. Can be retrieved using `SELECT SUSER_SID('<login_name>')`.",
		Type:                types.StringType,
	},
	"name": {
		MarkdownDescription: fmt.Sprintf("Login name. %s and cannot contain `\\ `", regularIdentifiersDoc),
		Type:                types.StringType,
		Validators:          validators.LoginNameValidators,
	},
	"must_change_password": {
		MarkdownDescription: "When true, password change will be forced on first logon.",
		Type:                types.BoolType,
	},
	"default_database_id": {
		MarkdownDescription: "ID of login's default DB. The ID can be retrieved using `mssql_database` data resource.",
		Type:                types.StringType,
	},
	"default_language": {
		Description: "Default language assigned to login.",
		Type:        types.StringType,
	},
	"check_password_expiration": {
		MarkdownDescription: "When `true`, password expiration policy is enforced for this login.",
		Type:                types.BoolType,
	},
	"check_password_policy": {
		MarkdownDescription: "When `true`, the Windows password policies of the computer on which SQL Server is running are enforced on this login.",
		Type:                types.BoolType,
	},
}

type sqlLoginDataSourceData struct {
	Id                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	MustChangePassword      types.Bool   `tfsdk:"must_change_password"`
	DefaultDatabaseId       types.String `tfsdk:"default_database_id"`
	DefaultLanguage         types.String `tfsdk:"default_language"`
	CheckPasswordExpiration types.Bool   `tfsdk:"check_password_expiration"`
	CheckPasswordPolicy     types.Bool   `tfsdk:"check_password_policy"`
}

func (d sqlLoginDataSourceData) withSettings(settings sql.SqlLoginSettings) sqlLoginDataSourceData {
	return sqlLoginDataSourceData{
		Id:                      d.Id,
		Name:                    types.String{Value: settings.Name},
		MustChangePassword:      types.Bool{Value: settings.MustChangePassword},
		DefaultDatabaseId:       types.String{Value: fmt.Sprint(settings.DefaultDatabaseId)},
		DefaultLanguage:         types.String{Value: settings.DefaultLanguage},
		CheckPasswordExpiration: types.Bool{Value: settings.CheckPasswordExpiration},
		CheckPasswordPolicy:     types.Bool{Value: settings.CheckPasswordPolicy},
	}
}
