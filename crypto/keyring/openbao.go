package keyring

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/ethereum/go-ethereum/crypto"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// OpenBaoClient is a client for interacting with OpenBao/Vault
type OpenBaoClient struct {
	address string
	token   string
	client  *http.Client
}

// NewOpenBaoClient creates a new OpenBao client
// It reads from config parameters first, then falls back to environment variables
// Parameters can be empty strings to use environment variables
func NewOpenBaoClient(configAddr, configTokenFile string) (*OpenBaoClient, error) {
	// Read address: config
	address := configAddr
	if address == "" {
		return nil, errors.New("OpenBao address must be set in client.toml (openbao-addr)")
	}

	// Read token: try token file config
	var token string
	tokenFile := configTokenFile

	if tokenFile != "" {
		// Read token from file
		tokenBytes, err := os.ReadFile(tokenFile)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read OpenBao token from file: %s", tokenFile)
		}
		token = strings.TrimSpace(string(tokenBytes))
	} else {
		return nil, errors.New("OpenBao token must be set in client.toml (openbao-token-file)")
	}

	return &OpenBaoClient{
		address: address,
		token:   token,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// SignData signs data using OpenBao Ethereum plugin
// This uses the endpoint: /v1/ethereum/key-managers/{vault-name}/sign
func (c *OpenBaoClient) SignData(vaultName, keyAddress string, data []byte) ([]byte, error) {
	// OpenBao Ethereum plugin endpoint
	url := fmt.Sprintf("%s/v1/ethereum/key-managers/%s/sign", c.address, vaultName)

	// Hash the data if it's not already 32 bytes (Keccak256)
	var hashToSign []byte
	if len(data) != crypto.DigestLength {
		hashToSign = crypto.Keccak256Hash(data).Bytes()
	} else {
		hashToSign = data
	}

	// Prepare request payload for Ethereum signing
	payload := map[string]string{
		"hash":    hex.EncodeToString(hashToSign),
		"address": keyAddress,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal payload")
	}

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	// Use Bearer token authentication (not X-Vault-Token)
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send request to OpenBao")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Newf("OpenBao returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response (OpenBao Ethereum plugin format)
	var result BaoSignResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal response")
	}

	// Check if signature exists
	if result.Data.Signature == "" {
		return nil, errors.New("empty signature returned from OpenBao")
	}

	// Decode hex signature
	signature, err := hex.DecodeString(result.Data.Signature)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode signature")
	}

	return signature, nil
}

// BaoSignResponse represents the response from OpenBao Ethereum plugin
type BaoSignResponse struct {
	Auth interface{} `json:"auth"`
	Data struct {
		Signature string `json:"signature"`
	} `json:"data"`
	LeaseDuration int         `json:"lease_duration"`
	LeaseID       string      `json:"lease_id"`
	Renewable     bool        `json:"renewable"`
	RequestID     string      `json:"request_id"`
	Warnings      interface{} `json:"warnings"`
	WrapInfo      interface{} `json:"wrap_info"`
}

// SignWithOpenBao signs a message using OpenBao Ethereum plugin
// configAddr and configTokenFile can be empty strings to use environment variables
func SignWithOpenBao(k *Record, msg []byte, configAddr, configTokenFile string) (sig []byte, pub cryptotypes.PubKey, err error) {
	openBaoInfo := k.GetOpenbao()
	if openBaoInfo == nil {
		return nil, nil, errors.New("not an OpenBao record")
	}

	client, err := NewOpenBaoClient(configAddr, configTokenFile)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create OpenBao client")
	}

	// Get the Ethereum address from the public key
	pub, err = k.GetPubKey()
	if err != nil {
		return nil, nil, err
	}

	// Extract the address from public key (first 20 bytes of Keccak256 hash)
	// The VaultPath field stores the vault name (key-manager name)
	// The KeyName field stores the Ethereum address
	vaultName := openBaoInfo.VaultPath
	address := openBaoInfo.KeyName

	sig, err = client.SignData(vaultName, address, msg)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to sign with OpenBao")
	}

	return sig, pub, nil
}
