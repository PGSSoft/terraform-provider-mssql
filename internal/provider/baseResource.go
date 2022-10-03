package provider

import (
	"context"
	"github.com/PGSSoft/terraform-provider-mssql/internal/provider/datasource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/provider/resource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
)

type BaseResource struct {
	conn sql.Connection
}

func (r *BaseResource) Configure(_ context.Context, req resource.ConfigureRequest) {
	r.conn = req.Conn
}

type BaseDataSource struct {
	conn sql.Connection
}

func (d *BaseDataSource) Configure(_ context.Context, req datasource.ConfigureRequest) {
	d.conn = req.Conn
}
