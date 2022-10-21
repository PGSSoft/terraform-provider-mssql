package sql

import (
	"context"
	"database/sql"
	"github.com/stretchr/testify/mock"
)

var _ Database = &dbMock{}

type dbMock struct {
	mock.Mock
}

func (m *dbMock) GetConnection(ctx context.Context) Connection {
	return m.Called(ctx).Get(0).(Connection)
}

func (m *dbMock) GetId(ctx context.Context) DatabaseId {
	return m.Called(ctx).Get(0).(DatabaseId)
}

func (m *dbMock) Exists(ctx context.Context) bool {
	return m.Called(ctx).Bool(0)
}

func (m *dbMock) GetSettings(ctx context.Context) DatabaseSettings {
	return m.Called(ctx).Get(0).(DatabaseSettings)
}

func (m *dbMock) Rename(ctx context.Context, name string) {
	m.Called(ctx, name)
}

func (m *dbMock) SetCollation(ctx context.Context, collation string) {
	m.Called(ctx, collation)
}

func (m *dbMock) Drop(ctx context.Context) {
	m.Called(ctx)
}

func (m *dbMock) CreateUser(ctx context.Context, settings UserSettings) User {
	return m.Called(ctx, settings).Get(0).(User)
}

func (m *dbMock) GetUser(ctx context.Context, id UserId) User {
	return m.Called(ctx, id).Get(0).(User)
}

func (m *dbMock) Query(ctx context.Context, query string) []map[string]string {
	return m.Called(ctx, query).Get(0).([]map[string]string)
}

func (m *dbMock) Exec(ctx context.Context, script string) {
	m.Called(ctx, script)
}

func (m dbMock) connect(ctx context.Context) *sql.DB {
	return m.Called(ctx).Get(0).(*sql.DB)
}
