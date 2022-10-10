package provider

import (
	"github.com/PGSSoft/terraform-provider-mssql/internal/provider/acctest"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

var testCtx = acctest.NewContext(func(connection sql.Connection) provider.Provider {
	return &mssqlProvider{
		Version: VersionTest,
		Db:      connection,
	}
})

type testRunner struct {
	m *testing.M
}

func (r *testRunner) Run() int {
	defer testCtx.Cleanup()
	return r.m.Run()
}

func TestMain(m *testing.M) {
	resource.TestMain(&testRunner{m: m})
}
