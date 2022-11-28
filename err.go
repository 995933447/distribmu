package distribmu

import "errors"

var ErrWaitTimeout = errors.New("wait timeout")

var ErrLockLost = errors.New("err lost lock")
