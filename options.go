package darwinx

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

func WithTableName(tn string) Option {
	return optionFn(func(darwinx *Darwinx) error {
		darwinx.tableName = tn
		return nil
	})
}
