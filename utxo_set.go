package main

import (
	"encoding/hex"
	"log"

	"github.com/syndtr/goleveldb/leveldb"
)

const utxoBucket = "chainstate"

// UTXOSet represents UTXO set
type UTXOSet struct {
	Blockchain *Blockchain
}

// getKey returns the key with the chainstate prefix
func getKey(txID string) []byte {
	return []byte(utxoBucket + "_" + txID)
}

// FindSpendableOutputs finds and returns unspent outputs to reference in inputs
func (u UTXOSet) FindSpendableOutputs(pubkeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	db := u.Blockchain.db

	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		key := iter.Key()
		// Check if the key has the chainstate prefix
		if len(key) > len(utxoBucket) && string(key[:len(utxoBucket)]) == utxoBucket {
			txID := string(key[len(utxoBucket)+1:]) // Extract the transaction ID
			outs := DeserializeOutputs(iter.Value())

			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(pubkeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				}
			}
		}
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		log.Panic(err)
	}

	return accumulated, unspentOutputs
}

// FindUTXO finds UTXO for a public key hash
func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	db := u.Blockchain.db

	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		key := iter.Key()
		// Check if the key has the chainstate prefix
		if len(key) > len(utxoBucket) && string(key[:len(utxoBucket)]) == utxoBucket {
			outs := DeserializeOutputs(iter.Value())

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		log.Panic(err)
	}

	return UTXOs
}

// CountTransactions returns the number of transactions in the UTXO set
func (u UTXOSet) CountTransactions() int {
	db := u.Blockchain.db
	counter := 0

	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		key := iter.Key()
		// Check if the key has the chainstate prefix
		if len(key) > len(utxoBucket) && string(key[:len(utxoBucket)]) == utxoBucket {
			counter++
		}
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		log.Panic(err)
	}

	return counter
}

// Reindex rebuilds the UTXO set
func (u UTXOSet) Reindex() {
	db := u.Blockchain.db

	// Clear the existing UTXO set by deleting all keys with the chainstate prefix
	batch := new(leveldb.Batch)
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		key := iter.Key()
		if len(key) > len(utxoBucket) && string(key[:len(utxoBucket)]) == utxoBucket {
			batch.Delete(key)
		}
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		log.Panic(err)
	}

	UTXO := u.Blockchain.FindUTXO()

	for txID, outs := range UTXO {
		key := getKey(txID)
		batch.Put(key, outs.Serialize())
	}

	if err := db.Write(batch, nil); err != nil {
		log.Panic(err)
	}
}

// Update updates the UTXO set with transactions from the Block
// The Block is considered to be the tip of a blockchain
func (u UTXOSet) Update(block *Block) {
	db := u.Blockchain.db

	batch := new(leveldb.Batch)
	for _, tx := range block.Transactions {
		if !tx.IsCoinbase() {
			for _, vin := range tx.Vin {
				key := getKey(hex.EncodeToString(vin.Txid))
				outsBytes, err := db.Get(key, nil)
				if err != nil && err != leveldb.ErrNotFound {
					log.Panic(err)
				}

				updatedOuts := TXOutputs{}
				outs := DeserializeOutputs(outsBytes)
				for outIdx, out := range outs.Outputs {
					if outIdx != vin.Vout {
						updatedOuts.Outputs = append(updatedOuts.Outputs, out)
					}
				}

				if len(updatedOuts.Outputs) == 0 {
					batch.Delete(key)
				} else {
					batch.Put(key, updatedOuts.Serialize())
				}
			}
		}

		newOutputs := TXOutputs{}
		for _, out := range tx.Vout {
			newOutputs.Outputs = append(newOutputs.Outputs, out)
		}

		key := getKey(hex.EncodeToString(tx.ID))
		batch.Put(key, newOutputs.Serialize())
	}

	if err := db.Write(batch, nil); err != nil {
		log.Panic(err)
	}
}