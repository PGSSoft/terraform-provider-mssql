package attrs

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"strconv"
	"strings"
)

type compIdType interface {
	getElementsCount() int
}

func CompositeIdType(elemCount int) types.StringTypable {
	var t types.StringTypable
	t = compositeIdType{elemCount: elemCount, valueFactory: func(id CompositeId) types.StringValuable {
		id.attrType = &t
		return id
	}}
	return t
}

type compositeIdType struct {
	elemCount    int
	valueFactory func(id CompositeId) types.StringValuable
}

func (t compositeIdType) TerraformType(context.Context) tftypes.Type {
	return tftypes.String
}

func (t compositeIdType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	if value.IsNull() {
		return t.valueFactory(CompositeId{isNull: true}), nil
	}

	if !value.IsKnown() {
		return t.valueFactory(CompositeId{isUnknown: true}), nil
	}

	var err error = nil
	var strVal string
	if err = value.As(&strVal); err != nil {
		return nil, err
	}

	if strVal == "" {
		return t.valueFactory(CompositeId{isNull: true}), nil
	}

	elems := strings.Split(strVal, "/")
	if len(elems) != t.elemCount {
		return nil, fmt.Errorf("unexpected ID elements count, expected %d, got %d", t.elemCount, len(elems))
	}

	return t.valueFactory(CompositeId{elems: elems}), nil
}

func (t compositeIdType) ValueFromString(_ context.Context, value types.String) (types.StringValuable, diag.Diagnostics) {
	if value.IsNull() {
		return t.valueFactory(CompositeId{isNull: true}), nil
	}

	if value.IsUnknown() {
		return t.valueFactory(CompositeId{isUnknown: true}), nil
	}

	if value.ValueString() == "" {
		return t.valueFactory(CompositeId{isNull: true}), nil
	}

	elems := strings.Split(value.ValueString(), "/")
	if len(elems) != t.elemCount {
		diags := diag.Diagnostics{}
		diags.AddError("Invalid ID format", fmt.Sprintf("unexpected ID elements count, expected %d, got %d", t.elemCount, len(elems)))
		return nil, diags
	}

	return t.valueFactory(CompositeId{elems: elems}), nil
}

func (t compositeIdType) ValueType(context.Context) attr.Value {
	return t.valueFactory(CompositeId{})
}

func (t compositeIdType) Equal(t2 attr.Type) bool {
	t2composite, ok := t2.(compIdType)
	return ok && t2composite.getElementsCount() == t.elemCount
}

func (t compositeIdType) String() string {
	return fmt.Sprintf("compositeIdType[%d]", t.elemCount)
}

func (t compositeIdType) ApplyTerraform5AttributePathStep(tftypes.AttributePathStep) (interface{}, error) {
	return nil, nil
}

func (t compositeIdType) getElementsCount() int {
	return t.elemCount
}

func CompositeIdValue(elems ...string) CompositeId {
	t := CompositeIdType(len(elems))
	return CompositeId{elems: elems, attrType: &t}
}

type compId interface {
	getElements() []string
}

type CompositeId struct {
	isNull    bool
	isUnknown bool
	elems     []string
	attrType  *types.StringTypable
}

func (id CompositeId) Type(context.Context) attr.Type {
	return *id.attrType
}

func (id CompositeId) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	strVal, _ := id.ToStringValue(ctx)
	return strVal.ToTerraformValue(ctx)
}

func (id CompositeId) ToStringValue(context.Context) (types.String, diag.Diagnostics) {
	switch {
	case id.isUnknown:
		return types.StringUnknown(), nil
	case id.isNull:
		return types.StringNull(), nil
	default:
		return types.StringValue(id.String()), nil
	}
}

func (id CompositeId) Equal(value attr.Value) bool {
	id2, ok := value.(compId)
	if !ok || len(id2.getElements()) != len(id.elems) {
		return false
	}

	for i, v := range id.elems {
		if v != id2.getElements()[i] {
			return false
		}
	}

	return true
}

func (id CompositeId) IsNull() bool {
	return id.isNull
}

func (id CompositeId) IsUnknown() bool {
	return id.isUnknown
}

func (id CompositeId) String() string {
	return strings.Join(id.elems, "/")
}

func (id CompositeId) GetString(i int) string {
	return id.elems[i]
}

func (id CompositeId) GetInt(ctx context.Context, i int) int {
	res, err := strconv.Atoi(id.GetString(i))
	utils.AddError(ctx, "Invalid numeric ID value", err)
	return res
}

func (id CompositeId) getElements() []string {
	return id.elems
}
