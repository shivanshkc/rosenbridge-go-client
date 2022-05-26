package rbclient

import (
	"errors"
)

// ErrConnectionClosure error is returned when a websocket connection is closed.
var ErrConnectionClosure = errors.New("connection closed")
