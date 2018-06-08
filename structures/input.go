package transaction

import (
	// "github.com/ethereum/go-ethereum/common"
	// "github.com/ethereum/go-ethereum/common/hexutil"
	
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/bankex/go-plasma/types/BigInt"
)


type TransactionInput struct {
	BlockNumber [BlockNumberLength]byte
	TransactionNumber [TransactionNumberLength]byte
	OutputNumber [OutputNumberLength]byte
	Value     [ValueLength]byte
}

// NewReceipt creates a new input sturcture
func NewInput(blockNumber BigInt, transactionNumber BigInt, outputNumber BigInt, value BigInt) *TransactionInput {
	blockNumberBytes = blockNumber.getLeftPaddedBytes(BlockNumberLength)
	transactionNumberBytes = transactionNumber.getLeftPaddedBytes(TransactionNumberLength)
	outputNumberBytes = outputNumber.getLeftPaddedBytes(OutputNumberLength)
	valueBytes = value.getLeftPaddedBytes(ValueLength)
	r := &TransactionInput{BlockNumber: blockNumberBytes,
							TransactionNumber: transactionNumberBytes,
							OutputNumber: outputNumberBytes,
							Value: valueBytes}
	return r
}

// EncodeRLP implements rlp.Encoder, and flattens the consensus fields of a receipt
// into an RLP stream. If no post state is present, byzantium fork is assumed.
func (input *TransactionInput) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, &TransactionInput{input.BlockNumber, input.TransactionNumber, input.OutputNumber, input.Value})
}

// // DecodeRLP implements rlp.Decoder, and loads the consensus fields of a receipt
// // from an RLP stream.
// func (r *Receipt) DecodeRLP(s *rlp.Stream) error {
// 	var dec receiptRLP
// 	if err := s.Decode(&dec); err != nil {
// 		return err
// 	}
// 	if err := r.setStatus(dec.PostStateOrStatus); err != nil {
// 		return err
// 	}
// 	r.CumulativeGasUsed, r.Bloom, r.Logs = dec.CumulativeGasUsed, dec.Bloom, dec.Logs
// 	return nil
// }