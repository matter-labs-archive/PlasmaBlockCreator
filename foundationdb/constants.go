package foundationdb

const (
	UTXOexistsButNotSpendable = byte{0x00}
	UTXOisReadyForSpending    = byte{0x01}
	depositIndexPrefix        = []byte{"deposit"}
	utxoIndexPrefix           = []byte{"utxo"}
	transactionIndexPrefix    = []byte{"ctr"}
)
