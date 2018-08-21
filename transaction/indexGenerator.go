package transaction

import (
	"encoding/binary"
	"errors"
	"math/big"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/directory"
	"github.com/apple/foundationdb/bindings/go/src/fdb/subspace"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"

	"github.com/shamatar/go-plasma/types"
	"github.com/ethereum/go-ethereum/common"
)

const (
	UTXOIndexLength = AddressLength + BlockNumberLength + TransactionNumberLength +
		OutputNumberLength + ValueLength

	ShortUTXOIndexLength = BlockNumberLength + TransactionNumberLength +
		OutputNumberLength
)

var (
	MaxUTXOIndex = big.NewInt(0).Lsh(big.NewInt(1), (BlockNumberLength+TransactionNumberLength+
		OutputNumberLength)*8)
)

type HumanReadableUTXOdetails struct {
	Owner             string
	BlockNumber       uint32
	TransactionNumber uint32
	OutputNumber      uint8
	Value             string
}

type ShortUTXOdetails struct {
	BlockNumber       uint32
	TransactionNumber uint32
	OutputNumber      uint8
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

func CreateFdbUTXOIndexForInput(db fdb.Database, tx *SignedTransaction, inputNumber int) (subspace.Subspace, error) {
	if inputNumber > len(tx.UnsignedTransaction.Inputs) {
		return nil, errors.New("Invalid input number")
	}
	input := tx.UnsignedTransaction.Inputs[inputNumber]
	from, err := tx.GetFrom()
	if err != nil {
		return nil, err
	}
	addressDirectory, err := directory.CreateOrOpen(db, []string{"utxo"}, nil)
	// fmt.Println(common.ToHex(addressDirectory.Bytes()))
	fullSubspace := addressDirectory.Sub(tuple.Tuple{from[:]})
	// fmt.Println(common.ToHex(fullSubspace.Bytes()))
	fullSubspace = fullSubspace.Sub(tuple.Tuple{input.BlockNumber[:]})
	// fmt.Println(common.ToHex(fullSubspace.Bytes()))
	fullSubspace = fullSubspace.Sub(tuple.Tuple{input.TransactionNumber[:]})
	// fmt.Println(common.ToHex(fullSubspace.Bytes()))
	fullSubspace = fullSubspace.Sub(tuple.Tuple{input.OutputNumber[:]})
	// fmt.Println(common.ToHex(fullSubspace.Bytes()))
	fullSubspace = fullSubspace.Sub(tuple.Tuple{input.Value[:]})
	// fmt.Println(common.ToHex(fullSubspace.Bytes()))
	return fullSubspace, nil
}

func CreateUTXOIndexForOutput(tx *SignedTransaction, blockNumber uint32, transactionNumber uint32, outputNumber int) ([UTXOIndexLength]byte, error) {
	if outputNumber > len(tx.UnsignedTransaction.Outputs) {
		return [UTXOIndexLength]byte{}, errors.New("Invalid output number")
	}
	output := tx.UnsignedTransaction.Outputs[outputNumber]
	index := []byte{}

	blockNumberBuffer := make([]byte, BlockNumberLength)
	binary.BigEndian.PutUint32(blockNumberBuffer, blockNumber)

	transactionNumberBuffer := make([]byte, TransactionNumberLength)
	binary.BigEndian.PutUint32(transactionNumberBuffer, transactionNumber)

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

func CreateShortUTXOIndexForOutput(tx *SignedTransaction, blockNumber uint32, transactionNumber uint32, outputNumber int) ([]byte, error) {
	if outputNumber > len(tx.UnsignedTransaction.Outputs) {
		return nil, errors.New("Invalid output number")
	}
	output := tx.UnsignedTransaction.Outputs[outputNumber]
	index := []byte{}

	blockNumberBuffer := make([]byte, BlockNumberLength)
	binary.BigEndian.PutUint32(blockNumberBuffer, blockNumber)

	transactionNumberBuffer := make([]byte, TransactionNumberLength)
	binary.BigEndian.PutUint32(transactionNumberBuffer, transactionNumber)

	index = append(index, blockNumberBuffer...)
	index = append(index, transactionNumberBuffer[:]...)
	index = append(index, output.OutputNumber[:]...)
	return index, nil
}

func CreateShortUTXOIndexForInput(tx *SignedTransaction, inputNumber int) ([]byte, error) {
	if inputNumber > len(tx.UnsignedTransaction.Outputs) {
		return nil, errors.New("Invalid input number")
	}
	input := tx.UnsignedTransaction.Inputs[inputNumber]
	index := []byte{}
	index = append(index, input.BlockNumber[:]...)
	index = append(index, input.TransactionNumber[:]...)
	index = append(index, input.OutputNumber[:]...)
	return index, nil
}

func PackUTXOnumber(blockNumber uint32, transactionNumber uint32, outputOrInputNumber uint8) []byte {
	index := []byte{}

	blockNumberBuffer := make([]byte, BlockNumberLength)
	binary.BigEndian.PutUint32(blockNumberBuffer, blockNumber)

	transactionNumberBuffer := make([]byte, TransactionNumberLength)
	binary.BigEndian.PutUint32(transactionNumberBuffer, transactionNumber)

	index = append(index, blockNumberBuffer...)
	index = append(index, transactionNumberBuffer[:]...)
	index = append(index, []byte{outputOrInputNumber}...)
	return index
}

func ParseUTXOnumber(index []byte) (uint32, uint32, uint8, error) {
	if len(index) != ShortUTXOIndexLength {
		return 0, 0, 0, errors.New("Invalid index length")
	}

	blockNumber := binary.BigEndian.Uint32(index[0:4])
	transactionNumber := binary.BigEndian.Uint32(index[4:8])
	outputNumber := index[8]
	return blockNumber, transactionNumber, outputNumber, nil
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

func ParseUTXOindexNumberIntoDetails(indexBN *types.BigInt) (*ShortUTXOdetails, error) {
	if indexBN.Bigint.Cmp(MaxUTXOIndex) != -1 {
		return nil, errors.New("Index is too large")
	}
	index, err := indexBN.GetLeftPaddedBytes(ShortUTXOIndexLength)
	if err != nil {
		return nil, err
	}
	idx := 0

	blockNumberBytes := index[idx : idx+BlockNumberLength]
	idx += BlockNumberLength

	transactionNumberBytes := index[idx : idx+TransactionNumberLength]
	idx += TransactionNumberLength

	outputNumberBytes := index[idx : idx+OutputNumberLength]
	idx += OutputNumberLength

	blockNumber := binary.BigEndian.Uint32(blockNumberBytes)
	transactionNumber := binary.BigEndian.Uint32(transactionNumberBytes)
	outputNumber := uint8(outputNumberBytes[0])
	return &ShortUTXOdetails{blockNumber, transactionNumber, outputNumber}, nil
}
