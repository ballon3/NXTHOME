package test

import (
	"strings"
	"testing"

	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	emulator "github.com/onflow/flow-emulator"
	sdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	sdktemplates "github.com/onflow/flow-go-sdk/templates"
	"github.com/onflow/flow-go-sdk/test"

	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-go-sdk"

	nft_contracts "github.com/onflow/flow-nft/lib/go/contracts"
)

const (
	nxtRootPath                   = "../../../cadence/nxt"
	nxtNxtPath             = nxtRootPath + "/contracts/Nxt.cdc"
	nxtSetupAccountPath           = nxtRootPath + "/transactions/setup_account.cdc"
	nxtMintKittyItemPath          = nxtRootPath + "/transactions/mint_kitty_item.cdc"
	nxtTransferKittyItemPath      = nxtRootPath + "/transactions/transfer_kitty_item.cdc"
	nxtInspectKittyItemSupplyPath = nxtRootPath + "/scripts/read_kitty_items_supply.cdc"
	nxtInspectCollectionLenPath   = nxtRootPath + "/scripts/read_collection_length.cdc"
	nxtInspectCollectionIdsPath   = nxtRootPath + "/scripts/read_collection_ids.cdc"

	typeID1 = 1000
	typeID2 = 2000
)

func NxtDeployContracts(b *emulator.Blockchain, t *testing.T) (flow.Address, flow.Address, crypto.Signer) {
	accountKeys := test.AccountKeyGenerator()

	// Should be able to deploy a contract as a new account with no keys.
	nftCode := loadNonFungibleToken()
	nftAddr, err := b.CreateAccount(
		nil,
		[]sdktemplates.Contract{
			{
				Name:   "NonFungibleToken",
				Source: string(nftCode),
			},
		})
	if !assert.NoError(t, err) {
		t.Log(err.Error())
	}
	_, err = b.CommitBlock()
	assert.NoError(t, err)

	// Should be able to deploy a contract as a new account with one key.
	nxtAccountKey, nxtSigner := accountKeys.NewWithSigner()
	nxtCode := loadNxt(nftAddr.String())
	nxtAddr, err := b.CreateAccount(
		[]*flow.AccountKey{nxtAccountKey},
		[]sdktemplates.Contract{
			{
				Name:   "Nxt",
				Source: string(nxtCode),
			},
		})
	if !assert.NoError(t, err) {
		t.Log(err.Error())
	}
	_, err = b.CommitBlock()
	assert.NoError(t, err)

	// Simplify the workflow by having the contract address also be our initial test collection.
	NxtSetupAccount(t, b, nxtAddr, nxtSigner, nftAddr, nxtAddr)

	return nftAddr, nxtAddr, nxtSigner
}

