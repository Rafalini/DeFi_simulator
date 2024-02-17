package main

import (
	"crypto"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"main/blockchainDataModel"
	"main/metrics"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	logFilePrefix                string  = "log/"
	keyChainFilePrefix           string  = "key/"
	pendingTrans                 int     = 0
	pendingValue                 float64 = 0
	balanceMap                   map[string]float64
	sigVerifyPort                string
	ammAdress                    string
	ammPort                      string
	oracleAdress                 string
	transactionBroadcastAddr     string
	localAddr                    string
	recievedTransactionHashes    []string
	transactionHandlingMulticast string
	logFile                      string
	privateKey                   *rsa.PrivateKey
	publicKey                    rsa.PublicKey
	startTime                    time.Time
	mu                           sync.Mutex
)

type RatesResponse struct {
	// BTC  map[string]float64 `json:"BTC"`
	ETH  map[string]float64 `json:"ETH"`
	MKR  map[string]float64 `json:"MKR"`
	XAUt map[string]float64 `json:"XAUt"`
	Time string             `json:"time"`
}

func setLocalVariables() {
	startTime = time.Now()
	balanceMap = make(map[string]float64)
	// balanceMap["BTC"] = 3000
	balanceMap["ETH"] = 300
	balanceMap["XAUt"] = 300
	balanceMap["MKR"] = 300

	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	os.Mkdir(logFilePrefix, os.ModePerm)
	os.Mkdir(keyChainFilePrefix, os.ModePerm)
	localAddr = strings.Split(conn.LocalAddr().(*net.UDPAddr).String(), ":")[0]
	transactionBroadcastAddr = os.Getenv("TRANSACTION_BROADCAST")
	ammAdress = os.Getenv("AMM_SERVER_ADDR")
	ammPort = os.Getenv("AMM_SERVER_PORT")
	oracleAdress = os.Getenv("ORACLE_SERVER_ADDR") + ":" + os.Getenv("ORACLE_SERVER_PORT")
	privateKey, publicKey = blockchainDataModel.GenerateKeyPairAndReturn(keyChainFilePrefix + localAddr)
	sigVerifyPort = os.Getenv("SIGNATURE_VERIFY_PORT")
	transactionHandlingMulticast = os.Getenv("TRANSACTION_BROADCAST")
	logFile = logFilePrefix + strings.Split(localAddr, ":")[0] + os.Getenv("METRICS_FILE")
	metrics.PrepareLog(balanceMap, getPrices(), logFile)
}

func getCurrentVal() float64 {
	prices := getPrices()
	sum := 0.0
	for key := range balanceMap {
		sum += balanceMap[key] * prices[key]
	}
	return sum
	// return sum + pendingValue
}

func getPrices() map[string]float64 {
	url := "http://" + oracleAdress + "/get-prices" // Replace with your API endpoint URL

	// Make the GET request
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error making GET request:", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Unexpected status code:", resp.StatusCode)
		return nil
	}

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil
	}

	var data []map[string]interface{}

	// Unmarshal JSON into a slice of maps
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	// Create a map to store symbol as key and usdprice as value
	symbolPriceMap := make(map[string]float64)

	// Iterate through the array and populate the map
	for _, item := range data {
		if symbol, ok := item["symbol"].(string); ok {
			if usdPrice, ok := item["usdprice"].(float64); ok {
				symbolPriceMap[symbol] = usdPrice
			}
		}
	}
	// fmt.Println(symbolPriceMap)
	return symbolPriceMap
}

func getRates() map[string]map[string]float64 {
	url := "http://" + ammAdress + ":" + ammPort + "/get-rates" // Replace with your API endpoint URL

	// Make the GET request
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error making GET request:", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Unexpected status code:", resp.StatusCode)
		return nil
	}

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil
	}

	var data RatesResponse

	// Unmarshal the JSON into the struct
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return nil
	}

	ratesMap := make(map[string]map[string]float64)
	// ratesMap["BTC"] = data.BTC
	ratesMap["ETH"] = data.ETH
	ratesMap["MKR"] = data.MKR
	ratesMap["XAUt"] = data.XAUt

	return ratesMap
}

