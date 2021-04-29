package test

import (
	"strings"
	"testing"

	"github.com/onflow/cadence"
	emulator "github.com/onflow/flow-emulator"
	"github.com/onflow/flow-go-sdk"
	sdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	sdktemplates "github.com/onflow/flow-go-sdk/templates"
	"github.com/onflow/flow-go-sdk/test"
	"github.com/stretchr/testify/assert"
)

const (
	nxtMarketRootPath             = "../../../cadence/NxtMarket"
	nxtMarketNxtMarketPath = nxtMarketRootPath + "/contracts/NxtMarket.cdc"
	nxtMarketSetupAccountPath     = nxtMarketRootPath + "/transactions/setup_account.cdc"
	nxtMarketSellItemPath         = nxtMarketRootPath + "/transactions/sell_market_item.cdc"
	nxtMarketBuyItemPath          = nxtMarketRootPath + "/transactions/buy_market_item.cdc"
	nxtMarketRemoveItemPath       = nxtMarketRootPath + "/transactions/remove_market_item.cdc"
)

const (
	typeID1337 = 1337
)

type TestContractsInfo struct {
	FTAddr                 flow.Address
	DropnAddr             flow.Address
	DropnSigner           crypto.Signer
	NFTAddr                flow.Address
	NxtAddr         flow.Address
	NxtSigner       crypto.Signer
	NxtMarketAddr   flow.Address
	NxtMarketSigner crypto.Signer
}

func NxtMarketDeployContracts(b *emulator.Blockchain, t *testing.T) TestContractsInfo {
	accountKeys := test.AccountKeyGenerator()

	ftAddr, dropnAddr, dropnSigner := DropnDeployContracts(b, t)
	nftAddr, nxtAddr, nxtSigner := NxtDeployContracts(b, t)

	// Should be able to deploy a contract as a new account with one key.
	nxtMarketAccountKey, nxtMarketSigner := accountKeys.NewWithSigner()
	nxtMarketCode := loadNxtMarket(
		ftAddr.String(),
		nftAddr.String(),
		dropnAddr.String(),
		nxtAddr.String(),
	)
	nxtMarketAddr, err := b.CreateAccount(
		[]*flow.AccountKey{nxtMarketAccountKey},
		[]sdktemplates.Contract{
			{
				Name:   "NxtMarket",
				Source: string(nxtMarketCode),
			},
		})
	if !assert.NoError(t, err) {
		t.Log(err.Error())
	}
	_, err = b.CommitBlock()
	assert.NoError(t, err)

	// Simplify the workflow by having contract addresses also be our initial test collections.
	NxtSetupAccount(t, b, nxtAddr, nxtSigner, nftAddr, nxtAddr)
	NxtMarketSetupAccount(b, t, nxtMarketAddr, nxtMarketSigner, nxtMarketAddr)

	return TestContractsInfo{
		ftAddr,
		dropnAddr,
		dropnSigner,
		nftAddr,
		nxtAddr,
		nxtSigner,
		nxtMarketAddr,
		nxtMarketSigner,
	}
}

