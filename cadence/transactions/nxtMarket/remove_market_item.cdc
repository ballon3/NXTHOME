import NxtMarket from "../../contracts/NxtMarket.cdc"

transaction(itemID: UInt64) {
    let marketCollection: &NxtMarket.Collection

    prepare(signer: AuthAccount) {
        self.marketCollection = signer.borrow<&NxtMarket.Collection>(from: NxtMarket.CollectionStoragePath)
            ?? panic("Missing or mis-typed NxtMarket Collection")
    }

    execute {
        let offer <-self.marketCollection.remove(itemID: itemID)
        destroy offer
    }
}
