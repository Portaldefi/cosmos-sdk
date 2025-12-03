package keyring

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/ethereum/go-ethereum/crypto"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// BaoClient is a client for interacting with Bao/Vault
type BaoClient struct {
	address string
	client  *http.Client
}

// NewBaoClient creates a new Bao client
// It reads address from configuration; token-based auth is not used.
func NewBaoClient(configAddr string) (*BaoClient, error) {
	// Read address from config
	address := configAddr
	if address == "" {
		return nil, errors.New("Bao address must be set in client.toml (bao-addr)")
	}

	return &BaoClient{
		address: address,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// SignData signs data using Bao Ethereum plugin
// This uses the endpoint: /ethereum/key-managers/{vault-name}/sign
func (c *BaoClient) SignData(vaultName, keyAddress string, data []byte) ([]byte, error) {
	// Bao Ethereum plugin endpoint
	url := fmt.Sprintf("%s/ethereum/key-managers/%s/sign", c.address, vaultName)

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

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send request to Bao")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Newf("Bao returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response (Bao Ethereum plugin format)
	var result BaoSignResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal response")
	}

	// Check if signature exists
	if result.Data.Signature == "" {
		return nil, errors.New("empty signature returned from Bao")
	}

	// Decode hex signature
	signature, err := hex.DecodeString(result.Data.Signature)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode signature")
	}

	return signature, nil
}

// BaoSignResponse represents the response from Bao Ethereum plugin
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

// SignWithBao signs a message using Bao Ethereum plugin
// configAddr is the Bao server address
func SignWithBao(k *Record, msg []byte, configAddr string) (sig []byte, pub cryptotypes.PubKey, err error) {
	baoInfo := k.GetBao()
	if baoInfo == nil {
		return nil, nil, errors.New("not an Bao record")
	}

	client, err := NewBaoClient(configAddr)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create Bao client")
	}

	// Get the Ethereum address from the public key
	pub, err = k.GetPubKey()
	if err != nil {
		return nil, nil, err
	}

	// Extract the address from public key (first 20 bytes of Keccak256 hash)
	// The VaultPath field stores the vault name (key-manager name)
	// The KeyName field stores the Ethereum address
	vaultName := baoInfo.VaultPath
	address := baoInfo.KeyName

	sig, err = client.SignData(vaultName, address, msg)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to sign with Bao")
	}

	return sig, pub, nil
}
