package rbclient

import (
	"context"
)

// Connection represents a connection with the Rosenbridge cluster.
// Use the available methods on this struct to interact with Rosenbridge.
type Connection struct{}

// NewConnection provides a new connection with Rosenbridge.
// This connection can be used to send requests to Rosenbridge and also to listen to all responses.
func NewConnection(ctx context.Context, params *ConnectionParams) (*Connection, error) {
	return nil, nil
}

// SendMessage sends a new message over Rosenbridge synchronously.
// Unlike SendMessageAsync, it blocks until Rosenbridge returns the OutgoingMessageRes.
func (c *Connection) SendMessage(ctx context.Context, req *OutgoingMessageReq) (*OutgoingMessageRes, error) {
	return nil, nil
}

// SendMessageAsync send a new message over Rosenbridge asynchronously.
// Unlike SendMessage, it does not wait for a response from Rosenbridge and returns immediately.
//
// The OutgoingMessageRes for messages sent by this method can be monitored by the OnOutgoingMessageResponse
// function.
func (c *Connection) SendMessageAsync(ctx context.Context, req *OutgoingMessageReq) error {
	return nil
}

// Disconnect closes the connection.
func (c *Connection) Disconnect(ctx context.Context) error {
	return nil
}
