package transaction

import (
	"errors"
	"fmt"
	"time"
)

type tr struct {
	opt *Option
}

type Option struct {
	RetryDelay time.Duration
	RetryCount int8
}

func New(opt *Option) *tr {
	return &tr{
		opt: opt,
	}
}

func (t *tr) Execute(steps ...*step) error {
	failedAt, err := t.commitPhase(steps...)
	if err != nil {
		err := t.rollbackPhase(failedAt, steps...)
		return fmt.Errorf("Commit failed: %w", err)
	}
	return nil
}

func (t *tr) commitPhase(steps ...*step) (int, error) {
	for i, step := range steps {
		for r := range t.opt.RetryCount + 1 {
			err := step.commit()
			if err == nil {
				break
			}

			if r == t.opt.RetryCount {
				return i, errors.New("Commit failed: " + err.Error())
			}

			time.Sleep(t.opt.RetryDelay)
		}
	}
	return -1, nil
}

func (t *tr) rollbackPhase(start int, steps ...*step) error {
	for i := start - 1; i >= 0; i-- {
		for r := range t.opt.RetryCount + 1 {
			if steps[i].rollback == nil {
				continue
			}

			err := steps[i].rollback()
			if err == nil {
				break
			}

			if r == t.opt.RetryCount {
				return errors.New("Rollback failed: " + err.Error())
			}

			time.Sleep(t.opt.RetryDelay)
		}
	}
	return nil
}

type step struct {
	commit   func() error
	rollback func() error
}

func (t *tr) Step() *step {
	return &step{}
}

func (s *step) CommitWith(commit func() error) *step {
	s.commit = commit
	return s
}

func (s *step) RollbackWith(rollback func() error) *step {
	s.rollback = rollback
	return s
}
