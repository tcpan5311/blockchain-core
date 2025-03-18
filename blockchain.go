package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/syndtr/goleveldb/leveldb"
)

const dbFile = "blockchain.db"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

// Blockchain implements interactions with a DB
type Blockchain struct {
	tip []byte
	db  *leveldb.DB
}

// CreateBlockchain creates a new blockchain DB
func CreateBlockchain(address string) *Blockchain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
	genesis := NewGenesisBlock(cbtx)
	db, err := leveldb.OpenFile(dbFile, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Put(genesis.Hash, genesis.Serialize(), nil)
	if err != nil {
		log.Panic(err)
	}
	err = db.Put([]byte("l"), genesis.Hash, nil)
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{genesis.Hash, db}
	return &bc
}

// NewBlockchain creates a new Blockchain with genesis Block
func NewBlockchain() *Blockchain {
	if !dbExists() {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}
	db, err := leveldb.OpenFile(dbFile, nil)
	if err != nil {
		log.Panic(err)
	}
	tip, err := db.Get([]byte("l"), nil)
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}
	return &bc
}

// FindTransaction finds a transaction by its ID
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Equal(tx.ID, ID) {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction is not found")
}

// FindUTXO finds all unspent transaction outputs
func (bc *Blockchain) FindUTXO() map[string]TXOutputs {
	UTXO := make(map[string]TXOutputs)
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}

				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return UTXO
}

func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{bc.tip, bc.db}
}


// MineBlock mines a new block with the provided transactions
func (bc *Blockchain) MineBlock(transactions []*Transaction, stake int64) *Block {
	var lastHash []byte

	for _, tx := range transactions {
		if !bc.VerifyTransaction(tx) {
			log.Panic("ERROR: Invalid transaction")
		}
	}

	lastHash, err := bc.db.Get([]byte("l"), nil)
	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, lastHash, stake)

	err = bc.db.Put(newBlock.Hash, newBlock.Serialize(), nil)
	if err != nil {
		log.Panic(err)
	}
	err = bc.db.Put([]byte("l"), newBlock.Hash, nil)
	if err != nil {
		log.Panic(err)
	}

	bc.tip = newBlock.Hash
	return newBlock
}

// SignTransaction signs inputs of a Transaction
func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

// VerifyTransaction verifies transaction input signatures
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}

func dbExists() bool {
	_, err := os.Stat(dbFile)
	return !os.IsNotExist(err)
}
