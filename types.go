package rbclient

import (
	"context"
)

// ConnectionParams are parameters required to establish a connection with Rosenbridge.
type ConnectionParams struct {
	// ClientID is the ID of the client who will own the connection.
	ClientID string

	// HTTPAddr is the list of HTTP addresses of Rosenbridge cluster nodes.
	HTTPAddr []string
	// WebsocketAddr is the list of Websocket addresses of Rosenbridge cluster nodes.
	WebsocketAddr []string

	// OnIncomingMessageReq is invoked when an incoming message request is received from Rosenbridge.
	OnIncomingMessageReq OnIncomingMessageReqFunc
	// OnOutgoingMessageRes is invoked when an outgoing message request is received from Rosenbridge.
	OnOutgoingMessageRes OnOutgoingMessageResFunc
	// OnError is invoked whenever an error occurs during any step of the message processing.
	OnError OnErrorFunc
}

// IncomingMessageReq is the schema for an incoming message from Rosenbridge, originally sent by another client.
type IncomingMessageReq struct{}

// OutgoingMessageReq is the schema for an outgoing message to Rosenbridge.
type OutgoingMessageReq struct{}

// OutgoingMessageRes is the response of an OutgoingMessageReq.
// It tells which clients successfully received the message and which did not, along with the reason.
type OutgoingMessageRes struct{}

// OnIncomingMessageReqFunc is the type of function that handles an incoming message request from Rosenbridge.
type OnIncomingMessageReqFunc func(ctx context.Context, req *IncomingMessageReq)

// OnOutgoingMessageResFunc is the type of function that handles an outgoing message response from Rosenbridge.
type OnOutgoingMessageResFunc func(ctx context.Context, res *OutgoingMessageRes)

// OnErrorFunc is the type of function that handles any errors occurred during message processing.
type OnErrorFunc func(ctx context.Context, err error)
