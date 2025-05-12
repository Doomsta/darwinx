package darwinx

import (
	"github.com/pkg/errors"
	"sync"
)

type Option interface {
	apply(darwinx *Darwinx) error
}

type optionFn func(darwinx *Darwinx) error

func (f optionFn) apply(darwinx *Darwinx) error {
	return f(darwinx)
}

func WithMigration(migrations []Migration) Option {
	return optionFn(func(darwinx *Darwinx) error {
		for _, m := range migrations {
			darwinx.migrations = append(darwinx.migrations, m)
		}
		return nil
	})
}

func WithNoTransaction() Option {
	return optionFn(func(darwinx *Darwinx) error {
		darwinx.transaction = false
		return nil
	})
}

func WithTableName(tn string) Option {
	return optionFn(func(darwinx *Darwinx) error {
		// only a-z, A-Z, 0-9 and _ are allowed
		for _, r := range tn {
			if !((r >= 'a' && r <= 'z') ||
				(r >= 'A' && r <= 'Z') ||
				(r >= '0' && r <= '9') ||
				r == '_') {
				return errors.New("invalid table name")
			}
		}
		darwinx.tableName = tn
		return nil
	})
}
