package lazy

import "errors"

var (
	ErrNilServer     = errors.New("lazy: nil server")
	ErrNothingToWork = errors.New("lazy: nothing to work")
	ErrJobMissing    = errors.New("lazy: cannot find job")
)
