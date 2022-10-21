package provider

import (
	"github.com/PGSSoft/terraform-provider-mssql/internal/core"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/azureADServicePrincipal"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/azureADUser"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/database"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/databaseRole"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/databaseRoleMember"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/script"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/sqlLogin"
	"github.com/PGSSoft/terraform-provider-mssql/internal/services/sqlUser"
)

func Services() []core.Service {
	return []core.Service{
		azureADServicePrincipal.Service(),
		azureADUser.Service(),

		database.Service(),
		databaseRole.Service(),
		databaseRoleMember.Service(),
		sqlLogin.Service(),
		sqlUser.Service(),

		script.Service(),
	}
}
