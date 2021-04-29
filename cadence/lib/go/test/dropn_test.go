package test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	emulator "github.com/onflow/flow-emulator"
	"github.com/onflow/flow-go-sdk"
	sdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/templates"
	"github.com/onflow/flow-go-sdk/test"

	ft_contracts "github.com/onflow/flow-ft/lib/go/contracts"
)

const (
	dropnRootPath           = "../../../cadence/dropn"
	dropnDropnPath         = dropnRootPath + "/contracts/Dropn.cdc"
	dropnSetupAccountPath   = dropnRootPath + "/transactions/setup_account.cdc"
	dropnTransferTokensPath = dropnRootPath + "/transactions/transfer_tokens.cdc"
	dropnMintTokensPath     = dropnRootPath + "/transactions/mint_tokens.cdc"
	dropnBurnTokensPath     = dropnRootPath + "/transactions/burn_tokens.cdc"
	dropnGetBalancePath     = dropnRootPath + "/scripts/get_balance.cdc"
	dropnGetSupplyPath      = dropnRootPath + "/scripts/get_supply.cdc"
)

func DropnDeployContracts(b *emulator.Blockchain, t *testing.T) (flow.Address, flow.Address, crypto.Signer) {
	accountKeys := test.AccountKeyGenerator()

	// Should be able to deploy a contract as a new account with no keys.
	fungibleTokenCode := loadFungibleToken()
	fungibleAddr, err := b.CreateAccount(
		[]*flow.AccountKey{},
		[]templates.Contract{templates.Contract{
			Name:   "FungibleToken",
			Source: string(fungibleTokenCode),
		}},
	)
	assert.NoError(t, err)

	_, err = b.CommitBlock()
	assert.NoError(t, err)

	dropnAccountKey, dropnSigner := accountKeys.NewWithSigner()
	dropnCode := loadDropn(fungibleAddr)

	dropnAddr, err := b.CreateAccount(
		[]*flow.AccountKey{dropnAccountKey},
		[]templates.Contract{templates.Contract{
			Name:   "Dropn",
			Source: string(dropnCode),
		}},
	)
	assert.NoError(t, err)

	_, err = b.CommitBlock()
	assert.NoError(t, err)

	// Simplify testing by having the contract address also be our initial Vault.
	DropnSetupAccount(t, b, dropnAddr, dropnSigner, fungibleAddr, dropnAddr)

	return fungibleAddr, dropnAddr, dropnSigner
}

func DropnSetupAccount(t *testing.T, b *emulator.Blockchain, userAddress sdk.Address, userSigner crypto.Signer, fungibleAddr sdk.Address, dropnAddr sdk.Address) {
	tx := flow.NewTransaction().
		SetScript(dropnGenerateSetupDropnAccountTransaction(fungibleAddr, dropnAddr)).
		SetGasLimit(100).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		AddAuthorizer(userAddress)

	signAndSubmit(
		t, b, tx,
		[]flow.Address{b.ServiceKey().Address, userAddress},
		[]crypto.Signer{b.ServiceKey().Signer(), userSigner},
		false,
	)
}

func DropnCreateAccount(t *testing.T, b *emulator.Blockchain, fungibleAddr sdk.Address, dropnAddr sdk.Address) (sdk.Address, crypto.Signer) {
	userAddress, userSigner, _ := createAccount(t, b)
	DropnSetupAccount(t, b, userAddress, userSigner, fungibleAddr, dropnAddr)
	return userAddress, userSigner
}

func DropnMint(t *testing.T, b *emulator.Blockchain, fungibleAddr sdk.Address, dropnAddr sdk.Address, dropnSigner crypto.Signer, recipientAddress flow.Address, amount string, shouldRevert bool) {
	tx := flow.NewTransaction().
		SetScript(dropnGenerateMintDropnTransaction(fungibleAddr, dropnAddr)).
		SetGasLimit(100).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		AddAuthorizer(dropnAddr)

	_ = tx.AddArgument(cadence.NewAddress(recipientAddress))
	_ = tx.AddArgument(CadenceUFix64(amount))

	signAndSubmit(
		t, b, tx,
		[]flow.Address{b.ServiceKey().Address, dropnAddr},
		[]crypto.Signer{b.ServiceKey().Signer(), dropnSigner},
		shouldRevert,
	)

}

