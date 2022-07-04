package rosenbridge

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Bridge represents a connection with the Rosenbridge cluster.
type Bridge struct {
	// params is the data required to form a bridge.
	params *BridgeParams

	// underlyingConnection is the low-level websocket connection for this bridge.
	underlyingConnection *websocket.Conn
	// httpClients maps addresses to their HTTP clients.
	httpClients map[string]*http.Client
	// httpClientsMutex makes sure that httpClients is thread safe to use.
	httpClientsMutex *sync.RWMutex
}

// NewBridge acts as a constructor for the Bridge type.
//
// Note that a separate Connect call has to be made to actually connect the bridge to Rosenbridge.
func NewBridge(params *BridgeParams) (*Bridge, error) {
	// Validating the params.
	if err := checkBridgeParams(params); err != nil {
		return nil, fmt.Errorf("error in checkBridgeParams call: %w", err)
	}

	return &Bridge{
		params:               params,
		underlyingConnection: nil,
		httpClients:          map[string]*http.Client{},
		httpClientsMutex:     &sync.RWMutex{},
	}, nil
}

// Connect actually connects the bridge with Rosenbridge.
func (b *Bridge) Connect() error {
	// Getting a random address for connection.
	addr := formatBaseURL(getRandomString(b.params.Addresses), b.params.EnableTLS, false)
	addr = fmt.Sprintf("%s/api/bridge", addr)

	// HTTP request headers.
	headers := http.Header{}
	headers.Set("x-client-id", b.params.ClientID)

	// Dialing the websocket address.
	underlyingConn, response, err := websocket.DefaultDialer.Dial(addr, headers)
	if err != nil {
		return fmt.Errorf("error in DefaultDialer.Dial call: %w", err)
	}
	if response.StatusCode > 399 {
		return fmt.Errorf("failed to upgrade websocket conn: %s", response.Status)
	}
	defer func() { _ = response.Body.Close() }()

	// Attaching the underlying connection to the bridge.
	b.underlyingConnection = underlyingConn
	go b.initListener()

	return nil
}

// Disconnect closes the bridge.
func (b *Bridge) Disconnect() error {
	if err := b.underlyingConnection.Close(); err != nil {
		return fmt.Errorf("error in underlyingConnection.Close call: %w", err)
	}
	return nil
}

// SendMessage can be used to send a message synchronously over Rosenbridge.
func (b *Bridge) SendMessage(ctx context.Context, message *BridgeMessage) (*OutgoingMessageRes, error) {
	// For now, clients can only send *OutgoingMessageReq type messages to Rosenbridge.
	if _, asserted := message.Body.(*OutgoingMessageReq); !asserted {
		return nil, fmt.Errorf("body of the message should be of type *OutgoingMessageReq")
	}

	// Converting the message to io.Reader to use in the HTTP Request body.
	reader, err := readerFromAny(message.Body)
	if err != nil {
		return nil, fmt.Errorf("error in readerFromAny call: %w", err)
	}

	// Getting a random address.
	addr := formatBaseURL(getRandomString(b.params.Addresses), b.params.EnableTLS, true)
	addr = fmt.Sprintf("%s/api/message", addr)

	// Getting an HTTP client for this address.
	httpClient := b.httpClientForAddr(addr)
	// Forming the HTTP request.
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, addr, reader)
	if err != nil {
		return nil, fmt.Errorf("error in http.NewRequestWithContext call: %w", err)
	}
	// Setting request headers.
	httpRequest.Header.Set("x-request-id", message.RequestID)
	httpRequest.Header.Set("x-client-id", b.params.ClientID)

	// Executing the request.
	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("error in httpClient.Do call: %w", err)
	}
	defer func() { _ = httpResponse.Body.Close() }()

	// The response message.
	responseMessage := &OutgoingMessageRes{}
	if err := json.NewDecoder(httpResponse.Body).Decode(responseMessage); err != nil {
		return nil, fmt.Errorf("error in json.NewDecoder(...).Decode call: %w", err)
	}

	return responseMessage, nil
}

