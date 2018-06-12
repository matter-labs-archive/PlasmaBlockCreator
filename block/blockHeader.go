package block

import (
	"encoding/binary"
	"errors"

	"github.com/bankex/go-plasma/transaction"

	common "github.com/ethereum/go-ethereum/common"
	crypto "github.com/ethereum/go-ethereum/crypto"
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
	if len(v) != VLength {
		return nil, errors.New("")
	}
	if len(r) != RLength {
		return nil, errors.New("")
	}
	if len(s) != SLength {
		return nil, errors.New("")
	}

	blockNumberBuffer := make([]byte, transaction.BlockNumberLength)
	err := binary.BigEndian.PutUint32(blockNumberBuffer, blockNumber)
	if err != nil {
		return nil, err
	}

	numTXBuffer := make([]byte, transaction.TransactionNumberLength)
	err = binary.BigEndian.PutUint32(numTXBuffer, numberOfTransactions)
	if err != nil {
		return nil, err
	}
	copy(header.BlockNumber[:], blockNumberBuffer)
	copy(numTXBuffer[:], numTXBuffer)
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
	err := binary.BigEndian.PutUint32(blockNumberBuffer, blockNumber)
	if err != nil {
		return nil, err
	}

	numTXBuffer := make([]byte, transaction.TransactionNumberLength)
	err = binary.BigEndian.PutUint32(numTXBuffer, numberOfTransactions)
	if err != nil {
		return nil, err
	}
	copy(header.BlockNumber[:], blockNumberBuffer)
	copy(numTXBuffer[:], numTXBuffer)
	copy(header.PreviousBlockHash[:], previousBlockHash)
	copy(header.MerkleTreeRoot[:], merkleTreeRoot)
	return header, nil
}

func (header *BlockHeader) GetHash() (common.Hash, error) {
	toHash := []byte{header.BlockNumber[:]}
	toHash = append(toHash, header.NumberOfTransactions[:]...)
	toHash = append(toHash, header.PreviousBlockHash[:]...)
	toHash = append(toHash, header.MerkleTreeRoot[:]...)
	personalHash := heplers.CreatePersonalHash(toHash)
	return personalHash, nil
}

func (header *BlockHeader) GetFrom() (common.Address, error) {
	if (header.from != common.Address{}) {
		return header.from, nil
	}
	sender, err := header.recoverSender()
	if err != nil {
		return common.Address{}, err
	}
	tx.from = sender
	return tx.from, nil
}

func (header *BlockHeader) recoverSender() (common.Address, error) {
	hash, err := header.Get()
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

func (header *BlockHeader) GetRaw() ([]byte, error) {
	if header.R[:] == make([]byte, transaction.RLength) {
		return nil, errors.New("Not signed header")
	}
	toReturn := []byte{header.BlockNumber[:]}
	toReturn = append(toReturn, header.NumberOfTransactions[:]...)
	toReturn = append(toReturn, header.PreviousBlockHash[:]...)
	toReturn = append(toReturn, header.MerkleTreeRoot[:]...)
	toReturn = append(toReturn, header.V[:]...)
	toReturn = append(toReturn, header.R[:]...)
	toReturn = append(toReturn, header.S[:]...)
	return toReturn, nil
}

func (header *BlockHeader) Sign(privateKey []byte) error {
	if len(privateKey) != 32 {
		return errors.New("Invalid private key length")
	}
	raw, err := header.GetHash()
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

	copy(header.R[:], sig[0:32])
	copy(header.S[:], sig[32:64])
	copy(header.V[:], []byte{sig[64]})
	tx.from = common.Address{}
	return nil
}
