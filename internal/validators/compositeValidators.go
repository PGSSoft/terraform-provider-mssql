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
