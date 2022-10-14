package core

import (
	"context"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
)

type ResourceContext struct {
	ConnFactory func(ctx context.Context) sql.Connection
}