func NxtMarketSetupAccount(b *emulator.Blockchain, t *testing.T, userAddress sdk.Address, userSigner crypto.Signer, nxtMarketAddr sdk.Address) {
	tx := flow.NewTransaction().
		SetScript(nxtMarketGenerateSetupAccountScript(nxtMarketAddr.String())).
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

// Create a new account with the Dropn and Nxt resources set up BUT no NxtMarket resource.
func NxtMarketCreatePurchaserAccount(b *emulator.Blockchain, t *testing.T, contracts TestContractsInfo) (sdk.Address, crypto.Signer) {
	userAddress, userSigner, _ := createAccount(t, b)
	DropnSetupAccount(t, b, userAddress, userSigner, contracts.FTAddr, contracts.DropnAddr)
	NxtSetupAccount(t, b, userAddress, userSigner, contracts.NFTAddr, contracts.NxtAddr)
	return userAddress, userSigner
}

// Create a new account with the Dropn, Nxt, and NxtMarket resources set up.
func NxtMarketCreateAccount(b *emulator.Blockchain, t *testing.T, contracts TestContractsInfo) (sdk.Address, crypto.Signer) {
	userAddress, userSigner := NxtMarketCreatePurchaserAccount(b, t, contracts)
	NxtMarketSetupAccount(b, t, userAddress, userSigner, contracts.NxtMarketAddr)
	return userAddress, userSigner
}

func NxtMarketListItem(b *emulator.Blockchain, t *testing.T, contracts TestContractsInfo, userAddress sdk.Address, userSigner crypto.Signer, tokenID uint64, price string, shouldFail bool) {
	tx := flow.NewTransaction().
		SetScript(nxtMarketGenerateSellItemScript(contracts)).
		SetGasLimit(100).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		AddAuthorizer(userAddress)
	tx.AddArgument(cadence.NewUInt64(tokenID))
	tx.AddArgument(CadenceUFix64(price))

	signAndSubmit(
		t, b, tx,
		[]flow.Address{b.ServiceKey().Address, userAddress},
		[]crypto.Signer{b.ServiceKey().Signer(), userSigner},
		shouldFail,
	)
}

func NxtMarketPurchaseItem(
	b *emulator.Blockchain,
	t *testing.T,
	contracts TestContractsInfo,
	userAddress sdk.Address,
	userSigner crypto.Signer,
	marketCollectionAddress sdk.Address,
	tokenID uint64,
	shouldFail bool,
) {
	tx := flow.NewTransaction().
		SetScript(nxtMarketGenerateBuyItemScript(contracts)).
		SetGasLimit(100).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		AddAuthorizer(userAddress)
	tx.AddArgument(cadence.NewUInt64(tokenID))
	tx.AddArgument(cadence.NewAddress(marketCollectionAddress))

	signAndSubmit(
		t, b, tx,
		[]flow.Address{b.ServiceKey().Address, userAddress},
		[]crypto.Signer{b.ServiceKey().Signer(), userSigner},
		shouldFail,
	)
}

func NxtMarketRemoveItem(
	b *emulator.Blockchain,
	t *testing.T,
	contracts TestContractsInfo,
	userAddress sdk.Address,
	userSigner crypto.Signer,
	marketCollectionAddress sdk.Address,
	tokenID uint64,
	shouldFail bool,
) {
	tx := flow.NewTransaction().
		SetScript(nxtMarketGenerateRemoveItemScript(contracts)).
		SetGasLimit(100).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		AddAuthorizer(userAddress)
	tx.AddArgument(cadence.NewUInt64(tokenID))

	signAndSubmit(
		t, b, tx,
		[]flow.Address{b.ServiceKey().Address, userAddress},
		[]crypto.Signer{b.ServiceKey().Signer(), userSigner},
		shouldFail,
	)
}

func TestNxtMarketDeployContracts(t *testing.T) {
	b := newEmulator()
	NxtMarketDeployContracts(b, t)
}

func TestNxtMarketSetupAccount(t *testing.T) {
	b := newEmulator()

	contracts := NxtMarketDeployContracts(b, t)

	t.Run("Should be able to create an empty Collection", func(t *testing.T) {
		userAddress, userSigner, _ := createAccount(t, b)
		NxtMarketSetupAccount(b, t, userAddress, userSigner, contracts.NxtMarketAddr)
	})
}

func TestNxtMarketCreateSaleOffer(t *testing.T) {
	b := newEmulator()

	contracts := NxtMarketDeployContracts(b, t)

	t.Run("Should be able to create a sale offer and list it", func(t *testing.T) {
		tokenToList := uint64(0)
		tokenPrice := "1.11"
		userAddress, userSigner := NxtMarketCreateAccount(b, t, contracts)
		// Contract mints item
		NxtMintItem(
			b,
			t,
			contracts.NFTAddr,
			contracts.NxtAddr,
			contracts.NxtSigner,
			typeID1337,
		)
		// Contract transfers item to another seller account (we don't need to do this)
		NxtTransferItem(
			b,
			t,
			contracts.NFTAddr,
			contracts.NxtAddr,
			contracts.NxtSigner,
			tokenToList,
			userAddress,
			false,
		)
		// Other seller account lists the item
		NxtMarketListItem(
			b,
			t,
			contracts,
			userAddress,
			userSigner,
			tokenToList,
			tokenPrice,
			false,
		)
	})

	t.Run("Should be able to accept a sale offer", func(t *testing.T) {
		tokenToList := uint64(1)
		tokenPrice := "1.11"
		userAddress, userSigner := NxtMarketCreateAccount(b, t, contracts)
		// Contract mints item
		NxtMintItem(
			b,
			t,
			contracts.NFTAddr,
			contracts.NxtAddr,
			contracts.NxtSigner,
			typeID1337,
		)
		// Contract transfers item to another seller account (we don't need to do this)
		NxtTransferItem(
			b,
			t,
			contracts.NFTAddr,
			contracts.NxtAddr,
			contracts.NxtSigner,
			tokenToList,
			userAddress,
			false,
		)
		// Other seller account lists the item
		NxtMarketListItem(
			b,
			t,
			contracts,
			userAddress,
			userSigner,
			tokenToList,
			tokenPrice,
			false,
		)
		buyerAddress, buyerSigner := NxtMarketCreatePurchaserAccount(b, t, contracts)
		// Fund the purchase
		DropnMint(
			t,
			b,
			contracts.FTAddr,
			contracts.DropnAddr,
			contracts.DropnSigner,
			buyerAddress,
			"100.0",
			false,
		)
		// Make the purchase
		NxtMarketPurchaseItem(
			b,
			t,
			contracts,
			buyerAddress,
			buyerSigner,
			userAddress,
			tokenToList,
			false,
		)
	})

	t.Run("Should be able to remove a sale offer", func(t *testing.T) {
		tokenToList := uint64(2)
		tokenPrice := "1.11"
		userAddress, userSigner := NxtMarketCreateAccount(b, t, contracts)
		// Contract mints item
		NxtMintItem(
			b,
			t,
			contracts.NFTAddr,
			contracts.NxtAddr,
			contracts.NxtSigner,
			typeID1337,
		)
		// Contract transfers item to another seller account (we don't need to do this)
		NxtTransferItem(
			b,
			t,
			contracts.NFTAddr,
			contracts.NxtAddr,
			contracts.NxtSigner,
			tokenToList,
			userAddress,
			false,
		)
		// Other seller account lists the item
		NxtMarketListItem(
			b,
			t,
			contracts,
			userAddress,
			userSigner,
			tokenToList,
			tokenPrice,
			false,
		)
		// Make the purchase
		NxtMarketRemoveItem(
			b,
			t,
			contracts,
			userAddress,
			userSigner,
			userAddress,
			tokenToList,
			false,
		)
	})
}

func replaceNxtMarketAddressPlaceholders(codeBytes []byte, contracts TestContractsInfo) []byte {
	code := string(codeBytes)

	code = strings.ReplaceAll(code, ftAddressPlaceholder, "0x"+contracts.FTAddr.String())
	code = strings.ReplaceAll(code, dropnAddressPlaceHolder, "0x"+contracts.DropnAddr.String())
	code = strings.ReplaceAll(code, nftAddressPlaceholder, "0x"+contracts.NFTAddr.String())
	code = strings.ReplaceAll(code, nxtAddressPlaceHolder, "0x"+contracts.NxtAddr.String())
	code = strings.ReplaceAll(code, nxtMarketPlaceholder, "0x"+contracts.NxtMarketAddr.String())

	return []byte(code)
}

func loadNxtMarket(ftAddr, nftAddr, dropnAddr, nxtAddr string) []byte {
	code := string(readFile(nxtMarketNxtMarketPath))

	code = strings.ReplaceAll(code, ftAddressPlaceholder, "0x"+ftAddr)
	code = strings.ReplaceAll(code, dropnAddressPlaceHolder, "0x"+dropnAddr)
	code = strings.ReplaceAll(code, nftAddressPlaceholder, "0x"+nftAddr)
	code = strings.ReplaceAll(code, nxtAddressPlaceHolder, "0x"+nxtAddr)

	return []byte(code)
}

func nxtMarketGenerateSetupAccountScript(nxtMarketAddr string) []byte {
	code := string(readFile(nxtMarketSetupAccountPath))

	code = strings.ReplaceAll(code, nxtMarketPlaceholder, "0x"+nxtMarketAddr)

	return []byte(code)
}

func nxtMarketGenerateSellItemScript(contracts TestContractsInfo) []byte {
	return replaceNxtMarketAddressPlaceholders(
		readFile(nxtMarketSellItemPath),
		contracts,
	)
}

func nxtMarketGenerateBuyItemScript(contracts TestContractsInfo) []byte {
	return replaceNxtMarketAddressPlaceholders(
		readFile(nxtMarketBuyItemPath),
		contracts,
	)
}

func nxtMarketGenerateRemoveItemScript(contracts TestContractsInfo) []byte {
	return replaceNxtMarketAddressPlaceholders(
		readFile(nxtMarketRemoveItemPath),
		contracts,
	)
}
