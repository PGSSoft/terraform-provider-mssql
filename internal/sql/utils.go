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
	defer conn.Close()

	return action(conn)
}
