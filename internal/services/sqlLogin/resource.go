package sqlLogin

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/resource"
	common2 "github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"strconv"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceData struct {
	Id                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	Password                types.String `tfsdk:"password"`
	MustChangePassword      types.Bool   `tfsdk:"must_change_password"`
	DefaultDatabaseId       types.String `tfsdk:"default_database_id"`
	DefaultLanguage         types.String `tfsdk:"default_language"`
	CheckPasswordExpiration types.Bool   `tfsdk:"check_password_expiration"`
	CheckPasswordPolicy     types.Bool   `tfsdk:"check_password_policy"`
	PrincipalId             types.String `tfsdk:"principal_id"`
}

func (d resourceData) toSettings(ctx context.Context) sql.SqlLoginSettings {
	var dbId int

	if common2.IsAttrSet(d.DefaultDatabaseId) {
		if id, err := strconv.Atoi(d.DefaultDatabaseId.ValueString()); err == nil {
			dbId = id
		} else {
			utils.AddError(ctx, "Failed to parse DB id", err)
		}
	}

	return sql.SqlLoginSettings{
		Name:                    d.Name.ValueString(),
		Password:                d.Password.ValueString(),
		MustChangePassword:      d.MustChangePassword.ValueBool(),
		DefaultDatabaseId:       sql.DatabaseId(dbId),
		DefaultLanguage:         d.DefaultLanguage.ValueString(),
		CheckPasswordExpiration: d.CheckPasswordExpiration.ValueBool(),
		CheckPasswordPolicy:     d.CheckPasswordPolicy.ValueBool() || d.CheckPasswordPolicy.IsNull() || d.CheckPasswordPolicy.IsUnknown(),
	}
}

func (d resourceData) withSettings(settings sql.SqlLoginSettings, isAzure bool) resourceData {
	d.Name = types.StringValue(settings.Name)
	d.PrincipalId = types.StringValue(fmt.Sprint(settings.PrincipalId))

	if isAzure {
		return d
	}

	if common2.IsAttrSet(d.MustChangePassword) {
		d.MustChangePassword = types.BoolValue(settings.MustChangePassword)
	}

	if common2.IsAttrSet(d.DefaultDatabaseId) {
		d.DefaultDatabaseId = types.StringValue(fmt.Sprint(settings.DefaultDatabaseId))
	}

	if common2.IsAttrSet(d.DefaultLanguage) {
		d.DefaultLanguage = types.StringValue(settings.DefaultLanguage)
	}

	if common2.IsAttrSet(d.CheckPasswordExpiration) {
		d.CheckPasswordExpiration = types.BoolValue(settings.CheckPasswordExpiration)
	}

	if common2.IsAttrSet(d.CheckPasswordPolicy) {
		d.CheckPasswordPolicy = types.BoolValue(settings.CheckPasswordPolicy)
	}

	return d
}

type res struct{}

func (r *res) GetName() string {
	return "sql_login"
}

func (r *res) GetSchema(context.Context) tfsdk.Schema {
	const azureSQLNote = "\n\n-> **Note** In case of Azure SQL, which does not support this feature, the flag will be ignored. "
	return tfsdk.Schema{
		Description: "Manages single login.",
		Attributes: map[string]tfsdk.Attribute{
			"id":   common2.ToResourceId(attributes["id"]),
			"name": common2.ToRequired(attributes["name"]),
			"password": {
				MarkdownDescription: "Password for the login. Must follow strong password policies defined for SQL server. " +
					"Passwords are case-sensitive, length must be 8-128 chars, can include all characters except `'` or `name`.\n\n" +
					"~> **Note** Password will be stored in the raw state as plain-text. [Read more about sensitive data in state](https://www.terraform.io/language/state/sensitive-data).",
				Type:      types.StringType,
				Required:  true,
				Sensitive: true,
			},
			"must_change_password": func() tfsdk.Attribute {
				attr := attributes["must_change_password"]
				attr.Optional = true
				attr.MarkdownDescription += " Defaults to `false`. \n\n" +
					"-> **Note** After password is changed, this flag is being reset to `false`, which will show as changes in Terraform plan. " +
					"Use `ignore_changes` block to prevent this behavior." + azureSQLNote
				return attr
			}(),
			"default_database_id": func() tfsdk.Attribute {
				attr := attributes["default_database_id"]
				attr.Optional = true
				attr.MarkdownDescription += " Defaults to ID of `master`." + azureSQLNote
				return attr
			}(),
			"default_language": func() tfsdk.Attribute {
				attr := attributes["default_language"]
				attr.Optional = true
				attr.Description += " Defaults to current default language of the server. " +
					"If the default language of the server is later changed, the default language of the login remains unchanged." + azureSQLNote
				return attr
			}(),
			"check_password_expiration": func() tfsdk.Attribute {
				attr := attributes["check_password_expiration"]
				attr.Optional = true
				attr.MarkdownDescription += " Defaults to `false`." + azureSQLNote
				return attr
			}(),
			"check_password_policy": func() tfsdk.Attribute {
				attr := attributes["check_password_policy"]
				attr.Optional = true
				attr.MarkdownDescription += " Defaults to `true`." + azureSQLNote
				return attr
			}(),
			"principal_id": func() tfsdk.Attribute {
				attr := attributes["principal_id"]
				attr.Computed = true
				return attr
			}(),
		},
	}
}

func (r *res) Create(ctx context.Context, req resource.CreateRequest[resourceData], resp *resource.CreateResponse[resourceData]) {
	var login sql.SqlLogin

	req.
		Then(func() { login = sql.CreateSqlLogin(ctx, req.Conn, req.Plan.toSettings(ctx)) }).
		Then(func() {
			resp.State = req.Plan.withSettings(login.GetSettings(ctx), req.Conn.IsAzure(ctx))
			resp.State.Id = types.StringValue(string(login.GetId(ctx)))
		})
}

func (r *res) Read(ctx context.Context, req resource.ReadRequest[resourceData], resp *resource.ReadResponse[resourceData]) {
	var (
		login  sql.SqlLogin
		exists bool
	)

	req.
		Then(func() { login = sql.GetSqlLogin(ctx, req.Conn, sql.LoginId(req.State.Id.ValueString())) }).
		Then(func() { exists = login.Exists(ctx) }).
		Then(func() {
			if exists {
				resp.SetState(req.State.withSettings(login.GetSettings(ctx), req.Conn.IsAzure(ctx)))
			}
		})
}

func (r *res) Update(ctx context.Context, req resource.UpdateRequest[resourceData], resp *resource.UpdateResponse[resourceData]) {
	var login sql.SqlLogin

	req.
		Then(func() { login = sql.GetSqlLogin(ctx, req.Conn, sql.LoginId(req.Plan.Id.ValueString())) }).
		Then(func() { login.UpdateSettings(ctx, req.Plan.toSettings(ctx)) }).
		Then(func() { resp.State = req.Plan.withSettings(login.GetSettings(ctx), req.Conn.IsAzure(ctx)) })
}

func (r *res) Delete(ctx context.Context, req resource.DeleteRequest[resourceData], _ *resource.DeleteResponse[resourceData]) {
	var login sql.SqlLogin

	req.
		Then(func() { login = sql.GetSqlLogin(ctx, req.Conn, sql.LoginId(req.State.Id.ValueString())) }).
		Then(func() { login.Drop(ctx) })
}
