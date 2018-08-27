package workers

import (
	"github.com/matterinc/PlasmaCommons/transaction"
	"github.com/matterinc/PlasmaCommons/types"
	"github.com/ethereum/go-ethereum/common"
)

type UTXOReader interface {
	CheckIfUTXOsExist(tx *transaction.SignedTransaction) error
}

type UTXOWriter interface {
	WriteTX(tx *transaction.SignedTransaction, counter int64) error
}

type FundingTXCreator interface {
	CreateFundingTX(to common.Address,
		amount *types.BigInt,
		counter int64,
		depositIndex *types.BigInt) error
}

type UTXOInserter interface {
	InsertUTXO(tx *transaction.NumberedTransaction, blockNumber uint32) error
}
