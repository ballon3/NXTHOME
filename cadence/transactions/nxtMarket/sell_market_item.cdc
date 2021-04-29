import FungibleToken from "../../contracts/FungibleToken.cdc"
import NonFungibleToken from "../../contracts/NonFungibleToken.cdc"
import Dropn from "../../contracts/Dropn.cdc"
import Nxt from "../../contracts/Nxt.cdc"
import NxtMarket from "../../contracts/NxtMarket.cdc"

transaction(itemID: UInt64, price: UFix64) {
    let dropnVault: Capability<&Dropn.Vault{FungibleToken.Receiver}>
    let nxtCollection: Capability<&Nxt.Collection{NonFungibleToken.Provider, Nxt.NxtCollectionPublic}>
    let marketCollection: &NxtMarket.Collection

    prepare(signer: AuthAccount) {
        // we need a provider capability, but one is not provided by default so we create one.
        let NxtCollectionProviderPrivatePath = /private/nxtCollectionProvider

        self.dropnVault = signer.getCapability<&Dropn.Vault{FungibleToken.Receiver}>(Dropn.ReceiverPublicPath)!
        assert(self.dropnVault.borrow() != nil, message: "Missing or mis-typed Dropn receiver")

        if !signer.getCapability<&Nxt.Collection{NonFungibleToken.Provider, Nxt.NxtCollectionPublic}>(NxtCollectionProviderPrivatePath)!.check() {
            signer.link<&Nxt.Collection{NonFungibleToken.Provider, Nxt.NxtCollectionPublic}>(NxtCollectionProviderPrivatePath, target: Nxt.CollectionStoragePath)
        }

        self.nxtCollection = signer.getCapability<&Nxt.Collection{NonFungibleToken.Provider, Nxt.NxtCollectionPublic}>(NxtCollectionProviderPrivatePath)!
        assert(self.nxtCollection.borrow() != nil, message: "Missing or mis-typed NxtCollection provider")

        self.marketCollection = signer.borrow<&NxtMarket.Collection>(from: NxtMarket.CollectionStoragePath)
            ?? panic("Missing or mis-typed NxtMarket Collection")
    }

    execute {
        let offer <- NxtMarket.createSaleOffer (
            sellerItemProvider: self.nxtCollection,
            itemID: itemID,
            typeID: self.nxtCollection.borrow()!.borrowKittyItem(id: itemID)!.typeID,
            sellerPaymentReceiver: self.dropnVault,
            price: price
        )
        self.marketCollection.insert(offer: <-offer)
    }
}
