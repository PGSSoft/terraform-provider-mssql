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
