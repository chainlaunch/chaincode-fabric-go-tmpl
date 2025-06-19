package main

import (
	"log"
	"os"
	"strconv"

	"github.com/chainlaunch/chaincode-fabric-go-tmpl/chaincode"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// serverConfig holds the configuration parameters needed to start the chaincode server.
// These values are typically provided through environment variables.
type serverConfig struct {
	CCID    string // Chaincode ID as registered with the fabric network
	Address string // Network address where the chaincode server will listen
}

// main initializes and starts the chaincode server.
func main() {
	// See chaincode.env.example for required variables
	config := serverConfig{
		CCID:    os.Getenv("CORE_CHAINCODE_ID"),
		Address: os.Getenv("CORE_CHAINCODE_ADDRESS"),
	}

	// Create a new chaincode instance with the SimpleChaincode
	// SimpleCo implements the business logic for storing and retrieving hash records
	chaincodeInstance, err := contractapi.NewChaincode(&chaincode.SimpleChaincode{})

	if err != nil {
		log.Panicf("error create  chaincode: %s", err)
	}

	// Configure the chaincode server with the appropriate settings
	server := &shim.ChaincodeServer{
		CCID:     config.CCID,        // Chaincode ID from environment
		Address:  config.Address,     // Network address from environment
		CC:       chaincodeInstance,  // The initialized chaincode
		TLSProps: getTLSProperties(), // TLS configuration
	}

	// Start the chaincode server
	// This will block until the server is shutdown or encounters an error
	if err := server.Start(); err != nil {
		log.Panicf("error starting  chaincode: %s", err)
	}
}

// getTLSProperties configures and returns the TLS settings for the chaincode server.
// It reads TLS configuration from environment variables and loads the necessary
// cryptographic materials (keys and certificates) when TLS is enabled.
// Returns a TLSProperties struct that can be used to configure the chaincode server.
func getTLSProperties() shim.TLSProperties {
	// Check if chaincode is TLS enabled by reading from environment variables
	tlsDisabledStr := getEnvOrDefault("CHAINCODE_TLS_DISABLED", "true")
	key := getEnvOrDefault("CHAINCODE_TLS_KEY", "")
	cert := getEnvOrDefault("CHAINCODE_TLS_CERT", "")
	clientCACert := getEnvOrDefault("CHAINCODE_CLIENT_CA_CERT", "")

	// convert tlsDisabledStr to boolean
	tlsDisabled := getBoolOrDefault(tlsDisabledStr, false)
	var keyBytes, certBytes, clientCACertBytes []byte
	var err error

	if !tlsDisabled {
		keyBytes, err = os.ReadFile(key)
		if err != nil {
			log.Panicf("error while reading the crypto file: %s", err)
		}
		certBytes, err = os.ReadFile(cert)
		if err != nil {
			log.Panicf("error while reading the crypto file: %s", err)
		}
	}
	// Did not request for the peer cert verification
	if clientCACert != "" {
		clientCACertBytes, err = os.ReadFile(clientCACert)
		if err != nil {
			log.Panicf("error while reading the crypto file: %s", err)
		}
	}

	return shim.TLSProperties{
		Disabled:      tlsDisabled,
		Key:           keyBytes,
		Cert:          certBytes,
		ClientCACerts: clientCACertBytes,
	}
}

// getEnvOrDefault retrieves the value of an environment variable or returns a default value if not set.
// Parameters:
//   - env: The name of the environment variable to retrieve
//   - defaultVal: The default value to return if the environment variable is not set
//
// Returns the value of the environment variable or the default value.
func getEnvOrDefault(env, defaultVal string) string {
	value, ok := os.LookupEnv(env)
	if !ok {
		value = defaultVal
	}
	return value
}

// getBoolOrDefault converts a string to a boolean value or returns a default value if parsing fails.
// Parameters:
//   - value: The string to convert to a boolean
//   - defaultVal: The default boolean value to return if parsing fails
//
// Returns the parsed boolean value or the default value if parsing fails.
// Note that the method returns default value if the string cannot be parsed!
func getBoolOrDefault(value string, defaultVal bool) bool {
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return defaultVal
	}
	return parsed
}
