import Dropn from "../../contracts/Dropn.cdc"

// This script returns the total amount of Dropn currently in existence.

pub fun main(): UFix64 {

    let supply = Dropn.totalSupply

    log(supply)

    return supply
}
