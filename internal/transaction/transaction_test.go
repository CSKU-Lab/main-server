package transaction_test

import (
	"errors"
	"testing"
	"time"

	"github.com/SornchaiTheDev/cs-lab-backend/internal/transaction"
	"github.com/stretchr/testify/assert"
)

func Add(a *int, amount int) error {
	*a += amount
	return nil
}

func Remove(a *int, amount int) error {
	*a -= amount
	return nil
}

func TestOneStepSuccess(t *testing.T) {
	a := 0

	tr := transaction.New(&transaction.Option{})

	err := tr.Execute(
		tr.Step().CommitWith(func() error {
			return Add(&a, 1)
		}).RollbackWith(func() error {
			return Remove(&a, 1)
		}),
	)

	assert.Nil(t, err, "There should be no error")
	assert.Equal(t, 1, a, "a should be 1 after commit")
}

func TestOneStepFailed(t *testing.T) {
	a := 0

	tr := transaction.New(&transaction.Option{})

	err := tr.Execute(
		tr.Step().CommitWith(func() error {
			return errors.New("this commit failed")
		}).RollbackWith(func() error {
			return Remove(&a, 1)
		}),
	)

	assert.Equal(t, 0, a, "a should be 0 as initial value")
	assert.NotNil(t, err, "There should be Commit Failed error")
}

func TestMultipleStepsCommitSuccess(t *testing.T) {
	a := 0

	tr := transaction.New(&transaction.Option{})

	tr.Execute(
		tr.Step().CommitWith(func() error {
			return Add(&a, 1)
		}).RollbackWith(func() error {
			return Remove(&a, 1)
		}),
		tr.Step().CommitWith(func() error {
			return Add(&a, 1)
		}).RollbackWith(func() error {
			return Remove(&a, 1)
		}),
	)

	assert.Equal(t, 2, a, "a should be 1 after commit")
}

func TestMultipleStepsCommitFailed(t *testing.T) {
	a := 0

	tr := transaction.New(&transaction.Option{})

	err := tr.Execute(
		tr.Step().CommitWith(func() error {
			return Add(&a, 1)
		}).RollbackWith(func() error {
			return Remove(&a, 1)
		}),
		tr.Step().CommitWith(func() error {
			return errors.New("this commit failed")
		}).RollbackWith(func() error {
			return Remove(&a, 1)
		}),
	)

	assert.Equal(t, 0, a, "a should be 0 as initial value")
	assert.NotNil(t, err, "There should be Commit Failed error")
}

func TestMultipleStepsRollbackFailed(t *testing.T) {
	a := 0

	tr := transaction.New(&transaction.Option{})

	err := tr.Execute(
		tr.Step().CommitWith(func() error {
			return Add(&a, 1)
		}).RollbackWith(func() error {
			return errors.New("this rollback failed")
		}),
		tr.Step().CommitWith(func() error {
			return errors.New("this commit failed")
		}).RollbackWith(func() error {
			return Remove(&a, 1)
		}),
	)

	assert.Equal(t, 1, a, "a should be 1 after commit and failed to rollback")
	assert.NotNil(t, err, "There should be an error")
}

func TestCommitWithRetry(t *testing.T) {
	a := 0

	tr := transaction.New(&transaction.Option{
		RetryCount: 3,
		RetryDelay: 1 * time.Second,
	})

	err := tr.Execute(
		tr.Step().CommitWith(func() error {
			return Add(&a, 1)
		}).RollbackWith(func() error {
			return Remove(&a, 1)
		}),
		tr.Step().CommitWith(func() error {
			return errors.New("this commit failed")
		}).RollbackWith(func() error {
			return Remove(&a, 1)
		}),
	)

	assert.Equal(t, 0, a, "a should be 0 as initial value")
	assert.NotNil(t, err, "There should be Commit Failed error")
}

func TestRollbackWithRetry(t *testing.T) {
	a := 0

	tr := transaction.New(&transaction.Option{
		RetryCount: 3,
		RetryDelay: 1 * time.Second,
	})

	err := tr.Execute(
		tr.Step().CommitWith(func() error {
			return Add(&a, 1)
		}).RollbackWith(func() error {
			return errors.New("this rollback failed")
		}),
		tr.Step().CommitWith(func() error {
			return errors.New("this commit failed")
		}).RollbackWith(func() error {
			return Remove(&a, 1)
		}),
	)

	assert.Equal(t, 1, a, "a should be 0 as initial value")
	assert.NotNil(t, err, "There should be Commit Failed error")
}

func TestStepNoRollback(t *testing.T) {
	a := 0

	tr := transaction.New(&transaction.Option{})

	err := tr.Execute(
		tr.Step().CommitWith(func() error {
			return Add(&a, 1)
		}),
		tr.Step().CommitWith(func() error {
			return Add(&a, 1)
		}),
		tr.Step().CommitWith(func() error {
			return errors.New("this commit failed")
		}),
	)

	assert.NotNil(t, err, "There should be an error")
	assert.Equal(t, 2, a, "a should be 1 after commit")
}

func TestGetCommitFailedError(t *testing.T) {
	tr := transaction.New(&transaction.Option{})

	err := tr.Execute(
		tr.Step().CommitWith(func() error {
			return errors.New("this commit failed")
		}),
	)

	assert.NotNil(t, err, "There should be an error")

	unwrappedErr := errors.Unwrap(err)
	assert.Equal(t, "this commit failed", unwrappedErr.Error(), "Unwrapped error should match the original error message")
}
