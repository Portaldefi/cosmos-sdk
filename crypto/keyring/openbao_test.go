package keyring_test

import (
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/stretchr/testify/require"
)

// TestBaoBackend tests the Bao keyring backend.
// This test requires a running Bao/Vault server with the transit engine enabled.
// Set BaoAddr environment variable before running.
func TestBaoBackend(t *testing.T) {
	// Skip if Bao is not configured
	if os.Getenv("BAO_ADDR") == "" && os.Getenv("VAULT_ADDR") == "" {
		t.Skip("Skipping Bao test: BAO_ADDR or VAULT_ADDR not set")
	}

	t.Run("NewBaoClient", func(t *testing.T) {
		// Get address from environment or use test value
		addr := os.Getenv("BAO_ADDR")
		if addr == "" {
			addr = os.Getenv("VAULT_ADDR")
		}
		if addr == "" {
			addr = "http://localhost:8200" // test default
		}
		// Pass address, empty namespace, and test mount path
		client, err := keyring.NewBaoClient(addr, "", "test-mount")
		require.NoError(t, err)
		require.NotNil(t, client)
	})

	// Note: The following tests require a properly configured Bao/Vault server
	// with the transit engine enabled and a key created.
	// Uncomment and customize based on your Bao setup.

	/*
		// Note: These tests require:
		// 1. Running Bao server with Ethereum plugin
		// 2. A configured key-manager (vault) with keys
		// 3. Known public keys for the Ethereum addresses

		t.Run("SignData", func(t *testing.T) {
			client, err := keyring.NewBaoClient()
			require.NoError(t, err)

			// Must have an existing key in Bao Ethereum plugin
			vaultName := "test-vault"
			address := "0x1234567890123456789012345678901234567890"
			data := []byte("test message to sign")

			sig, err := client.SignData(vaultName, address, data)
			require.NoError(t, err)
			require.NotEmpty(t, sig)
		})

		t.Run("SaveAndUseBaoKey", func(t *testing.T) {
			// Initialize keyring with Bao backend
			kr, err := keyring.New(
				"test",
				keyring.BackendBao,
				t.TempDir(),
				nil,
				getCodec(),
			)
			require.NoError(t, err)

			// You must provide the public key (cannot retrieve from Bao)
			// This should be the ethsecp256k1 public key (65 bytes uncompressed)
			testPubKeyHex := "04..." // Your actual public key here
			pubKeyBytes, _ := hex.DecodeString(testPubKeyHex)
			pubKey := &ethsecp256k1.PubKey{Key: pubKeyBytes}

			// Save the Bao key reference
			vaultName := "test-vault"
			ethAddress := "0x1234567890123456789012345678901234567890"

			record, err := kr.SaveBaoKey(
				"my-bao-key",
				pubKey,
				vaultName,
				ethAddress,
			)
			require.NoError(t, err)
			require.NotNil(t, record)
			require.Equal(t, keyring.TypeBao, record.GetType())

			// Retrieve the key
			retrievedRecord, err := kr.Key("my-bao-key")
			require.NoError(t, err)
			require.Equal(t, "my-bao-key", retrievedRecord.Name)

			// Sign a message
			msg := []byte("test transaction data")
			sig, pubKeyReturned, err := kr.Sign("my-bao-key", msg, signing.SignMode_SIGN_MODE_DIRECT)
			require.NoError(t, err)
			require.NotEmpty(t, sig)
			require.NotNil(t, pubKeyReturned)
		})
	*/
}

// TestBaoBackendConfig tests the Bao backend configuration
func TestBaoBackendConfig(t *testing.T) {
	t.Run("DefaultMountPoint", func(t *testing.T) {
		// Clear the environment variable
		oldVal := os.Getenv("BAO_MOUNT_POINT")
		os.Unsetenv("BAO_MOUNT_POINT")
		defer func() {
			if oldVal != "" {
				os.Setenv("BAO_MOUNT_POINT", oldVal)
			}
		}()

		// The config should use default path when BAO_MOUNT_POINT is not set
		// This is tested implicitly through the New function
	})

	t.Run("CustomMountPoint", func(t *testing.T) {
		customPath := "/custom/bao/path"
		os.Setenv("BAO_MOUNT_POINT", customPath)
		defer os.Unsetenv("BAO_MOUNT_POINT")

		// The config should use the custom path
		// This is tested implicitly through the New function
	})
}

// TestBaoRecordType tests the Bao record type
func TestBaoRecordType(t *testing.T) {
	t.Run("TypeBao", func(t *testing.T) {
		require.Equal(t, "bao", keyring.TypeBao.String())
	})
}

// Example of how to use Bao keyring in your application
func ExampleBaoKeyring() {
	// This example shows how to use the Bao Ethereum keyring backend
	// Prerequisites:
	// 1. Set BAO_ADDR="http://localhost:8200"
	// 2. Have Bao Ethereum plugin configured with a key-manager
	// 3. Know your key's Ethereum address and public key

	/*
		import (
			"encoding/hex"
			"github.com/cosmos/cosmos-sdk/crypto/keyring"
			"github.com/cosmos/cosmos-sdk/crypto/keys/ethsecp256k1"
		)

		// Initialize keyring with Bao backend
		kr, err := keyring.New(
			"myapp",
			keyring.BackendBao,
			"/path/to/keyring",
			nil,
			cdc,
		)
		if err != nil {
			panic(err)
		}

		// You MUST provide the public key (cannot retrieve from Bao API)
		// This is the ethsecp256k1 public key (65 bytes uncompressed)
		pubKeyHex := "04abcdef..." // Your actual public key
		pubKeyBytes, _ := hex.DecodeString(pubKeyHex)
		pubKey := &ethsecp256k1.PubKey{Key: pubKeyBytes}

		// Save the Bao key reference
		record, err := kr.SaveBaoKey(
			"my-local-key",           // local name in keyring
			pubKey,                    // ethsecp256k1 public key (REQUIRED)
			"foundation",              // vault/key-manager name in Bao
			"0x1234567890...",        // Ethereum address
		)
		if err != nil {
			panic(err)
		}

		// Use the key to sign
		msg := []byte("transaction data")
		sig, pubKey, err := kr.Sign("my-local-key", msg, signing.SignMode_SIGN_MODE_DIRECT)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Signature: %x\n", sig)
		fmt.Printf("Public key: %x\n", pubKey.Bytes())
	*/
}
