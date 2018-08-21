package foundationdb

import (
	"errors"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	common "github.com/ethereum/go-ethereum/common"
	commonConst "github.com/shamatar/go-plasma/common"
	"github.com/shamatar/go-plasma/transaction"
	types "github.com/shamatar/go-plasma/types"
)

type WithdrawTXMarker struct {
	db     *fdb.Database
	lister *UTXOlister
}

func NewWithdrawTXMarker(db *fdb.Database) *WithdrawTXMarker {
	lister := NewUTXOlister(db)
	marker := &WithdrawTXMarker{db: db, lister: lister}
	return marker
}

func (r *WithdrawTXMarker) MarkTX(to common.Address,
	index *types.BigInt) (bool, error) {
	details, err := transaction.ParseUTXOindexNumberIntoDetails(index)
	if err != nil {
		return false, err
	}
	existingUTXO, err := r.lister.GetUTXOsForAddress(to, details.BlockNumber, details.TransactionNumber, details.OutputNumber, 1)
	if err != nil {
		return false, err
	}
	if len(existingUTXO) != 1 {
		return false, nil
	}
	utxoIndex := []byte{}
	utxoIndex = append(utxoIndex, commonConst.UtxoIndexPrefix...)
	utxoIndex = append(utxoIndex, existingUTXO[0][:]...)
	_, err = r.db.Transact(func(tr fdb.Transaction) (interface{}, error) {
		existing := tr.Get(fdb.Key(utxoIndex)).MustGet()
		if len(existing) != 1 {
			return nil, errors.New("Invalid UTXO state")
		}
		if existing[0] == commonConst.UTXOexistsButNotSpendable {
			return nil, nil
		}
		if existing[0] != commonConst.UTXOisReadyForSpending {
			return nil, errors.New("Invalid UTXO state")
		}
		tr.Set(fdb.Key(utxoIndex), []byte{commonConst.UTXOexistsButNotSpendable})
		existing = tr.Get(fdb.Key(utxoIndex)).MustGet()
		if len(existing) != 1 {
			return nil, errors.New("Invalid UTXO state")
		}
		if existing[0] == commonConst.UTXOexistsButNotSpendable {
			return nil, nil
		}
		if existing[0] != commonConst.UTXOisReadyForSpending {
			tr.Reset()
			return nil, errors.New("Invalid UTXO state")
		}
		return nil, nil
	})
	if err != nil {
		return false, err
	}
	return true, nil
}