func handleTransactions() {
	addr, _ := net.ResolveUDPAddr("udp", transactionHandlingMulticast)
	conn, _ := net.ListenMulticastUDP("udp", nil, addr)

	defer conn.Close()

	buffer := make([]byte, 1024)

	fmt.Println("Listening for multicast transactions on", transactionHandlingMulticast)

	// Infinite loop to listen for multicast messages
	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading from UDP connection:", err)
			continue
		}

		var transaction = blockchainDataModel.Transaction{}
		json.Unmarshal(buffer[:n], &transaction)

		if !hasTransactionBeenUsed(transaction.TransactionHash) && transaction.Reciever == localAddr {
			// fmt.Printf("Recieved transaction: " + transaction.Token + " " + transaction.Amount)
			mu.Lock()
			val, _ := strconv.ParseFloat(transaction.Amount, 64)
			balanceMap[transaction.Token] += val

			priceMap := getPrices()
			pendingValue -= val * priceMap[transaction.Token]

			pendingTrans -= 1
			recievedTransactionHashes = append(recievedTransactionHashes, transaction.TransactionHash)
			mu.Unlock()
		}
	}
}

func hasTransactionBeenUsed(target string) bool {
	for _, element := range recievedTransactionHashes {
		if element == target {
			return true // Found the target string in the list
		}
	}
	return false // Target string not found in the list
}

func broadcastTransaction(transaction blockchainDataModel.Transaction) {
	udpAddr, err := net.ResolveUDPAddr("udp", transactionBroadcastAddr)
	if err != nil {
		fmt.Println("Error resolving UDP address:", err)
		os.Exit(1)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Println("Error creating UDP connection:", err)
		os.Exit(1)
	}
	defer conn.Close()

	blockStr, _ := json.Marshal(transaction)
	_, err = conn.Write(blockStr)
	if err != nil {
		fmt.Println("Error sending data:", err)
		os.Exit(1)
	}
}

func calculateMaxGain() (string, string, float64, float64) {
	maxToken := ""
	maxEchangeToken := ""
	maxGain := 100.0

	priceMap := getPrices() // key - value
	rateMap := getRates()   // key - (key - value)

	for token := range priceMap {
		for referenceToken := range priceMap {
			if token == referenceToken {
				continue
			}

			gain := priceMap[token]*balanceMap[token] - balanceMap[referenceToken]*priceMap[referenceToken]*rateMap[token][referenceToken]
			// gain := priceMap[token]*balanceMap[token] - balanceMap[referenceToken]*priceMap[referenceToken]
			if maxGain < gain {
				maxGain = gain
				maxToken = token
				maxEchangeToken = referenceToken
			}

		}
	}
	min := 1.0
	max := balanceMap[maxToken] / 4
	amount := min + rand.Float64()*(max-min)
	pendingValue += amount * priceMap[maxToken] * rateMap[maxToken][maxEchangeToken]
	return maxEchangeToken, maxToken, amount, rateMap[maxToken][maxEchangeToken]
}

