package provider

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/PGSSoft/terraform-provider-mssql/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
)

var databaseAttributes = map[string]tfsdk.Attribute{
	"id": {
		MarkdownDescription: "Database ID. Can be retrieved using `SELECT DB_ID('<db_name>')`.",
		Type:                types.StringType,
	},
	"name": {
		MarkdownDescription: fmt.Sprintf("Database name. %s.", regularIdentifiersDoc),
		Type:                types.StringType,
		Validators:          validators.DatabaseNameValidators,
	},
	"collation": {
		Description: "Default collation name. Can be either a Windows collation name or a SQL collation name.",
		Type:        types.StringType,
	},
}

func getDB(ctx context.Context, getter utils.DataGetter) (databaseResourceData, sql.DatabaseId) {
	data := utils.GetData[databaseResourceData](ctx, getter)
	if utils.HasError(ctx) {
		return data, sql.NullDatabaseId
	}

	if data.Id.Unknown || data.Id.Null {
		return data, sql.NullDatabaseId
	} else {
		id, err := strconv.Atoi(data.Id.Value)
		if err != nil {
			utils.AddError(ctx, fmt.Sprintf("Failed to convert resource ID '%s'", data.Id.Value), err)
		}

		return data, sql.DatabaseId(id)
	}
}

type databaseResourceData struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Collation types.String `tfsdk:"collation"`
}

func (d databaseResourceData) getDbId(ctx context.Context) sql.DatabaseId {
	if d.Id.Unknown || d.Id.Null {
		return sql.NullDatabaseId
	}

	id, err := strconv.Atoi(d.Id.Value)

	if err != nil {
		utils.AddError(ctx, fmt.Sprintf("Failed to convert resource ID '%s'", d.Id.Value), err)
	}

	return sql.DatabaseId(id)
}

func (d databaseResourceData) toSettings() sql.DatabaseSettings {
	return sql.DatabaseSettings{
		Name:      d.Name.Value,
		Collation: d.Collation.Value,
	}
}

func (d databaseResourceData) withSettings(settings sql.DatabaseSettings) databaseResourceData {
	return databaseResourceData{
		Id:   d.Id,
		Name: types.String{Value: settings.Name},

		Collation: types.String{
			Value: settings.Collation,
			Null:  settings.Collation == "",
		},
	}
}
