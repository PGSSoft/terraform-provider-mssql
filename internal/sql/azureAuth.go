package sql

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/microsoft/go-mssqldb/azuread"
	"net/url"
)

type ConnectionAuthAzure struct {
	ClientId     string
	ClientSecret string
	TenantId     string
}

func (auth ConnectionAuthAzure) configure(_ context.Context, u *url.URL) diag.Diagnostics {
	q := u.Query()

	if auth.ClientId == "" || auth.ClientSecret == "" {
		q.Set("fedauth", "ActiveDirectoryDefault")
		u.RawQuery = q.Encode()
		return nil
	}

	q.Set("fedauth", "ActiveDirectoryServicePrincipal")
	q.Set("password", auth.ClientSecret)

	if auth.TenantId != "" {
		q.Set("user id", fmt.Sprintf("%s@%s", auth.ClientId, auth.TenantId))
	} else {
		q.Set("user id", auth.ClientId)
	}

	u.RawQuery = q.Encode()
	return nil
}

func (ConnectionAuthAzure) getDriverName() string {
	return azuread.DriverName
}
