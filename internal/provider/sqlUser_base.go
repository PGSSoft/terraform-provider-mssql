package provider

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/PGSSoft/terraform-provider-mssql/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
)

var sqlUserAttributes = map[string]tfsdk.Attribute{
	"id": {
		MarkdownDescription: "`<database_id>/<user_id>`. User ID can be retrieved using `SELECT DATABASE_PRINCIPAL_ID('<user_name>')`.",
		Type:                types.StringType,
	},
	"name": {
		MarkdownDescription: fmt.Sprintf("User name. %s and cannot be longer than 128 chars.", regularIdentifiersDoc),
		Type:                types.StringType,
		Validators:          validators.UserNameValidators,
	},
	"database_id": {
		MarkdownDescription: fmt.Sprintf("ID of database. Can be retrieved using `mssql_database` or `SELECT DB_ID('<db_name>')`."),
		Type:                types.StringType,
	},
	"login_id": {
		MarkdownDescription: "SID of SQL login. Can be retrieved using `mssql_sql_login` or `SELECT SUSER_SID('<login_name>')`.",
		Type:                types.StringType,
	},
}

type sqlUserResourceData struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	DatabaseId types.String `tfsdk:"database_id"`
	LoginId    types.String `tfsdk:"login_id"`
}

func (d sqlUserResourceData) toSettings() sql.UserSettings {
	return sql.UserSettings{
		Name:    d.Name.Value,
		LoginId: sql.LoginId(d.LoginId.Value),
	}
}

func (d sqlUserResourceData) withSettings(settings sql.UserSettings) sqlUserResourceData {
	return sqlUserResourceData{
		Id:         d.Id,
		DatabaseId: d.DatabaseId,
		Name:       types.String{Value: settings.Name},
		LoginId:    types.String{Value: fmt.Sprint(settings.LoginId)},
	}
}

func (d sqlUserResourceData) withIds(dbId sql.DatabaseId, userId sql.UserId) sqlUserResourceData {
	return sqlUserResourceData{
		Id:         types.String{Value: fmt.Sprintf("%v/%v", dbId, userId)},
		DatabaseId: types.String{Value: fmt.Sprint(dbId)},
		Name:       d.Name,
		LoginId:    d.LoginId,
	}
}

type sqlUserResourceBase struct {
	Resource
}

func (r sqlUserResourceBase) getDb(ctx context.Context, dbId string) sql.Database {
	if dbId == "" {
		return r.Db.GetDatabaseByName(ctx, "master")
	}

	id, err := strconv.Atoi(dbId)
	if err != nil {
		utils.AddError(ctx, "Failed to convert DB ID", err)
		return nil
	}

	return r.Db.GetDatabase(ctx, sql.DatabaseId(id))
}
