package types

type TxType string

const (
	EOA    TxType = "EOA"
	ERC20  TxType = "ERC20"
	ERC721 TxType = "ERC721"
)