// SendMessageAsync can be used to send a message asynchronously over Rosenbridge.
func (b *Bridge) SendMessageAsync(ctx context.Context, message *BridgeMessage) error {
	// Setting the type of the message.
	message.Type = MessageOutgoingReq

	// Marshalling message to byte array.
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error in json.Marshal call: %w", err)
	}

	// Writing the message to the connection.
	if err := b.underlyingConnection.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
		return fmt.Errorf("error in underlyingConnection.WriteMessage call: %w", err)
	}

	return nil
}

// httpClientForAddr provides an HTTP client for the given address.
//
// First, it tries to find a client in the httpClients map which is persisted in the Connection struct.
// If a client is not found, a new one is created, persisted and returned.
//
// It is safe for concurrent use.
func (b *Bridge) httpClientForAddr(addr string) *http.Client {
	// Ensuring thread safety.
	b.httpClientsMutex.Lock()
	defer b.httpClientsMutex.Unlock()

	// Checking if the HTTP client is already present for this addr, otherwise creating one.
	httpClient, exists := b.httpClients[addr]
	if !exists {
		httpClient = &http.Client{}
		b.httpClients[addr] = httpClient
	}

	return httpClient
}

// initListener sets up a loop that keeps listening to the websocket messages and calls appropriate handlers.
//
// This method blocks and should be called from within a goroutine.
func (b *Bridge) initListener() {
	// Starting a loop to process all websocket communication.
	// This loop breaks when the connection is closed.
main:
	for {
		wsMessageType, message, err := b.underlyingConnection.ReadMessage()
		if err != nil {
			// Creating a new CnR with the error reason and notifying the client.
			b.params.OnErrorRes("", &CodeAndReason{Code: BridgeCloseCnR.Code, Reason: err.Error()})
			// Breaking out of the loop.
			break main
		}

		// Handling different websocket message types.
		switch wsMessageType {
		case websocket.CloseMessage:
			// This means graceful connection closure.
			// Notifying the client.
			b.params.OnErrorRes("", BridgeCloseCnR)
			// Breaking out of the loop.
			break main
		case websocket.TextMessage:
			// Unmarshalling the received message into a bridge message.
			bridgeMessage := &BridgeMessage{}
			if err := json.Unmarshal(message, bridgeMessage); err != nil {
				// Creating a new CnR with the error reason and notifying the client.
				b.params.OnErrorRes("", &CodeAndReason{Code: BadMessageCnR.Code, Reason: err.Error()})
				// Continuing with the main loop.
				continue main
			}
			// Handling different message types.
			switch bridgeMessage.Type {
			case MessageIncomingReq:
				// Unmarshalling the message body into an incoming message request type.
				inMessage, err := incomingMessageReqFromAny(bridgeMessage.Body)
				if err != nil {
					// Creating a new CnR with the error reason and notifying the client.
					b.params.OnErrorRes("", &CodeAndReason{Code: BadMessageCnR.Code, Reason: err.Error()})
					// Continuing with the main loop.
					continue main
				}
				// Notifying the user.
				b.params.OnIncomingMessageReq(bridgeMessage.RequestID, inMessage)
			case MessageOutgoingRes:
				// Unmarshalling the message body into an outgoing message response type.
				outRes, err := outgoingMessageResFromAny(bridgeMessage.Body)
				if err != nil {
					// Creating a new CnR with the error reason and notifying the client.
					b.params.OnErrorRes("", &CodeAndReason{Code: BadMessageCnR.Code, Reason: err.Error()})
					// Continuing with the main loop.
					continue main
				}
				// Notifying the user.
				b.params.OnOutgoingMessageRes(bridgeMessage.RequestID, outRes)
			case MessageErrorRes:
				// Unmarshalling the message body into a code and reason type.
				cnr, err := codeAndReasonFromAny(bridgeMessage.Body)
				if err != nil {
					// Creating a new CnR with the error reason and notifying the client.
					b.params.OnErrorRes("", &CodeAndReason{Code: BadMessageCnR.Code, Reason: err.Error()})
					// Continuing with the main loop.
					continue main
				}
				// Notifying the user.
				b.params.OnErrorRes(bridgeMessage.RequestID, cnr)
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