func trade() {
	duration, _ := strconv.Atoi(os.Getenv("DURATION"))
	for time.Since(startTime) < time.Duration(duration-5)*time.Second {
		// fmt.Printf("Current val: %.2f\n", getCurrentVal())
		mu.Lock()
		exchangeToken, token, amount, rate := calculateMaxGain()
		if token != "" && exchangeToken != "" {
			balanceMap[token] -= amount

			var newTransaction = blockchainDataModel.Transaction{}
			newTransaction.Sender = localAddr
			newTransaction.Reciever = ammAdress
			var timeStamp = time.Now()
			newTransaction.TimeStamp = timeStamp.Format("2006-01-02T15:04:00.000Z")
			newTransaction.Amount = fmt.Sprintf("%f", amount)
			newTransaction.Token = token
			newTransaction.Metadata.ExchangeToken = exchangeToken
			newTransaction.Metadata.MaxSlippage = fmt.Sprintf("%f", 60.0)
			newTransaction.Metadata.ExchangeRate = fmt.Sprintf("%f", rate)

			transactionbytes, _ := json.Marshal(newTransaction)
			transactionhash := blockchainDataModel.ReturnHash(transactionbytes)
			newTransaction.TransactionHash = hex.EncodeToString(transactionhash)
			sig, _ := rsa.SignPKCS1v15(crand.Reader, privateKey, crypto.SHA256, transactionhash)
			mu.Unlock()
			newTransaction.SenderSignature = hex.EncodeToString(sig)

			// fmt.Println("Exchange of: " + newTransaction.Token + " for " + newTransaction.Metadata.ExchangeToken + " of " + newTransaction.Amount)

			broadcastTransaction(newTransaction)
			pendingTrans += 1
		}

		// min, max := 0.1, 0.8
		time.Sleep(time.Duration(250 * time.Millisecond))
	}
	fmt.Println("Exiting trade loop")
	finalizeTransactions()
	time.Sleep(5 * time.Second)
}

func finalizeTransactions() {
	for i := 0; i < 5; i++ {
		token, exchangeToken, _, _ := calculateMaxGain()
		var newTransaction = blockchainDataModel.Transaction{}
		newTransaction.Sender = localAddr
		newTransaction.Reciever = ammAdress
		var timeStamp = time.Now()
		newTransaction.TimeStamp = timeStamp.Format("2006-01-02T15:04:00.000Z")
		newTransaction.Amount = fmt.Sprintf("%f", 0.0)
		newTransaction.Token = token
		newTransaction.Metadata.ExchangeToken = exchangeToken
		newTransaction.Metadata.MaxSlippage = fmt.Sprintf("%f", 999.0)
		newTransaction.Metadata.ExchangeRate = fmt.Sprintf("%f", 1.0)

		transactionbytes, _ := json.Marshal(newTransaction)
		transactionhash := blockchainDataModel.ReturnHash(transactionbytes)
		newTransaction.TransactionHash = hex.EncodeToString(transactionhash)
		sig, _ := rsa.SignPKCS1v15(crand.Reader, privateKey, crypto.SHA256, transactionhash)

		newTransaction.SenderSignature = hex.EncodeToString(sig)

		broadcastTransaction(newTransaction)
		pendingTrans += 1

		time.Sleep(time.Duration(pendingTrans) * 500 * time.Millisecond)
	}
}

func getKeysExcept(token string) []string {
	keys := make([]string, 0, len(balanceMap)-1)
	for key := range balanceMap {
		if key != token {
			keys = append(keys, key)
		}
	}
	shuffleKeys(keys)
	return keys
}

func shuffleKeys(keys []string) {
	rand.Shuffle(len(keys), func(i, j int) {
		keys[i], keys[j] = keys[j], keys[i]
	})
}

func publicKeyHandler(w http.ResponseWriter, r *http.Request) {
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&publicKey),
	})
	w.Header().Set("Content-Type", "application/x-pem-file") // Set the appropriate content type
	w.Write(pemBytes)
}

func httpHandler() {
	http.HandleFunc("/get-public-key", publicKeyHandler)
	fmt.Println("Server is running: /get-public-key")
	err := http.ListenAndServe(":"+sigVerifyPort, nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func periodicSave() {
	duration, _ := strconv.Atoi(os.Getenv("DURATION"))
	for time.Since(startTime) < time.Duration(duration)*time.Second {
		metrics.UpdateLog(startTime, getCurrentVal(), balanceMap, getPrices(), logFile)
		time.Sleep(1 * time.Second)
	}
}

func main() {
	delay, _ := strconv.Atoi(os.Getenv("DELAY"))
	time.Sleep(time.Duration(delay) * time.Second)
	setLocalVariables()
	go httpHandler() //public key
	go handleTransactions()
	go trade()
	periodicSave()
}
