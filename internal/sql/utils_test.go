package sql

import (
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"regexp"
	"strings"
)

type SqlMock interface {
	ExpectQuery(query string) *sqlmock.ExpectedQuery
	ExpectExec(query string) *sqlmock.ExpectedExec
}

func expectExactQuery(mock SqlMock, queryFmt string, fmtArgs ...any) *sqlmock.ExpectedQuery {
	return mock.ExpectQuery(formatExactSql(queryFmt, fmtArgs))
}

func expectExactExec(mock SqlMock, execFmt string, fmtArgs ...any) *sqlmock.ExpectedExec {
	return mock.ExpectExec(formatExactSql(execFmt, fmtArgs))
}

func newRows(cols ...string) *sqlmock.Rows {
	return sqlmock.NewRows(cols)
}

func formatExactSql(fmtSql string, args []any) string {
	return fmt.Sprintf("^%s$", regexp.QuoteMeta(strings.TrimSpace(fmt.Sprintf(fmtSql, args...))))
}
