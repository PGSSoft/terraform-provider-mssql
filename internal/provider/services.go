package provider

import (
	"github.com/PGSSoft/terraform-provider-mssql/internal/core"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/azureADServicePrincipal"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/azureADUser"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/database"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/databasePermission"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/databaseRole"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/databaseRoleMember"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/schema"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/schemaPermission"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/script"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/serverPermission"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/serverRole"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/serverRoleMember"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/sqlLogin"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/sqlUser"
)

func Services() []core.Service {
	return []core.Service{
		azureADServicePrincipal.Service(),
		azureADUser.Service(),

		database.Service(),
		databasePermission.Service(),
		databaseRole.Service(),
		databaseRoleMember.Service(),
		sqlLogin.Service(),
		sqlUser.Service(),
		schema.Service(),
		schemaPermission.Service(),
		serverRole.Service(),
		serverRoleMember.Service(),
		serverPermission.Service(),

		script.Service(),
	}
}
