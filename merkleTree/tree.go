package merkletree

import (
	"bytes"
	"errors"
	"fmt"

	sha3 "github.com/matterinc/PlasmaCommons/crypto/sha3"
)

//Content represents the data that is stored and verified by the tree. A type that
//implements this interface can be used as an item in the tree.
type Content interface {
	CalculateHash() []byte
	Equals(other Content) bool
}

//MerkleTree is the container for the tree. It holds a pointer to the root of the tree,
//a list of pointers to the leaf nodes, and the merkle root.
type MerkleTree struct {
	Root       *Node
	merkleRoot []byte
	Leafs      []*Node
}

//Node represents a node, root, or leaf in the tree. It stores pointers to its immediate
//relationships, a hash, the content stored if it is a leaf, and other metadata.
type Node struct {
	Parent *Node
	Left   *Node
	Right  *Node
	leaf   bool
	Hash   []byte
	C      Content
}

//verifyNode walks down the tree until hitting a leaf, calculating the hash at each level
//and returning the resulting hash of Node n.
func (n *Node) verifyNode() []byte {
	if n.leaf {
		return n.C.CalculateHash()
	}
	h := sha3.Keccak256(append(n.Left.verifyNode(), n.Right.verifyNode()...))
	return h
}

//calculateNodeHash is a helper function that calculates the hash of the node.
func (n *Node) calculateNodeHash() []byte {
	if n.leaf {
		return n.C.CalculateHash()
	}
	hash := sha3.Keccak256(append(n.Left.Hash, n.Right.Hash...))
	return hash
}

//NewTree creates a new Merkle Tree using the content cs.
func NewTree(cs []Content) (*MerkleTree, error) {
	root, leafs, err := buildWithContent(cs)
	if err != nil {
		return nil, err
	}
	t := &MerkleTree{
		Root:       root,
		merkleRoot: root.Hash,
		Leafs:      leafs,
	}
	return t, nil
}

//buildWithContent is a helper function that for a given set of Contents, generates a
//corresponding tree and returns the root node, a list of leaf nodes, and a possible error.
//Returns an error if cs contains no Contents.
func buildWithContent(cs []Content) (*Node, []*Node, error) {
	if len(cs) == 0 {
		return nil, nil, errors.New("Error: cannot construct tree with no content.")
	}
	var leafs []*Node
	for _, c := range cs {
		leafs = append(leafs, &Node{
			Hash: c.CalculateHash(),
			C:    c,
			leaf: true,
		})
	}
	root := buildIntermediate(leafs)
	return root, leafs, nil
}

//buildIntermediate is a helper function that for a given list of leaf nodes, constructs
//the intermediate and root levels of the tree. Returns the resulting root node of the tree.
func buildIntermediate(nl []*Node) *Node {
	var nodes []*Node
	numItems := len(nl)
	if numItems == 1 {
		nodes = append(nodes, nl[0])
		return nodes[0]
	}
	for i := 0; i < numItems-(numItems%2); i += 2 {
		var left, right int = i, i + 1
		chash := append(nl[left].Hash, nl[right].Hash...)
		h := sha3.Keccak256(chash)
		n := &Node{
			Left:  nl[left],
			Right: nl[right],
			Hash:  h,
		}
		nodes = append(nodes, n)
		nl[left].Parent = n
		nl[right].Parent = n
		if numItems == 2 {
			return n
		}
	}
	if numItems%2 == 1 {
		left := numItems - 1
		paddingEl := NewPaddingNode()
		paddingNode := &Node{
			Hash: paddingEl.CalculateHash(),
			C:    paddingEl,
			leaf: true,
		}
		chash := append(nl[left].Hash, paddingNode.Hash...)
		h := sha3.Keccak256(chash)
		n := &Node{
			Left:  nl[left],
			Right: paddingNode,
			Hash:  h,
		}
		nodes = append(nodes, n)
		nl[left].Parent = n
		paddingNode.Parent = n
	}
	return buildIntermediate(nodes)
}

//MerkleRoot returns the unverified Merkle Root (hash of the root node) of the tree.
func (m *MerkleTree) MerkleRoot() []byte {
	return m.merkleRoot
}

//RebuildTree is a helper function that will rebuild the tree reusing only the content that
//it holds in the leaves.
func (m *MerkleTree) RebuildTree() error {
	var cs []Content
	for _, c := range m.Leafs {
		cs = append(cs, c.C)
	}
	root, leafs, err := buildWithContent(cs)
	if err != nil {
		return err
	}
	m.Root = root
	m.Leafs = leafs
	m.merkleRoot = root.Hash
	return nil
}

