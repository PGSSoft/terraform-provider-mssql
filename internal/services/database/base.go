package database

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/PGSSoft/terraform-provider-mssql/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
)

var attributes = map[string]tfsdk.Attribute{
	"id": {
		MarkdownDescription: "Database ID. Can be retrieved using `SELECT DB_ID('<db_name>')`.",
		Type:                types.StringType,
	},
	"name": {
		MarkdownDescription: fmt.Sprintf("Database name. %s.", common.RegularIdentifiersDoc),
		Type:                types.StringType,
		Validators:          validators.DatabaseNameValidators,
	},
	"collation": {
		Description: "Default collation name. Can be either a Windows collation name or a SQL collation name.",
		Type:        types.StringType,
	},
}

type resourceData struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Collation types.String `tfsdk:"collation"`
}

func (d resourceData) getDbId(ctx context.Context) sql.DatabaseId {
	if d.Id.Unknown || d.Id.Null {
		return sql.NullDatabaseId
	}

	id, err := strconv.Atoi(d.Id.Value)

	if err != nil {
		utils.AddError(ctx, fmt.Sprintf("Failed to convert resource ID '%s'", d.Id.Value), err)
	}

	return sql.DatabaseId(id)
}

func (d resourceData) toSettings() sql.DatabaseSettings {
	return sql.DatabaseSettings{
		Name:      d.Name.Value,
		Collation: d.Collation.Value,
	}
}

func (d resourceData) withSettings(settings sql.DatabaseSettings) resourceData {
	return resourceData{
		Id:   d.Id,
		Name: types.String{Value: settings.Name},

		Collation: types.String{
			Value: settings.Collation,
			Null:  settings.Collation == "",
		},
	}
}
