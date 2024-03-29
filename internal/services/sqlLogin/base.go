package sqlLogin

import (
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var attrDescriptions = map[string]string{
	"id":                        "Login SID. Can be retrieved using `SELECT SUSER_SID('<login_name>')`.",
	"name":                      fmt.Sprintf("Login name. %s and cannot contain `\\ `", common.RegularIdentifiersDoc),
	"must_change_password":      "When true, password change will be forced on first logon.",
	"default_database_id":       "ID of login's default DB. The ID can be retrieved using `mssql_database` data resource.",
	"default_language":          "Default language assigned to login.",
	"check_password_expiration": "When `true`, password expiration policy is enforced for this login.",
	"check_password_policy":     "When `true`, the Windows password policies of the computer on which SQL Server is running are enforced on this login.",
	"principal_id":              "ID used to reference SQL Login in other resources, e.g. `server_role`. Can be retrieved from `sys.sql_logins`.",
}

type dataSourceData struct {
	Id                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	MustChangePassword      types.Bool   `tfsdk:"must_change_password"`
	DefaultDatabaseId       types.String `tfsdk:"default_database_id"`
	DefaultLanguage         types.String `tfsdk:"default_language"`
	CheckPasswordExpiration types.Bool   `tfsdk:"check_password_expiration"`
	CheckPasswordPolicy     types.Bool   `tfsdk:"check_password_policy"`
	PrincipalId             types.String `tfsdk:"principal_id"`
}

func (d dataSourceData) withSettings(settings sql.SqlLoginSettings) dataSourceData {
	return dataSourceData{
		Id:                      d.Id,
		Name:                    types.StringValue(settings.Name),
		MustChangePassword:      types.BoolValue(settings.MustChangePassword),
		DefaultDatabaseId:       types.StringValue(fmt.Sprint(settings.DefaultDatabaseId)),
		DefaultLanguage:         types.StringValue(settings.DefaultLanguage),
		CheckPasswordExpiration: types.BoolValue(settings.CheckPasswordExpiration),
		CheckPasswordPolicy:     types.BoolValue(settings.CheckPasswordPolicy),
		PrincipalId:             types.StringValue(fmt.Sprint(settings.PrincipalId)),
	}
}
