package rbclient

import (
	"errors"
)

var (
	// ErrConnectionClosure error is returned when a websocket connection is closed.
	ErrConnectionClosure = errors.New("connection closed")
	// ErrUnknownMessageType error is returned when the message received from Rosenbridge contains an unknown or invalid
	// message type, or does not contain one at all.
	ErrUnknownMessageType = errors.New("unknown, invalid or absent message type")
)

// Types of data sent/received over the connection to/from Rosenbridge.
const (
	typeIncomingMessageReq string = "INCOMING_MESSAGE_REQ"
	typeOutgoingMessageReq string = "OUTGOING_MESSAGE_REQ"
	typeOutgoingMessageRes string = "OUTGOING_MESSAGE_RES"
)
