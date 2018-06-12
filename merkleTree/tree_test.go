package merkletree

import (
	"bytes"
	"fmt"
	"testing"
)

func TestMerkleTreeCreation(t *testing.T) {
	content0 := NewTransactionContent([]byte{0x00})
	content1 := NewTransactionContent([]byte{0x01})
	var list []Content
	list = append(list, content0)
	list = append(list, content1)
	merkleTree, err := NewTree(list)
	if err != nil {
		t.Errorf("Failed to make tree")
	}
	proof, err := merkleTree.ProvideBinaryProof(0)
	if bytes.Compare(proof, append([]byte{0x01}, content1.CalculateHash()...)) != 0 {
		t.Errorf("Failed to make proof")
	}
	validProof, err := merkleTree.VerifyBinaryProof(merkleTree.MerkleRoot(), proof, content0)
	if err != nil || !validProof {
		t.Errorf("Failed to validate proof")
	}
}

func TestOddMerkleTreeCreation(t *testing.T) {
	content0 := NewTransactionContent([]byte{0x00})
	content1 := NewTransactionContent([]byte{0x01})
	content2 := NewTransactionContent([]byte{0x02})
	var list []Content
	list = append(list, content0)
	list = append(list, content1)
	list = append(list, content2)
	merkleTree, err := NewTree(list)
	if err != nil {
		t.Errorf("Failed to make tree")
	}
	proof0, err := merkleTree.ProvideBinaryProofForContent(content0)
	validProof0, err := merkleTree.VerifyBinaryProof(merkleTree.MerkleRoot(), proof0, content0)
	if err != nil || !validProof0 {
		t.Errorf("Failed to validate proof 0")
	}

	proof1, err := merkleTree.ProvideBinaryProofForContent(content1)
	fmt.Println(proof1)
	validProof1, err := merkleTree.VerifyBinaryProof(merkleTree.MerkleRoot(), proof1, content1)
	if err != nil || !validProof1 {
		t.Errorf("Failed to validate proof 1")
	}

	proof2, err := merkleTree.ProvideBinaryProofForContent(content2)
	validProof2, err := merkleTree.VerifyBinaryProof(merkleTree.MerkleRoot(), proof2, content2)
	if err != nil || !validProof2 {
		t.Errorf("Failed to validate proof 2")
	}
}
