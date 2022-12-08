package databaseRole

import (
	"context"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/datasource"
	common2 "github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

type dataSource struct{}

func (d *dataSource) GetName() string {
	return "database_role"
}

func (d *dataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.MarkdownDescription = "Obtains information about single database role."
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: roleAttributeDescriptions["id"],
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: roleAttributeDescriptions["name"],
			Required:            true,
		},
		"database_id": schema.StringAttribute{
			MarkdownDescription: common2.AttributeDescriptions["database_id"] + " Defaults to ID of `master`.",
			Optional:            true,
		},
		"owner_id": schema.StringAttribute{
			MarkdownDescription: roleAttributeDescriptions["owner_id"],
			Computed:            true,
		},
		"members": schema.SetNestedAttribute{
			MarkdownDescription: "Set of role members",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						MarkdownDescription: "`<database_id>/<member_id>`. Member ID can be retrieved using `SELECT DATABASE_PRINCIPAL_ID('<member_name>')",
						Computed:            true,
					},
					"name": schema.StringAttribute{
						Description: "Name of the database principal.",
						Computed:    true,
					},
					"type": schema.StringAttribute{
						Description: "One of: `SQL_USER`, `DATABASE_ROLE`, `AZUREAD_USER`",
						Computed:    true,
					},
				},
			},
		},
	}
}

func (d *dataSource) Read(ctx context.Context, req datasource.ReadRequest[dataSourceData], resp *datasource.ReadResponse[dataSourceData]) {
	var (
		db   sql.Database
		role sql.DatabaseRole
	)

	req.
		Then(func() { db = common2.GetResourceDb(ctx, req.Conn, req.Config.DatabaseId.ValueString()) }).
		Then(func() { role = sql.GetDatabaseRoleByName(ctx, db, req.Config.Name.ValueString()) }).
		Then(func() { resp.SetState(req.Config.withRoleData(ctx, role)) })
}
