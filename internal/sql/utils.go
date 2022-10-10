package sql

import (
	"context"
	"database/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
)

func WithConnection[T any](ctx context.Context, connectionFactory func(context.Context) *sql.DB, action func(*sql.DB) T) T {
	conn := connectionFactory(ctx)
	if utils.HasError(ctx) {
		var result T
		return result
	}

	return action(conn)
}

func getPrincipalName[T DatabasePrincipalId](ctx context.Context, conn *sql.DB, id T) string {
	var res string

	if err := conn.QueryRowContext(ctx, "SELECT USER_NAME(@p1)", id).Scan(&res); err != nil {
		utils.AddError(ctx, "Failed to retrieve DB principal name", err)
	}

	return res
}

func getCurrentUserName(ctx context.Context, conn *sql.DB) string {
	var res string

	if err := conn.QueryRowContext(ctx, "SELECT USER_NAME()").Scan(&res); err != nil {
		utils.AddError(ctx, "Failed to retrieve current user name", err)
	}

	return res
}
