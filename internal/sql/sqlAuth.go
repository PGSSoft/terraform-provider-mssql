package sql

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"net/url"
)

type ConnectionAuthSql struct {
	Username string
	Password string
}

func (auth ConnectionAuthSql) configure(_ context.Context, u *url.URL) diag.Diagnostics {
	u.User = url.UserPassword(auth.Username, auth.Password)
	return nil
}

func (ConnectionAuthSql) getDriverName() string {
	return "sqlserver"
}
