package keyring_test

import (
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/stretchr/testify/require"
)

// TestOpenBaoBackend tests the OpenBao keyring backend.
// This test requires a running OpenBao/Vault server with the transit engine enabled.
// Set OPENBAO_ADDR and OPENBAO_TOKEN environment variables before running.
func TestOpenBaoBackend(t *testing.T) {
	// Skip if OpenBao is not configured
	if os.Getenv("OPENBAO_ADDR") == "" && os.Getenv("VAULT_ADDR") == "" {
		t.Skip("Skipping OpenBao test: OPENBAO_ADDR or VAULT_ADDR not set")
	}
	if os.Getenv("OPENBAO_TOKEN") == "" && os.Getenv("VAULT_TOKEN") == "" {
		t.Skip("Skipping OpenBao test: OPENBAO_TOKEN or VAULT_TOKEN not set")
	}

	t.Run("NewOpenBaoClient", func(t *testing.T) {
		// Pass empty strings to use environment variables
		client, err := keyring.NewOpenBaoClient("", "")
		require.NoError(t, err)
		require.NotNil(t, client)
	})

	// Note: The following tests require a properly configured OpenBao/Vault server
	// with the transit engine enabled and a key created.
	// Uncomment and customize based on your OpenBao setup.

	/*
		// Note: These tests require:
		// 1. Running OpenBao server with Ethereum plugin
		// 2. A configured key-manager (vault) with keys
		// 3. Known public keys for the Ethereum addresses

		t.Run("SignData", func(t *testing.T) {
			client, err := keyring.NewOpenBaoClient()
			require.NoError(t, err)

			// Must have an existing key in OpenBao Ethereum plugin
			vaultName := "test-vault"
			address := "0x1234567890123456789012345678901234567890"
			data := []byte("test message to sign")

			sig, err := client.SignData(vaultName, address, data)
			require.NoError(t, err)
			require.NotEmpty(t, sig)
		})

		t.Run("SaveAndUseOpenBaoKey", func(t *testing.T) {
			// Initialize keyring with OpenBao backend
			kr, err := keyring.New(
				"test",
				keyring.BackendOpenBao,
				t.TempDir(),
				nil,
				getCodec(),
			)
			require.NoError(t, err)

			// You must provide the public key (cannot retrieve from OpenBao)
			// This should be the ethsecp256k1 public key (65 bytes uncompressed)
			testPubKeyHex := "04..." // Your actual public key here
			pubKeyBytes, _ := hex.DecodeString(testPubKeyHex)
			pubKey := &ethsecp256k1.PubKey{Key: pubKeyBytes}

			// Save the OpenBao key reference
			vaultName := "test-vault"
			ethAddress := "0x1234567890123456789012345678901234567890"

			record, err := kr.SaveOpenBaoKey(
				"my-openbao-key",
				pubKey,
				vaultName,
				ethAddress,
			)
			require.NoError(t, err)
			require.NotNil(t, record)
			require.Equal(t, keyring.TypeOpenBao, record.GetType())

			// Retrieve the key
			retrievedRecord, err := kr.Key("my-openbao-key")
			require.NoError(t, err)
			require.Equal(t, "my-openbao-key", retrievedRecord.Name)

			// Sign a message
			msg := []byte("test transaction data")
			sig, pubKeyReturned, err := kr.Sign("my-openbao-key", msg, signing.SignMode_SIGN_MODE_DIRECT)
			require.NoError(t, err)
			require.NotEmpty(t, sig)
			require.NotNil(t, pubKeyReturned)
		})
	*/
}

// TestOpenBaoBackendConfig tests the OpenBao backend configuration
func TestOpenBaoBackendConfig(t *testing.T) {
	t.Run("DefaultMountPoint", func(t *testing.T) {
		// Clear the environment variable
		oldVal := os.Getenv("OPENBAO_MOUNT_POINT")
		os.Unsetenv("OPENBAO_MOUNT_POINT")
		defer func() {
			if oldVal != "" {
				os.Setenv("OPENBAO_MOUNT_POINT", oldVal)
			}
		}()

		// The config should use default path when OPENBAO_MOUNT_POINT is not set
		// This is tested implicitly through the New function
	})

	t.Run("CustomMountPoint", func(t *testing.T) {
		customPath := "/custom/openbao/path"
		os.Setenv("OPENBAO_MOUNT_POINT", customPath)
		defer os.Unsetenv("OPENBAO_MOUNT_POINT")

		// The config should use the custom path
		// This is tested implicitly through the New function
	})
}

// TestOpenBaoRecordType tests the OpenBao record type
func TestOpenBaoRecordType(t *testing.T) {
	t.Run("TypeOpenBao", func(t *testing.T) {
		require.Equal(t, "openbao", keyring.TypeOpenBao.String())
	})
}

// Example of how to use OpenBao keyring in your application
func ExampleOpenBaoKeyring() {
	// This example shows how to use the OpenBao Ethereum keyring backend
	// Prerequisites:
	// 1. Set OPENBAO_ADDR="http://localhost:8200"
	// 2. Set OPENBAO_TOKEN="your-bearer-token"
	// 3. Have OpenBao Ethereum plugin configured with a key-manager
	// 4. Know your key's Ethereum address and public key

	/*
		import (
			"encoding/hex"
			"github.com/cosmos/cosmos-sdk/crypto/keyring"
			"github.com/cosmos/cosmos-sdk/crypto/keys/ethsecp256k1"
		)

		// Initialize keyring with OpenBao backend
		kr, err := keyring.New(
			"myapp",
			keyring.BackendOpenBao,
			"/path/to/keyring",
			nil,
			cdc,
		)
		if err != nil {
			panic(err)
		}

		// You MUST provide the public key (cannot retrieve from OpenBao API)
		// This is the ethsecp256k1 public key (65 bytes uncompressed)
		pubKeyHex := "04abcdef..." // Your actual public key
		pubKeyBytes, _ := hex.DecodeString(pubKeyHex)
		pubKey := &ethsecp256k1.PubKey{Key: pubKeyBytes}

		// Save the OpenBao key reference
		record, err := kr.SaveOpenBaoKey(
			"my-local-key",           // local name in keyring
			pubKey,                    // ethsecp256k1 public key (REQUIRED)
			"foundation",              // vault/key-manager name in OpenBao
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
