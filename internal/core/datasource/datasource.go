package datasource

import (
	"context"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type StateSetter[TData any] func(state TData)

type ConfigureRequest struct {
	Conn sql.Connection
}

var _ utils.ErrorMonad = ReadRequest[any]{}

type ReadRequest[TData any] struct {
	monad  utils.ErrorMonad
	Conn   sql.Connection
	Config TData
}

func (r ReadRequest[TData]) Then(f func()) utils.ErrorMonad {
	return r.monad.Then(f)
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
	GetSchema(ctx context.Context) tfsdk.Schema
	Read(ctx context.Context, req ReadRequest[TData], resp *ReadResponse[TData])
}
