package block

import (
	"bytes"
	"encoding/binary"
	"errors"

	helpers "github.com/shamatar/go-plasma/common"
	crypto "github.com/shamatar/go-plasma/crypto"
	secp256k1 "github.com/shamatar/go-plasma/crypto/secp256k1"
	"github.com/shamatar/go-plasma/transaction"
	common "github.com/ethereum/go-ethereum/common"
)

const (
	BlockHeaderLength = transaction.BlockNumberLength + transaction.TransactionNumberLength + PreviousBlockHashLength +
		MerkleTreeRootLength + transaction.VLength + transaction.RLength + transaction.SLength
)

// TransactionInput is one of the inputs into Plasma transaction
type BlockHeader struct {
	BlockNumber          [transaction.BlockNumberLength]byte
	NumberOfTransactions [transaction.TransactionNumberLength]byte
	PreviousBlockHash    [PreviousBlockHashLength]byte
	MerkleTreeRoot       [MerkleTreeRootLength]byte
	V                    [transaction.VLength]byte
	R                    [transaction.RLength]byte
	S                    [transaction.SLength]byte
	from                 common.Address
}

func NewBlockHeader(blockNumber uint32,
	numberOfTransactions uint32,
	previousBlockHash []byte,
	merkleTreeRoot []byte,
	v []byte, r []byte, s []byte) (*BlockHeader, error) {
	header := &BlockHeader{}
	if len(previousBlockHash) != PreviousBlockHashLength {
		return nil, errors.New("")
	}
	if len(merkleTreeRoot) != MerkleTreeRootLength {
		return nil, errors.New("")
	}
	if len(v) != transaction.VLength {
		return nil, errors.New("")
	}
	if len(r) != transaction.RLength {
		return nil, errors.New("")
	}
	if len(s) != transaction.SLength {
		return nil, errors.New("")
	}

	blockNumberBuffer := make([]byte, transaction.BlockNumberLength)
	binary.BigEndian.PutUint32(blockNumberBuffer, blockNumber)

	numTXBuffer := make([]byte, transaction.TransactionNumberLength)
	binary.BigEndian.PutUint32(numTXBuffer, numberOfTransactions)

	copy(header.BlockNumber[:], blockNumberBuffer)
	copy(header.NumberOfTransactions[:], numTXBuffer)
	copy(header.PreviousBlockHash[:], previousBlockHash)
	copy(header.MerkleTreeRoot[:], merkleTreeRoot)
	copy(header.V[:], v)
	copy(header.R[:], r)
	copy(header.S[:], s)
	return header, nil
}

func NewUnsignedBlockHeader(blockNumber uint32,
	numberOfTransactions uint32,
	previousBlockHash []byte,
	merkleTreeRoot []byte) (*BlockHeader, error) {
	header := &BlockHeader{}
	if len(previousBlockHash) != PreviousBlockHashLength {
		return nil, errors.New("")
	}
	if len(merkleTreeRoot) != MerkleTreeRootLength {
		return nil, errors.New("")
	}

	blockNumberBuffer := make([]byte, transaction.BlockNumberLength)
	binary.BigEndian.PutUint32(blockNumberBuffer, blockNumber)

	numTXBuffer := make([]byte, transaction.TransactionNumberLength)
	binary.BigEndian.PutUint32(numTXBuffer, numberOfTransactions)
	copy(header.BlockNumber[:], blockNumberBuffer)
	copy(header.NumberOfTransactions[:], numTXBuffer)
	copy(header.PreviousBlockHash[:], previousBlockHash)
	copy(header.MerkleTreeRoot[:], merkleTreeRoot)
	return header, nil
}

