import NonFungibleToken from "../../contracts/NonFungibleToken.cdc"
import Nxt from "../../contracts/Nxt.cdc"

// This transaction configures an account to hold Kitty Items.

transaction {
    prepare(signer: AuthAccount) {
        // if the account doesn't already have a collection
        if signer.borrow<&Nxt.Collection>(from: Nxt.CollectionStoragePath) == nil {

            // create a new empty collection
            let collection <- Nxt.createEmptyCollection()
            
            // save it to the account
            signer.save(<-collection, to: Nxt.CollectionStoragePath)

            // create a public capability for the collection
            signer.link<&Nxt.Collection{NonFungibleToken.CollectionPublic, Nxt.NxtCollectionPublic}>(Nxt.CollectionPublicPath, target: Nxt.CollectionStoragePath)
        }
    }
}
