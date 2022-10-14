package databaseRoleMember

import (
	"github.com/PGSSoft/terraform-provider-mssql/internal/core"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core/resource"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	sdkResource "github.com/hashicorp/terraform-plugin-framework/resource"
)

func Service() core.Service {
	return service{}
}

type service struct{}

func (s service) Name() string {
	return "database_role_member"
}

func (s service) Resources() []func() sdkResource.ResourceWithConfigure {
	return []func() sdkResource.ResourceWithConfigure{
		resource.NewResource[resourceData](&res{}),
	}
}

func (s service) DataSources() []func() datasource.DataSourceWithConfigure {
	return []func() datasource.DataSourceWithConfigure{}
}

func (s service) Tests() core.AccTests {
	return core.AccTests{
		Resource: testResource,
	}
}
