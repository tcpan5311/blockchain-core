package main

import (
	"fmt"
	"math/big"
	"bytes"
	"crypto/sha256"
)

// ProofOfStake represents a proof-of-stake
type ProofOfStake struct {
	block *Block
}

// NewProofOfStake creates and returns a ProofOfStake
func NewProofOfStake(b *Block) *ProofOfStake {
	pos := &ProofOfStake{b}
	return pos
}

func (pos *ProofOfStake) prepareData() []byte {
	data := bytes.Join(
		[][]byte{
			pos.block.PrevBlockHash,
			pos.block.HashTransactions(),
			IntToHex(pos.block.Timestamp),
		},
		[]byte{},
	)

	return data
}

// Run performs a placeholder proof-of-stake validation
func (pos *ProofOfStake) Run() []byte {

	var hashInt big.Int
	var hash [32]byte
	
	// Placeholder validation logic: consider stake > threshold as valid
	data := pos.prepareData()
	hash = sha256.Sum256(data)

	fmt.Printf("\r%x", hash)
	hashInt.SetBytes(hash[:])

	// Ensure stake meets the threshold
	if pos.block.Stake < 50 {
		fmt.Println("Block rejected due to insufficient stake")
	}

	// If threshold is above 50, always return true
	if pos.block.Stake > 50 {
		fmt.Println("Block created successfully")
	}
	
	return hash[:]
}

// Validate checks if the block follows PoS rules
func (pos *ProofOfStake) Validate() bool {
    var hashInt big.Int

    data := pos.prepareData()
    hash := sha256.Sum256(data)
    hashInt.SetBytes(hash[:])

    // Ensure stake meets the threshold
    if pos.block.Stake < 50 {
        return false
    }

    // If threshold is above 50, always return true
    if pos.block.Stake > 50 {
        return true
    }

	return false

}

