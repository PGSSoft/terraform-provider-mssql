package azureADUser

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/datasource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"strings"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataSource struct{}

func (d *dataSource) GetName() string {
	return "azuread_user"
}

func (d *dataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.MarkdownDescription = "Obtains information about single Azure AD database user."
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: attrDescriptions["id"],
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: attrDescriptions["name"],
			Validators:          validators.UserNameValidators,
			Optional:            true,
			Computed:            true,
		},
		"database_id": schema.StringAttribute{
			MarkdownDescription: common.AttributeDescriptions["database_id"],
			Required:            true,
		},
		"user_object_id": schema.StringAttribute{
			MarkdownDescription: attrDescriptions["user_object_id"],
			Optional:            true,
			Computed:            true,
		},
	}
}

func (d *dataSource) Read(ctx context.Context, req datasource.ReadRequest[resourceData], resp *datasource.ReadResponse[resourceData]) {
	var (
		db   sql.Database
		user sql.User
	)

	req.
		Then(func() { db = common.GetResourceDb(ctx, req.Conn, req.Config.DatabaseId.ValueString()) }).
		Then(func() {
			if !req.Config.Name.IsNull() && !req.Config.Name.IsUnknown() {
				user = sql.GetUserByName(ctx, db, req.Config.Name.ValueString())
				return
			}

			for _, u := range sql.GetUsers(ctx, db) {
				settings := u.GetSettings(ctx)
				if settings.Type == sql.USER_TYPE_AZUREAD && strings.ToUpper(fmt.Sprint(settings.AADObjectId)) == strings.ToUpper(req.Config.UserObjectId.ValueString()) {
					user = u
					return
				}
			}

			utils.AddError(ctx, "User does not exist", fmt.Errorf("could not find user with name=%q and object_id=%q", req.Config.Name.ValueString(), req.Config.UserObjectId.ValueString()))
		}).
		Then(func() {
			state := req.Config.withSettings(ctx, user.GetSettings(ctx))
			state.Id = types.StringValue(common.DbObjectId[sql.UserId]{DbId: db.GetId(ctx), ObjectId: user.GetId(ctx)}.String())
			resp.SetState(state)
		})
}
