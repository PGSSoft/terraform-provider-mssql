package schema

import (
	"github.com/PGSSoft/terraform-provider-mssql/internal/core"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/resource"
	sdkdatasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	sdkresource "github.com/hashicorp/terraform-plugin-framework/resource"
)

func Service() core.Service {
	return service{}
}

type service struct{}

func (s service) Name() string {
	return "schema"
}

func (s service) Resources() []func() sdkresource.ResourceWithConfigure {
	return []func() sdkresource.ResourceWithConfigure{
		resource.NewResource[resourceData](&res{}),
	}
}

func (s service) DataSources() []func() sdkdatasource.DataSourceWithConfigure {
	return []func() sdkdatasource.DataSourceWithConfigure{
		//datasource.NewDataSource[resourceData](&dataSource{}),
		//datasource.NewDataSource[listDataSourceData](&listDataSource{}),
	}
}

func (s service) Tests() core.AccTests {
	return core.AccTests{
		//DataSource:     testDataSource,
		//ListDataSource: testListDataSource,
		Resource: testResource,
	}
}
