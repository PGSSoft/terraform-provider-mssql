package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var databaseIdAttribute = tfsdk.Attribute{
	MarkdownDescription: fmt.Sprintf("ID of database. Can be retrieved using `mssql_database` or `SELECT DB_ID('<db_name>')`."),
	Type:                types.StringType,
}

var databaseIdResourceAttribute = databaseIdAttribute

func init() {
	databaseIdResourceAttribute.Optional = true
	databaseIdResourceAttribute.Computed = true
	databaseIdResourceAttribute.MarkdownDescription += " Defaults to ID of `master`."
	databaseIdResourceAttribute.PlanModifiers = tfsdk.AttributePlanModifiers{
		tfsdk.RequiresReplace(),
	}
}

func toResourceId(attr tfsdk.Attribute) tfsdk.Attribute {
	attr.Computed = true
	attr.PlanModifiers = tfsdk.AttributePlanModifiers{
		tfsdk.UseStateForUnknown(),
	}
	return attr
}

func toRequired(attr tfsdk.Attribute) tfsdk.Attribute {
	attr.Required = true
	return attr
}
