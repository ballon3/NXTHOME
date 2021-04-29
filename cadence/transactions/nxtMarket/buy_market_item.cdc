import FungibleToken from "../../contracts/FungibleToken.cdc"
import NonFungibleToken from "../../contracts/NonFungibleToken.cdc"
import Dropn from "../../contracts/Dropn.cdc"
import Nxt from "../../contracts/Nxt.cdc"
import NxtMarket from "../../contracts/NxtMarket.cdc"

transaction(itemID: UInt64, marketCollectionAddress: Address) {
    let paymentVault: @FungibleToken.Vault
    let nxtCollection: &Nxt.Collection{NonFungibleToken.Receiver}
    let marketCollection: &NxtMarket.Collection{NxtMarket.CollectionPublic}

    prepare(signer: AuthAccount) {
        self.marketCollection = getAccount(marketCollectionAddress)
            .getCapability<&NxtMarket.Collection{NxtMarket.CollectionPublic}>(
                NxtMarket.CollectionPublicPath
            )!
            .borrow()
            ?? panic("Could not borrow market collection from market address")

        let saleItem = self.marketCollection.borrowSaleItem(itemID: itemID)
                    ?? panic("No item with that ID")
        let price = saleItem.price

        let mainDropnVault = signer.borrow<&Dropn.Vault>(from: Dropn.VaultStoragePath)
            ?? panic("Cannot borrow Dropn vault from acct storage")
        self.paymentVault <- mainDropnVault.withdraw(amount: price)

        self.nxtCollection = signer.borrow<&Nxt.Collection{NonFungibleToken.Receiver}>(
            from: Nxt.CollectionStoragePath
        ) ?? panic("Cannot borrow Nxt collection receiver from acct")
    }

    execute {
        self.marketCollection.purchase(
            itemID: itemID,
            buyerCollection: self.nxtCollection,
            buyerPayment: <- self.paymentVault
        )
    }
}
