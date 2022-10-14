package core

import (
	"github.com/PGSSoft/terraform-provider-mssql/internal/acctest"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

type AccTest func(testCtx *acctest.TestContext)

type AccTests struct {
	Resource       AccTest
	DataSource     AccTest
	ListDataSource AccTest
}

type Service interface {
	Name() string
	Resources() []func() resource.ResourceWithConfigure
	DataSources() []func() datasource.DataSourceWithConfigure
	Tests() AccTests
}
