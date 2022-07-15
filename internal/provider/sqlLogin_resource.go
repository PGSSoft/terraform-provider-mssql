package provider

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"strconv"
)

// To ensure resource types fully satisfy framework interfaces
var (
	_ tfsdk.ResourceType            = SqlLoginResourceType{}
	_ tfsdk.Resource                = SqlLoginResource{}
	_ tfsdk.ResourceWithImportState = SqlLoginResource{}
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

type SqlLoginResourceType struct{}

func (l SqlLoginResourceType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
	}, nil
}

func (l SqlLoginResourceType) NewResource(ctx context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return newResource(ctx, p, func(base Resource) SqlLoginResource {
		return SqlLoginResource{Resource: base}
	})
}

type SqlLoginResource struct {
	Resource
}

func (l SqlLoginResource) Create(ctx context.Context, request tfsdk.CreateResourceRequest, response *tfsdk.CreateResourceResponse) {
	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	data := utils.GetData[sqlLoginResourceData](ctx, request.Plan)
	if utils.HasError(ctx) {
		return
	}

	login := sql.CreateSqlLogin(ctx, l.Db, data.toSettings(ctx))
	if utils.HasError(ctx) {
		return
	}

	data = data.withSettings(login.GetSettings(ctx), l.Db.IsAzure(ctx))
	data.Id = types.String{Value: string(login.GetId(ctx))}
	if utils.HasError(ctx) {
		return
	}

	utils.SetData(ctx, &response.State, data)
}

func (l SqlLoginResource) Read(ctx context.Context, request tfsdk.ReadResourceRequest, response *tfsdk.ReadResourceResponse) {
	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	data := utils.GetData[sqlLoginResourceData](ctx, request.State)
	if utils.HasError(ctx) {
		return
	}

	login := sql.GetSqlLogin(ctx, l.Db, sql.LoginId(data.Id.Value))
	if utils.HasError(ctx) {
		return
	}

	loginExists := login.Exists(ctx)
	if utils.HasError(ctx) {
		return
	}

	if !loginExists {
		response.State.RemoveResource(ctx)
		return
	}

	data = data.withSettings(login.GetSettings(ctx), l.Db.IsAzure(ctx))
	if utils.HasError(ctx) {
		return
	}

	utils.SetData(ctx, &response.State, data)
}

func (l SqlLoginResource) Update(ctx context.Context, request tfsdk.UpdateResourceRequest, response *tfsdk.UpdateResourceResponse) {
	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	data := utils.GetData[sqlLoginResourceData](ctx, request.Plan)
	if utils.HasError(ctx) {
		return
	}

	login := sql.GetSqlLogin(ctx, l.Db, sql.LoginId(data.Id.Value))
	if utils.HasError(ctx) {
		return
	}

	if login.UpdateSettings(ctx, data.toSettings(ctx)); utils.HasError(ctx) {
		return
	}

	if data = data.withSettings(login.GetSettings(ctx), l.Db.IsAzure(ctx)); utils.HasError(ctx) {
		return
	}

	utils.SetData(ctx, &response.State, data)
}

func (l SqlLoginResource) Delete(ctx context.Context, request tfsdk.DeleteResourceRequest, response *tfsdk.DeleteResourceResponse) {
	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)
	data := utils.GetData[sqlLoginResourceData](ctx, request.State)
	if utils.HasError(ctx) {
		return
	}

	login := sql.GetSqlLogin(ctx, l.Db, sql.LoginId(data.Id.Value))
	if utils.HasError(ctx) {
		return
	}

	if login.Drop(ctx); utils.HasError(ctx) {
		return
	}

	response.State.RemoveResource(ctx)
}

func (l SqlLoginResource) ImportState(ctx context.Context, request tfsdk.ImportResourceStateRequest, response *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), request, response)
}
