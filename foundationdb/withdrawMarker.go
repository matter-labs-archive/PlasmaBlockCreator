package foundationdb

import (
	"errors"

	fdb "github.com/apple/foundationdb/bindings/go/src/fdb"
	common "github.com/ethereum/go-ethereum/common"
	commonConst "github.com/matterinc/PlasmaCommons/common"
	"github.com/matterinc/PlasmaCommons/transaction"
	types "github.com/matterinc/PlasmaCommons/types"
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
	// check for unspent one
	existingUTXO, err := r.lister.GetExactUTXOsForAddress(to, details.BlockNumber, details.TransactionNumber, details.OutputNumber, 1, false)
	if err != nil {
		return false, err
	}
	if len(existingUTXO) != 1 {
		// looks like there is no unspent, so test for a spent one
		existingUTXO, err = r.lister.GetExactUTXOsForAddress(to, details.BlockNumber, details.TransactionNumber, details.OutputNumber, 1, true)
		if err != nil {
			return false, err
		}
		if len(existingUTXO) != 1 {
			return false, nil
			// definatelly was spent!
		}
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
