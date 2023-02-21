package attrs

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func NumericIdType[T sql.NumericObjectId]() basetypes.StringTypable {
	var t basetypes.StringTypable
	t = numericIdType[T]{
		compositeIdType{
			elemCount: 1,
			valueFactory: func(id CompositeId) basetypes.StringValuable {
				id.attrType = &t
				return NumericId[T]{id}
			},
		},
	}
	return t
}

type numericIdType[T sql.NumericObjectId] struct {
	compositeIdType
}

func NumericIdValue[T sql.NumericObjectId](id T) NumericId[T] {
	t := NumericIdType[T]()

	return NumericId[T]{
		CompositeId{
			attrType: &t,
			elems:    []string{fmt.Sprint(id)},
		},
	}
}

type NumericId[T sql.NumericObjectId] struct {
	CompositeId
}

func (id NumericId[T]) Id(ctx context.Context) T {
	return T(id.GetInt(ctx, 0))
}
