package transaction

import (
	"errors"
	"io"

	"github.com/ethereum/go-ethereum/common"
	// "github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/matterinc/PlasmaCommons/types"
	"github.com/ethereum/go-ethereum/rlp"
)

// TransactionInput is one of the inputs into Plasma transaction
type TransactionOutput struct {
	OutputNumber [OutputNumberLength]byte
	To           [AddressLength]byte
	Value        [ValueLength]byte
}

type rlpTransactionOutput struct {
	OutputNumber []byte
	To           []byte
	Value        []byte
}

// SetFields creates a new input sturcture
func (output *TransactionOutput) SetFields(outputNumber *types.BigInt, address common.Address, value *types.BigInt) error {
	outputNumberBytes, err := outputNumber.GetLeftPaddedBytes(OutputNumberLength)
	if err != nil {
		return errors.New("Output number is too long")
	}
	valueBytes, err := value.GetLeftPaddedBytes(ValueLength)
	if err != nil {
		return errors.New("Value is too long")
	}
	addressBytes := address.Bytes()
	copy(output.OutputNumber[:], outputNumberBytes)
	copy(output.To[:], addressBytes)
	copy(output.Value[:], valueBytes)
	return nil
}

func (output *TransactionOutput) GetValue() *types.BigInt {
	bi := types.NewBigInt(0)
	bi.SetBytes(output.Value[:])
	return bi
}

// EncodeRLP implements rlp.Encoder, and flattens the consensus fields of a receipt
// into an RLP stream. If no post state is present, byzantium fork is assumed.
func (output *TransactionOutput) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, rlpTransactionOutput{output.OutputNumber[:], output.To[:], output.Value[:]})
}

// Get address as Ethereum Address type
func (output *TransactionOutput) GetToAddress() common.Address {
	return output.To
}

// DecodeRLP implements rlp.Decoder, and loads the consensus fields of a receipt
// from an RLP stream.
func (output *TransactionOutput) DecodeRLP(s *rlp.Stream) error {
	var dec rlpTransactionOutput
	if err := s.Decode(&dec); err != nil {
		return err
	}
	if len(dec.OutputNumber) != OutputNumberLength {
		return errors.New("Invalid output number length")
	}
	if len(dec.To) != AddressLength {
		return errors.New("Invalid address length")
	}
	if len(dec.Value) != ValueLength {
		return errors.New("Invalid value length")
	}
	copy(output.OutputNumber[:], dec.OutputNumber)
	copy(output.To[:], dec.To)
	copy(output.Value[:], dec.Value)
	// output.OutputNumber = dec.OutputNumber
	// output.To = dec.To
	// output.Value = dec.Value
	return nil
}
