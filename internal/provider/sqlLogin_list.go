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
	_ tfsdk.DataSourceType = SqlLoginListDataSourceType{}
	_ tfsdk.DataSource     = sqlLoginList{}
)

type SqlLoginListDataSourceType struct{}

func (l SqlLoginListDataSourceType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	attrs := map[string]tfsdk.Attribute{}
	for n, attribute := range sqlLoginAttributes {
		attribute.Computed = true
		attrs[n] = attribute
	}

	return tfsdk.Schema{
		Description: "Obtains information about all SQL logins found in SQL Server instance.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:        types.StringType,
				Computed:    true,
				Description: "ID of the resource used only internally by the provider.",
			},
			"logins": {
				Description: "Set of SQL login objects",
				Attributes:  tfsdk.SetNestedAttributes(attrs),
				Computed:    true,
			},
		},
	}, nil
}

func (l SqlLoginListDataSourceType) NewDataSource(ctx context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	return newResource(ctx, p, func(base Resource) sqlLoginList {
		return sqlLoginList{Resource: base}
	})
}

type sqlLoginList struct {
	Resource
}

func (l sqlLoginList) Read(ctx context.Context, _ tfsdk.ReadDataSourceRequest, response *tfsdk.ReadDataSourceResponse) {
	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)

	logins := sql.GetSqlLogins(ctx, l.Db)
	if utils.HasError(ctx) {
		return
	}

	result := struct {
		Id     types.String             `tfsdk:"id"`
		Logins []sqlLoginDataSourceData `tfsdk:"logins"`
	}{
		Id: types.String{Value: ""},
	}

	for id, login := range logins {
		s := login.GetSettings(ctx)

		if utils.HasError(ctx) {
			return
		}

		r := sqlLoginDataSourceData{Id: types.String{Value: fmt.Sprint(id)}}
		result.Logins = append(result.Logins, r.withSettings(s))
	}

	utils.SetData(ctx, &response.State, result)
}
