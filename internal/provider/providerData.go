package provider

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"os"
)

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
	SqlAuth   *sqlAuth     `tfsdk:"sql_auth"`
	AzureAuth *azureAuth   `tfsdk:"azure_auth"`
}

func (pd providerData) asConnectionDetails(context.Context) (sql.ConnectionDetails, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	var addComputedError = func(summary string) {
		diags.AddError(summary, "SQL connection details must be known during plan execution")
	}

	if pd.Hostname.IsUnknown() {
		addComputedError("Hostname cannot be a computed value")
	}

	connDetails := sql.ConnectionDetails{
		Host: os.Getenv("MSSQL_HOSTNAME"),
	}

	if !pd.Hostname.IsNull() {
		connDetails.Host = pd.Hostname.ValueString()
	}

	if !pd.Port.IsNull() {
		connDetails.Host = fmt.Sprintf("%s:%d", connDetails.Host, pd.Port.ValueInt64())
	} else if envPort := os.Getenv("MSSQL_PORT"); envPort != "" {
		connDetails.Host = fmt.Sprintf("%s:%s", connDetails.Host, envPort)
	}

	if pd.SqlAuth != nil {
		if pd.SqlAuth.Username.IsUnknown() {
			addComputedError("SQL username cannot be a computed value")
		}

		if pd.SqlAuth.Password.IsUnknown() {
			addComputedError("SQL password cannot be a computed value")
		}

		connDetails.Auth = sql.ConnectionAuthSql{Username: pd.SqlAuth.Username.ValueString(), Password: pd.SqlAuth.Password.ValueString()}
	}

	if pd.AzureAuth != nil {
		if pd.AzureAuth.ClientId.IsUnknown() {
			addComputedError("Azure AD Service Principal client_id cannot be a computed value")
		}

		if pd.AzureAuth.ClientSecret.IsUnknown() {
			addComputedError("Azure AD Service Principal client_secret cannot be a computed value")
		}

		if pd.AzureAuth.TenantId.IsUnknown() {
			addComputedError("Azure AD Service Principal tenant_id cannot be a computed value")
		}

		connAuth := sql.ConnectionAuthAzure{
			ClientId:     pd.AzureAuth.ClientId.ValueString(),
			ClientSecret: pd.AzureAuth.ClientSecret.ValueString(),
			TenantId:     pd.AzureAuth.TenantId.ValueString(),
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
