package provider

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// To ensure provider fully satisfies framework interfaces
var (
	_ tfsdk.Provider  = &provider{}
	_ resourceFactory = &provider{}
)

const (
	VersionDev  = "dev"
	VersionTest = "test"
)

type Resource struct {
	Version string
	Db      sql.Connection
}

type resourceFactory interface {
	NewResource() Resource
}

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

	connDetails := sql.ConnectionDetails{}

	if pd.Port.Null {
		connDetails.Host = pd.Hostname.Value
	} else {
		connDetails.Host = fmt.Sprintf("%s:%d", pd.Hostname.Value, pd.Port.Value)
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

		connDetails.Auth = sql.ConnectionAuthAzure{
			ClientId:     auth.ClientId.Value,
			ClientSecret: auth.ClientSecret.Value,
			TenantId:     auth.TenantId.Value,
		}
	}

	return connDetails, diags
}

type provider struct {
	Version string
	Db      sql.Connection
}

func (p *provider) Configure(ctx context.Context, request tfsdk.ConfigureProviderRequest, response *tfsdk.ConfigureProviderResponse) {
	if p.Version == VersionTest {
		return
	}

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

func (p provider) GetResources(context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
		"mssql_database": DatabaseResourceType{},
	}, nil
}

func (p provider) GetDataSources(context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{
		"mssql_database":  DatabaseDataSourceType{},
		"mssql_databases": DatabaseListDataSourceType{},
	}, nil
}

func (p provider) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	if p.Version == VersionTest {
		return tfsdk.Schema{}, nil
	}

	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"hostname": {
				Description: "FQDN or IP address of the SQL endpoint.",
				Type:        types.StringType,
				Required:    true,
			},
			"port": {
				MarkdownDescription: "TCP port of SQL endpoint. Defaults to `1433`.",
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

func New(version string) func() tfsdk.Provider {
	return func() tfsdk.Provider {
		return &provider{
			Version: version,
		}
	}
}

func (p *provider) NewResource() Resource {
	return Resource{
		Version: p.Version,
		Db:      p.Db,
	}
}

func convertProviderType(in tfsdk.Provider) (resourceFactory, diag.Diagnostics) {
	var diags diag.Diagnostics

	p, ok := in.(resourceFactory)

	if !ok {
		diags.AddError("Unexpected provider instance type", fmt.Sprintf("Unexpected provider type (%T). This is bug in the provider code and should be reported", p))
	}

	if p == nil {
		diags.AddError("Unexpected empty provider instance", "Unexpected empty provider instance. This is bug in the provider code and should be reported")
	}

	return p, diags
}

func newResource[T any](_ context.Context, in tfsdk.Provider, ctor func(base Resource) T) (T, diag.Diagnostics) {
	p, diags := convertProviderType(in)
	if diags.HasError() {
		var result T
		return result, diags
	}

	return ctor(p.NewResource()), diags
}
