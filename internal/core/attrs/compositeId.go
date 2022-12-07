package attrs

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"strconv"
	"strings"
)

type compIdType interface {
	getElementsCount() int
}

func CompositeIdType(elemCount int) attr.Type {
	var t attr.Type
	t = compositeIdType{elemCount: elemCount, valueFactory: func(id CompositeId) attr.Value {
		id.attrType = &t
		return id
	}}
	return t
}

type compositeIdType struct {
	elemCount    int
	valueFactory func(id CompositeId) attr.Value
}

func (t compositeIdType) TerraformType(context.Context) tftypes.Type {
	return tftypes.String
}

func (t compositeIdType) ValueFromTerraform(_ context.Context, value tftypes.Value) (attr.Value, error) {
	if value.IsNull() {
		return t.valueFactory(CompositeId{isNull: true}), nil
	}

	if !value.IsKnown() {
		return t.valueFactory(CompositeId{isUnknown: true}), nil
	}

	var strVal string
	if err := value.As(&strVal); err != nil {
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
	attrType  *attr.Type
}

func (id CompositeId) Type(context.Context) attr.Type {
	return *id.attrType
}

func (id CompositeId) ToTerraformValue(context.Context) (tftypes.Value, error) {
	if id.isUnknown {
		return tftypes.NewValue(tftypes.String, tftypes.UnknownValue), nil
	}

	if id.isNull {
		return tftypes.NewValue(tftypes.String, nil), nil
	}

	return tftypes.NewValue(tftypes.String, id.String()), nil
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
