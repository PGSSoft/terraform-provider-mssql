package validators

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var DatabaseNameValidators = []tfsdk.AttributeValidator{
	sqlIdentifierValidator{},
	stringLengthValidator{1, 128},
}

var LoginNameValidators = []tfsdk.AttributeValidator{
	sqlIdentifierValidator{},
}

var UserNameValidators = []tfsdk.AttributeValidator{
	stringLengthValidator{Min: 1, Max: 128},
}

var SchemaNameValidators = []tfsdk.AttributeValidator{
	stringLengthValidator{Min: 1, Max: 128},
}
