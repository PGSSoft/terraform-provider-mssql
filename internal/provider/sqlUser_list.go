package provider

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// To ensure resource types fully satisfy framework interfaces
var (
	_ tfsdk.DataSourceType = SqlUserListDataSourceType{}
	_ tfsdk.DataSource     = sqlUserList{}
)

type SqlUserListDataSourceType struct{}

func (s SqlUserListDataSourceType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	attrs := map[string]tfsdk.Attribute{}
	for n, attr := range sqlUserAttributes {
		attr.Computed = true
		attrs[n] = attr
	}

	return tfsdk.Schema{
		Description: "Obtains information about all SQL users found in a database",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:        types.StringType,
				Computed:    true,
				Description: "ID of the resource, equals to database ID",
			},
			"database_id": func() tfsdk.Attribute {
				attr := sqlUserAttributes["database_id"]
				attr.Optional = true
				attr.MarkdownDescription += " Defaults to ID of `master`."
				return attr
			}(),
			"users": {
				Description: "Set of SQL user objects",
				Attributes:  tfsdk.SetNestedAttributes(attrs),
				Computed:    true,
			},
		},
	}, nil
}

func (s SqlUserListDataSourceType) NewDataSource(ctx context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return newResource(ctx, p, func(base Resource) sqlUserList {
		return sqlUserList{sqlUserResourceBase: sqlUserResourceBase{Resource: base}}
	})
}

type sqlUserListData struct {
	Id         types.String          `tfsdk:"id"`
	DatabaseId types.String          `tfsdk:"database_id"`
	Users      []sqlUserResourceData `tfsdk:"users"`
}

type sqlUserList struct {
	sqlUserResourceBase
}

func (s sqlUserList) Read(ctx context.Context, request tfsdk.ReadDataSourceRequest, response *tfsdk.ReadDataSourceResponse) {
	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)

	data := utils.GetData[sqlUserListData](ctx, request.Config)
	if utils.HasError(ctx) {
		return
	}

	db := getResourceDb(ctx, s.Db, data.DatabaseId.Value)
	dbId := db.GetId(ctx)
	data.DatabaseId = types.String{Value: fmt.Sprint(dbId)}
	data.Id = data.DatabaseId
	if utils.HasError(ctx) {
		return
	}

	for id, user := range sql.GetUsers(ctx, db) {
		s := user.GetSettings(ctx)
		if utils.HasError(ctx) {
			return
		}

		data.Users = append(data.Users, sqlUserResourceData{}.withIds(dbId, id).withSettings(s))
	}

	utils.SetData(ctx, &response.State, data)
}
