package provider

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"strconv"
	"strings"
)

func getResourceDb(ctx context.Context, conn sql.Connection, dbId string) sql.Database {
	if dbId == "" {
		return sql.GetDatabaseByName(ctx, conn, "master")
	}

	id, err := strconv.Atoi(dbId)
	if err != nil {
		utils.AddError(ctx, "Failed to convert DB ID", err)
		return nil
	}

	return sql.GetDatabase(ctx, conn, sql.DatabaseId(id))
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

type dbObjectId[T sql.NumericObjectId] struct {
	DbId     sql.DatabaseId
	ObjectId T
	IsEmpty  bool
}

func parseDbObjectId[T sql.NumericObjectId](ctx context.Context, s string) dbObjectId[T] {
	var res dbObjectId[T]
	segments := parseIdSegments(ctx, s)

	if len(segments) < 2 || utils.HasError(ctx) {
		res.IsEmpty = true
	} else {
		res.DbId = sql.DatabaseId(segments[0])
		res.ObjectId = T(segments[1])
	}

	return res
}

func (id dbObjectId[T]) String() string {
	return fmt.Sprintf("%v/%v", id.DbId, id.ObjectId)
}

type dbObjectMemberId[TObject sql.NumericObjectId, TMember sql.NumericObjectId] struct {
	dbObjectId[TObject]
	MemberId TMember
}

func parseDbObjectMemberId[TObject sql.NumericObjectId, TMember sql.NumericObjectId](ctx context.Context, s string) dbObjectMemberId[TObject, TMember] {
	res := dbObjectMemberId[TObject, TMember]{dbObjectId: parseDbObjectId[TObject](ctx, s)}
	segments := parseIdSegments(ctx, s)

	if len(segments) < 3 || utils.HasError(ctx) {
		res.IsEmpty = true
	} else {
		res.MemberId = TMember(segments[2])
	}

	return res
}

func (id dbObjectMemberId[TObject, TMember]) String() string {
	return fmt.Sprintf("%s/%d", id.dbObjectId, id.MemberId)
}

func (id dbObjectMemberId[TObject, TMember]) getMemberId() dbObjectId[TMember] {
	return dbObjectId[TMember]{DbId: id.DbId, ObjectId: id.MemberId}
}

func isAttrSet[T attr.Value](attr T) bool {
	return !attr.IsUnknown() && !attr.IsNull()
}
