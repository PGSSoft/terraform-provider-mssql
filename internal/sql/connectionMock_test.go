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

func (c *connectionMock) GetPermissions(ctx context.Context, principalId GenericServerPrincipalId) ServerPermissions {
	return c.Called(ctx, principalId).Get(0).(ServerPermissions)
}

func (c *connectionMock) GrantPermission(ctx context.Context, principalId GenericServerPrincipalId, permission ServerPermission) {
	c.Called(ctx, principalId, permission)
}

func (c *connectionMock) RevokePermission(ctx context.Context, principalId GenericServerPrincipalId, permission string) {
	c.Called(ctx, principalId, permission)
}

func (c *connectionMock) exec(ctx context.Context, query string, args ...any) sql.Result {
	res, err := ExecContextWithRetry(ctx, c.db, query, args...)
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

func (c *connectionMock) lookupServerPrincipalName(ctx context.Context, id GenericServerPrincipalId) string {
	return c.Called(ctx, id).String(0)
}

func (c *connectionMock) lookupServerPrincipalId(ctx context.Context, name string) GenericServerPrincipalId {
	return c.Called(ctx, name).Get(0).(GenericServerPrincipalId)
}
