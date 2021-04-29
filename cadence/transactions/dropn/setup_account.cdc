import FungibleToken from "../../contracts/FungibleToken.cdc"
import Dropn from "../../contracts/Dropn.cdc"

// This transaction is a template for a transaction
// to add a Vault resource to their account
// so that they can use the Dropn

transaction {

    prepare(signer: AuthAccount) {

        if signer.borrow<&Dropn.Vault>(from: Dropn.VaultStoragePath) == nil {
            // Create a new Dropn Vault and put it in storage
            signer.save(<-Dropn.createEmptyVault(), to: Dropn.VaultStoragePath)

            // Create a public capability to the Vault that only exposes
            // the deposit function through the Receiver interface
            signer.link<&Dropn.Vault{FungibleToken.Receiver}>(
                Dropn.ReceiverPublicPath,
                target: Dropn.VaultStoragePath
            )

            // Create a public capability to the Vault that only exposes
            // the balance field through the Balance interface
            signer.link<&Dropn.Vault{FungibleToken.Balance}>(
                Dropn.BalancePublicPath,
                target: Dropn.VaultStoragePath
            )
        }
    }
}
