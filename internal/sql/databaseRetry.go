package sql

import (
	"context"
	"database/sql"
	"slices"
	"time"

	mssql "github.com/microsoft/go-mssqldb"
	"github.com/sethvargo/go-retry"
)

// These are all error codes by the (mssql|Azure SQL) Server that can be retried
var retryableErrors = []int32{539, 617, 952, 956, 988, 1205, 1807, 3055, 3762, 5034, 5059, 5061, 5065, 5295, 8628, 8645, 10922, 10930, 12111, 14258, 16528, 19510, 20689, 22380, 22498, 22754, 22758, 22759, 22760, 25003, 25738, 25740, 27118, 27230, 30024, 30026, 30085, 33115, 33116, 33136, 40602, 40613, 40642, 40648, 40671, 40675, 40806, 40807, 40825, 40938, 41828, 41838, 42104, 42106, 45156, 45157, 45161, 45168, 45169, 45182, 45509, 45541, 45727, 47132, 49510, 49518, 49918}

// This backoff func is configured to wait a maximum of 100 seconds.
// A common retryable errors is a paused db, that needs some time to
// wake up to execute the query.
// The auto-resume of paused db is "in the order of one minute".
// See:
// (https://learn.microsoft.com/en-us/azure/azure-sql/database/serverless-tier-overview?view=azuresql&tabs=general-purpose#latency)
func ExpBackoff() retry.Backoff {
	backoff := retry.NewFibonacci(1 * time.Second)
	backoff = retry.WithMaxDuration(100*time.Second, backoff)
	return backoff
}

// The err that is returned by executing a query is checked against the list of
// all retryable errors, and if so the err is marked as retryable.
func CheckIfRetryable(err error) (checkedErr error) {
	if mssqldb, ok := err.(mssql.Error); ok {
		if slices.Contains(retryableErrors, mssqldb.Number) {
			return retry.RetryableError(err)
		}
	}
	return err
}

// This is "conn.ExecContext" wrapped with a retry mechanism for retrying transient
// errors. Should behave in all other regards the same as the original.
func ExecContextWithRetry(ctx context.Context, conn *sql.DB, query string, args ...any) (res sql.Result, err error) {
	backoff := ExpBackoff()

	if err := retry.Do(ctx, backoff,
		func(ctx context.Context) error {
			if res, err = conn.ExecContext(ctx, query, args...); err != nil {
				return CheckIfRetryable(err)
			}
			return nil
		}); err != nil {
		return res, err
	}
	return res, nil
}

// This is "conn.QueryRowContext" wrapped with a retry mechanism for retrying transient
// errors. Should behave in all other regards the same as the original.
func QueryRowContextWithRetry(ctx context.Context, conn *sql.DB, query string, args ...any) *sql.Row {
	backoff := ExpBackoff()

	var row *sql.Row
	retry.Do(ctx, backoff,
		func(ctx context.Context) (err error) {
			row = conn.QueryRowContext(ctx, query, args...)
			err = row.Err()
			return CheckIfRetryable(err)
		})
	return row
}
