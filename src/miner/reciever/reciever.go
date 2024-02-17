package reciever

import (
	"encoding/json"
	"main/blockchain"
	"os"
)

var (
	blocks []blockchain.Block
)

func SaveLocalBlockchain() {
	jsons, _ := json.Marshal(blocks)
	_ = os.WriteFile("blockchain.json", jsons, 0644)
}

func addBlock(block blockchain.Block) {

}

func getLastBlockHash() []byte {
	return blocks[len(blocks)-1].Hash
}
