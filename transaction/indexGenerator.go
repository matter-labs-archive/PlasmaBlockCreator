package transaction

import (
	"encoding/binary"
	"errors"

	"github.com/bankex/go-plasma/types"
	"github.com/ethereum/go-ethereum/common"
)

const (
	UTXOIndexLength = AddressLength + BlockNumberLength + TransactionNumberLength +
		OutputNumberLength + ValueLength

	// UTXOIndexLength = AddressLength + BlockNumberLength + TransactionNumberLength +
	// OutputNumberLength
)

type HumanReadableUTXOdetails struct {
	Owner             string
	BlockNumber       uint32
	TransactionNumber uint32
	OutputNumber      uint8
	Value             string
}

func CreateCorrespondingUTXOIndexForInput(tx *SignedTransaction, inputNumber int) ([UTXOIndexLength]byte, error) {
	if inputNumber > len(tx.UnsignedTransaction.Inputs) {
		return [UTXOIndexLength]byte{}, errors.New("Invalid input number")
	}
	input := tx.UnsignedTransaction.Inputs[inputNumber]
	from, err := tx.GetFrom()
	if err != nil {
		return [UTXOIndexLength]byte{}, err
	}
	index := []byte{}
	index = append(index, from[:]...)
	index = append(index, input.BlockNumber[:]...)
	index = append(index, input.TransactionNumber[:]...)
	index = append(index, input.OutputNumber[:]...)
	index = append(index, input.Value[:]...)
	if len(index) != UTXOIndexLength {
		return [UTXOIndexLength]byte{}, errors.New("Index length mismatch")
	}
	indexCopy := [UTXOIndexLength]byte{}
	copy(indexCopy[:], index)
	return indexCopy, nil
}

func CreateUTXOIndexForOutput(tx *NumberedTransaction, outputNumber int, blockNumber uint32) ([UTXOIndexLength]byte, error) {
	if outputNumber > len(tx.SignedTransaction.UnsignedTransaction.Outputs) {
		return [UTXOIndexLength]byte{}, errors.New("Invalid output number")
	}
	output := tx.SignedTransaction.UnsignedTransaction.Outputs[outputNumber]
	index := []byte{}

	blockNumberBuffer := make([]byte, BlockNumberLength)
	binary.BigEndian.PutUint32(blockNumberBuffer, blockNumber)

	transactionNumberBuffer := tx.TransactionNumber

	index = append(index, output.To[:]...)
	index = append(index, blockNumberBuffer...)
	index = append(index, transactionNumberBuffer[:]...)
	index = append(index, output.OutputNumber[:]...)
	index = append(index, output.Value[:]...)
	if len(index) != UTXOIndexLength {
		return [UTXOIndexLength]byte{}, errors.New("Index length mismatch")
	}
	indexCopy := [UTXOIndexLength]byte{}
	copy(indexCopy[:], index)
	return indexCopy, nil
}

func ParseIndexIntoUTXOdetails(index [UTXOIndexLength]byte) HumanReadableUTXOdetails {
	idx := 0
	ownerBytes := index[idx : idx+AddressLength]
	idx += AddressLength

	blockNumberBytes := index[idx : idx+BlockNumberLength]
	idx += BlockNumberLength

	transactionNumberBytes := index[idx : idx+TransactionNumberLength]
	idx += TransactionNumberLength

	outputNumberBytes := index[idx : idx+OutputNumberLength]
	idx += OutputNumberLength

	valueBytes := index[idx : idx+ValueLength]

	blockNumber := binary.BigEndian.Uint32(blockNumberBytes)
	transactionNumber := binary.BigEndian.Uint32(transactionNumberBytes)
	outputNumber := uint8(outputNumberBytes[0])
	value := types.NewBigInt(0)
	value.SetBytes(valueBytes)
	owner := common.ToHex(ownerBytes)
	return HumanReadableUTXOdetails{owner, blockNumber, transactionNumber, outputNumber, value.GetString(10)}
}
