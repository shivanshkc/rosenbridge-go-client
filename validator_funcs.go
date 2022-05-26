package rbclient

import (
	"errors"
)

// checkConnectionParams checks if the provided ConnectionParams are valid.
// TODO: We should have more aggressive validations rather than just nil or empty checks.
func checkConnectionParams(params *ConnectionParams) error {
	// The params itself should not be nil.
	if params == nil {
		return errors.New("connection params cannot be nil")
	}

	// The client ID cannot be an empty string.
	// More aggressive client ID validations can be done by the backend.
	if params.ClientID == "" {
		return errors.New("client id cannot be empty")
	}

	// The addresses should not be empty.
	if len(params.HTTPAddr) == 0 {
		return errors.New("http address list cannot be empty")
	}
	if len(params.WebsocketAddr) == 0 {
		return errors.New("websocket address list cannot be empty")
	}

	// The handlers should not be nil.
	if params.OnIncomingMessageReq == nil {
		return errors.New("incoming message request handler cannot be nil")
	}
	if params.OnOutgoingMessageRes == nil {
		return errors.New("outgoing message response handler cannot be nil")
	}
	if params.OnError == nil {
		return errors.New("error handler cannot be nil")
	}

	return nil
}
