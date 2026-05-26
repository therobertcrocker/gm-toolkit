package builder

import (
	"errors"
	"fmt"
)

var (
	ErrMissingSource = errors.New("missing source file")
	ErrOutputExists  = errors.New("canonical output exists")
)

func layoutErrorf(line, col int, format string, args ...any) error {
	return fmt.Errorf("layout.txt:%d:%d: "+format, append([]any{line, col}, args...)...)
}

func dataErrorf(context, format string, args ...any) error {
	var prefix string
	if context == "" {
		prefix = "data.toml: "
	} else {
		prefix = "data.toml:" + context + ": "
	}
	return fmt.Errorf(prefix+format, args...)
}
