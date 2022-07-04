package provider

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
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

type DbObjectId[T sql.NumericObjectId] struct {
	DbId     sql.DatabaseId
	ObjectId T
	IsEmpty  bool
}

func parseDbObjectId[T sql.NumericObjectId](ctx context.Context, s string) DbObjectId[T] {
	var res DbObjectId[T]
	var segments []int

	for _, seg := range strings.Split(s, "/") {
		if seg != "" {
			num, err := strconv.Atoi(seg)
			if err != nil {
				utils.AddError(ctx, fmt.Sprintf("Failed to parse DB object ID %q", s), err)
				return res
			}

			segments = append(segments, num)
		}
	}

	if len(segments) < 2 {
		res.IsEmpty = true
	} else {
		res.DbId = sql.DatabaseId(segments[0])
		res.ObjectId = T(segments[1])
	}

	return res
}

func (id DbObjectId[T]) String() string {
	return fmt.Sprintf("%v/%v", id.DbId, id.ObjectId)
}
