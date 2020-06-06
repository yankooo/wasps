package wasps

type taskOptions struct {
	RecoverFn func(r interface{})
	Args      []interface{}
}

// TaskOption configures how we set up the connection.
type TaskOption interface {
	apply(*taskOptions)
}

// funcTaskOption wraps a function that modifies taskOptions into an
// implementation of the TaskOption interface.
type funcTaskOption struct {
	f func(*taskOptions)
}

func (fdo *funcTaskOption) apply(do *taskOptions) {
	fdo.f(do)
}

func newFuncTaskOption(f func(*taskOptions)) *funcTaskOption {
	return &funcTaskOption{
		f: f,
	}
}

func WithRecoverFn(f func(r interface{})) TaskOption {
	return newFuncTaskOption(func(o *taskOptions) {
		o.RecoverFn = f
	})
}

func WithArgs(args ...interface{}) TaskOption {
	return newFuncTaskOption(func(o *taskOptions) {
		o.Args = args
	})
}

var defaultTaskOptions = &taskOptions{
	RecoverFn: func(r interface{}) {},
}

type poolOptions struct {
}

// PoolOption configures how we set up the connection.
type PoolOption interface {
	apply(*poolOptions)
}

// funcPoolOption wraps a function that modifies poolOptions into an
// implementation of the PoolOption interface.
type funcPoolOption struct {
	f func(*poolOptions)
}

func (fdo *funcPoolOption) apply(do *poolOptions) {
	fdo.f(do)
}

func newFuncPoolOption(f func(*poolOptions)) *funcPoolOption {
	return &funcPoolOption{
		f: f,
	}
}

func WithXXX(f func(r interface{})) PoolOption {
	return newFuncPoolOption(func(o *poolOptions) {
	})
}

var defaultPoolOptions = &poolOptions{}
