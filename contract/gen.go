package contract

// Generate Token contract bindings
//go:generate abigen --bin=./abi/ZexCoinERC20.bin --abi=./abi/ZexCoinERC20.abi --pkg erc20 --out=erc20/erc20.go

//go:generate abigen --bin=./abi/ZexNFTs.bin --abi=./abi/ZexNFTs.abi --pkg erc721 --out=erc721/erc721.go
