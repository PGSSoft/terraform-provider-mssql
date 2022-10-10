package provider

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"os"
)

// To ensure provider fully satisfies framework interfaces
var (
	_ provider.ProviderWithMetadata = &mssqlProvider{}
)

const (
	VersionDev  = "dev"
	VersionTest = "test"
)

const regularIdentifiersDoc = "Must follow [Regular Identifiers rules](https://docs.microsoft.com/en-us/sql/relational-databases/databases/database-identifiers#rules-for-regular-identifiers)"

type sqlAuth struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

type azureAuth struct {
	ClientId     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	TenantId     types.String `tfsdk:"tenant_id"`
}

type providerData struct {
	Hostname  types.String `tfsdk:"hostname"`
	Port      types.Int64  `tfsdk:"port"`
	SqlAuth   types.Object `tfsdk:"sql_auth"`
	AzureAuth types.Object `tfsdk:"azure_auth"`
}

func (pd providerData) asConnectionDetails(ctx context.Context) (sql.ConnectionDetails, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	var addComputedError = func(summary string) {
		diags.AddError(summary, "SQL connection details must be known during plan execution")
	}

	if pd.Hostname.Unknown {
		addComputedError("Hostname cannot be a computed value")
	}

	connDetails := sql.ConnectionDetails{
		Host: os.Getenv("MSSQL_HOSTNAME"),
	}

	if !pd.Hostname.Null {
		connDetails.Host = pd.Hostname.Value
	}

	if !pd.Port.Null {
		connDetails.Host = fmt.Sprintf("%s:%d", connDetails.Host, pd.Port.Value)
	} else if envPort := os.Getenv("MSSQL_PORT"); envPort != "" {
		connDetails.Host = fmt.Sprintf("%s:%s", connDetails.Host, envPort)
	}

	if !pd.SqlAuth.Null {
		var auth sqlAuth
		diags.Append(pd.SqlAuth.As(ctx, &auth, types.ObjectAsOptions{})...)

		if auth.Username.Unknown {
			addComputedError("SQL username cannot be a computed value")
		}

		if auth.Password.Unknown {
			addComputedError("SQL password cannot be a computed value")
		}

		connDetails.Auth = sql.ConnectionAuthSql{Username: auth.Username.Value, Password: auth.Password.Value}
	}

	if !pd.AzureAuth.Null {
		var auth azureAuth
		diags.Append(pd.AzureAuth.As(ctx, &auth, types.ObjectAsOptions{})...)

		if auth.ClientId.Unknown {
			addComputedError("Azure AD Service Principal client_id cannot be a computed value")
		}

		if auth.ClientSecret.Unknown {
			addComputedError("Azure AD Service Principal client_secret cannot be a computed value")
		}

		if auth.TenantId.Unknown {
			addComputedError("Azure AD Service Principal tenant_id cannot be a computed value")
		}

		connAuth := sql.ConnectionAuthAzure{
			ClientId:     auth.ClientId.Value,
			ClientSecret: auth.ClientSecret.Value,
			TenantId:     auth.TenantId.Value,
		}

		if connAuth.ClientId == "" {
			connAuth.ClientId = os.Getenv("ARM_CLIENT_ID")
		}

		if connAuth.ClientSecret == "" {
			connAuth.ClientSecret = os.Getenv("ARM_CLIENT_SECRET")
		}

		if connAuth.TenantId == "" {
			connAuth.TenantId = os.Getenv("ARM_TENANT_ID")
		}

		connDetails.Auth = connAuth
	}

	return connDetails, diags
}

type mssqlProvider struct {
	Version string
	Db      sql.Connection
}

func (p *mssqlProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "mssql"
	resp.Version = p.Version
}

func (p *mssqlProvider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	if p.Version != VersionTest {
		var data providerData
		diags := request.Config.Get(ctx, &data)

		if response.Diagnostics.Append(diags...); response.Diagnostics.HasError() {
			return
		}

		connDetails, diags := data.asConnectionDetails(ctx)

		if response.Diagnostics.Append(diags...); response.Diagnostics.HasError() {
			return
		}

		p.Db, diags = connDetails.Open(ctx)
		response.Diagnostics.Append(diags...)
	}

	response.DataSourceData = p.Db
	response.ResourceData = p.Db
}

func (p mssqlProvider) Resources(ctx context.Context) []func() resource.Resource {
	ctors := []func() resource.Resource{
		p.NewAzureADServicePrincipalResource(),
		p.NewAzureADUserResource(),
		p.NewDatabaseResource(),
		p.NewDatabaseRoleResource(),
		p.NewDatabaseRoleMemberResource(),
		p.NewSqlLoginResource(),
		p.NewSqlUserResource(),
	}

	for _, ctor := range ctors {
		res := ctor()

		if _, ok := res.(resource.ResourceWithConfigure); !ok {
			panic(fmt.Sprintf("Resource %T does not implement ResourceWithConfigure", res))
		}
	}

	return ctors
}

func (p mssqlProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	ctors := []func() datasource.DataSource{
		p.NewAzureADServicePrincipalDataSource(),
		p.NewAzureADUserDataSource(),
		p.NewDatabaseDataSource(),
		p.NewDatabaseListDataSource(),
		p.NewDatabaseRoleDataSource(),
		p.NewDatabaseRoleListDataSource(),
		p.NewSqlLoginDataSource(),
		p.NewSqlLoginListDataSource(),
		p.NewSqlUserDataSource(),
		p.NewSqlUserListDataSource(),
	}

	for _, ctor := range ctors {
		data := ctor()

		if _, ok := data.(datasource.DataSourceWithConfigure); !ok {
			panic(fmt.Sprintf("Data source %T does not implmenet DataSourceWithConfigure", data))
		}
	}

	return ctors
}

func (p mssqlProvider) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if p.Version == VersionTest {
		return tfsdk.Schema{}, nil
	}

	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"hostname": {
				Description: "FQDN or IP address of the SQL endpoint. Can be also set using `MSSQL_HOSTNAME` environment variable.",
				Type:        types.StringType,
				Optional:    true,
			},
			"port": {
				MarkdownDescription: "TCP port of SQL endpoint. Defaults to `1433`. Can be also set using `MSSQL_PORT` environment variable.",
				Type:                types.Int64Type,
				Optional:            true,
			},
			"sql_auth": {
				Description: "When provided, SQL authentication will be used when connecting.",
				Optional:    true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"username": {
						Description: "User name for SQL authentication.",
						Type:        types.StringType,
						Required:    true,
					},
					"password": {
						Description: "Password for SQL authentication.",
						Type:        types.StringType,
						Required:    true,
						Sensitive:   true,
					},
				}),
			},
			"azure_auth": {
				Description: "When provided, Azure AD authentication will be used when connecting.",
				Optional:    true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"client_id": {
						Description: "Service Principal client (application) ID. When omitted, default, chained set of credentials will be used.",
						Type:        types.StringType,
						Optional:    true,
					},
					"client_secret": {
						Description: "Service Principal secret. When omitted, default, chained set of credentials will be used.",
						Type:        types.StringType,
						Sensitive:   true,
						Optional:    true,
					},
					"tenant_id": {
						Description: "Azure AD tenant ID. Required only if Azure SQL Server's tenant is different than Service Principal's.",
						Type:        types.StringType,
						Optional:    true,
					},
				}),
			},
		},
	}, nil
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &mssqlProvider{
			Version: version,
		}
	}
}
