import NxtMarket from "../../contracts/NxtMarket.cdc"

// This transaction configures an account to hold SaleOffer items.

transaction {
    prepare(signer: AuthAccount) {

        // if the account doesn't already have a collection
        if signer.borrow<&NxtMarket.Collection>(from: NxtMarket.CollectionStoragePath) == nil {

            // create a new empty collection
            let collection <- NxtMarket.createEmptyCollection() as! @NxtMarket.Collection
            
            // save it to the account
            signer.save(<-collection, to: NxtMarket.CollectionStoragePath)

            // create a public capability for the collection
            signer.link<&NxtMarket.Collection{NxtMarket.CollectionPublic}>(NxtMarket.CollectionPublicPath, target: NxtMarket.CollectionStoragePath)
        }
    }
}
