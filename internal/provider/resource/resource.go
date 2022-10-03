package resource

import (
	"context"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type StateSetter[TData any] func(state TData)

var _ utils.ErrorMonad = monadRequest{}

type monadRequest struct {
	monad utils.ErrorMonad
}

func (r monadRequest) Then(f func()) utils.ErrorMonad {
	return r.monad.Then(f)
}

type stateResponse[TData any] struct {
	State TData
}

type ConfigureRequest struct {
	Conn sql.Connection
}

type ReadRequest[TData any] struct {
	monadRequest
	State TData
}

type ReadResponse[TData any] struct {
	state  TData
	exists bool
}

func (r *ReadResponse[TData]) SetState(state TData) {
	r.state = state
	r.exists = true
}

type CreateRequest[TData any] struct {
	monadRequest
	Plan TData
}

type CreateResponse[TData any] stateResponse[TData]

type UpdateRequest[TData any] struct {
	monadRequest
	Plan  TData
	State TData
}

type UpdateResponse[TData any] stateResponse[TData]

type DeleteRequest[TData any] struct {
	monadRequest
	State TData
}

type DeleteResponse[TData any] struct{}

type Resource[TData any] interface {
	GetName() string
	Configure(ctx context.Context, req ConfigureRequest)
	GetSchema(ctx context.Context) tfsdk.Schema
	Read(ctx context.Context, req ReadRequest[TData], resp *ReadResponse[TData])
	Create(ctx context.Context, req CreateRequest[TData], resp *CreateResponse[TData])
	Update(ctx context.Context, req UpdateRequest[TData], resp *UpdateResponse[TData])
	Delete(ctx context.Context, req DeleteRequest[TData], resp *DeleteResponse[TData])
}
