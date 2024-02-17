package metrics

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"
)

func PrepareLog(balances map[string]float64, prices map[string]float64, fileName string) {
	file, err := os.Create(fileName)
	check(err)
	writer := csv.NewWriter(file)
	s0 := []string{"Timestamp"}
	s1 := mapKeysToArray(balances)
	s2 := mapKeysToValueArray(prices)
	s3 := []string{"USDvalue"}

	s0 = append(s0, s1...)
	s0 = append(s0, s2...)
	s0 = append(s0, s3...)

	err = writer.Write(s0)
	check(err)
	defer writer.Flush()
}

func mapKeysToArray(m map[string]float64) []string {
	var keys []string

	// Iterate over the map and append keys to the slice
	for key := range m {
		keys = append(keys, key)
	}

	return keys
}

func mapKeysToValueArray(m map[string]float64) []string {
	var keys []string

	// Iterate over the map and append keys to the slice
	for key := range m {
		keys = append(keys, key+"_value")
	}

	return keys
}

func mapToArr(m map[string]float64) []string {
	var keys []string
	for _, value := range m {
		keys = append(keys, fmt.Sprintf("%f", value))

	}

	return keys
}

func UpdateLog(startTime time.Time, value float64, balances map[string]float64, prices map[string]float64, fileName string) {

	_, err := os.Stat(fileName)
	check(err)

	var file *os.File
	file, err = os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	check(err)
	defer file.Close()

	// Create a CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Get current timestamp
	// timestamp := time.Now().Format("2006-01-02 15:04:05")
	elapsedTime := time.Since(startTime).Milliseconds()
	s0 := []string{fmt.Sprintf("%d", elapsedTime)}
	s1 := []string{fmt.Sprintf("%f", value)}
	s0 = append(s0, mapToArr(balances)...)
	s0 = append(s0, mapToArr(prices)...)
	data := append(s0, s1...)

	// Prepare data to write to CSV
	// data := []string{timestamp, fmt.Sprintf("%s, %s, %.2f", balancesString, pricesString, sum)}

	// Append data to the CSV file
	if err := writer.Write(data); err != nil {
		log.Fatal("Error writing data to CSV:", err)
	}

	// Flush data to ensure it is written immediately
	writer.Flush()
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
