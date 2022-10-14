package common

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"strconv"
	"strings"
)

type DbObjectId[T sql.NumericObjectId] struct {
	DbId     sql.DatabaseId
	ObjectId T
	IsEmpty  bool
}

func (id DbObjectId[T]) String() string {
	return fmt.Sprintf("%v/%v", id.DbId, id.ObjectId)
}

func ParseDbObjectId[T sql.NumericObjectId](ctx context.Context, s string) DbObjectId[T] {
	var res DbObjectId[T]
	segments := parseIdSegments(ctx, s)

	if len(segments) < 2 || utils.HasError(ctx) {
		res.IsEmpty = true
	} else {
		res.DbId = sql.DatabaseId(segments[0])
		res.ObjectId = T(segments[1])
	}

	return res
}

type DbObjectMemberId[TObject sql.NumericObjectId, TMember sql.NumericObjectId] struct {
	DbObjectId[TObject]
	MemberId TMember
}

func ParseDbObjectMemberId[TObject sql.NumericObjectId, TMember sql.NumericObjectId](ctx context.Context, s string) DbObjectMemberId[TObject, TMember] {
	res := DbObjectMemberId[TObject, TMember]{DbObjectId: ParseDbObjectId[TObject](ctx, s)}
	segments := parseIdSegments(ctx, s)

	if len(segments) < 3 || utils.HasError(ctx) {
		res.IsEmpty = true
	} else {
		res.MemberId = TMember(segments[2])
	}

	return res
}

func (id DbObjectMemberId[TObject, TMember]) String() string {
	return fmt.Sprintf("%s/%d", id.DbObjectId, id.MemberId)
}

func (id DbObjectMemberId[TObject, TMember]) GetMemberId() DbObjectId[TMember] {
	return DbObjectId[TMember]{DbId: id.DbId, ObjectId: id.MemberId}
}

func parseIdSegments(ctx context.Context, s string) []int {
	var segments []int

	for _, seg := range strings.Split(s, "/") {
		if seg != "" {
			num, err := strconv.Atoi(seg)
			if err != nil {
				utils.AddError(ctx, fmt.Sprintf("Failed to parse DB object ID %q", s), err)
				return nil
			}

			segments = append(segments, num)
		}
	}

	return segments
}
