import NxtDropContract from 0xf8d6e0586b0a20c7

transaction {
  let receiverRef: &{NxtDropContract.NFTReceiver}
  let minterRef: &NxtDropContract.NFTMinter

  prepare(acct: AuthAccount) {
      self.receiverRef = acct.getCapability<&{NxtDropContract.NFTReceiver}>(/public/NFTReceiver)
          .borrow()
          ?? panic("Could not borrow receiver reference")        
      
      self.minterRef = acct.borrow<&NxtDropContract.NFTMinter>(from: /storage/NFTMinter)
          ?? panic("could not borrow minter reference")
  }

  execute {
      let metadata : {String : String} = {
          "name": "Way97",
          "athlete": "Danny Way", 
          "location": "California, USA",
          "date":"dec, 1997",
          "uri": "ipfs://Qmetdo9kn3LzemMTkB9DSM5VjpvxLeGWWqgvN1Q7nithas"
      }
      let newNFT <- self.minterRef.mintNFT()
  
      self.receiverRef.deposit(token: <-newNFT, metadata: metadata)

      log("NFT Minted and deposited to Account 2's Collection")
  }
}
 