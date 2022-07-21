package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/PGSSoft/terraform-provider-mssql/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var databaseRoleAttributes = map[string]tfsdk.Attribute{
	"id": {
		MarkdownDescription: "`<database_id>/<role_id>`. Role ID can be retrieved using `SELECT DATABASE_PRINCIPAL_ID('<role_name>')`",
		Type:                types.StringType,
	},
	"name": {
		MarkdownDescription: fmt.Sprintf("Role name. %s and cannot be longer than 128 chars.", regularIdentifiersDoc),
		Type:                types.StringType,
		Validators:          validators.UserNameValidators,
	},
	"database_id": databaseIdAttribute,
	"owner_id": {
		MarkdownDescription: "ID of another database role or user owning this role. Can be retrieved using `mssql_database_role` or `mssql_sql_user`.",
		Type:                types.StringType,
	},
}

var databaseRoleMemberSetAttributes = map[string]tfsdk.Attribute{
	"id": {
		MarkdownDescription: "`<database_id>/<member_id>`. Member ID can be retrieved using `SELECT DATABASE_PRINCIPAL_ID('<member_name>')",
		Type:                types.StringType,
		Computed:            true,
	},
	"name": {
		Description: "Name of the database principal.",
		Type:        types.StringType,
		Computed:    true,
	},
	"type": {
		Description: "One of: `SQL_USER`, `DATABASE_ROLE`, `AZUREAD_USER`",
		Type:        types.StringType,
		Computed:    true,
	},
}

type databaseRoleResourceData struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	DatabaseId types.String `tfsdk:"database_id"`
	OwnerId    types.String `tfsdk:"owner_id"`
}

func (d databaseRoleResourceData) withRoleData(ctx context.Context, role sql.DatabaseRole) databaseRoleResourceData {
	dbId := role.GetDb(ctx).GetId(ctx)

	return databaseRoleResourceData{
		Id:         types.String{Value: dbObjectId[sql.DatabaseRoleId]{DbId: dbId, ObjectId: role.GetId(ctx)}.String()},
		Name:       types.String{Value: role.GetName(ctx)},
		DatabaseId: types.String{Value: fmt.Sprint(dbId)},
		OwnerId:    types.String{Value: dbObjectId[sql.GenericDatabasePrincipalId]{DbId: dbId, ObjectId: role.GetOwnerId(ctx)}.String()},
	}
}

type databaseRoleDataResourceMembersData struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
}

type databaseRoleDataResourceData struct {
	Id         types.String                          `tfsdk:"id"`
	Name       types.String                          `tfsdk:"name"`
	DatabaseId types.String                          `tfsdk:"database_id"`
	OwnerId    types.String                          `tfsdk:"owner_id"`
	Members    []databaseRoleDataResourceMembersData `tfsdk:"members"`
}

func (d databaseRoleDataResourceData) withRoleData(ctx context.Context, role sql.DatabaseRole) databaseRoleDataResourceData {
	dbId := role.GetDb(ctx).GetId(ctx)
	data := databaseRoleDataResourceData{
		Id:         types.String{Value: dbObjectId[sql.DatabaseRoleId]{DbId: dbId, ObjectId: role.GetId(ctx)}.String()},
		Name:       types.String{Value: role.GetName(ctx)},
		DatabaseId: types.String{Value: fmt.Sprint(dbId)},
		OwnerId:    types.String{Value: dbObjectId[sql.GenericDatabasePrincipalId]{DbId: dbId, ObjectId: role.GetOwnerId(ctx)}.String()},
	}

	mapType := func(typ sql.DatabasePrincipalType) string {
		switch typ {
		case sql.DATABASE_ROLE:
			return "DATABASE_ROLE"
		case sql.SQL_USER:
			return "SQL_USER"
		case sql.AZUREAD_USER:
			return "AZUREAD_USER"
		default:
			utils.AddError(ctx, "Unknown member type", errors.New(fmt.Sprintf("member type %d unknown", typ)))
			return ""
		}
	}

	for id, member := range role.GetMembers(ctx) {
		memberData := databaseRoleDataResourceMembersData{
			Id:   types.String{Value: dbObjectId[sql.GenericDatabasePrincipalId]{DbId: dbId, ObjectId: id}.String()},
			Name: types.String{Value: member.Name},
			Type: types.String{Value: mapType(member.Type)},
		}
		data.Members = append(data.Members, memberData)
	}

	return data
}
