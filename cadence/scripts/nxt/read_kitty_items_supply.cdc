import Nxt from "../../contracts/Nxt.cdc"

// This scripts returns the number of Nxt currently in existence.

pub fun main(): UInt64 {    
    return Nxt.totalSupply
}
