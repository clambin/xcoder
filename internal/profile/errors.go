package profile

import "errors"

var ErrSourceInTargetCodec = errors.New("video already in target codec")

type ErrSourceRejected struct {
	Reason string
}

func (e ErrSourceRejected) Error() string {
	return e.Reason
}

func (e ErrSourceRejected) Is(e2 error) bool {
	var err ErrSourceRejected
	ok := errors.As(e2, &err)
	return ok
}
