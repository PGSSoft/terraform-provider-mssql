package serverRole

import (
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/acctest"
)

func fetchPrincipalId(testCtx *acctest.TestContext, name string) string {
	var id string
	err := testCtx.GetMasterDBConnection().
		QueryRow(fmt.Sprintf("SELECT [principal_id] FROM sys.server_principals WHERE [name]=%s", name)).Scan(&id)
	testCtx.Require.NoError(err, "Fetching IDs")

	return id
}
