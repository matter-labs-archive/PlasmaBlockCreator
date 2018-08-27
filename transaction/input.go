package transaction

import (
	"errors"
	"io"
	// "github.com/ethereum/go-ethereum/common"
	// "github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/matterinc/PlasmaCommons/types"
	"github.com/ethereum/go-ethereum/rlp"
)

// TransactionInput is one of the inputs into Plasma transaction
type TransactionInput struct {
	BlockNumber       [BlockNumberLength]byte
	TransactionNumber [TransactionNumberLength]byte
	OutputNumber      [OutputNumberLength]byte
	Value             [ValueLength]byte
}

type rlpTransactionInput struct {
	BlockNumber       []byte
	TransactionNumber []byte
	OutputNumber      []byte
	Value             []byte
}

// SetFields creates a new input sturcture
func (input *TransactionInput) SetFields(blockNumber *types.BigInt, transactionNumber *types.BigInt, outputNumber *types.BigInt, value *types.BigInt) error {
	blockNumberBytes, err := blockNumber.GetLeftPaddedBytes(BlockNumberLength)
	if err != nil {
		return errors.New("Block number is too long")
	}
	transactionNumberBytes, err := transactionNumber.GetLeftPaddedBytes(TransactionNumberLength)
	if err != nil {
		return errors.New("Transaction number is too long")
	}
	outputNumberBytes, err := outputNumber.GetLeftPaddedBytes(OutputNumberLength)
	if err != nil {
		return errors.New("Output number is too long")
	}
	valueBytes, err := value.GetLeftPaddedBytes(ValueLength)
	if err != nil {
		return errors.New("Value is too long")
	}

	copy(input.BlockNumber[:], blockNumberBytes)
	copy(input.TransactionNumber[:], transactionNumberBytes)
	copy(input.OutputNumber[:], outputNumberBytes)
	copy(input.Value[:], valueBytes)
	return nil
}

// EncodeRLP implements rlp.Encoder, and flattens the consensus fields of a receipt
// into an RLP stream. If no post state is present, byzantium fork is assumed.
func (input *TransactionInput) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, rlpTransactionInput{input.BlockNumber[:], input.TransactionNumber[:], input.OutputNumber[:], input.Value[:]})
}

func (input *TransactionInput) GetValue() *types.BigInt {
	bi := types.NewBigInt(0)
	bi.SetBytes(input.Value[:])
	return bi
}

func (input *TransactionInput) GetReferedUTXO() *types.BigInt {
	bi := types.NewBigInt(0)
	fullIndex := []byte{}
	fullIndex = append(fullIndex, input.BlockNumber[:]...)
	fullIndex = append(fullIndex, input.TransactionNumber[:]...)
	fullIndex = append(fullIndex, input.OutputNumber[:]...)
	bi.SetBytes(fullIndex)
	return bi
}

func (input *TransactionInput) DecodeRLP(s *rlp.Stream) error {
	var dec rlpTransactionInput
	if err := s.Decode(&dec); err != nil {
		return err
	}
	if len(dec.BlockNumber) != BlockNumberLength {
		return errors.New("Invalid output number length")
	}
	if len(dec.TransactionNumber) != TransactionNumberLength {
		return errors.New("Invalid output number length")
	}
	if len(dec.OutputNumber) != OutputNumberLength {
		return errors.New("Invalid output number length")
	}
	if len(dec.Value) != ValueLength {
		return errors.New("Invalid value length")
	}
	copy(input.BlockNumber[:], dec.BlockNumber)
	copy(input.TransactionNumber[:], dec.TransactionNumber)
	copy(input.OutputNumber[:], dec.OutputNumber)
	copy(input.Value[:], dec.Value)
	return nil
}