func NewBlockHeaderFromBytes(serializedHeader []byte) (*BlockHeader, error) {
	if len(serializedHeader) != BlockHeaderLength {
		return nil, errors.New("Invalid header length")
	}
	header := &BlockHeader{}
	idx := 0

	blockNumberBuffer := serializedHeader[idx : idx+transaction.BlockNumberLength]
	idx += transaction.BlockNumberLength
	numTXBuffer := serializedHeader[idx : idx+transaction.TransactionNumberLength]
	idx += transaction.TransactionNumberLength
	previousHashBuffer := serializedHeader[idx : idx+PreviousBlockHashLength]
	idx += PreviousBlockHashLength
	merkleRootBuffer := serializedHeader[idx : idx+MerkleTreeRootLength]
	idx += MerkleTreeRootLength
	vBuffer := serializedHeader[idx : idx+transaction.VLength]
	idx += transaction.VLength
	rBuffer := serializedHeader[idx : idx+transaction.RLength]
	idx += transaction.RLength
	sBuffer := serializedHeader[idx : idx+transaction.SLength]
	idx += transaction.RLength
	copy(header.BlockNumber[:], blockNumberBuffer)
	copy(header.NumberOfTransactions[:], numTXBuffer)
	copy(header.PreviousBlockHash[:], previousHashBuffer)
	copy(header.MerkleTreeRoot[:], merkleRootBuffer)
	copy(header.V[:], vBuffer)
	copy(header.R[:], rBuffer)
	copy(header.S[:], sBuffer)
	return header, nil
}

func (header *BlockHeader) GetHashToSign() (common.Hash, error) {
	toHash := []byte{}
	toHash = append(toHash, header.BlockNumber[:]...)
	toHash = append(toHash, header.NumberOfTransactions[:]...)
	toHash = append(toHash, header.PreviousBlockHash[:]...)
	toHash = append(toHash, header.MerkleTreeRoot[:]...)
	personalHash := helpers.CreatePersonalHash(toHash)
	return personalHash, nil
}

func (header *BlockHeader) GetHash() (common.Hash, error) {
	toHash := header.GetRaw()
	personalHash := helpers.CreatePersonalHash(toHash)
	return personalHash, nil
}

func (header *BlockHeader) GetFrom() (common.Address, error) {
	if bytes.Compare(header.from[:], EmptyAddress[:]) != 0 {
		return header.from, nil
	}
	sender, err := header.recoverSender()
	if err != nil {
		return common.Address{}, err
	}
	header.from = sender
	return header.from, nil
}

func (header *BlockHeader) recoverSender() (common.Address, error) {
	hash, err := header.GetHashToSign()
	if err != nil {
		return common.Address{}, err
	}
	fullSignature := []byte{}
	fullSignature = append(fullSignature, header.R[:]...)
	fullSignature = append(fullSignature, header.S[:]...)
	if header.V[0] >= 27 {
		V := header.V[0] - 27
		fullSignature = append(fullSignature, []byte{V}...)
	} else {
		fullSignature = append(fullSignature, header.V[:]...)
	}
	senderPubKey, err := secp256k1.RecoverPubkey(hash[:], fullSignature)
	if err != nil {
		return common.Address{}, err
	}
	sender := crypto.PubkeyToAddress(senderPubKey)
	if (sender == common.Address{}) {
		return common.Address{}, errors.New("")
	}
	return sender, nil
}

func (header *BlockHeader) GetRaw() []byte {
	toReturn := []byte{}
	toReturn = append(toReturn, header.BlockNumber[:]...)
	toReturn = append(toReturn, header.NumberOfTransactions[:]...)
	toReturn = append(toReturn, header.PreviousBlockHash[:]...)
	toReturn = append(toReturn, header.MerkleTreeRoot[:]...)
	toReturn = append(toReturn, header.V[:]...)
	toReturn = append(toReturn, header.R[:]...)
	toReturn = append(toReturn, header.S[:]...)
	return toReturn
}

func (header *BlockHeader) Sign(privateKey []byte) error {
	if len(privateKey) != 32 {
		return errors.New("Invalid private key length")
	}
	raw, err := header.GetHashToSign()
	if err != nil {
		return err
	}
	sig, err := secp256k1.Sign(raw[:], privateKey)
	if err != nil {
		return err
	}
	v := sig[64]
	if v < 27 {
		v = v + 27
	}

	copy(header.R[:], sig[0:32])
	copy(header.S[:], sig[32:64])
	copy(header.V[:], []byte{v})
	header.from = common.Address{}
	return nil
}
