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
type IncomingMessageReq struct {
	// SenderID is the ID of the client who sent the message.
	SenderID string
	// Message is the main message content.
	Message []byte
	// Persist is the persistence criteria of this message set by the sender.
	Persist Persistence
}

// OutgoingMessageReq is the schema for an outgoing message to Rosenbridge.
type OutgoingMessageReq struct {
	// ReceiverIDs is the list of IDs of clients who are intended to receive this message.
	ReceiverIDs []string
	// Message is the main message content.
	Message []byte
	// Persist is the persistence criteria of this message set by the sender.
	Persist Persistence
}

// OutgoingMessageRes is the response of an OutgoingMessageReq.
// It tells which clients successfully received the message and which did not, along with the reason.
type OutgoingMessageRes struct {
	// Code is the global response code. For example: OK
	Code string
	// Reason is the loggable or human-readable reason for failures, if any.
	Reason string
	// PerReceiver is the response data per receiver.
	PerReceiver []struct {
		// ReceiverID is the ID of the receiver to whom this slice element belongs.
		ReceiverID string
		// Code is the response code for this receiver.
		Code string
		// Reason is the loggable or human-readable reason for failures, if any, for this receiver.
		Reason string
	}
}

// OnIncomingMessageReqFunc is the type of function that handles an incoming message request from Rosenbridge.
type OnIncomingMessageReqFunc func(ctx context.Context, req *IncomingMessageReq)

// OnOutgoingMessageResFunc is the type of function that handles an outgoing message response from Rosenbridge.
type OnOutgoingMessageResFunc func(ctx context.Context, res *OutgoingMessageRes)

// OnErrorFunc is the type of function that handles any errors occurred during message processing.
type OnErrorFunc func(ctx context.Context, err error)

// Persistence is a type for the various message persistence criterion provided by Rosenbridge.
type Persistence string

const (
	// True always persists the message.
	True Persistence = "true"
	// False never persists the message. If the receiver is offline, the message is lost forever.
	False Persistence = "false"
	// IfOffline persists the message only if the receiver is offline.
	IfOffline Persistence = "if_offline"
)
