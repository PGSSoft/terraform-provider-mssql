package sql

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
)

type ServerRoleSettings struct {
	Name    string
	OwnerId GenericServerPrincipalId
}

type ServerRoleMember struct {
	Id   GenericServerPrincipalId
	Name string
	Type ServerPrincipalType
}

type ServerRoleMembers map[GenericServerPrincipalId]ServerRoleMember

type ServerRole interface {
	GetId(ctx context.Context) ServerRoleId
	GetSettings(ctx context.Context) ServerRoleSettings
	Rename(ctx context.Context, name string)
	Drop(ctx context.Context)
	HasMember(ctx context.Context, id GenericServerPrincipalId) bool
	AddMember(ctx context.Context, id GenericServerPrincipalId)
	RemoveMember(ctx context.Context, id GenericServerPrincipalId)
	GetMembers(ctx context.Context) ServerRoleMembers
}

type ServerRoles map[ServerRoleId]ServerRole

func GetServerRole(_ context.Context, conn Connection, id ServerRoleId) ServerRole {
	return serverRole{conn: conn, id: id}
}

func GetServerRoleByName(ctx context.Context, conn Connection, name string) ServerRole {
	return GetServerRole(ctx, conn, ServerRoleId(conn.lookupServerPrincipalId(ctx, name)))
}

func GetServerRoles(ctx context.Context, conn Connection) ServerRoles {
	sqlConn := conn.getSqlConnection(ctx)

	if utils.HasError(ctx) {
		return nil
	}

	res, err := sqlConn.QueryContext(ctx, "SELECT [principal_id] FROM sys.server_principals WHERE [type]='R'")

	roles := ServerRoles{}
	switch err {
	case sql.ErrNoRows:
		return roles
	case nil:
		for res.Next() {
			var id ServerRoleId
			err := res.Scan(&id)
			utils.AddError(ctx, "Failed to parse server roles result", err)
			roles[id] = GetServerRole(ctx, conn, id)
		}
	default:
		utils.AddError(ctx, "Failed to fetch server roles", err)
	}

	return roles
}

func CreateServerRole(ctx context.Context, conn Connection, settings ServerRoleSettings) ServerRole {
	stat := fmt.Sprintf("CREATE SERVER ROLE [%s]", settings.Name)

	if settings.OwnerId != EmptyServerPrincipalId {
		ownerName := conn.lookupServerPrincipalName(ctx, settings.OwnerId)
		stat += fmt.Sprintf(" AUTHORIZATION [%s]", ownerName)
	}

	var role ServerRole
	utils.StopOnError(ctx).
		Then(func() { conn.exec(ctx, stat) }).
		Then(func() { role = GetServerRoleByName(ctx, conn, settings.Name) })

	return role
}

var _ ServerRole = serverRole{}

type serverRole struct {
	conn Connection
	id   ServerRoleId
}

func (s serverRole) GetId(context.Context) ServerRoleId {
	return s.id
}

func (s serverRole) GetSettings(ctx context.Context) ServerRoleSettings {
	settings := ServerRoleSettings{}
	conn := s.conn.getSqlConnection(ctx)

	utils.StopOnError(ctx).
		Then(func() {
			err := conn.
				QueryRowContext(ctx, "SELECT [name], [owning_principal_id] FROM sys.server_principals WHERE [principal_id]=@p1", s.id).
				Scan(&settings.Name, &settings.OwnerId)

			utils.AddError(ctx, "Failed to retrieve server role settings", err)
		})

	return settings
}

func (s serverRole) Rename(ctx context.Context, name string) {
	var oldName string
	conn := s.conn.getSqlConnection(ctx)

	utils.StopOnError(ctx).
		Then(func() { oldName = s.conn.lookupServerPrincipalName(ctx, GenericServerPrincipalId(s.id)) }).
		Then(func() {
			_, err := conn.ExecContext(ctx, fmt.Sprintf("ALTER SERVER ROLE [%s] WITH NAME = [%s]", oldName, name))
			utils.AddError(ctx, "Failed to rename server role", err)
		})
}

func (s serverRole) Drop(ctx context.Context) {
	var name string
	conn := s.conn.getSqlConnection(ctx)

	utils.StopOnError(ctx).
		Then(func() { name = s.conn.lookupServerPrincipalName(ctx, GenericServerPrincipalId(s.id)) }).
		Then(func() {
			_, err := conn.ExecContext(ctx, fmt.Sprintf("DROP SERVER ROLE [%s]", name))
			utils.AddError(ctx, "Failed to drop server role", err)
		})
}

func (s serverRole) HasMember(ctx context.Context, id GenericServerPrincipalId) bool {
	conn := s.conn.getSqlConnection(ctx)
	var result bool

	utils.StopOnError(ctx).
		Then(func() {
			err := conn.QueryRowContext(ctx, "SELECT 1 FROM sys.server_role_members WHERE [role_principal_id]=@p1 AND [member_principal_id]=@p2", s.id, id).Err()

			switch err {
			case sql.ErrNoRows:
				result = false
			case nil:
				result = true
			default:
				utils.AddError(ctx, "Failed to check role membership", err)
			}
		})

	return result
}

func (s serverRole) AddMember(ctx context.Context, id GenericServerPrincipalId) {
	var roleName, memberName string
	conn := s.conn.getSqlConnection(ctx)

	utils.StopOnError(ctx).
		Then(func() {
			roleName = s.conn.lookupServerPrincipalName(ctx, GenericServerPrincipalId(s.id))
			memberName = s.conn.lookupServerPrincipalName(ctx, id)
		}).
		Then(func() {
			_, err := conn.ExecContext(ctx, fmt.Sprintf("ALTER SERVER ROLE [%s] ADD MEMBER [%s]", roleName, memberName))
			utils.AddError(ctx, "Failed to add role member", err)
		})
}

func (s serverRole) RemoveMember(ctx context.Context, id GenericServerPrincipalId) {
	var roleName, memberName string
	conn := s.conn.getSqlConnection(ctx)

	utils.StopOnError(ctx).
		Then(func() {
			roleName = s.conn.lookupServerPrincipalName(ctx, GenericServerPrincipalId(s.id))
			memberName = s.conn.lookupServerPrincipalName(ctx, id)
		}).
		Then(func() {
			_, err := conn.ExecContext(ctx, fmt.Sprintf("ALTER SERVER ROLE [%s] DROP MEMBER [%s]", roleName, memberName))
			utils.AddError(ctx, "Failed to remove role member", err)
		})
}

func (s serverRole) GetMembers(ctx context.Context) ServerRoleMembers {
	conn := s.conn.getSqlConnection(ctx)
	if utils.HasError(ctx) {
		return nil
	}

	result := ServerRoleMembers{}
	rs, err := conn.QueryContext(ctx, `
SELECT [principal_id], [name], [type] FROM sys.server_role_members
INNER JOIN sys.server_principals ON [member_principal_id] = [principal_id]
WHERE [role_principal_id]=@p1 AND [type] IN ('S', 'R')`, s.id)

	switch err {
	case sql.ErrNoRows:
		return result
	case nil:
		for rs.Next() {
			var mType string
			member := ServerRoleMember{Type: UNKNOWN}
			err := rs.Scan(&member.Id, &member.Name, &mType)
			utils.AddError(ctx, "Failed to parse member result", err)

			switch mType {
			case "S":
				member.Type = SQL_LOGIN
			case "R":
				member.Type = SERVER_ROLE
			}

			result[member.Id] = member
		}
	default:
		utils.AddError(ctx, "Failed to retrieve role members", err)
	}

	return result
}
