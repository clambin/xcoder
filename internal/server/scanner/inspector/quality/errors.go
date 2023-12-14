package quality

import "errors"

type ErrSourceRejected struct {
	Reason string
}

func (e ErrSourceRejected) Error() string {
	return "source rejected: " + e.Reason
}

func (e ErrSourceRejected) Is(e2 error) bool {
	var err ErrSourceRejected
	ok := errors.As(e2, &err)
	return ok
}
