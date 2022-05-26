package rbclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
)

// rbGetConnectionURL provides the URL for the Get-Connection API of Rosenbridge.
func rbGetConnectionURL(baseURL string, clientID string) string {
	// Removing troublemaker suffixes.
	baseURL = strings.TrimSuffix(baseURL, "/api/")
	baseURL = strings.TrimSuffix(baseURL, "/api")
	baseURL = strings.TrimSuffix(baseURL, "/")

	// Creating and returning the URL.
	return fmt.Sprintf("%s/api/clients/%s/connections", baseURL, clientID)
}

// rbPostMessageURL provides the URL for the Post-Message API of Rosenbridge.
func rbPostMessageURL(baseURL string, clientID string) string {
	// Removing troublemaker suffixes.
	baseURL = strings.TrimSuffix(baseURL, "/api/")
	baseURL = strings.TrimSuffix(baseURL, "/api")
	baseURL = strings.TrimSuffix(baseURL, "/")

	// Creating and returning the URL.
	return fmt.Sprintf("%s/api/clients/%s/messages", baseURL, clientID)
}

// getRandomAddr provides a random address from the list of provided addresses.
func getRandomAddr(addr []string) string {
	// Generating a random index.
	// nolint:gosec // gosec recommends the use of crypto/rand instead of math/rand but IMO, that's not required.
	randIndex := rand.Intn(len(addr))
	return addr[randIndex]
}

// unmarshalMessageType is a pure function that provides the type of the message.
func unmarshalMessageType(message []byte) (string, error) {
	// Decoding into a simple map.
	decoded := map[string]interface{}{}
	if err := json.Unmarshal(message, &decoded); err != nil {
		return "", fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// Checking if there's a type key.
	mTypeInterface, exists := decoded["type"]
	if !exists {
		return "", fmt.Errorf("no type key present")
	}

	// Checking if the message type is string.
	mType, asserted := mTypeInterface.(string)
	if !asserted {
		return "", fmt.Errorf("message type value is not string: %v", mTypeInterface)
	}

	return mType, nil
}

// unmarshalIncomingMessageReq decodes the provided byte slice into an IncomingMessageReq type.
func unmarshalIncomingMessageReq(message []byte) (*IncomingMessageReq, error) {
	inMessage := &IncomingMessageReq{}
	if err := json.Unmarshal(message, inMessage); err != nil {
		return nil, fmt.Errorf("error in json.Unmarshal call: %w", err)
	}
	return inMessage, nil
}

// unmarshalOutgoingMessageRes decodes the provided byte slice into an OutgoingMessageRes type.
func unmarshalOutgoingMessageRes(message []byte) (*OutgoingMessageRes, error) {
	outMessageResp := &OutgoingMessageRes{}
	if err := json.Unmarshal(message, outMessageResp); err != nil {
		return nil, fmt.Errorf("error in json.Unmarshal call: %w", err)
	}
	return outMessageResp, nil
}

// unmarshalHTTPResponse decodes a http response body into its struct.
func unmarshalHTTPResponse(response *http.Response) (*httpResponseBody, error) {
	// Closing the body upon function return.
	defer func() { _ = response.Body.Close() }()

	// Reading into a byte slice.
	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Unmarshalling into the struct.
	responseBody := &httpResponseBody{}
	if err := json.Unmarshal(bodyBytes, responseBody); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return responseBody, nil
}

func toOutgoingMessageRes(data interface{}) (*OutgoingMessageRes, error) {
	// Marshalling into json to later unmarshal into struct.
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("error in json.Marshal call: %w", err)
	}

	// Unmarshalling into struct.
	outMessageRes := &OutgoingMessageRes{}
	if err := json.Unmarshal(dataBytes, outMessageRes); err != nil {
		return nil, fmt.Errorf("error in json.Unmarshal call: %w", err)
	}

	return outMessageRes, nil
}
