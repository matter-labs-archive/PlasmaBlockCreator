package common

const (
	UTXOexistsButNotSpendable = byte(0x00)
	UTXOisReadyForSpending    = byte(0x01)
)

var (
	DepositIndexPrefix     = []byte("deposit")
	UtxoIndexPrefix        = []byte("utxo")
	TransactionIndexPrefix = []byte("ctr")
	BlockNumberKey         = []byte("blockNumber")
	TransactionNumberKey   = []byte("txNumber")
	SpendingIndexKey       = []byte("spend")
)
