package sql

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestConfigure(t *testing.T) {
	var u url.URL
	auth := ConnectionAuthSql{Username: "test_username", Password: "test_password"}

	auth.configure(context.Background(), &u)

	assert.Equal(t, fmt.Sprintf("%s:%s", auth.Username, auth.Password), u.User.String())
}
