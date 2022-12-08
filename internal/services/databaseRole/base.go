package databaseRole

import (
	"context"
	"errors"
	"fmt"
	common2 "github.com/PGSSoft/terraform-provider-mssql/internal/services/common"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var roleAttributeDescriptions = map[string]string{
	"id":       "`<database_id>/<role_id>`. Role ID can be retrieved using `SELECT DATABASE_PRINCIPAL_ID('<role_name>')`",
	"name":     fmt.Sprintf("Role name. %s and cannot be longer than 128 chars.", common2.RegularIdentifiersDoc),
	"owner_id": "ID of another database role or user owning this role. Can be retrieved using `mssql_database_role` or `mssql_sql_user`.",
}

type resourceData struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	DatabaseId types.String `tfsdk:"database_id"`
	OwnerId    types.String `tfsdk:"owner_id"`
}

func (d resourceData) withRoleData(ctx context.Context, role sql.DatabaseRole) resourceData {
	dbId := role.GetDb(ctx).GetId(ctx)

	return resourceData{
		Id:         types.StringValue(common2.DbObjectId[sql.DatabaseRoleId]{DbId: dbId, ObjectId: role.GetId(ctx)}.String()),
		Name:       types.StringValue(role.GetName(ctx)),
		DatabaseId: types.StringValue(fmt.Sprint(dbId)),
		OwnerId:    types.StringValue(common2.DbObjectId[sql.GenericDatabasePrincipalId]{DbId: dbId, ObjectId: role.GetOwnerId(ctx)}.String()),
	}
}

type resourceRoleMembersData struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
}

type dataSourceData struct {
	Id         types.String              `tfsdk:"id"`
	Name       types.String              `tfsdk:"name"`
	DatabaseId types.String              `tfsdk:"database_id"`
	OwnerId    types.String              `tfsdk:"owner_id"`
	Members    []resourceRoleMembersData `tfsdk:"members"`
}

func (d dataSourceData) withRoleData(ctx context.Context, role sql.DatabaseRole) dataSourceData {
	dbId := role.GetDb(ctx).GetId(ctx)
	data := dataSourceData{
		Id:         types.StringValue(common2.DbObjectId[sql.DatabaseRoleId]{DbId: dbId, ObjectId: role.GetId(ctx)}.String()),
		Name:       types.StringValue(role.GetName(ctx)),
		DatabaseId: types.StringValue(fmt.Sprint(dbId)),
		OwnerId:    types.StringValue(common2.DbObjectId[sql.GenericDatabasePrincipalId]{DbId: dbId, ObjectId: role.GetOwnerId(ctx)}.String()),
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
		memberData := resourceRoleMembersData{
			Id:   types.StringValue(common2.DbObjectId[sql.GenericDatabasePrincipalId]{DbId: dbId, ObjectId: id}.String()),
			Name: types.StringValue(member.Name),
			Type: types.StringValue(mapType(member.Type)),
		}
		data.Members = append(data.Members, memberData)
	}

	return data
}
