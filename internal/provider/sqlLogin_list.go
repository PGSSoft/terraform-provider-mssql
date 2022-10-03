package provider

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/provider/datasource"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	sdkdatasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type sqlLoginListData struct {
	Id     types.String             `tfsdk:"id"`
	Logins []sqlLoginDataSourceData `tfsdk:"logins"`
}

type sqlLoginList struct {
	BaseDataSource
}

func (p mssqlProvider) NewSqlLoginListDataSource() func() sdkdatasource.DataSource {
	return func() sdkdatasource.DataSource {
		return datasource.WrapDataSource[sqlLoginListData](&sqlLoginList{})
	}
}

func (l *sqlLoginList) GetName() string {
	return "sql_logins"
}

func (l *sqlLoginList) GetSchema(context.Context) tfsdk.Schema {
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
	}
}

func (l *sqlLoginList) Read(ctx context.Context, req datasource.ReadRequest[sqlLoginListData], resp *datasource.ReadResponse[sqlLoginListData]) {
	var logins map[sql.LoginId]sql.SqlLogin

	req.
		Then(func() { logins = sql.GetSqlLogins(ctx, l.conn) }).
		Then(func() {
			result := sqlLoginListData{
				Id: types.String{Value: ""},
			}

			for id, login := range logins {
				s := login.GetSettings(ctx)
				r := sqlLoginDataSourceData{Id: types.String{Value: fmt.Sprint(id)}}
				result.Logins = append(result.Logins, r.withSettings(s))
			}

			resp.SetState(result)
		})
}
