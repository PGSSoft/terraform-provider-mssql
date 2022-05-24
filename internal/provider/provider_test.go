package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tkielar/terraform-provider-mssql/internal/sql"
	"testing"
)

func TestProviderDataAsConnectionDetails(t *testing.T) {
	ctx := context.Background()

	computedErrorCases := map[string]struct {
		pd         providerData
		errSummary string
	}{
		"Hostname": {
			pd:         providerData{Hostname: types.String{Unknown: true}},
			errSummary: "Hostname cannot be a computed value",
		},

		"SQL Username": {
			pd: providerData{SqlAuth: types.Object{
				AttrTypes: map[string]attr.Type{
					"username": types.StringType,
					"password": types.StringType,
				},
				Attrs: map[string]attr.Value{
					"username": types.String{Unknown: true},
					"password": types.String{Null: true},
				},
			}},
			errSummary: "SQL username cannot be a computed value",
		},

		"SQL Password": {
			pd: providerData{SqlAuth: types.Object{
				AttrTypes: map[string]attr.Type{
					"username": types.StringType,
					"password": types.StringType,
				},
				Attrs: map[string]attr.Value{
					"username": types.String{Null: true},
					"password": types.String{Unknown: true},
				},
			}},
			errSummary: "SQL password cannot be a computed value",
		},

		"Azure auth ClientId": {
			pd: providerData{AzureAuth: types.Object{
				AttrTypes: map[string]attr.Type{
					"client_id":     types.StringType,
					"client_secret": types.StringType,
					"tenant_id":     types.StringType,
				},
				Attrs: map[string]attr.Value{
					"client_id":     types.String{Unknown: true},
					"client_secret": types.String{Null: true},
					"tenant_id":     types.String{Null: true},
				},
			}},
			errSummary: "Azure AD Service Principal client_id cannot be a computed value",
		},

		"Azure auth ClientSecret": {
			pd: providerData{AzureAuth: types.Object{
				AttrTypes: map[string]attr.Type{
					"client_id":     types.StringType,
					"client_secret": types.StringType,
					"tenant_id":     types.StringType,
				},
				Attrs: map[string]attr.Value{
					"client_id":     types.String{Null: true},
					"client_secret": types.String{Unknown: true},
					"tenant_id":     types.String{Null: true},
				},
			}},
			errSummary: "Azure AD Service Principal client_secret cannot be a computed value",
		},

		"Azure auth TenantId": {
			pd: providerData{AzureAuth: types.Object{
				AttrTypes: map[string]attr.Type{
					"client_id":     types.StringType,
					"client_secret": types.StringType,
					"tenant_id":     types.StringType,
				},
				Attrs: map[string]attr.Value{
					"client_id":     types.String{Null: true},
					"client_secret": types.String{Null: true},
					"tenant_id":     types.String{Unknown: true},
				},
			}},
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
	}{
		"With port": {
			pd: providerData{
				Hostname: types.String{Value: "test_hostname"},
				Port:     types.Int64{Value: 123},
			},
			host: "test_hostname:123",
		},
		"Without port": {
			pd: providerData{
				Hostname: types.String{Value: "test_hostname2"},
				Port:     types.Int64{Null: true},
			},
			host: "test_hostname2",
		},
	}

	for name, tc := range hostCases {
		name, tc := name, tc
		t.Run(fmt.Sprintf("Host %s", name), func(t *testing.T) {
			cd, _ := tc.pd.asConnectionDetails(ctx)
			assert.Equal(t, tc.host, cd.Host)
		})
	}

	t.Run("SQL auth", func(t *testing.T) {
		pd := providerData{
			SqlAuth: types.Object{
				AttrTypes: map[string]attr.Type{
					"username": types.StringType,
					"password": types.StringType,
				},
				Attrs: map[string]attr.Value{
					"username": types.String{Value: "test_username"},
					"password": types.String{Value: "test_password"},
				},
			},
			AzureAuth: types.Object{Null: true},
		}

		cd, _ := pd.asConnectionDetails(ctx)

		sqlAuth, ok := cd.Auth.(sql.ConnectionAuthSql)
		require.True(t, ok, "Connection auth not set to SQL")
		assert.Equal(t, "test_username", sqlAuth.Username, "username")
		assert.Equal(t, "test_password", sqlAuth.Password, "password")
	})

	t.Run("Azure auth", func(t *testing.T) {
		pd := providerData{
			SqlAuth: types.Object{Null: true},
			AzureAuth: types.Object{
				AttrTypes: map[string]attr.Type{
					"client_id":     types.StringType,
					"client_secret": types.StringType,
					"tenant_id":     types.StringType,
				},
				Attrs: map[string]attr.Value{
					"client_id":     types.String{Value: "test_client_id"},
					"client_secret": types.String{Value: "test_client_secret"},
					"tenant_id":     types.String{Value: "test_tenant_id"},
				},
			},
		}

		cd, _ := pd.asConnectionDetails(ctx)

		azureAuth, ok := cd.Auth.(sql.ConnectionAuthAzure)
		require.True(t, ok, "Connection auth not set to Azure")
		assert.Equal(t, "test_client_id", azureAuth.ClientId, "client_id")
		assert.Equal(t, "test_client_secret", azureAuth.ClientSecret, "client_secret")
		assert.Equal(t, "test_tenant_id", azureAuth.TenantId, "tenant_id")
	})
}
