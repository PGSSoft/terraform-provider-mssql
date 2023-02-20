package provider

import (
	"context"
	"github.com/PGSSoft/terraform-provider-mssql/internal/core"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/pkg/errors"
)

// To ensure provider fully satisfies framework interfaces
var _ provider.ProviderWithValidateConfig = &mssqlProvider{}

const (
	VersionDev  = "dev"
	VersionTest = "test"
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &mssqlProvider{
			Version: version,
		}
	}
}

type mssqlProvider struct {
	Version string
	Db      sql.Connection
}

func (p *mssqlProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "mssql"
	resp.Version = p.Version
}

func (p *mssqlProvider) Configure(_ context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	resCtx := core.ResourceContext{}

	if p.Version == VersionTest {
		resCtx.ConnFactory = func(ctx context.Context) sql.Connection {
			return p.Db
		}
	} else {
		resCtx.ConnFactory = func(ctx context.Context) sql.Connection {
			var data providerData

			d := request.Config.Get(ctx, &data)

			if utils.AppendDiagnostics(ctx, d...); utils.HasError(ctx) {
				return nil
			}

			connDetails, d := data.asConnectionDetails(ctx)

			if utils.AppendDiagnostics(ctx, d...); utils.HasError(ctx) {
				return nil
			}

			conn, d := connDetails.Open(ctx)
			utils.AppendDiagnostics(ctx, d...)

			return conn
		}

	}

	response.DataSourceData = resCtx
	response.ResourceData = resCtx
}

func (p *mssqlProvider) Resources(context.Context) []func() resource.Resource {
	var ctors []func() resource.Resource

	for _, svc := range Services() {
		for _, svcRes := range svc.Resources() {
			ctor := svcRes
			ctors = append(ctors, func() resource.Resource { return ctor() })
		}
	}

	return ctors
}

func (p *mssqlProvider) DataSources(context.Context) []func() datasource.DataSource {
	var ctors []func() datasource.DataSource

	for _, svc := range Services() {
		for _, svcDataSource := range svc.DataSources() {
			ctor := svcDataSource
			ctors = append(ctors, func() datasource.DataSource { return ctor() })
		}
	}

	return ctors
}

func (p *mssqlProvider) Schema(_ context.Context, _ provider.SchemaRequest, response *provider.SchemaResponse) {
	if p.Version == VersionTest {
		return
	}

	response.Schema.Attributes = map[string]schema.Attribute{
		"hostname": schema.StringAttribute{
			Description: "FQDN or IP address of the SQL endpoint. Can be also set using `MSSQL_HOSTNAME` environment variable.",
			Optional:    true,
		},
		"port": schema.Int64Attribute{
			MarkdownDescription: "TCP port of SQL endpoint. Defaults to `1433`. Can be also set using `MSSQL_PORT` environment variable.",
			Optional:            true,
		},
		"sql_auth": schema.SingleNestedAttribute{
			Description: "When provided, SQL authentication will be used when connecting.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"username": schema.StringAttribute{
					Description: "User name for SQL authentication.",
					Required:    true,
				},
				"password": schema.StringAttribute{
					Description: "Password for SQL authentication.",
					Required:    true,
					Sensitive:   true,
				},
			},
		},
		"azure_auth": schema.SingleNestedAttribute{
			Description: "When provided, Azure AD authentication will be used when connecting.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"client_id": schema.StringAttribute{
					Description: "Service Principal client (application) ID. When omitted, default, chained set of credentials will be used.",
					Optional:    true,
				},
				"client_secret": schema.StringAttribute{
					Description: "Service Principal secret. When omitted, default, chained set of credentials will be used.",
					Sensitive:   true,
					Optional:    true,
				},
				"tenant_id": schema.StringAttribute{
					Description: "Azure AD tenant ID. Required only if Azure SQL Server's tenant is different than Service Principal's.",
					Optional:    true,
				},
			},
		},
	}
}

func (p *mssqlProvider) ValidateConfig(ctx context.Context, request provider.ValidateConfigRequest, response *provider.ValidateConfigResponse) {
	if p.Version == VersionTest {
		return
	}

	var data providerData

	ctx = utils.WithDiagnostics(ctx, &response.Diagnostics)

	utils.StopOnError(ctx).
		Then(func() { data = utils.GetData[providerData](ctx, request.Config) }).
		Then(func() {
			if data.AzureAuth == nil && data.SqlAuth == nil {
				utils.AddError(ctx, "Missing SQL authentication config", errors.New("One of authentication methods must be provided: sql_auth, azure_auth"))
			}
		})
}
