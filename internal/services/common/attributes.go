package common

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const RegularIdentifiersDoc = "Must follow [Regular Identifiers rules](https://docs.microsoft.com/en-us/sql/relational-databases/databases/database-identifiers#rules-for-regular-identifiers)"

var DatabaseIdAttribute = tfsdk.Attribute{
	MarkdownDescription: fmt.Sprintf("ID of database. Can be retrieved using `mssql_database` or `SELECT DB_ID('<db_name>')`."),
	Type:                types.StringType,
}

var DatabaseIdResourceAttribute = DatabaseIdAttribute

func init() {
	DatabaseIdResourceAttribute.Optional = true
	DatabaseIdResourceAttribute.Computed = true
	DatabaseIdResourceAttribute.MarkdownDescription += " Defaults to ID of `master`."
	DatabaseIdResourceAttribute.PlanModifiers = tfsdk.AttributePlanModifiers{
		resource.RequiresReplace(),
	}
}

func ToResourceId(attr tfsdk.Attribute) tfsdk.Attribute {
	attr.Computed = true
	attr.PlanModifiers = tfsdk.AttributePlanModifiers{
		resource.UseStateForUnknown(),
	}
	return attr
}

func ToRequired(attr tfsdk.Attribute) tfsdk.Attribute {
	attr.Required = true
	return attr
}

func ToRequiredImmutable(attr tfsdk.Attribute) tfsdk.Attribute {
	attr.Required = true
	attr.PlanModifiers = tfsdk.AttributePlanModifiers{resource.RequiresReplace()}
	return attr
}
