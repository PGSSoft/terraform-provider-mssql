package datasource

import (
	"context"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

type StateSetter[TData any] func(state TData)

type ConfigureRequest struct {
	Conn sql.Connection
}

var _ utils.ErrorMonad = MonadRequest{}

type MonadRequest struct {
	monad utils.ErrorMonad
}

func (r MonadRequest) Then(f func()) utils.ErrorMonad {
	return r.monad.Then(f)
}

type SchemaRequest struct{}

type SchemaResponse struct {
	Schema schema.Schema
}

type ReadRequest[TData any] struct {
	MonadRequest
	Conn   sql.Connection
	Config TData
}

type ReadResponse[TData any] struct {
	state  TData
	exists bool
}

func (r *ReadResponse[TData]) SetState(state TData) {
	r.state = state
	r.exists = true
}

type DataSource[TData any] interface {
	GetName() string
	Schema(ctx context.Context, req SchemaRequest, resp *SchemaResponse)
	Read(ctx context.Context, req ReadRequest[TData], resp *ReadResponse[TData])
}

type ValidateRequest[TData any] struct {
	MonadRequest
	Config TData
}

type ValidateResponse[TData any] struct{}

type DataSourceWithValidation[TData any] interface {
	Validate(ctx context.Context, req ValidateRequest[TData], resp *ValidateResponse[TData])
}
