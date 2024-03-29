package serverPermission

import (
	"github.com/PGSSoft/terraform-provider-mssql/internal/core"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/datasource"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/resource"
	sdkdatasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	sdkresource "github.com/hashicorp/terraform-plugin-framework/resource"
)

func Service() core.Service {
	return service{}
}

type service struct{}

func (s service) Name() string {
	return "server_permission"
}

func (s service) Resources() []func() sdkresource.ResourceWithConfigure {
	return []func() sdkresource.ResourceWithConfigure{
		resource.NewResource[resourceData](&res{}),
	}
}

func (s service) DataSources() []func() sdkdatasource.DataSourceWithConfigure {
	return []func() sdkdatasource.DataSourceWithConfigure{
		datasource.NewDataSource[listDataSourceData](&listDataSource{}),
	}
}

func (s service) Tests() core.AccTests {
	return core.AccTests{
		ListDataSource: testListDataSource,
		Resource:       testResource,
	}
}
