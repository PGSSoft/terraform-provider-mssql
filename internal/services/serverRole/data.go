package serverRole

import (
	"context"
	"errors"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/datasource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSourceDataMember struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
}

type dataSourceData struct {
	Id      types.String           `tfsdk:"id"`
	Name    types.String           `tfsdk:"name"`
	OwnerId types.String           `tfsdk:"owner_id"`
	Members []dataSourceDataMember `tfsdk:"members"`
}

var _ datasource.DataSourceWithValidation[dataSourceData] = dataSource{}

type dataSource struct{}

func (d dataSource) GetName() string {
	return "server_role"
}

func (d dataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	const requiredNote = " Either `name` or `id` must be provided."
	resp.Schema.MarkdownDescription = "Obtains information about single server role."
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: attrDescriptions["id"] + requiredNote,
			Computed:            true,
			Optional:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: attrDescriptions["name"] + requiredNote,
			Computed:            true,
			Optional:            true,
		},
		"owner_id": schema.StringAttribute{
			MarkdownDescription: attrDescriptions["owner_id"],
			Computed:            true,
		},
		"members": schema.SetNestedAttribute{
			MarkdownDescription: "Set of role members",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						MarkdownDescription: "ID of the member principal",
						Computed:            true,
					},
					"name": schema.StringAttribute{
						MarkdownDescription: "Name of the server principal",
						Computed:            true,
					},
					"type": schema.StringAttribute{
						MarkdownDescription: "One of: `SQL_LOGIN`, `SERVER_ROLE`",
						Computed:            true,
					},
				},
			},
		},
	}
}

func (d dataSource) Read(ctx context.Context, req datasource.ReadRequest[dataSourceData], resp *datasource.ReadResponse[dataSourceData]) {
	id := sql.ServerRoleId(0)

	if common.IsAttrSet(req.Config.Id) {
		id = parseId(ctx, req.Config.Id)
	}

	var role sql.ServerRole
	var settings sql.ServerRoleSettings
	var members sql.ServerRoleMembers

	req.
		Then(func() {
			if common.IsAttrSet(req.Config.Id) {
				role = sql.GetServerRole(ctx, req.Conn, id)
			} else {
				role = sql.GetServerRoleByName(ctx, req.Conn, req.Config.Name.ValueString())
			}
		}).
		Then(func() {
			settings = role.GetSettings(ctx)
			members = role.GetMembers(ctx)
		}).
		Then(func() {
			state := dataSourceData{
				Id:      types.StringValue(fmt.Sprint(role.GetId(ctx))),
				Name:    types.StringValue(settings.Name),
				OwnerId: types.StringValue(fmt.Sprint(settings.OwnerId)),
				Members: []dataSourceDataMember{},
			}

			for _, m := range members {
				member := dataSourceDataMember{
					Id:   types.StringValue(fmt.Sprint(m.Id)),
					Name: types.StringValue(m.Name),
				}

				switch m.Type {
				case sql.SQL_LOGIN:
					member.Type = types.StringValue("SQL_LOGIN")
				case sql.SERVER_ROLE:
					member.Type = types.StringValue("SERVER_ROLE")
				default:
					utils.AddError(ctx, "Unknown server principal type", fmt.Errorf("received unexpected principal type %d", m.Type))
				}

				state.Members = append(state.Members, member)
			}

			resp.SetState(state)
		})
}

func (d dataSource) Validate(ctx context.Context, req datasource.ValidateRequest[dataSourceData], _ *datasource.ValidateResponse[dataSourceData]) {
	if !common.IsAttrSet(req.Config.Id) && !common.IsAttrSet(req.Config.Name) {
		utils.AddError(ctx, "Either name or id must be provided", errors.New("both name and id are empty values"))
	}
}
