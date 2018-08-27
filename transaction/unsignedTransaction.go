package transaction

import (
	"bytes"
	"errors"
	"io"
	// "github.com/ethereum/go-ethereum/common"
	// "github.com/ethereum/go-ethereum/common/hexutil"

	plasmaCommon "github.com/matterinc/PlasmaCommons/common"
	"github.com/matterinc/PlasmaCommons/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

// TransactionInput is one of the inputs into Plasma transaction
type UnsignedTransaction struct {
	TransactionType [TransactionTypeLength]byte
	Inputs          []*TransactionInput
	Outputs         []*TransactionOutput
}

type rlpUnsignedTransaction struct {
	TransactionType []byte
	Inputs          []TransactionInput
	Outputs         []TransactionOutput
}

func NewUnsignedTransaction(txType byte, inputs []*TransactionInput, outputs []*TransactionOutput) (*UnsignedTransaction, error) {
	tx := &UnsignedTransaction{}
	switch txType {
	case TransactionTypeSplit:
		tx.TransactionType = [TransactionTypeLength]byte{txType}
	case TransactionTypeMerge:
		tx.TransactionType = [TransactionTypeLength]byte{txType}
	case TransactionTypeFund:
		tx.TransactionType = [TransactionTypeLength]byte{txType}
	default:
		return nil, errors.New("Invalid transaction type")
	}
	if len(inputs) == 0 || len(outputs) == 0 {
		return nil, errors.New("Empty inputs or outputs")
	}
	tx.Inputs = inputs
	tx.Outputs = outputs
	return tx, nil
}

func (tx *UnsignedTransaction) Validate() error {
	numInputs := len(tx.Inputs)
	numOutput := len(tx.Outputs)
	if numInputs == 0 || numOutput == 0 {
		return errors.New("Empty inputs or outputs")
	}
	switch tx.TransactionType[0] {
	case TransactionTypeSplit:
		if numInputs != 1 || numOutput > 3 {
			return errors.New("Invalid number of inputs or outputs")
		}
	case TransactionTypeMerge:
		if numInputs != 2 || numOutput != 1 {
			return errors.New("Invalid number of inputs or outputs")
		}
	case TransactionTypeFund:
		if numInputs != 1 || numOutput != 1 {
			return errors.New("Invalid number of inputs or outputs")
		}
	default:
		return errors.New("Invalid transaction type")
	}

	if tx.TransactionType[0] != TransactionTypeFund {
		totalIn := types.NewBigInt(0)
		totalOut := types.NewBigInt(0)
		for _, input := range tx.Inputs {
			totalIn.Bigint.Add(totalIn.Bigint, input.GetValue().Bigint)
		}
		for idx, output := range tx.Outputs {
			if int(output.OutputNumber[0]) != idx {
				return errors.New("Invalid output numbering")
			}
			totalOut.Bigint.Add(totalOut.Bigint, output.GetValue().Bigint)
		}
		if totalIn.Bigint.Cmp(totalOut.Bigint) != 0 {
			return errors.New("Inputs value is not equal to outputs value")
		}
	} else {
		if tx.Inputs[0].GetReferedUTXO().Bigint.Cmp(types.NewBigInt(0).Bigint) != 0 {
			return errors.New("Invalid funding transaction input")
		}
		if int(tx.Outputs[0].OutputNumber[0]) != 0 {
			return errors.New("Invalid output numbering")
		}
	}
	return nil
}

// EncodeRLP implements rlp.Encoder, and flattens the consensus fields of a receipt
// into an RLP stream. If no post state is present, byzantium fork is assumed.
func (tx *UnsignedTransaction) EncodeRLP(w io.Writer) error {
	rlpTX := rlpUnsignedTransaction{}
	rlpTX.TransactionType = tx.TransactionType[:]
	rlpTX.Inputs = make([]TransactionInput, len(tx.Inputs))
	rlpTX.Outputs = make([]TransactionOutput, len(tx.Outputs))
	for i := 0; i < len(tx.Inputs); i++ {
		rlpTX.Inputs[i] = *(tx.Inputs[i])
	}
	for i := 0; i < len(tx.Outputs); i++ {
		rlpTX.Outputs[i] = *(tx.Outputs[i])
	}
	return rlp.Encode(w, rlpTX)
}

// DecodeRLP implements rlp.Decoder, and loads the consensus fields of a receipt
// from an RLP stream.
func (tx *UnsignedTransaction) DecodeRLP(s *rlp.Stream) error {
	var dec rlpUnsignedTransaction
	if err := s.Decode(&dec); err != nil {
		return err
	}
	if len(dec.TransactionType) != TransactionTypeLength {
		return errors.New("Transaction type length is invalid")
	}
	switch dec.TransactionType[0] {
	case TransactionTypeSplit:
		copy(tx.TransactionType[:], dec.TransactionType)
	case TransactionTypeMerge:
		copy(tx.TransactionType[:], dec.TransactionType)
	case TransactionTypeFund:
		copy(tx.TransactionType[:], dec.TransactionType)
	default:
		return errors.New("Invalid transaction type")
	}
	tx.Inputs = make([]*TransactionInput, len(dec.Inputs))
	tx.Outputs = make([]*TransactionOutput, len(dec.Outputs))
	for i := 0; i < len(dec.Inputs); i++ {
		tx.Inputs[i] = &(dec.Inputs[i])
	}
	for i := 0; i < len(dec.Outputs); i++ {
		tx.Outputs[i] = &(dec.Outputs[i])
	}
	return nil
}

func (tx *UnsignedTransaction) GetHash() (common.Hash, error) {
	var b bytes.Buffer
	i := io.Writer(&b)
	err := tx.EncodeRLP(i)
	if err != nil {
		return common.Hash{}, err
	}
	encoding := b.Bytes()
	hash := plasmaCommon.CreatePersonalHash(encoding)
	return hash, nil
}
