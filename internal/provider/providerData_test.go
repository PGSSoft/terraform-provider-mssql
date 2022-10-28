package provider

import (
	"context"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestProviderDataAsConnectionDetails(t *testing.T) {
	ctx := context.Background()

	computedErrorCases := map[string]struct {
		pd         providerData
		errSummary string
	}{
		"Hostname": {
			pd: providerData{
				Hostname: types.StringUnknown(),
			},
			errSummary: "Hostname cannot be a computed value",
		},

		"SQL Username": {
			pd: providerData{
				SqlAuth: &sqlAuth{
					Username: types.StringUnknown(),
					Password: types.StringNull(),
				},
			},
			errSummary: "SQL username cannot be a computed value",
		},

		"SQL Password": {
			pd: providerData{
				SqlAuth: &sqlAuth{
					Username: types.StringNull(),
					Password: types.StringUnknown(),
				},
			},
			errSummary: "SQL password cannot be a computed value",
		},

		"Azure auth ClientId": {
			pd: providerData{
				AzureAuth: &azureAuth{
					ClientId:     types.StringUnknown(),
					ClientSecret: types.StringNull(),
					TenantId:     types.StringNull(),
				},
			},
			errSummary: "Azure AD Service Principal client_id cannot be a computed value",
		},

		"Azure auth ClientSecret": {
			pd: providerData{
				AzureAuth: &azureAuth{
					ClientId:     types.StringNull(),
					ClientSecret: types.StringUnknown(),
					TenantId:     types.StringNull(),
				},
			},
			errSummary: "Azure AD Service Principal client_secret cannot be a computed value",
		},

		"Azure auth TenantId": {
			pd: providerData{
				AzureAuth: &azureAuth{
					ClientId:     types.StringNull(),
					ClientSecret: types.StringNull(),
					TenantId:     types.StringUnknown(),
				},
			},
			errSummary: "Azure AD Service Principal tenant_id cannot be a computed value",
		},
	}

	for name, tc := range computedErrorCases {
		name, tc := name, tc
		t.Run(fmt.Sprintf("Unknown %s", name), func(t *testing.T) {
			_, diags := tc.pd.asConnectionDetails(ctx)

			for _, d := range diags {
				if d.Severity() == diag.SeverityError && d.Summary() == tc.errSummary && d.Detail() == "SQL connection details must be known during plan execution" {
					return
				}
			}

			t.Errorf("Could not find expected diags error '%s'", tc.errSummary)
		})
	}

	hostCases := map[string]struct {
		pd   providerData
		host string
		env  map[string]string
	}{
		"With port": {
			pd: providerData{
				Hostname: types.StringValue("test_hostname"),
				Port:     types.Int64Value(123),
			},
			host: "test_hostname:123",
		},
		"Without port": {
			pd: providerData{
				Hostname: types.StringValue("test_hostname2"),
				Port:     types.Int64Null(),
			},
			host: "test_hostname2",
		},
		"Env variable hostname": {
			pd: providerData{
				Hostname: types.StringNull(),
				Port:     types.Int64Null(),
			},
			env: map[string]string{
				"MSSQL_HOSTNAME": "env_test_hostname",
			},
			host: "env_test_hostname",
		},
		"Env variable hostname and port": {
			pd: providerData{
				Hostname: types.StringNull(),
				Port:     types.Int64Null(),
			},
			env: map[string]string{
				"MSSQL_HOSTNAME": "env_test_hostname2",
				"MSSQL_PORT":     "321",
			},
			host: "env_test_hostname2:321",
		},
		"Env variables and attributes": {
			pd: providerData{
				Hostname: types.StringValue("test_hostname"),
				Port:     types.Int64Value(123),
			},
			env: map[string]string{
				"MSSQL_HOSTNAME": "env_test_hostname2",
				"MSSQL_PORT":     "321",
			},
			host: "test_hostname:123",
		},
	}

	for name, tc := range hostCases {
		name, tc := name, tc
		t.Run(fmt.Sprintf("Host %s", name), func(t *testing.T) {
			for n, v := range tc.env {
				os.Setenv(n, v)
			}

			defer func() {
				for n := range tc.env {
					os.Unsetenv(n)
				}
			}()

			cd, _ := tc.pd.asConnectionDetails(ctx)
			assert.Equal(t, tc.host, cd.Host)
		})
	}

	t.Run("SQL auth", func(t *testing.T) {
		pd := providerData{
			SqlAuth: &sqlAuth{
				Username: types.StringValue("test_username"),
				Password: types.StringValue("test_password"),
			},
		}

		cd, _ := pd.asConnectionDetails(ctx)

		sqlAuth, ok := cd.Auth.(sql.ConnectionAuthSql)
		require.True(t, ok, "Connection auth not set to SQL")
		assert.Equal(t, "test_username", sqlAuth.Username, "username")
		assert.Equal(t, "test_password", sqlAuth.Password, "password")
	})

	t.Run("Azure auth", func(t *testing.T) {
		pd := providerData{
			AzureAuth: &azureAuth{
				ClientId:     types.StringValue("test_client_id"),
				ClientSecret: types.StringValue("test_client_secret"),
				TenantId:     types.StringValue("test_tenant_id"),
			},
		}

		cd, _ := pd.asConnectionDetails(ctx)

		azureAuth, ok := cd.Auth.(sql.ConnectionAuthAzure)
		require.True(t, ok, "Connection auth not set to Azure")
		assert.Equal(t, "test_client_id", azureAuth.ClientId, "client_id")
		assert.Equal(t, "test_client_secret", azureAuth.ClientSecret, "client_secret")
		assert.Equal(t, "test_tenant_id", azureAuth.TenantId, "tenant_id")
	})

	t.Run("Azure auth env variables", func(t *testing.T) {
		pd := providerData{
			AzureAuth: &azureAuth{},
		}
		os.Setenv("ARM_CLIENT_ID", "env_test_client_id")
		os.Setenv("ARM_CLIENT_SECRET", "env_test_client_secret")
		os.Setenv("ARM_TENANT_ID", "env_test_tenant_id")

		cd, _ := pd.asConnectionDetails(ctx)

		azureAuth, ok := cd.Auth.(sql.ConnectionAuthAzure)
		require.True(t, ok, "Connection auth not set to Azure")
		assert.Equal(t, "env_test_client_id", azureAuth.ClientId, "client_id")
		assert.Equal(t, "env_test_client_secret", azureAuth.ClientSecret, "client_secret")
		assert.Equal(t, "env_test_tenant_id", azureAuth.TenantId, "tenant_id")
	})

	t.Run("Azure auth and env variables", func(t *testing.T) {
		pd := providerData{
			AzureAuth: &azureAuth{
				ClientId:     types.StringValue("test_client_id"),
				ClientSecret: types.StringValue("test_client_secret"),
				TenantId:     types.StringValue("test_tenant_id"),
			},
		}
		os.Setenv("ARM_CLIENT_ID", "env_test_client_id")
		os.Setenv("ARM_CLIENT_SECRET", "env_test_client_secret")
		os.Setenv("ARM_TENANT_ID", "env_test_tenant_id")

		cd, _ := pd.asConnectionDetails(ctx)

		azureAuth, ok := cd.Auth.(sql.ConnectionAuthAzure)
		require.True(t, ok, "Connection auth not set to Azure")
		assert.Equal(t, "test_client_id", azureAuth.ClientId, "client_id")
		assert.Equal(t, "test_client_secret", azureAuth.ClientSecret, "client_secret")
		assert.Equal(t, "test_tenant_id", azureAuth.TenantId, "tenant_id")
	})
}