func TestDropnDeployment(t *testing.T) {
	b := newEmulator()

	fungibleAddr, dropnAddr, _ := DropnDeployContracts(b, t)

	t.Run("Should have initialized Supply field correctly", func(t *testing.T) {
		supply := executeScriptAndCheck(t, b, dropnGenerateGetSupplyScript(fungibleAddr, dropnAddr), nil)
		expectedSupply, expectedSupplyErr := cadence.NewUFix64("0.0")
		assert.NoError(t, expectedSupplyErr)
		assert.Equal(t, expectedSupply, supply.(cadence.UFix64))
	})
}

func TestDropnSetupAccount(t *testing.T) {
	b := newEmulator()

	t.Run("Should be able to create empty Vault that doesn't affect supply", func(t *testing.T) {

		fungibleAddr, dropnAddr, _ := DropnDeployContracts(b, t)

		userAddress, _ := DropnCreateAccount(t, b, fungibleAddr, dropnAddr)

		balance := executeScriptAndCheck(t, b, dropnGenerateGetBalanceScript(fungibleAddr, dropnAddr), [][]byte{jsoncdc.MustEncode(cadence.Address(userAddress))})
		assert.Equal(t, CadenceUFix64("0.0"), balance)

		supply := executeScriptAndCheck(t, b, dropnGenerateGetSupplyScript(fungibleAddr, dropnAddr), nil)
		assert.Equal(t, CadenceUFix64("0.0"), supply.(cadence.UFix64))
	})
}

func TestDropnMinting(t *testing.T) {
	b := newEmulator()

	fungibleAddr, dropnAddr, dropnSigner := DropnDeployContracts(b, t)

	userAddress, _ := DropnCreateAccount(t, b, fungibleAddr, dropnAddr)

	t.Run("Shouldn't be able to mint zero tokens", func(t *testing.T) {
		DropnMint(t, b, fungibleAddr, dropnAddr, dropnSigner, userAddress, "0.0", true)
	})

	t.Run("Should mint tokens, deposit, and update balance and total supply", func(t *testing.T) {
		DropnMint(t, b, fungibleAddr, dropnAddr, dropnSigner, userAddress, "50.0", false)

		// Assert that the vault's balance is correct
		result, err := b.ExecuteScript(dropnGenerateGetBalanceScript(fungibleAddr, dropnAddr), [][]byte{jsoncdc.MustEncode(cadence.Address(userAddress))})
		require.NoError(t, err)
		if !assert.True(t, result.Succeeded()) {
			t.Log(result.Error.Error())
		}
		balance := result.Value
		assert.Equal(t, CadenceUFix64("50.0"), balance.(cadence.UFix64))

		// Make sure that the total supply is correct
		supply := executeScriptAndCheck(t, b, dropnGenerateGetSupplyScript(fungibleAddr, dropnAddr), nil)
		assert.Equal(t, CadenceUFix64("50.0"), supply.(cadence.UFix64))
	})
}

