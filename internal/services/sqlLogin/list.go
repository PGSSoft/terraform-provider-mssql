package sqlLogin

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/datasource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type listDataSourceData struct {
	Id     types.String     `tfsdk:"id"`
	Logins []dataSourceData `tfsdk:"logins"`
}

type listDataSource struct{}

func (l *listDataSource) GetName() string {
	return "sql_logins"
}

func (l *listDataSource) GetSchema(context.Context) tfsdk.Schema {
	attrs := map[string]tfsdk.Attribute{}
	for n, attribute := range attributes {
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

func (l *listDataSource) Read(ctx context.Context, req datasource.ReadRequest[listDataSourceData], resp *datasource.ReadResponse[listDataSourceData]) {
	var logins map[sql.LoginId]sql.SqlLogin

	req.
		Then(func() { logins = sql.GetSqlLogins(ctx, req.Conn) }).
		Then(func() {
			result := listDataSourceData{
				Id: types.String{Value: ""},
			}

			for id, login := range logins {
				s := login.GetSettings(ctx)
				r := dataSourceData{Id: types.String{Value: fmt.Sprint(id)}}
				result.Logins = append(result.Logins, r.withSettings(s))
			}

			resp.SetState(result)
		})
}
