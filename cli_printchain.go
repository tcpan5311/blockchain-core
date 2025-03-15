package main
import (
	"fmt"
	"strconv"
)

func (cli *CLI) printChain() {
	bc := NewBlockchain("")
	defer bc.db.Close()

	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("============ Block %x ============\n", block.Hash)
		fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)
		pos := NewProofOfStake(block)
		fmt.Printf("Stake: %d\n", block.Stake)
		fmt.Printf("PoS: %s\n\n", strconv.FormatBool(pos.Validate()))
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Printf("\n\n")

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}