package provider

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/provider/resource"
	"strconv"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	sdkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type sqlLoginResourceData struct {
	Id                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	Password                types.String `tfsdk:"password"`
	MustChangePassword      types.Bool   `tfsdk:"must_change_password"`
	DefaultDatabaseId       types.String `tfsdk:"default_database_id"`
	DefaultLanguage         types.String `tfsdk:"default_language"`
	CheckPasswordExpiration types.Bool   `tfsdk:"check_password_expiration"`
	CheckPasswordPolicy     types.Bool   `tfsdk:"check_password_policy"`
}

func (d sqlLoginResourceData) toSettings(ctx context.Context) sql.SqlLoginSettings {
	var dbId int

	if !d.DefaultDatabaseId.Null && !d.DefaultDatabaseId.Unknown {
		if id, err := strconv.Atoi(d.DefaultDatabaseId.Value); err == nil {
			dbId = id
		} else {
			utils.AddError(ctx, "Failed to parse DB id", err)
		}
	}

	return sql.SqlLoginSettings{
		Name:                    d.Name.Value,
		Password:                d.Password.Value,
		MustChangePassword:      d.MustChangePassword.Value,
		DefaultDatabaseId:       sql.DatabaseId(dbId),
		DefaultLanguage:         d.DefaultLanguage.Value,
		CheckPasswordExpiration: d.CheckPasswordExpiration.Value,
		CheckPasswordPolicy:     d.CheckPasswordPolicy.Value || d.CheckPasswordPolicy.Null || d.CheckPasswordPolicy.Unknown,
	}
}

func (d sqlLoginResourceData) withSettings(settings sql.SqlLoginSettings, isAzure bool) sqlLoginResourceData {
	d.Name = types.String{Value: settings.Name}

	if isAzure {
		return d
	}

	if isAttrSet(d.MustChangePassword) {
		d.MustChangePassword.Value = settings.MustChangePassword
	}

	if isAttrSet(d.DefaultDatabaseId) {
		d.DefaultDatabaseId = types.String{Value: fmt.Sprint(settings.DefaultDatabaseId)}
	}

	if isAttrSet(d.DefaultLanguage) {
		d.DefaultLanguage = types.String{Value: settings.DefaultLanguage}
	}

	if isAttrSet(d.CheckPasswordExpiration) {
		d.CheckPasswordExpiration = types.Bool{Value: settings.CheckPasswordExpiration}
	}

	if isAttrSet(d.CheckPasswordPolicy) {
		d.CheckPasswordPolicy = types.Bool{Value: settings.CheckPasswordPolicy}
	}

	return d
}

type sqlLoginResource struct {
	BaseResource
}

func (p mssqlProvider) NewSqlLoginResource() func() sdkresource.Resource {
	return func() sdkresource.Resource {
		return resource.WrapResource[sqlLoginResourceData](&sqlLoginResource{})
	}
}

func (r *sqlLoginResource) GetName() string {
	return "sql_login"
}

func (r *sqlLoginResource) GetSchema(context.Context) tfsdk.Schema {
	const azureSQLNote = "\n\n-> **Note** In case of Azure SQL, which does not support this feature, the flag will be ignored. "
	return tfsdk.Schema{
		Description: "Manages single login.",
		Attributes: map[string]tfsdk.Attribute{
			"id":   toResourceId(sqlLoginAttributes["id"]),
			"name": toRequired(sqlLoginAttributes["name"]),
			"password": {
				MarkdownDescription: "Password for the login. Must follow strong password policies defined for SQL server. " +
					"Passwords are case-sensitive, length must be 8-128 chars, can include all characters except `'` or `name`.\n\n" +
					"~> **Note** Password will be stored in the raw state as plain-text. [Read more about sensitive data in state](https://www.terraform.io/language/state/sensitive-data).",
				Type:      types.StringType,
				Required:  true,
				Sensitive: true,
			},
			"must_change_password": func() tfsdk.Attribute {
				attr := sqlLoginAttributes["must_change_password"]
				attr.Optional = true
				attr.MarkdownDescription += " Defaults to `false`. \n\n" +
					"-> **Note** After password is changed, this flag is being reset to `false`, which will show as changes in Terraform plan. " +
					"Use `ignore_changes` block to prevent this behavior." + azureSQLNote
				return attr
			}(),
			"default_database_id": func() tfsdk.Attribute {
				attr := sqlLoginAttributes["default_database_id"]
				attr.Optional = true
				attr.MarkdownDescription += " Defaults to ID of `master`." + azureSQLNote
				return attr
			}(),
			"default_language": func() tfsdk.Attribute {
				attr := sqlLoginAttributes["default_language"]
				attr.Optional = true
				attr.Description += " Defaults to current default language of the server. " +
					"If the default language of the server is later changed, the default language of the login remains unchanged." + azureSQLNote
				return attr
			}(),
			"check_password_expiration": func() tfsdk.Attribute {
				attr := sqlLoginAttributes["check_password_expiration"]
				attr.Optional = true
				attr.MarkdownDescription += " Defaults to `false`." + azureSQLNote
				return attr
			}(),
			"check_password_policy": func() tfsdk.Attribute {
				attr := sqlLoginAttributes["check_password_policy"]
				attr.Optional = true
				attr.MarkdownDescription += " Defaults to `true`." + azureSQLNote
				return attr
			}(),
		},
	}
}

func (r *sqlLoginResource) Create(ctx context.Context, req resource.CreateRequest[sqlLoginResourceData], resp *resource.CreateResponse[sqlLoginResourceData]) {
	var login sql.SqlLogin

	req.
		Then(func() { login = sql.CreateSqlLogin(ctx, r.conn, req.Plan.toSettings(ctx)) }).
		Then(func() {
			resp.State = req.Plan.withSettings(login.GetSettings(ctx), r.conn.IsAzure(ctx))
			resp.State.Id = types.String{Value: string(login.GetId(ctx))}
		})
}

func (r *sqlLoginResource) Read(ctx context.Context, req resource.ReadRequest[sqlLoginResourceData], resp *resource.ReadResponse[sqlLoginResourceData]) {
	var (
		login  sql.SqlLogin
		exists bool
	)

	req.
		Then(func() { login = sql.GetSqlLogin(ctx, r.conn, sql.LoginId(req.State.Id.Value)) }).
		Then(func() { exists = login.Exists(ctx) }).
		Then(func() {
			if exists {
				resp.SetState(req.State.withSettings(login.GetSettings(ctx), r.conn.IsAzure(ctx)))
			}
		})
}

func (r *sqlLoginResource) Update(ctx context.Context, req resource.UpdateRequest[sqlLoginResourceData], resp *resource.UpdateResponse[sqlLoginResourceData]) {
	var login sql.SqlLogin

	req.
		Then(func() { login = sql.GetSqlLogin(ctx, r.conn, sql.LoginId(req.Plan.Id.Value)) }).
		Then(func() { login.UpdateSettings(ctx, req.Plan.toSettings(ctx)) }).
		Then(func() { resp.State = req.Plan.withSettings(login.GetSettings(ctx), r.conn.IsAzure(ctx)) })
}

func (r *sqlLoginResource) Delete(ctx context.Context, req resource.DeleteRequest[sqlLoginResourceData], _ *resource.DeleteResponse[sqlLoginResourceData]) {
	var login sql.SqlLogin

	req.
		Then(func() { login = sql.GetSqlLogin(ctx, r.conn, sql.LoginId(req.State.Id.Value)) }).
		Then(func() { login.Drop(ctx) })
}
