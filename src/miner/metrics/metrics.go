package metrics

import (
	"encoding/json"
	"main/blockchainDataModel"
	"os"
)

type Stats struct {
	Avg_attempts float32
	Times        int
	Attempts     int
}

func UpdateStats(stats Stats, filename string) {

	_, error := os.Stat(filename)
	// check if error is "file not exists"
	if !os.IsNotExist(error) {
		byteValue, _ := os.ReadFile(filename)
		var result Stats
		json.Unmarshal([]byte(byteValue), &result)

		stats.Attempts += result.Attempts
		stats.Times += result.Times
		stats.Avg_attempts = float32(stats.Attempts) / float32(stats.Times)
	}

	rankingsJson, _ := json.Marshal(stats)
	_ = os.WriteFile(filename, rankingsJson, 0644)

}

func SaveBlockChain(root *blockchainDataModel.TreeNode, filename string) {
	f, err := os.Create(filename)
	check(err)
	root.SaveToFile(f, 0)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
