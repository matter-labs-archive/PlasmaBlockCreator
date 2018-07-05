package block

import (
	"bytes"
	"errors"
	"io"

	"github.com/bankex/go-plasma/merkleTree"
	"github.com/cornelk/hashmap"

	"github.com/bankex/go-plasma/transaction"
	common "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	// "github.com/ethereum/go-ethereum/common/hexutil"
)

// TransactionInput is one of the inputs into Plasma transaction
type Block struct {
	BlockHeader  *BlockHeader
	Transactions []*transaction.NumberedTransaction
	MerkleTree   *merkletree.MerkleTree
}

type rlpBlockTransactions struct {
	Transactions []*transaction.NumberedTransaction
}

func treeFromTransactions(txes []*transaction.NumberedTransaction) (*merkletree.MerkleTree, error) {
	contents := make([]merkletree.Content, len(txes))
	for i, tx := range txes {
		raw, err := tx.GetRaw()
		if err != nil {
			return nil, err
		}
		contents[i] = merkletree.NewTransactionContent(raw)
	}
	tree, err := merkletree.NewTree(contents)
	if err != nil {
		return nil, err
	}
	return tree, nil
}

func NewBlock(blockNumber uint32, txes []*transaction.SignedTransaction, previousBlockHash []byte) (*Block, error) {
	block := &Block{}
	validTXes := make([]*transaction.NumberedTransaction, 0)
	enumeratingCounter := uint32(0)
	for _, tx := range txes {
		err := tx.Validate()
		if err != nil {
			continue
		}
		numberedTX, err := transaction.NewNumberedTransaction(tx, enumeratingCounter)
		if err != nil {
			continue
		}
		enumeratingCounter++
		validTXes = append(validTXes, numberedTX)
	}

	inputLookupHashmap := &hashmap.HashMap{}
	outputLookupHashmap := &hashmap.HashMap{}
	for _, tx := range validTXes {
		for _, input := range tx.SignedTransaction.UnsignedTransaction.Inputs {
			key := input.GetReferedUTXO().GetBytes()
			val, _ := inputLookupHashmap.Get(key)
			if val == nil {
				inputLookupHashmap.Set(key, []byte{0x01})
			} else {
				return nil, errors.New("Potential doublespend")
			}
		}
		for j := range tx.SignedTransaction.UnsignedTransaction.Outputs {
			key, err := transaction.CreateShortUTXOIndexForOutput(tx, j, blockNumber)
			if err != nil {
				return nil, errors.New("Transaction numbering is incorrect")
			}
			val, _ := inputLookupHashmap.Get(key)
			if val == nil {
				outputLookupHashmap.Set(key, []byte{0x01})
			} else {
				return nil, errors.New("Transaction numbering is incorrect")
			}
		}
	}

	tree, err := treeFromTransactions(validTXes)
	if err != nil {
		return nil, err
	}

	merkleRoot := tree.MerkleRoot()

	if len(validTXes) > 4294967295 {
		return nil, errors.New("Too many transactions in block")
	}

	header, err := NewUnsignedBlockHeader(blockNumber, uint32(len(validTXes)), previousBlockHash, merkleRoot)
	if err != nil {
		return nil, err
	}
	block.BlockHeader = header
	block.Transactions = validTXes
	block.MerkleTree = tree
	return block, nil
}

// signature is [R || S || V]
func (block *Block) Validate() error {
	_, err := block.BlockHeader.GetFrom()
	if err != nil {
		return err
	}

	// newTree, err := treeFromTransactions(block.Transactions)
	// if err != nil {
	// 	return err
	// }
	// if bytes.Compare(newTree.MerkleRoot(), block.BlockHeader.MerkleTreeRoot[:]) != 0 {
	// 	return errors.New("Merkle tree root mismatch")
	// }
	return nil
}

func (block *Block) GetFrom() (common.Address, error) {
	return block.BlockHeader.GetFrom()
}

func (block *Block) Sign(privateKey []byte) error {
	return block.BlockHeader.Sign(privateKey)
}

func (block *Block) Serialize() ([]byte, error) {
	err := block.Validate()
	if err != nil {
		return nil, err
	}
	headerBytes := block.BlockHeader.GetRaw()
	var b bytes.Buffer
	i := io.Writer(&b)
	err = block.EncodeRLP(i)
	if err != nil {
		return nil, err
	}
	rawTransactionsArray := b.Bytes()
	fullBlock := []byte{}
	fullBlock = append(fullBlock, headerBytes...)
	fullBlock = append(fullBlock, rawTransactionsArray...)
	return fullBlock, nil
}

func NewBlockFromBytes(rawBlock []byte) (*Block, error) {
	if len(rawBlock) <= BlockHeaderLength {
		return nil, errors.New("Data is too short")
	}
	headerBytes := rawBlock[0:BlockHeaderLength]
	header, err := NewBlockHeaderFromBytes(headerBytes)
	if err != nil {
		return nil, err
	}
	blockBytes := rawBlock[BlockHeaderLength:]
	block := &Block{}
	err = rlp.DecodeBytes(blockBytes, block)
	if err != nil {
		return nil, err
	}

	newTree, err := treeFromTransactions(block.Transactions)
	if err != nil {
		return nil, err
	}
	block.BlockHeader = header
	if bytes.Compare(newTree.MerkleRoot(), block.BlockHeader.MerkleTreeRoot[:]) != 0 {
		return nil, errors.New("Merkle tree root mismatch")
	}
	block.MerkleTree = newTree
	return block, nil
}

func (block *Block) EncodeRLP(w io.Writer) error {
	transactionBodies := make([]*transaction.NumberedTransaction, len(block.Transactions))
	for i, tx := range block.Transactions {
		transactionBodies[i] = tx
	}
	rlpBlockBody := rlpBlockTransactions{transactionBodies}
	return rlp.Encode(w, rlpBlockBody)
}

// DecodeRLP implements rlp.Decoder, and loads the consensus fields of a receipt
// from an RLP stream.
func (block *Block) DecodeRLP(s *rlp.Stream) error {
	var dec rlpBlockTransactions
	if err := s.Decode(&dec); err != nil {
		return err
	}
	transactionBodies := make([]*transaction.NumberedTransaction, len(dec.Transactions))
	for i, tx := range dec.Transactions {
		transactionBodies[i] = tx
	}
	block.Transactions = transactionBodies
	return nil
}
