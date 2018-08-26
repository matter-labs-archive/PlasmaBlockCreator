package block

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/shamatar/go-plasma/merkleTree"

	common "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/shamatar/go-plasma/transaction"
	// "github.com/ethereum/go-ethereum/common/hexutil"
)

// TransactionInput is one of the inputs into Plasma transaction
type Block struct {
	BlockHeader  *BlockHeader
	Transactions []*transaction.SignedTransaction
	MerkleTree   *merkletree.MerkleTree
}

type rlpBlockTransactions struct {
	Transactions []*transaction.SignedTransaction
}

func treeFromTransactions(txes []*transaction.SignedTransaction) (*merkletree.MerkleTree, error) {
	start := time.Now()
	contents := make([]merkletree.Content, len(txes))
	for i, tx := range txes {
		raw, err := tx.GetRaw()
		if err != nil {
			return nil, err
		}
		contents[i] = merkletree.NewTransactionContent(raw)
	}
	elapsed := time.Since(start)
	fmt.Println("Tree content preparation for " + strconv.Itoa(len(txes)) + " is " + fmt.Sprintf("%f", elapsed.Seconds()))

	start = time.Now()
	tree, err := merkletree.NewTree(contents)
	if err != nil {
		return nil, err
	}
	elapsed = time.Since(start)
	fmt.Println("Tree creation for " + strconv.Itoa(len(txes)) + " is " + fmt.Sprintf("%f", elapsed.Seconds()))
	return tree, nil
}

func NewBlock(blockNumber uint32, txes []*transaction.SignedTransaction, previousBlockHash []byte) (*Block, error) {
	block := &Block{}
	// validTXes := make([]*transaction.SignedTransaction, 0)

	// start := time.Now()
	// for _, tx := range txes {
	// 	err := tx.Validate()
	// 	if err != nil {
	// 		fmt.Println(err.Error())
	// 		continue
	// 	}
	// 	validTXes = append(validTXes, tx)
	// }
	// elapsed := time.Since(start)
	// fmt.Println("Transaction validation for " + strconv.Itoa(len(txes)) + " is " + fmt.Sprintf("%f", elapsed.Seconds()))

	// // try parallel validation

	// type empty struct{}
	// res := make([]*transaction.SignedTransaction, len(txes))
	// sem := make(chan empty, len(txes)) // semaphore pattern
	// validTXes := make([]*transaction.SignedTransaction, 0)

	// start := time.Now()
	// for i, tx := range txes {
	// 	go func(j int, signedTX *transaction.SignedTransaction) {
	// 		defer func() { sem <- empty{} }()
	// 		err := signedTX.Validate()
	// 		if err != nil {
	// 			fmt.Println(err.Error())
	// 			return
	// 		}
	// 		res[j] = signedTX
	// 	}(i, tx)
	// }
	// // wait for goroutines to finish
	// for i := 0; i < len(txes); i++ {
	// 	<-sem
	// 	if res[i] != nil {
	// 		validTXes = append(validTXes, res[i])
	// 	} else {
	// 		fmt.Println("Invalid tx encountered for " + strconv.Itoa(i))
	// 	}
	// }

	// elapsed := time.Since(start)
	// fmt.Println("Parallel transaction validation for " + strconv.Itoa(len(txes)) + " is " + fmt.Sprintf("%f", elapsed.Seconds()))

	// for some reason the above function fails and does NOT wait for all checks to finish
	validTXes := make([]*transaction.SignedTransaction, 0)
	res := make([]*transaction.SignedTransaction, len(txes))
	var wg sync.WaitGroup
	start := time.Now()
	wg.Add(len(txes))
	for i, tx := range txes {
		go func(j int, signedTX *transaction.SignedTransaction) {
			defer wg.Done()
			err := signedTX.Validate()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			res[j] = signedTX
		}(i, tx)
	}
	wg.Wait()

	// wait for goroutines to finish
	for i := 0; i < len(txes); i++ {
		if res[i] != nil {
			validTXes = append(validTXes, res[i])
		} else {
			fmt.Println("Invalid tx encountered for " + strconv.Itoa(i))
		}
	}

	elapsed := time.Since(start)
	fmt.Println("Parallel transaction validation for " + strconv.Itoa(len(txes)) + " is " + fmt.Sprintf("%f", elapsed.Seconds()))

	// disable hashmapping check for now

	// start = time.Now()
	// inputLookupHashmap := hashmap.New(uintptr(2 * len(validTXes)))
	// outputLookupHashmap := hashmap.New(uintptr(2 * len(validTXes)))
	// for i, tx := range validTXes {
	// 	for _, input := range tx.UnsignedTransaction.Inputs {
	// 		key := input.GetReferedUTXO().GetBytes()
	// 		val, _ := inputLookupHashmap.Get(key)
	// 		if val == nil {
	// 			inputLookupHashmap.Set(key, 1)
	// 		} else {
	// 			return nil, errors.New("Potential doublespend")
	// 		}
	// 	}
	// 	for j := range tx.UnsignedTransaction.Outputs {
	// 		key, err := transaction.CreateShortUTXOIndexForOutput(tx, blockNumber, uint32(i), j)
	// 		if err != nil {
	// 			return nil, errors.New("Transaction numbering is incorrect")
	// 		}
	// 		val, _ := outputLookupHashmap.Get(key)
	// 		if val == nil {
	// 			outputLookupHashmap.Set(key, 1)
	// 		} else {
	// 			return nil, errors.New("Transaction numbering is incorrect")
	// 		}
	// 	}
	// }
	// elapsed = time.Since(start)
	// fmt.Println("Hashmapping for " + strconv.Itoa(len(validTXes)) + " is " + fmt.Sprintf("%f", elapsed.Seconds()))

	tree, err := treeFromTransactions(validTXes)
	if err != nil {
		return nil, err
	}

	start = time.Now()
	merkleRoot := tree.MerkleRoot()
	elapsed = time.Since(start)
	fmt.Println("Merkle root obtained for " + strconv.Itoa(len(validTXes)) + " is " + fmt.Sprintf("%f", elapsed.Seconds()))

	if len(validTXes) > 4294967296 { // 2**32
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
	transactionBodies := make([]*transaction.SignedTransaction, len(block.Transactions))
	for i, tx := range block.Transactions {
		transactionBodies[i] = tx
	}
	// rlpBlockBody := rlpBlockTransactions{transactionBodies}
	// return rlp.Encode(w, rlpBlockBody)
	return rlp.Encode(w, transactionBodies)
}

// DecodeRLP implements rlp.Decoder, and loads the consensus fields of a receipt
// from an RLP stream.
func (block *Block) DecodeRLP(s *rlp.Stream) error {
	var dec []*transaction.SignedTransaction
	if err := s.Decode(&dec); err != nil {
		return err
	}
	transactionBodies := make([]*transaction.SignedTransaction, len(dec))
	for i, tx := range dec {
		transactionBodies[i] = tx
	}
	block.Transactions = transactionBodies
	return nil
}

// // DecodeRLP implements rlp.Decoder, and loads the consensus fields of a receipt
// // from an RLP stream.
// func (block *Block) DecodeRLP(s *rlp.Stream) error {
// 	var dec rlpBlockTransactions
// 	if err := s.Decode(&dec); err != nil {
// 		return err
// 	}
// 	transactionBodies := make([]*transaction.SignedTransaction, len(dec.Transactions))
// 	for i, tx := range dec.Transactions {
// 		transactionBodies[i] = tx
// 	}
// 	block.Transactions = transactionBodies
// 	return nil
// }
