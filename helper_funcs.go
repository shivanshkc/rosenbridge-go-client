package rbclient

import (
	"fmt"
	"math/rand"
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

// getRandomAddr provides a random address from the list of provided addresses.
func getRandomAddr(addr []string) string {
	// Generating a random index.
	// nolint:gosec // gosec recommends the use of crypto/rand instead of math/rand but IMO, that's not required.
	randIndex := rand.Intn(len(addr))
	return addr[randIndex]
}