//RebuildTreeWith replaces the content of the tree and does a complete rebuild; while the root of
//the tree will be replaced the MerkleTree completely survives this operation. Returns an error if the
//list of content cs contains no entries.
func (m *MerkleTree) RebuildTreeWith(cs []Content) error {
	root, leafs, err := buildWithContent(cs)
	if err != nil {
		return err
	}
	m.Root = root
	m.Leafs = leafs
	m.merkleRoot = root.Hash
	return nil
}

//VerifyTree verify tree validates the hashes at each level of the tree and returns true if the
//resulting hash at the root of the tree matches the resulting root hash; returns false otherwise.
func (m *MerkleTree) VerifyTree() bool {
	calculatedMerkleRoot := m.Root.verifyNode()
	if bytes.Compare(m.merkleRoot, calculatedMerkleRoot) == 0 {
		return true
	}
	return false
}

//VerifyContent indicates whether a given content is in the tree and the hashes are valid for that content.
//Returns true if the expected Merkle Root is equivalent to the Merkle root calculated on the critical path
//for a given content. Returns true if valid and false otherwise.
func (m *MerkleTree) VerifyContent(expectedMerkleRoot []byte, content Content) bool {
	for _, l := range m.Leafs {
		if l.C.Equals(content) {
			currentParent := l.Parent
			for currentParent != nil {
				if currentParent.Left.leaf && currentParent.Right.leaf {
					h := sha3.Keccak256(append(currentParent.Left.calculateNodeHash(), currentParent.Right.calculateNodeHash()...))
					if bytes.Compare(h, currentParent.Hash) != 0 {
						return false
					}
					currentParent = currentParent.Parent
				} else {
					h := sha3.Keccak256(append(currentParent.Left.calculateNodeHash(), currentParent.Right.calculateNodeHash()...))
					if bytes.Compare(h, currentParent.Hash) != 0 {
						return false
					}
					currentParent = currentParent.Parent
				}
			}
			return true
		}
	}
	return false
}

//VerifyBinaryProof verifies if the content is in the tree
func (m *MerkleTree) VerifyBinaryProof(expectedMerkleRoot []byte, proof []byte, content Content) (bool, error) {
	if len(proof)%33 != 0 {
		return false, errors.New("Invalid proof length")
	}
	numItems := len(proof) / 33
	h := content.CalculateHash()
	if numItems == 0 {
		return bytes.Compare(h, expectedMerkleRoot) == 0, nil
	}
	proofPiece := make([]byte, 32)
	for i := 0; i < numItems; i++ {
		slice := proof[i*33 : (i+1)*33]
		leftOrRight := slice[0]
		copy(proofPiece, slice[1:33])
		if leftOrRight == byte(0x00) { // provided leaf is on the left
			h = sha3.Keccak256(append(proofPiece, h...))
		} else {
			h = sha3.Keccak256(append(h, proofPiece...))
		}
	}
	return bytes.Compare(h, expectedMerkleRoot) == 0, nil
}

//ProvideBinaryProof generates byte encoded proof for the element
func (m *MerkleTree) ProvideBinaryProof(contentIndex int) ([]byte, error) {
	if contentIndex >= len(m.Leafs) {
		return nil, errors.New("Invalid index")
	}
	leaf := m.Leafs[contentIndex]
	if leaf.Parent == nil {
		if bytes.Compare(leaf.C.CalculateHash(), m.merkleRoot) == 0 {
			return []byte{}, nil
		}
		return []byte{}, errors.New("Root hash mismatch")
	}
	proof := []byte{}
	for leaf.Parent != nil {
		if leaf.Parent.Left == leaf {
			proof = append(proof, []byte{0x01}...)
			proof = append(proof, leaf.Parent.Right.Hash...)
		} else if leaf.Parent.Right == leaf {
			proof = append(proof, []byte{0x00}...)
			proof = append(proof, leaf.Parent.Left.Hash...)
		} else {
			return []byte{}, errors.New("Root hash mismatch")
		}
		leaf = leaf.Parent
	}
	return proof, nil
}

func (m *MerkleTree) ProvideBinaryProofForContent(content Content) ([]byte, error) {
	if len(m.Leafs) == 1 {
		if bytes.Compare(content.CalculateHash(), m.merkleRoot) == 0 {
			return []byte{}, nil
		}
	}
	for index, l := range m.Leafs {
		if l.C.Equals(content) {
			return m.ProvideBinaryProof(index)
		}
	}
	return nil, errors.New("Not in the tree")
}

//String returns a string representation of the tree. Only leaf nodes are included
//in the output.
func (m *MerkleTree) String() string {
	s := ""
	for _, l := range m.Leafs {
		s += fmt.Sprint(l)
		s += "\n"
	}
	return s
}
