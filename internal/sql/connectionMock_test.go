package sql

import (
	"context"
	"database/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/stretchr/testify/mock"
)

type SQLEdition string

const (
	EDITION_AZURE_SQL  SQLEdition = "SQL Azure"
	EDITION_ENTERPRISE SQLEdition = "Enterprise Edition"
)

var _ Connection = &connectionMock{}

type connectionMock struct {
	mock.Mock
	db      *sql.DB
	edition SQLEdition
}

func (c *connectionMock) IsAzure(ctx context.Context) bool {
	return c.edition == EDITION_AZURE_SQL
}

func (c *connectionMock) exec(ctx context.Context, query string, args ...any) sql.Result {
	res, err := c.db.ExecContext(ctx, query, args...)
	if err != nil {
		utils.AddError(ctx, "mock error", err)
	}
	return res
}

func (c *connectionMock) getConnectionDetails(ctx context.Context) ConnectionDetails {
	return c.Called(ctx).Get(0).(ConnectionDetails)
}

func (c *connectionMock) getSqlConnection(ctx context.Context) *sql.DB {
	return c.db
}

func (c *connectionMock) getDBSqlConnection(ctx context.Context, dbName string) *sql.DB {
	return c.db
}
