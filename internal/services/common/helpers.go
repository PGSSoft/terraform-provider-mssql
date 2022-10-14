package common

import (
	"context"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"strconv"
)

func GetResourceDb(ctx context.Context, conn sql.Connection, dbId string) sql.Database {
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

func IsAttrSet[T attr.Value](attr T) bool {
	return !attr.IsUnknown() && !attr.IsNull()
}
