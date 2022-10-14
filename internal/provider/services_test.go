package provider

import (
	"github.com/PGSSoft/terraform-provider-mssql/internal/acctest"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"testing"
)

func TestServices(t *testing.T) {
	ctx := acctest.NewContext(t, func(connection sql.Connection) provider.Provider {
		return &mssqlProvider{
			Version: VersionTest,
			Db:      connection,
		}
	})

	defer ctx.Cleanup()

	for _, svc := range Services() {
		ctx.Run(svc.Name(), func(svcCtx *acctest.TestContext) {
			if test := svc.Tests().Resource; test != nil {
				svcCtx.Run("resource", test)
			}

			if test := svc.Tests().DataSource; test != nil {
				svcCtx.Run("data_source", test)
			}

			if test := svc.Tests().ListDataSource; test != nil {
				svcCtx.Run("list", test)
			}
		})
	}
}
