package rosenbridge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"strings"
)

func formatBaseURL(addr string, enableTLS bool, isHTTP bool) string {
	// Removing troublemaker suffixes.
	addr = strings.TrimSuffix(addr, "/api/")
	addr = strings.TrimSuffix(addr, "/api")
	addr = strings.TrimSuffix(addr, "/")

	// Removing troublemaker prefixes.
	addr = strings.TrimPrefix(addr, "http://")
	addr = strings.TrimPrefix(addr, "https://")
	addr = strings.TrimPrefix(addr, "ws://")
	addr = strings.TrimPrefix(addr, "wss://")

	// Deciding on the protocol.
	protocol := "ws"
	if isHTTP {
		protocol = "http"
	}
	if enableTLS {
		protocol += "s"
	}

	// Attaching the protocol and returning.
	return fmt.Sprintf("%s://%s", protocol, addr)
}

// getRandomString provides a random value from the provided slice.
func getRandomString(values []string) string {
	// Generating a random index.
	// nolint:gosec // gosec recommends the use of crypto/rand instead of math/rand but IMO, that's not required.
	randIndex := rand.Intn(len(values))
	return values[randIndex]
}

// readerFromAny.
func readerFromAny(value interface{}) (io.Reader, error) {
	// Marshalling the value to byte array.
	requestBytes, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("error in json.Marshal call: %w", err)
	}
	// Converting the value byte array to io.Reader.
	return bytes.NewReader(requestBytes), nil
}

// incomingMessageReqFromAny converts the provided interface into an IncomingMessageReq.
func incomingMessageReqFromAny(value interface{}) (*IncomingMessageReq, error) {
	// Marshalling the value to byte array.
	requestBytes, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("error in json.Marshal call: %w", err)
	}

	// Unmarshalling the byte array into an IncomingMessageReq.
	inMessage := &IncomingMessageReq{}
	if err := json.Unmarshal(requestBytes, inMessage); err != nil {
		return nil, fmt.Errorf("error in json.Unmarshal call: %w", err)
	}

	return inMessage, nil
}

// outgoingMessageResFromAny converts the provided interface into an OutgoingMessageRes.
func outgoingMessageResFromAny(value interface{}) (*OutgoingMessageRes, error) {
	// Marshalling the value to byte array.
	requestBytes, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("error in json.Marshal call: %w", err)
	}

	// Unmarshalling the byte array into an OutgoingMessageRes.
	outRes := &OutgoingMessageRes{}
	if err := json.Unmarshal(requestBytes, outRes); err != nil {
		return nil, fmt.Errorf("error in json.Unmarshal call: %w", err)
	}

	return outRes, nil
}

// codeAndReasonFromAny converts the provided interface into a CodeAndReason.
func codeAndReasonFromAny(value interface{}) (*CodeAndReason, error) {
	// Marshalling the value to byte array.
	requestBytes, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("error in json.Marshal call: %w", err)
	}

	// Unmarshalling the byte array into an OutgoingMessageRes.
	cnr := &CodeAndReason{}
	if err := json.Unmarshal(requestBytes, cnr); err != nil {
		return nil, fmt.Errorf("error in json.Unmarshal call: %w", err)
	}

	return cnr, nil
}
