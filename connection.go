package rbclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

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
	// httpClientsMutex makes sure that httpClients is thread safe to use.
	httpClientsMutex *sync.RWMutex
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
		httpClients:          map[string]*http.Client{},
		httpClientsMutex:     &sync.RWMutex{},
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
	// Setting the type of the message.
	req.Type = typeOutgoingMessageReq

	// Marshalling request to byte array.
	requestBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	// Converting the request byte array to io.Reader for the http client.
	bodyReader := bytes.NewReader(requestBytes)

	// Getting a random address.
	randomAddr := getRandomAddr(c.params.HTTPAddr)
	// Getting an HTTP client for this address.
	httpClient := c.httpClientForAddr(randomAddr)

	// Getting the REST endpoint.
	endpoint := rbPostMessageURL(randomAddr, c.params.ClientID)
	// Forming the HTTP request.
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to form the http request: %w", err)
	}

	// Executing the request.
	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to execute http request: %w", err)
	}

	// Getting the response body.
	responseBody, err := unmarshalHTTPResponse(httpResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode the response body: %w", err)
	}

	// Converting the responseBody.data type into the OutgoingMessageRes type.
	outMessageRes, err := toOutgoingMessageRes(responseBody)
	if err != nil {
		return nil, fmt.Errorf("failed to convert http response body to outgoing message response: %w", err)
	}

	// If the http response data did not contain a request ID, we fetch it from the headers.
	if outMessageRes.RequestID == "" {
		outMessageRes.RequestID = httpResponse.Header.Get("x-request-id")
	}

	return outMessageRes, nil
}

// SendMessageAsync send a new message over Rosenbridge asynchronously.
// Unlike SendMessage, it does not wait for a response from Rosenbridge and returns immediately.
//
// The OutgoingMessageRes for messages sent by this method can be monitored by the OnOutgoingMessageResponse
// function.
func (c *Connection) SendMessageAsync(ctx context.Context, req *OutgoingMessageReq) error {
	// Setting the type of the message.
	req.Type = typeOutgoingMessageReq

	// Marshalling request to byte array.
	requestBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Writing the message to the connection.
	if err := c.underlyingConnection.WriteMessage(websocket.TextMessage, requestBytes); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

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

// httpClientForAddr provides an HTTP client for the given address.
//
// First, it tries to find a client in the httpClients map which is persisted in the Connection struct.
// If a client is not found, a new one is created, persisted and returned.
//
// It is safe for concurrent use.
func (c *Connection) httpClientForAddr(addr string) *http.Client {
	// Ensuring thread safety.
	c.httpClientsMutex.Lock()
	defer c.httpClientsMutex.Unlock()

	// Checking if the HTTP client is already present for this addr, otherwise creating one.
	httpClient, exists := c.httpClients[addr]
	if !exists {
		httpClient = &http.Client{}
		c.httpClients[addr] = httpClient
	}

	return httpClient
}
