package transaction

import (
	"fmt"
	"testing"

	"bytes"
	"io"

	types "github.com/matterinc/PlasmaCommons/types"
	common "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

func TestInputCreation(t *testing.T) {
	bn := types.NewBigInt(0)
	tn := types.NewBigInt(1)
	in := types.NewBigInt(2)
	v := types.NewBigInt(3)
	var b bytes.Buffer
	i := io.Writer(&b)
	input := &TransactionInput{}
	err := input.SetFields(bn, tn, in, v)
	if err != nil {
		t.Errorf("Failed to set fields")
	}
	err = input.EncodeRLP(i)
	if err != nil {
		t.Errorf("Failed to encode")
	}
	a := b.Bytes()
	str := common.ToHex(a)
	fmt.Println(str)
	freshInput := &TransactionInput{}
	err = rlp.DecodeBytes(a, freshInput)
	if err != nil {
		t.Errorf("Failed to decode")
	}
}

func TestOutputCreation(t *testing.T) {
	bn := types.NewBigInt(0)
	to := &common.Address{}
	v := types.NewBigInt(3)
	var b bytes.Buffer
	i := io.Writer(&b)
	output := &TransactionOutput{}
	err := output.SetFields(bn, to, v)
	if err != nil {
		t.Errorf("Failed to set fields")
	}
	err = output.EncodeRLP(i)
	if err != nil {
		t.Errorf("Failed to encode")
	}
	a := b.Bytes()
	str := common.ToHex(a)
	fmt.Println(str)
	freshOutput := &TransactionOutput{}
	err = rlp.DecodeBytes(a, freshOutput)
	if err != nil {
		t.Errorf("Failed to decode")
	}
}

func TestTransactionCreation(t *testing.T) {

	bn := types.NewBigInt(0)
	tn := types.NewBigInt(1)
	in := types.NewBigInt(2)
	v := types.NewBigInt(3)
	input := &TransactionInput{}
	err := input.SetFields(bn, tn, in, v)
	if err != nil {
		t.Errorf("Failed to set fields")
	}

	bn = types.NewBigInt(0)
	to := &common.Address{}
	v = types.NewBigInt(12)

	output := &TransactionOutput{}
	err = output.SetFields(bn, to, v)
	if err != nil {
		t.Errorf("Failed to set fields")
	}

	inputs := []*TransactionInput{input}
	outputs := []*TransactionOutput{output}
	txType := TransactionTypeFund
	tx, err := NewUnsignedTransaction(txType, inputs, outputs)

	var b bytes.Buffer
	i := io.Writer(&b)
	err = tx.EncodeRLP(i)
	if err != nil {
		t.Errorf("Failed to encode")
	}
	a := b.Bytes()
	str := common.ToHex(a)
	fmt.Println(str)
	fresh := &UnsignedTransaction{}
	err = rlp.DecodeBytes(a, fresh)
	if err != nil {
		t.Errorf("Failed to decode")
	}
	fmt.Println(fresh)
}

func TestSignedTransactionCreation(t *testing.T) {

	bn := types.NewBigInt(0)
	tn := types.NewBigInt(1)
	in := types.NewBigInt(2)
	v := types.NewBigInt(3)
	input := &TransactionInput{}
	err := input.SetFields(bn, tn, in, v)
	if err != nil {
		t.Errorf("Failed to set fields")
	}

	bn = types.NewBigInt(0)
	to := &common.Address{}
	v = types.NewBigInt(12)

	output := &TransactionOutput{}
	err = output.SetFields(bn, to, v)
	if err != nil {
		t.Errorf("Failed to set fields")
	}

	inputs := []*TransactionInput{input}
	outputs := []*TransactionOutput{output}
	txType := TransactionTypeFund
	tx, err := NewUnsignedTransaction(txType, inputs, outputs)

	emptyBytes := [32]byte{}
	signed, err := NewSignedTransaction(tx, []byte{0x00}, emptyBytes[:], emptyBytes[:])

	var b bytes.Buffer
	i := io.Writer(&b)
	err = signed.EncodeRLP(i)
	if err != nil {
		t.Errorf("Failed to encode")
	}
	a := b.Bytes()
	str := common.ToHex(a)
	fmt.Println(str)
	fresh := &SignedTransaction{}
	err = rlp.DecodeBytes(a, fresh)
	if err != nil {
		t.Errorf("Failed to decode")
	}
	fmt.Println(fresh)

}
