package transaction

import (
	"bytes"
	"errors"
	"io"

	"github.com/bankex/go-plasma/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	// "github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/ethereum/go-ethereum/rlp"
)

// TransactionInput is one of the inputs into Plasma transaction
type SignedTransaction struct {
	UnsignedTransaction *UnsignedTransaction
	V                   [VLength]byte
	R                   [RLength]byte
	S                   [SLength]byte
	from                common.Address
	RawValue            []byte
}

type rlpSignedTransaction struct {
	UnsignedTransaction *UnsignedTransaction
	V                   []byte
	R                   []byte
	S                   []byte
}

func NewSignedTransaction(unsignedTX *UnsignedTransaction, v []byte, r []byte, s []byte) (*SignedTransaction, error) {
	tx := &SignedTransaction{}
	if len(v) != VLength {
		return nil, errors.New("")
	}
	if len(r) != RLength {
		return nil, errors.New("")
	}
	if len(s) != SLength {
		return nil, errors.New("")
	}
	copy(tx.V[:], v)
	copy(tx.R[:], r)
	copy(tx.S[:], s)
	tx.UnsignedTransaction = unsignedTX
	return tx, nil
}

// signature is [R || S || V]
func (tx *SignedTransaction) Validate() error {
	err := tx.UnsignedTransaction.Validate()
	if err != nil {
		return err
	}
	_, err = tx.GetFrom()
	if err != nil {
		return err
	}
	return nil
}

func (tx *SignedTransaction) GetFrom() (common.Address, error) {
	if (tx.from != common.Address{}) {
		return tx.from, nil
	}
	sender, err := tx.recoverSender()
	if err != nil {
		return common.Address{}, err
	}
	tx.from = sender
	return tx.from, nil
}

func (tx *SignedTransaction) recoverSender() (common.Address, error) {
	hash, err := tx.UnsignedTransaction.GetHash()
	if err != nil {
		return common.Address{}, err
	}
	fullSignature := []byte{}
	fullSignature = append(fullSignature, tx.R[:]...)
	fullSignature = append(fullSignature, tx.S[:]...)
	if tx.V[0] >= 27 {
		V := tx.V[0] - 27
		fullSignature = append(fullSignature, []byte{V}...)
	} else {
		fullSignature = append(fullSignature, tx.V[:]...)
	}
	senderPubKey, err := crypto.Ecrecover(hash[:], fullSignature)
	if err != nil {
		return common.Address{}, err
	}
	pubKey := crypto.ToECDSAPub(senderPubKey)
	sender := crypto.PubkeyToAddress(*pubKey)
	if (sender == common.Address{}) {
		return common.Address{}, errors.New("")
	}
	return sender, nil
}

// EncodeRLP implements rlp.Encoder, and flattens the consensus fields of a receipt
// into an RLP stream. If no post state is present, byzantium fork is assumed.
func (tx *SignedTransaction) EncodeRLP(w io.Writer) error {
	rlpSigned := rlpSignedTransaction{tx.UnsignedTransaction, tx.V[:], tx.R[:], tx.S[:]}
	return rlp.Encode(w, rlpSigned)
}

// DecodeRLP implements rlp.Decoder, and loads the consensus fields of a receipt
// from an RLP stream.
func (tx *SignedTransaction) DecodeRLP(s *rlp.Stream) error {
	var dec rlpSignedTransaction
	if err := s.Decode(&dec); err != nil {
		return err
	}
	if len(dec.V) != VLength {
		return errors.New("Invalid V length")
	}
	if len(dec.R) != RLength {
		return errors.New("Invalid R length")
	}
	if len(dec.S) != SLength {
		return errors.New("Invalid S length")
	}
	tx.UnsignedTransaction = dec.UnsignedTransaction
	copy(tx.V[:], dec.V)
	copy(tx.R[:], dec.R)
	copy(tx.S[:], dec.S)
	return nil
}

func (tx *SignedTransaction) GetRaw() ([]byte, error) {
	var b bytes.Buffer
	i := io.Writer(&b)
	err := tx.EncodeRLP(i)
	if err != nil {
		return nil, err
	}
	raw := b.Bytes()
	return raw, nil
}

func (tx *SignedTransaction) Sign(privateKey []byte) error {
	if len(privateKey) != 32 {
		return errors.New("Invalid private key length")
	}
	raw, err := tx.UnsignedTransaction.GetHash()
	if err != nil {
		return err
	}
	key, err := crypto.ToECDSA(privateKey)
	if err != nil {
		return err
	}
	sig, err := crypto.Sign(raw[:], key)
	if err != nil {
		return err
	}

	copy(tx.R[:], sig[0:31])
	copy(tx.S[:], sig[32:64])
	copy(tx.V[:], []byte{sig[64]})
	tx.from = common.Address{}
	return nil
}

func CreateRawFundingTX(to common.Address,
	value *types.BigInt,
	depositIndex *types.BigInt,
	signingKey []byte) (*SignedTransaction, error) {

	//input

	iBlockNumber := types.NewBigInt(0)
	iTransactionNumber := types.NewBigInt(0)
	iOutputNumber := types.NewBigInt(0)
	iValue := depositIndex
	input := &TransactionInput{}
	err := input.SetFields(iBlockNumber, iTransactionNumber, iOutputNumber, iValue)
	if err != nil {
		return nil, err
	}
	// output
	oOutputNumber := types.NewBigInt(0)
	oTo := to
	oValue := value
	output := &TransactionOutput{}
	err = output.SetFields(oOutputNumber, oTo, oValue)
	if err != nil {
		return nil, err
	}

	inputs := []*TransactionInput{input}
	outputs := []*TransactionOutput{output}
	txType := TransactionTypeFund
	unsignedTX, err := NewUnsignedTransaction(txType, inputs, outputs)
	if err != nil {
		return nil, err
	}
	emptyBytes := [32]byte{}
	signedTX, err := NewSignedTransaction(unsignedTX, []byte{0x00}, emptyBytes[:], emptyBytes[:])
	if err != nil {
		return nil, err
	}
	err = signedTX.Sign(signingKey)
	if err != nil {
		return nil, err
	}
	return signedTX, nil
}
