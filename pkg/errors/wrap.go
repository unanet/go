package errors

import (
	"database/sql/driver"
	"fmt"

	"github.com/pkg/errors"
)

type cause interface {
	Cause() error
}

type unanetError interface {
	IsUnanetError() bool
}

func Wrap(err error, args ...interface{}) error {
	if err == nil {
		return nil
	}
	if ee, ok := err.(unanetError); ok && ee.IsUnanetError() {
		return err
	} else if _, ok := err.(cause); ok {
		return err
	} else {
		if len(args) == 0 {
			return errors.Wrap(err, err.Error())
		} else if len(args) == 1 {
			return errors.Wrap(err, fmt.Sprintf("%v", args[0]))
		} else {
			return errors.Wrap(err, fmt.Sprintf(fmt.Sprintf("%v", args[0]), args[1:]...))
		}
	}
}

func Wrapf(format string, a ...interface{}) error {
	return Wrap(fmt.Errorf(format, a...))
}

func WrapTx(tx driver.Tx, err error) error {
	if tx == nil {
		return Wrap(err)
	}
	_ = tx.Rollback()
	return Wrap(err)
}
