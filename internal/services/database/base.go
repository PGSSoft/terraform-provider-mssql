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

var attrDescriptions = map[string]string{
	"id":        "Database ID. Can be retrieved using `SELECT DB_ID('<db_name>')`.",
	"name":      fmt.Sprintf("Database name. %s.", common.RegularIdentifiersDoc),
	"collation": "Default collation name. Can be either a Windows collation name or a SQL collation name.",
}

type resourceData struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Collation types.String `tfsdk:"collation"`
}

func (d resourceData) getDbId(ctx context.Context) sql.DatabaseId {
	if !common.IsAttrSet(d.Id) {
		return sql.NullDatabaseId
	}

	id, err := strconv.Atoi(d.Id.ValueString())

	if err != nil {
		utils.AddError(ctx, fmt.Sprintf("Failed to convert resource ID '%s'", d.Id.ValueString()), err)
	}

	return sql.DatabaseId(id)
}

func (d resourceData) toSettings() sql.DatabaseSettings {
	return sql.DatabaseSettings{
		Name:      d.Name.ValueString(),
		Collation: d.Collation.ValueString(),
	}
}

func (d resourceData) withSettings(settings sql.DatabaseSettings) resourceData {
	resData := resourceData{
		Id:        d.Id,
		Name:      types.StringValue(settings.Name),
		Collation: types.StringValue(settings.Collation),
	}

	if settings.Collation == "" {
		resData.Collation = types.StringNull()
	}

	return resData
}
