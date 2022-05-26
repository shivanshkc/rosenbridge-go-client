package rbclient

import (
	"context"
	"fmt"

	"github.com/gorilla/websocket"
)

// Connection represents a connection with the Rosenbridge cluster.
// Use the available methods on this struct to interact with Rosenbridge.
type Connection struct {
	// params are the ConnectionParams that were used to make this connection.
	params *ConnectionParams

	// underlyingConnection is the low-level websocket connection.
	underlyingConnection *websocket.Conn
}

// NewConnection provides a new connection object.
//
// Note that this function does not automatically connect the websocket. For that, the Connect method has to be called
// explicitly. This is because some Rosenbridge operations do not require a websocket connection.
//
// As long as the Connect method is not called, the websocket connection will not be made, and the client will remain
// OFFLINE (assuming they do not have any other connections).
func NewConnection(ctx context.Context, params *ConnectionParams) (*Connection, error) {
	// Validating the external input.
	if err := checkConnectionParams(params); err != nil {
		return nil, fmt.Errorf("invalid connection params: %w", err)
	}

	// Creating and returning the connection object.
	return &Connection{
		params:               params,
		underlyingConnection: nil,
	}, nil
}

// Connect establishes the websocket connection with Rosenbridge.
func (c *Connection) Connect(ctx context.Context) error {
	return nil
}

// Disconnect closes the websocket connection with Rosenbridge.
func (c *Connection) Disconnect(ctx context.Context) error {
	return nil
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