func TestDropnTransfers(t *testing.T) {
	b := newEmulator()

	fungibleAddr, dropnAddr, dropnSigner := DropnDeployContracts(b, t)

	userAddress, _ := DropnCreateAccount(t, b, fungibleAddr, dropnAddr)

	DropnMint(t, b, fungibleAddr, dropnAddr, dropnSigner, dropnAddr, "1000.0", false)

	t.Run("Shouldn't be able to withdraw more than the balance of the Vault", func(t *testing.T) {
		tx := flow.NewTransaction().
			SetScript(dropnGenerateTransferVaultScript(fungibleAddr, dropnAddr)).
			SetGasLimit(100).
			SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
			SetPayer(b.ServiceKey().Address).
			AddAuthorizer(dropnAddr)

		_ = tx.AddArgument(CadenceUFix64("30000.0"))
		_ = tx.AddArgument(cadence.NewAddress(userAddress))

		signAndSubmit(
			t, b, tx,
			[]flow.Address{b.ServiceKey().Address, dropnAddr},
			[]crypto.Signer{b.ServiceKey().Signer(), dropnSigner},
			true,
		)

		// Assert that the vaults' balances are correct
		result, err := b.ExecuteScript(dropnGenerateGetBalanceScript(fungibleAddr, dropnAddr), [][]byte{jsoncdc.MustEncode(cadence.Address(dropnAddr))})
		require.NoError(t, err)
		if !assert.True(t, result.Succeeded()) {
			t.Log(result.Error.Error())
		}
		balance := result.Value
		assert.Equal(t, balance.(cadence.UFix64), CadenceUFix64("1000.0"))

		result, err = b.ExecuteScript(dropnGenerateGetBalanceScript(fungibleAddr, dropnAddr), [][]byte{jsoncdc.MustEncode(cadence.Address(userAddress))})
		require.NoError(t, err)
		if !assert.True(t, result.Succeeded()) {
			t.Log(result.Error.Error())
		}
		balance = result.Value
		assert.Equal(t, balance.(cadence.UFix64), CadenceUFix64("0.0"))
	})

	t.Run("Should be able to withdraw and deposit tokens from a vault", func(t *testing.T) {
		tx := flow.NewTransaction().
			SetScript(dropnGenerateTransferVaultScript(fungibleAddr, dropnAddr)).
			SetGasLimit(100).
			SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
			SetPayer(b.ServiceKey().Address).
			AddAuthorizer(dropnAddr)

		_ = tx.AddArgument(CadenceUFix64("300.0"))
		_ = tx.AddArgument(cadence.NewAddress(userAddress))

		signAndSubmit(
			t, b, tx,
			[]flow.Address{b.ServiceKey().Address, dropnAddr},
			[]crypto.Signer{b.ServiceKey().Signer(), dropnSigner},
			false,
		)

		// Assert that the vaults' balances are correct
		result, err := b.ExecuteScript(dropnGenerateGetBalanceScript(fungibleAddr, dropnAddr), [][]byte{jsoncdc.MustEncode(cadence.Address(dropnAddr))})
		require.NoError(t, err)
		if !assert.True(t, result.Succeeded()) {
			t.Log(result.Error.Error())
		}
		balance := result.Value
		assert.Equal(t, balance.(cadence.UFix64), CadenceUFix64("700.0"))

		result, err = b.ExecuteScript(dropnGenerateGetBalanceScript(fungibleAddr, dropnAddr), [][]byte{jsoncdc.MustEncode(cadence.Address(userAddress))})
		require.NoError(t, err)
		if !assert.True(t, result.Succeeded()) {
			t.Log(result.Error.Error())
		}
		balance = result.Value
		assert.Equal(t, balance.(cadence.UFix64), CadenceUFix64("300.0"))

		supply := executeScriptAndCheck(t, b, dropnGenerateGetSupplyScript(fungibleAddr, dropnAddr), nil)
		assert.Equal(t, supply.(cadence.UFix64), CadenceUFix64("1000.0"))
	})
}

func dropnReplaceAddressPlaceholders(code string, fungibleAddress, dropnAddress string) []byte {
	return []byte(replaceStrings(
		code,
		map[string]string{
			ftAddressPlaceholder:     "0x" + fungibleAddress,
			dropnAddressPlaceHolder: "0x" + dropnAddress,
		},
	))
}

func loadFungibleToken() []byte {
	return ft_contracts.FungibleToken()
}

func loadDropn(fungibleAddr flow.Address) []byte {
	return []byte(strings.ReplaceAll(
		string(readFile(dropnDropnPath)),
		ftAddressPlaceholder,
		"0x"+fungibleAddr.String(),
	))
}

func dropnGenerateGetSupplyScript(fungibleAddr, dropnAddr flow.Address) []byte {
	return dropnReplaceAddressPlaceholders(
		string(readFile(dropnGetSupplyPath)),
		fungibleAddr.String(),
		dropnAddr.String(),
	)
}

func dropnGenerateGetBalanceScript(fungibleAddr, dropnAddr flow.Address) []byte {
	return dropnReplaceAddressPlaceholders(
		string(readFile(dropnGetBalancePath)),
		fungibleAddr.String(),
		dropnAddr.String(),
	)
}
func dropnGenerateTransferVaultScript(fungibleAddr, dropnAddr flow.Address) []byte {
	return dropnReplaceAddressPlaceholders(
		string(readFile(dropnTransferTokensPath)),
		fungibleAddr.String(),
		dropnAddr.String(),
	)
}

func dropnGenerateSetupDropnAccountTransaction(fungibleAddr, dropnAddr flow.Address) []byte {
	return dropnReplaceAddressPlaceholders(
		string(readFile(dropnSetupAccountPath)),
		fungibleAddr.String(),
		dropnAddr.String(),
	)
}

func dropnGenerateMintDropnTransaction(fungibleAddr, dropnAddr flow.Address) []byte {
	return dropnReplaceAddressPlaceholders(
		string(readFile(dropnMintTokensPath)),
		fungibleAddr.String(),
		dropnAddr.String(),
	)
}
