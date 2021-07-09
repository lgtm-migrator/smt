package smt

import (
	"bytes"
	"errors"
	"hash"
)

// ErrBadProof is returned when an invalid Merkle proof is supplied.
var ErrBadProof = errors.New("bad proof")

// DeepSparseMerkleSubTree is a deep Sparse Merkle subtree for working on only a few leafs.
type DeepSparseMerkleSubTree struct {
	*SparseMerkleTree
}

// NewDeepSparseMerkleSubTree creates a new deep Sparse Merkle subtree on an empty MapStore.
func NewDeepSparseMerkleSubTree(nodes, values MapStore, hasher hash.Hash, root []byte) *DeepSparseMerkleSubTree {
	return &DeepSparseMerkleSubTree{
		SparseMerkleTree: ImportSparseMerkleTree(nodes, values, hasher, root),
	}
}

// AddBranch adds a branch to the tree.
// These branches are generated by smt.ProveForRoot.
// If the proof is invalid, a ErrBadProof is returned.
//
// If the leaf may be updated (e.g. during a state transition fraud proof),
// an updatable proof should be used. See SparseMerkleTree.ProveUpdatable.
func (dsmst *DeepSparseMerkleSubTree) AddBranch(proof SparseMerkleProof, key []byte, value []byte) error {
	result, updates := verifyProofWithUpdates(proof, dsmst.Root(), key, value, dsmst.th.hasher)
	if !result {
		return ErrBadProof
	}

	if !bytes.Equal(value, defaultValue) { // Membership proof.
		if err := dsmst.values.Set(dsmst.th.path(key), value); err != nil {
			return err
		}
	}

	// Update nodes along branch
	for _, update := range updates {
		err := dsmst.nodes.Set(update[0], update[1])
		if err != nil {
			return err
		}
	}

	// Update sibling node
	if proof.SiblingData != nil {
		if proof.SideNodes != nil && len(proof.SideNodes) > 0 {
			err := dsmst.nodes.Set(proof.SideNodes[0], proof.SiblingData)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
