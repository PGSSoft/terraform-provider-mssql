package azureADServicePrincipal

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/datasource"
	common2 "github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
)

type dataSource struct{}

func (d *dataSource) GetName() string {
	return "azuread_service_principal"
}

func (d *dataSource) GetSchema(_ context.Context) tfsdk.Schema {
	attrs := map[string]tfsdk.Attribute{}
	for n, attr := range attributes {
		attr.Required = n == "database_id"
		attr.Optional = n != "id" && !attr.Required
		attr.Computed = !attr.Required
		attrs[n] = attr
	}

	return tfsdk.Schema{
		Description: "Obtains information about single Azure AD Service Principal database user.",
		Attributes:  attrs,
	}
}

func (d *dataSource) Read(ctx context.Context, request datasource.ReadRequest[resourceData], response *datasource.ReadResponse[resourceData]) {
	var (
		db   sql.Database
		user sql.User
	)

	request.
		Then(func() { db = common2.GetResourceDb(ctx, request.Conn, request.Config.DatabaseId.ValueString()) }).
		Then(func() {
			if !request.Config.Name.IsNull() && !request.Config.Name.IsUnknown() {
				user = sql.GetUserByName(ctx, db, request.Config.Name.ValueString())
				return
			}

			for _, u := range sql.GetUsers(ctx, db) {
				settings := u.GetSettings(ctx)
				if settings.Type == sql.USER_TYPE_AZUREAD && strings.EqualFold(fmt.Sprint(settings.AADObjectId), request.Config.ClientId.ValueString()) {
					user = u
					return
				}
			}

			utils.AddError(ctx, "User does not exist", fmt.Errorf("could not find user with name=%q and client_id=%q", request.Config.Name.ValueString(), request.Config.ClientId.ValueString()))
		}).
		Then(func() {
			state := request.Config
			state.Id = types.StringValue(common2.DbObjectId[sql.UserId]{DbId: db.GetId(ctx), ObjectId: user.GetId(ctx)}.String())
			state = state.withSettings(ctx, user.GetSettings(ctx))
			response.SetState(state)
		})
}
