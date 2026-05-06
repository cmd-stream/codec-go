package codec

type Options struct {
	maxLen int
}

type SetOption func(*Options)

func WithMaxLen(maxLen int) SetOption {
	return func(o *Options) {
		o.maxLen = maxLen
	}
}

func Apply(opts []SetOption, o *Options) {
	for _, opt := range opts {
		opt(o)
	}
}
