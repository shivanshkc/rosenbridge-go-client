package rosenbridge

// BridgeParams are the params required to create a bridge.
type BridgeParams struct {
	// ClientID is the ID of the client for whom the bridge is being created.
	ClientID string
	// Addresses is the list of addresses of the nodes in the Rosenbridge cluster.
	Addresses []string
	// EnableTLS tells whether to use secured protocol or not.
	EnableTLS bool

	// OnIncomingMessageReq is invoked when an IncomingMessageReq is received over the bridge.
	OnIncomingMessageReq func(requestID string, message *IncomingMessageReq)
	// OnOutgoingMessageRes is invoked when an OutgoingMessageRes is received over the bridge.
	OnOutgoingMessageRes func(requestID string, message *OutgoingMessageRes)
	// OnErrorRes is invoked when an error CodeAndReason is received over the bridge.
	OnErrorRes func(requestID string, message *CodeAndReason)
}

// BridgeMessage represents a message sent/received over a bridge.
type BridgeMessage struct {
	// Type helps differentiate and route different kinds of messages.
	Type string `json:"type"`
	// RequestID is the identifier of this message.
	RequestID string `json:"request_id"`
	// Body is the content of the message.
	Body interface{} `json:"body"`
}

// OutgoingMessageReq represents a request from the client to send a message.
type OutgoingMessageReq struct {
	// Message is the main message content that needs to be delivered.
	Message string `json:"message"`
	// ReceiverIDs is the list of client IDs that are intended to receive this message.
	ReceiverIDs []string `json:"receiver_ids"`
	// Persist is the persistence criteria for the message.
	Persist string `json:"persist"`
}

// OutgoingMessageRes is the response of a client's outgoing message request.
// It gives the final status of the message delivery, including exhaustive error information, if any.
type OutgoingMessageRes struct {
	// The global code and reason.
	//
	// If something causes the entire request to fail, which means the message does not get delivered to even a single
	// bridge, then this code and reason will reflect that error and cause.
	*CodeAndReason
	// Persistence tells whether the messages were successfully persisted or not.
	Persistence *CodeAndReason `json:"persistence"`
	// Bridges is the list of statuses of all bridges that were triggered as part of the request.
	Bridges []*BridgeStatus `json:"bridges"`
}

// IncomingMessageReq represents an incoming message for a client.
// It is called "incoming message request" because the naming is done from the client's perspective.
type IncomingMessageReq struct {
	// SenderID is the ID of the client who sent the message.
	SenderID string `json:"sender_id"`
	// Message is the main message content.
	Message string `json:"message"`
	// Persist is the persistence criteria of the message.
	Persist string `json:"persist"`
}

// BridgeIdentity is the information required to uniquely identify a bridge.
type BridgeIdentity struct {
	// ClientID is the ID of the client to which the bridge belongs.
	ClientID string `json:"client_id" bson:"client_id"`
	// BridgeID is unique for all bridges for a given client.
	// But two bridges, belonging to two different clients may have the same BridgeID.
	BridgeID string `json:"bridge_id" bson:"bridge_id"`
}

// BridgeStatus represents any operation result on a bridge.
type BridgeStatus struct {
	*BridgeIdentity
	*CodeAndReason
}

// CodeAndReason represent the response of an operation.
type CodeAndReason struct {
	// Code is the response code. For example: OK, CONFLICT, OFFLINE etc.
	Code string `json:"code"`
	// Reason is the human-readable error reason.
	Reason string `json:"reason"`
}
