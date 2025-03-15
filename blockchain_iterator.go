package main

import (
	"log"
	"github.com/syndtr/goleveldb/leveldb"
)

// BlockchainIterator is used to iterate over blockchain blocks
type BlockchainIterator struct {
	currentHash []byte
	db          *leveldb.DB
}

// Next returns next block starting from the tip
func (i *BlockchainIterator) Next() *Block {
	encodedBlock, err := i.db.Get(i.currentHash, nil)
	if err != nil {
		log.Panic(err)
	}

	block := DeserializeBlock(encodedBlock)
	i.currentHash = block.PrevBlockHash

	return block
}