func NxtSetupAccount(t *testing.T, b *emulator.Blockchain, userAddress sdk.Address, userSigner crypto.Signer, nftAddr sdk.Address, nxtAddr sdk.Address) {
	tx := flow.NewTransaction().
		SetScript(nxtGenerateSetupAccountScript(nftAddr.String(), nxtAddr.String())).
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

func NxtCreateAccount(t *testing.T, b *emulator.Blockchain, nftAddr sdk.Address, nxtAddr sdk.Address) (sdk.Address, crypto.Signer) {
	userAddress, userSigner, _ := createAccount(t, b)
	NxtSetupAccount(t, b, userAddress, userSigner, nftAddr, nxtAddr)
	return userAddress, userSigner
}

func NxtMintItem(b *emulator.Blockchain, t *testing.T, nftAddr, nxtAddr flow.Address, nxtSigner crypto.Signer, typeID uint64) {
	tx := flow.NewTransaction().
		SetScript(nxtGenerateMintKittyItemScript(nftAddr.String(), nxtAddr.String())).
		SetGasLimit(100).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		AddAuthorizer(nxtAddr)
	tx.AddArgument(cadence.NewAddress(nxtAddr))
	tx.AddArgument(cadence.NewUInt64(typeID))

	signAndSubmit(
		t, b, tx,
		[]flow.Address{b.ServiceKey().Address, nxtAddr},
		[]crypto.Signer{b.ServiceKey().Signer(), nxtSigner},
		false,
	)
}

func NxtTransferItem(b *emulator.Blockchain, t *testing.T, nftAddr, nxtAddr flow.Address, nxtSigner crypto.Signer, typeID uint64, recipientAddr flow.Address, shouldFail bool) {
	tx := flow.NewTransaction().
		SetScript(nxtGenerateTransferKittyItemScript(nftAddr.String(), nxtAddr.String())).
		SetGasLimit(100).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		AddAuthorizer(nxtAddr)
	tx.AddArgument(cadence.NewAddress(recipientAddr))
	tx.AddArgument(cadence.NewUInt64(typeID))

	signAndSubmit(
		t, b, tx,
		[]flow.Address{b.ServiceKey().Address, nxtAddr},
		[]crypto.Signer{b.ServiceKey().Signer(), nxtSigner},
		shouldFail,
	)
}

func TestNxtDeployContracts(t *testing.T) {
	b := newEmulator()
	NxtDeployContracts(b, t)
}

func TestCreateKittyItem(t *testing.T) {
	b := newEmulator()

	nftAddr, nxtAddr, nxtSigner := NxtDeployContracts(b, t)

	supply := executeScriptAndCheck(t, b, nxtGenerateInspectKittyItemSupplyScript(nftAddr.String(), nxtAddr.String()), nil)
	assert.Equal(t, cadence.NewUInt64(0), supply.(cadence.UInt64))

	len := executeScriptAndCheck(
		t,
		b,
		nxtGenerateInspectCollectionLenScript(nftAddr.String(), nxtAddr.String()),
		[][]byte{jsoncdc.MustEncode(cadence.NewAddress(nxtAddr))},
	)
	assert.Equal(t, cadence.NewInt(0), len.(cadence.Int))

	t.Run("Should be able to mint a nxt", func(t *testing.T) {
		NxtMintItem(b, t, nftAddr, nxtAddr, nxtSigner, typeID1)

		// Assert that the account's collection is correct
		len := executeScriptAndCheck(
			t,
			b,
			nxtGenerateInspectCollectionLenScript(nftAddr.String(), nxtAddr.String()),
			[][]byte{jsoncdc.MustEncode(cadence.NewAddress(nxtAddr))},
		)
		assert.Equal(t, cadence.NewInt(1), len.(cadence.Int))

		// Assert that the token type is correct
		/*typeID := executeScriptAndCheck(
			t,
			b,
			nxtGenerateInspectKittyItemTypeIDScript(nftAddr.String(), nxtAddr.String()),
			// Cheat: We know it's token ID 0
			[][]byte{jsoncdc.MustEncode(cadence.NewUInt64(0))},
		)
		assert.Equal(t, cadence.NewUInt64(typeID1), typeID.(cadence.UInt64))*/
	})

	/*t.Run("Shouldn't be able to borrow a reference to an NFT that doesn't exist", func(t *testing.T) {
		// Assert that the account's collection is correct
		result, err := b.ExecuteScript(nxtGenerateInspectCollectionScript(nftAddr, nxtAddr, nxtAddr, "Nxt", "NxtCollection", 5), nil)
		require.NoError(t, err)
		assert.True(t, result.Reverted())
	})*/
}

func TestTransferNFT(t *testing.T) {
	b := newEmulator()

	nftAddr, nxtAddr, nxtSigner := NxtDeployContracts(b, t)

	userAddress, userSigner, _ := createAccount(t, b)

	// create a new Collection
	t.Run("Should be able to create a new empty NFT Collection", func(t *testing.T) {
		NxtSetupAccount(t, b, userAddress, userSigner, nftAddr, nxtAddr)

		len := executeScriptAndCheck(
			t,
			b, nxtGenerateInspectCollectionLenScript(nftAddr.String(), nxtAddr.String()),
			[][]byte{jsoncdc.MustEncode(cadence.NewAddress(userAddress))},
		)
		assert.Equal(t, cadence.NewInt(0), len.(cadence.Int))

	})

	t.Run("Shouldn't be able to withdraw an NFT that doesn't exist in a collection", func(t *testing.T) {
		NxtTransferItem(b, t, nftAddr, nxtAddr, nxtSigner, 3333333, userAddress, true)

		//executeScriptAndCheck(t, b, nxtGenerateInspectCollectionLenScript(nftAddr, nxtAddr, userAddress, "Nxt", "NxtCollection", 0))

		// Assert that the account's collection is correct
		//executeScriptAndCheck(t, b, nxtGenerateInspectCollectionLenScript(nftAddr, nxtAddr, nxtAddr, "Nxt", "NxtCollection", 1))
	})

	// transfer an NFT
	t.Run("Should be able to withdraw an NFT and deposit to another accounts collection", func(t *testing.T) {
		NxtMintItem(b, t, nftAddr, nxtAddr, nxtSigner, typeID1)
		// Cheat: we have minted one item, its ID will be zero
		NxtTransferItem(b, t, nftAddr, nxtAddr, nxtSigner, 0, userAddress, false)

		// Assert that the account's collection is correct
		//executeScriptAndCheck(t, b, nxtGenerateInspectCollectionScript(nftAddr, nxtAddr, userAddress, "Nxt", "NxtCollection", 0))

		//executeScriptAndCheck(t, b, nxtGenerateInspectCollectionLenScript(nftAddr, nxtAddr, userAddress, "Nxt", "NxtCollection", 1))

		// Assert that the account's collection is correct
		//executeScriptAndCheck(t, b, nxtGenerateInspectCollectionLenScript(nftAddr, nxtAddr, nxtAddr, "Nxt", "NxtCollection", 0))
	})

	// transfer an NFT
	/*t.Run("Should be able to withdraw an NFT and destroy it, not reducing the supply", func(t *testing.T) {
		tx := flow.NewTransaction().
			SetScript(nxtGenerateDestroyScript(nftAddr, nxtAddr, "Nxt", "NxtCollection", 0)).
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

		executeScriptAndCheck(t, b, nxtGenerateInspectCollectionLenScript(nftAddr, nxtAddr, userAddress, "Nxt", "NxtCollection", 0))

		// Assert that the account's collection is correct
		executeScriptAndCheck(t, b, nxtGenerateInspectCollectionLenScript(nftAddr, nxtAddr, nxtAddr, "Nxt", "NxtCollection", 0))

		executeScriptAndCheck(t, b, nxtGenerateInspectNFTSupplyScript(nftAddr, nxtAddr, "Nxt", 1))

	})*/
}

func replaceNxtAddressPlaceholders(code, nftAddress, nxtAddress string) []byte {
	return []byte(replaceStrings(
		code,
		map[string]string{
			nftAddressPlaceholder:        "0x" + nftAddress,
			nxtAddressPlaceHolder: "0x" + nxtAddress,
		},
	))
}

func loadNonFungibleToken() []byte {
	return nft_contracts.NonFungibleToken()
}

func loadNxt(nftAddr string) []byte {
	return []byte(strings.ReplaceAll(
		string(readFile(nxtNxtPath)),
		nftAddressPlaceholder,
		"0x"+nftAddr,
	))
}

func nxtGenerateSetupAccountScript(nftAddr, nxtAddr string) []byte {
	return replaceNxtAddressPlaceholders(
		string(readFile(nxtSetupAccountPath)),
		nftAddr,
		nxtAddr,
	)
}

func nxtGenerateMintKittyItemScript(nftAddr, nxtAddr string) []byte {
	return replaceNxtAddressPlaceholders(
		string(readFile(nxtMintKittyItemPath)),
		nftAddr,
		nxtAddr,
	)
}

func nxtGenerateTransferKittyItemScript(nftAddr, nxtAddr string) []byte {
	return replaceNxtAddressPlaceholders(
		string(readFile(nxtTransferKittyItemPath)),
		nftAddr,
		nxtAddr,
	)
}

func nxtGenerateInspectKittyItemSupplyScript(nftAddr, nxtAddr string) []byte {
	return replaceNxtAddressPlaceholders(
		string(readFile(nxtInspectKittyItemSupplyPath)),
		nftAddr,
		nxtAddr,
	)
}

func nxtGenerateInspectCollectionLenScript(nftAddr, nxtAddr string) []byte {
	return replaceNxtAddressPlaceholders(
		string(readFile(nxtInspectCollectionLenPath)),
		nftAddr,
		nxtAddr,
	)
}

func nxtGenerateInspectCollectionIdsScript(nftAddr, nxtAddr string) []byte {
	return replaceNxtAddressPlaceholders(
		string(readFile(nxtInspectCollectionIdsPath)),
		nftAddr,
		nxtAddr,
	)
}
