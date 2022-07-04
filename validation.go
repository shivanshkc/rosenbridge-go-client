package rosenbridge

import (
	"errors"
)

// checkBridgeParams checks if the provided BridgeParams are valid.
// TODO: We should have more aggressive validations rather than just nil or empty checks.
func checkBridgeParams(params *BridgeParams) error {
	// The params itself should not be nil.
	if params == nil {
		return errors.New("connection params cannot be nil")
	}

	// The client ID cannot be an empty string.
	// More aggressive client ID validations can be done by the backend.
	if params.ClientID == "" {
		return errors.New("client id cannot be empty")
	}

	// The address list should not be empty.
	if len(params.Addresses) == 0 {
		return errors.New("address list cannot be empty")
	}
	// An address should not be empty.
	for _, addr := range params.Addresses {
		if len(addr) == 0 {
			return errors.New("address cannot be empty")
		}
	}

	// The handlers should not be nil.
	if params.OnIncomingMessageReq == nil {
		return errors.New("incoming message request handler cannot be nil")
	}
	if params.OnOutgoingMessageRes == nil {
		return errors.New("outgoing message response handler cannot be nil")
	}
	if params.OnErrorRes == nil {
		return errors.New("error response handler cannot be nil")
	}

	return nil
}
