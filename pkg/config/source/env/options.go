package env

type Option func(*env)

func WithParseArray(seg string) Option {
	if seg == "" {
		seg = ";"
	}
	return func(e *env) {
		e.parseArray = true
		e.arraySep = seg
	}
}

func SetKeyDelimiter(delimiter string) Option {
	if delimiter == "" {
		delimiter = "_"
	}
	return func(e *env) {
		e.delimiter = delimiter
	}
}
