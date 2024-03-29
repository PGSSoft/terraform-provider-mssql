package sql

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net/url"
	"testing"
)

func TestConnectionTestSuite(t *testing.T) {
	s := &ConnectionTestSuite{}
	var (
		_ suite.SetupTestSuite = s
		_ suite.AfterTest      = s
	)

	suite.Run(t, s)
}

type ConnectionTestSuite struct {
	suite.Suite
	auth              mockedAuth
	authConfigureCall *mock.Call
	connDetails       ConnectionDetails
}

func (s *ConnectionTestSuite) SetupTest() {
	s.auth = mockedAuth{}
	s.authConfigureCall = s.auth.On("configure", mock.IsType(context.Background()), mock.IsType(&url.URL{})).Return(diag.Diagnostics{})
	s.connDetails = ConnectionDetails{Auth: s.auth}
}

func (s *ConnectionTestSuite) AfterTest(string, string) {
	s.auth.AssertExpectations(s.T())
}

func (s *ConnectionTestSuite) TestGetConnectionStringSetsParameters() {
	s.connDetails.Host = "hostname_test"

	cs, _ := s.getConnectionString()

	s.Equal("sqlserver", cs.Scheme, "scheme")
	s.Equal("hostname_test", cs.Host, "host")
}

func (s *ConnectionTestSuite) TestGetConnectionStringWhenDatabaseNotProvided() {
	cs, _ := s.getConnectionString()

	s.False(cs.Query().Has("database"), "database")
}

func (s *ConnectionTestSuite) TestGetConnectionStringReturnsParamsSetByAuthProvider() {
	testDiag := diag.NewErrorDiagnostic("Test error", "Test error details")
	s.authConfigureCall.
		Run(func(args mock.Arguments) {
			u := args.Get(1).(*url.URL)
			u.User = url.UserPassword("test_username", "test_password")
			query := u.Query()
			query.Set("test_param", "test_value")
			u.RawQuery = query.Encode()
		}).
		Return(diag.Diagnostics{testDiag})
	s.connDetails.Auth = s.auth

	cs, diags := s.getConnectionString()

	s.Equal("test_username:test_password", cs.User.String(), "user")
	s.Equal("test_value", cs.Query().Get("test_param"), "extra param")
	s.True(diags.Contains(testDiag), "diagnostics")
}

func TestGrantPermission(t *testing.T) {
	cases := map[string]struct {
		stat string
		perm ServerPermission
	}{
		"without_grant": {
			stat: "GRANT TEST PERMISSION TO [test_user]",
			perm: ServerPermission{Name: "TEST PERMISSION"},
		},
		"with_grant": {
			stat: "GRANT TEST PERMISSION GRANT TO [test_user] WITH GRANT OPTION",
			perm: ServerPermission{Name: "TEST PERMISSION GRANT", WithGrantOption: true},
		},
	}

	for name, tc := range cases {
		testCase := tc
		t.Run(name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err, "SQL mock")
			expectExactQuery(mock, "SELECT [name] FROM sys.server_principals WHERE [principal_id]=@p1").
				WithArgs(12).
				WillReturnRows(newRows("name").AddRow("test_user"))
			expectExactExec(mock, testCase.stat).WillReturnResult(sqlmock.NewResult(0, 1))
			conn := connection{conn: db}
			diags := diag.Diagnostics{}

			conn.GrantPermission(utils.WithDiagnostics(context.Background(), &diags), 12, testCase.perm)

			assert.NoError(t, mock.ExpectationsWereMet())
			assert.Len(t, diags, 0)
		})
	}
}

func TestGetPermissions(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "SQL mock")
	expectExactQuery(mock, "SELECT [permission_name], [state] FROM sys.server_permissions WHERE [class]=100 AND [grantee_principal_id]=@p1").
		WithArgs(24).
		WillReturnRows(newRows("permission_name", "state").AddRow("TEST PERM", "G").AddRow("TEST PERM2", "W"))
	conn := connection{conn: db}
	diags := diag.Diagnostics{}

	perms := conn.GetPermissions(utils.WithDiagnostics(context.Background(), &diags), 24)

	assert.Len(t, diags, 0)
	require.NoError(t, mock.ExpectationsWereMet())
	assert.Equal(t, ServerPermissions{
		"TEST PERM":  {Name: "TEST PERM"},
		"TEST PERM2": {Name: "TEST PERM2", WithGrantOption: true},
	}, perms)
}

func TestRevokePermission(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "SQL mock")
	expectExactQuery(mock, "SELECT [name] FROM sys.server_principals WHERE [principal_id]=@p1").
		WithArgs(5).
		WillReturnRows(newRows("name").AddRow("test_user"))
	expectExactExec(mock, "REVOKE TEST PERM RVK FROM [test_user] CASCADE").WillReturnResult(sqlmock.NewResult(0, 1))
	conn := connection{conn: db}
	diags := diag.Diagnostics{}

	conn.RevokePermission(utils.WithDiagnostics(context.Background(), &diags), 5, "TEST PERM RVK")

	assert.Len(t, diags, 0)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestIsAzure(t *testing.T) {
	cases := map[string]bool{
		"Enterprise Edition":  false,
		"Enterprise (64-bit)": false,
		"SQL Azure":           true,
		"SQL Azure (64-bit)":  true,
	}

	for edition, expected := range cases {
		t.Run(edition, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			defer db.Close()
			require.NoError(t, err, "creating SQL mock")
			expectExactQuery(mock, "SELECT SERVERPROPERTY('edition')").WillReturnRows(newRows("prop").AddRow(edition))
			c := connection{conn: db}
			diags := diag.Diagnostics{}
			ctx := utils.WithDiagnostics(context.Background(), &diags)

			result := c.IsAzure(ctx)

			if diags.HasError() {
				for _, d := range diags {
					if d.Severity() == diag.SeverityError {
						t.Error(errors.New(d.Detail()))
					}
				}
			}
			assert.Equal(t, expected, result)
			assert.NoError(t, mock.ExpectationsWereMet(), "mock expectations")
		})
	}
}

func TestExec(t *testing.T) {
	db, mock, err := sqlmock.New()
	defer db.Close()
	require.NoError(t, err, "creating SQL mock")
	c := connection{conn: db}
	diags := diag.Diagnostics{}
	ctx := utils.WithDiagnostics(context.Background(), &diags)

	mock.ExpectExec("INVALID QUERY").WithArgs(1, "foo").WillReturnError(errors.New("test error"))

	c.exec(ctx, "INVALID QUERY", 1, "foo")

	assert.NoError(t, mock.ExpectationsWereMet(), "SQL query")

	for _, d := range diags {
		if d.Severity() == diag.SeverityError && d.Detail() == "test error" {
			return
		}
	}

	t.Error("Error returned by SQL provider not added to Diagnostics")
}

var _ ConnectionAuth = mockedAuth{}

type mockedAuth struct {
	mock.Mock
}

func (a mockedAuth) configure(ctx context.Context, u *url.URL) diag.Diagnostics {
	args := a.Called(ctx, u)
	return args.Get(0).(diag.Diagnostics)
}

func (a mockedAuth) getDriverName() string {
	return a.Called().String(0)
}

func (s *ConnectionTestSuite) getConnectionString() (*url.URL, diag.Diagnostics) {
	cs, diags := s.connDetails.getConnectionString(context.Background())

	connString, err := url.Parse(cs)
	s.Require().NoError(err, "Failed to parse URL")

	return connString, diags
}
