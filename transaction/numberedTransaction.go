package transaction

import (
	"encoding/binary"
	"errors"
	"io"
	// "github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/ethereum/go-ethereum/rlp"
)

// TransactionInput is one of the inputs into Plasma transaction
type NumberedTransaction struct {
	TransactionNumber [TransactionNumberLength]byte
	SignedTransaction *SignedTransaction
}

type rlpNumberedTransaction struct {
	TransactionNumber []byte
	SignedTransaction SignedTransaction
}

func NewNumberedTransaction(signedTX *SignedTransaction, transactionNumber uint32) (*NumberedTransaction, error) {
	tx := &NumberedTransaction{}
	bs := make([]byte, TransactionNumberLength)
	binary.BigEndian.PutUint32(bs, transactionNumber)
	copy(tx.TransactionNumber[:], bs)
	tx.SignedTransaction = signedTX
	return tx, nil
}

// signature is [R || S || V]
func (tx *NumberedTransaction) Validate() error {
	err := tx.SignedTransaction.Validate()
	if err != nil {
		return err
	}
	return nil
}

// EncodeRLP implements rlp.Encoder, and flattens the consensus fields of a receipt
// into an RLP stream. If no post state is present, byzantium fork is assumed.
func (tx *NumberedTransaction) EncodeRLP(w io.Writer) error {
	rlpNumbered := rlpNumberedTransaction{tx.TransactionNumber[:], *tx.SignedTransaction}
	return rlp.Encode(w, rlpNumbered)
}

// DecodeRLP implements rlp.Decoder, and loads the consensus fields of a receipt
// from an RLP stream.
func (tx *NumberedTransaction) DecodeRLP(s *rlp.Stream) error {
	var dec rlpNumberedTransaction
	if err := s.Decode(&dec); err != nil {
		return err
	}
	if len(dec.TransactionNumber) != TransactionNumberLength {
		return errors.New("Invalid transaction number length")
	}
	tx.SignedTransaction = &dec.SignedTransaction
	copy(tx.TransactionNumber[:], dec.TransactionNumber)
	return nil
}
