import Dropn from "../../contracts/Dropn.cdc"
import FungibleToken from "../../contracts/FungibleToken.cdc"

// This script returns an account's Dropn balance.

pub fun main(address: Address): UFix64 {
    let account = getAccount(address)
    
    let vaultRef = account.getCapability(Dropn.BalancePublicPath)!.borrow<&Dropn.Vault{FungibleToken.Balance}>()
        ?? panic("Could not borrow Balance reference to the Vault")

    return vaultRef.balance
}
