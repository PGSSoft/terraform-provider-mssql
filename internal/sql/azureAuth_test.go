package sql

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestConfigureDefault(t *testing.T) {
	u := url.URL{}
	auth := ConnectionAuthAzure{}

	auth.configure(context.Background(), &u)

	assert.Equal(t, "ActiveDirectoryDefault", u.Query().Get("fedauth"))
}

func TestConfigureServicePrincipal(t *testing.T) {
	u := url.URL{}
	auth := ConnectionAuthAzure{
		ClientId:     "test_client_id",
		ClientSecret: "test_client_secret",
	}

	auth.configure(context.Background(), &u)

	assert.Equal(t, "ActiveDirectoryServicePrincipal", u.Query().Get("fedauth"))
	assert.Equal(t, auth.ClientId, u.Query().Get("user id"), "user id")
	assert.Equal(t, auth.ClientSecret, u.Query().Get("password"), "password")
}

func TestConfigureServicePrincipalWithTenant(t *testing.T) {
	u := url.URL{}
	auth := ConnectionAuthAzure{
		ClientId:     "test_client_id",
		ClientSecret: "test_client_secret",
		TenantId:     "test_tenant_id",
	}

	auth.configure(context.Background(), &u)

	assert.Equal(t, fmt.Sprintf("%s@%s", auth.ClientId, auth.TenantId), u.Query().Get("user id"))
}
