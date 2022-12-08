package validators

import (
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var DatabaseNameValidators = []validator.String{
	sqlIdentifierValidator{},
	stringLengthValidator{1, 128},
}

var LoginNameValidators = []validator.String{
	sqlIdentifierValidator{},
}

var UserNameValidators = []validator.String{
	stringLengthValidator{Min: 1, Max: 128},
}

var SchemaNameValidators = []validator.String{
	stringLengthValidator{Min: 1, Max: 128},
}
