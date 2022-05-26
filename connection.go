package rbclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

// Connection represents a connection with the Rosenbridge cluster.
// Use the available methods on this struct to interact with Rosenbridge.
type Connection struct {
	// params are the ConnectionParams that were used to make this connection.
	params *ConnectionParams

	// underlyingConnection is the low-level websocket connection.
	underlyingConnection *websocket.Conn
	// httpClients maps addresses to their HTTP clients.
	httpClients map[string]*http.Client
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
	// Getting the REST endpoint.
	endpoint := rbGetConnectionURL(getRandomAddr(c.params.HTTPAddr), c.params.ClientID)

	// Establishing websocket connection.
	underlyingConn, response, err := websocket.DefaultDialer.Dial(endpoint, nil)
	if err != nil {
		return fmt.Errorf("error in websocket.Dial: %w", err)
	}
	defer func() { _ = response.Body.Close() }()

	// Persisting the connection.
	c.underlyingConnection = underlyingConn
	// Listening to websocket messages.
	go c.initListener(ctx)

	return nil
}

// Disconnect closes the websocket connection with Rosenbridge.
func (c *Connection) Disconnect(ctx context.Context) error {
	if err := c.underlyingConnection.Close(); err != nil {
		return fmt.Errorf("failed to close underlying connection: %w", err)
	}
	return nil
}

// SendMessage sends a new message over Rosenbridge synchronously.
// Unlike SendMessageAsync, it blocks until Rosenbridge returns the OutgoingMessageRes.
func (c *Connection) SendMessage(ctx context.Context, req *OutgoingMessageReq) (*OutgoingMessageRes, error) {
	return nil, errors.New("todo")
}

// SendMessageAsync send a new message over Rosenbridge asynchronously.
// Unlike SendMessage, it does not wait for a response from Rosenbridge and returns immediately.
//
// The OutgoingMessageRes for messages sent by this method can be monitored by the OnOutgoingMessageResponse
// function.
func (c *Connection) SendMessageAsync(ctx context.Context, req *OutgoingMessageReq) error {
	return nil
}

// initListener sets up a loop that keeps listening to the websocket messages and calls appropriate handlers.
//
// This method blocks and should be called from within a goroutine.
func (c *Connection) initListener(ctx context.Context) {
	// Starting a loop to process all websocket communication.
	// This loop panics when the connection is closed.
main:
	for {
		wsMessageType, message, err := c.underlyingConnection.ReadMessage()
		if err != nil {
			// Forming the wrapped error.
			err = fmt.Errorf("%w: error in ReadMessage: %v", ErrConnectionClosure, err)
			// Invoking the OnError function with the formed error.
			c.params.OnError(ctx, nil, err)
			// Breaking out of the loop.
			break main
		}

		// Handling different websocket message types.
		switch wsMessageType {
		case websocket.CloseMessage:
			// This means graceful connection closure.
			c.params.OnError(ctx, nil, nil)
			// Breaking out of the loop.
			break main
		case websocket.TextMessage:
			messageType, err := unmarshalMessageType(message)
			if err != nil {
				// Forming the wrapped error.
				err = fmt.Errorf("%w: %v", ErrUnknownMessageType, err)
				// Invoking the OnError function with the formed error.
				c.params.OnError(ctx, message, err)
				// Continuing with the loop.
				continue main
			}

			// Handling different message types.
			switch messageType {
			case typeIncomingMessageReq:
				// Decoding the message into the IncomingMessageReq type.
				inMessageReq, err := unmarshalIncomingMessageReq(message)
				if err != nil {
					c.params.OnError(ctx, message, fmt.Errorf("failed to decode incoming message req: %w", err))
					continue main
				}

				// Invoking the handler.
				c.params.OnIncomingMessageReq(ctx, inMessageReq)
			case typeOutgoingMessageRes:
				outMessageRes, err := unmarshalOutgoingMessageRes(message)
				if err != nil {
					c.params.OnError(ctx, message, fmt.Errorf("failed to decode outgoing message res: %w", err))
					continue main
				}

				// Invoking the handler.
				c.params.OnOutgoingMessageRes(ctx, outMessageRes)
			default:
				// Unknown message types are simply ignored.
			}
		case websocket.BinaryMessage:
		case websocket.PingMessage:
		case websocket.PongMessage:
		default:
		}
	}
}
