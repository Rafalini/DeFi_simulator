package main

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
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
	"strings"
	"sync"
	"time"
)

var (
	logFilePrefix                string = "log/"
	blockchainFilePrefix         string = "blockchain/"
	lastBlock                    blockchainDataModel.Block
	miner                        string
	metricsFile                  string
	blockChainFile               string
	localAddr                    string
	ammAdress                    string
	sigVerifyPort                string
	transactionHandlingMulticast string
	blockHandlingMulticast       string
	root                         blockchainDataModel.TreeNode
	rootLock                     sync.Mutex
	miningLock                   sync.Mutex
	queueLock                    sync.Mutex
	breakHashSearch              = true
	pendingTransactions          []blockchainDataModel.Transaction
)

const (
	BlockType       string = "block"
	TransactionType        = "transaction"
	blockSize       int    = 6
)

func setLocalVariables() {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr = conn.LocalAddr().(*net.UDPAddr).String()
	miner, _ = os.Hostname()
	os.Mkdir(logFilePrefix, os.ModePerm)
	os.Mkdir(blockchainFilePrefix, os.ModePerm)
	metricsFile = blockchainFilePrefix + strings.Split(localAddr, ":")[0] + os.Getenv("METRICS_FILE")
	blockChainFile = blockchainFilePrefix + strings.Split(localAddr, ":")[0] + os.Getenv("BLOCK_CHAIN_FILE")
	sigVerifyPort = os.Getenv("SIGNATURE_VERIFY_PORT")
	transactionHandlingMulticast = os.Getenv("TRANSACTION_BROADCAST")
	blockHandlingMulticast = os.Getenv("NODE_BROADCAST")
	ammAdress = os.Getenv("AMM_SERVER_ADDR")
	root = *blockchainDataModel.NewRoot()
}

func printBlock(block blockchainDataModel.Block) {
	fmt.Println("Block:")
	for i, s := range block.Transactions {
		fmt.Println(i, s)
	}
	printHash(block.Hash)
}

func printHash(hash []byte) {
	// for i, v := range hash {
	// 	fmt.Printf("%08b ", v)
	// 	if i > 4 {
	// 		// break
	// 	}
	// }
	// fmt.Println("")
	fmt.Println(fmt.Sprintf("%x", hash))
}

func searchHash() {
	if len(pendingTransactions) < blockSize {
		return
	}

	block := blockchainDataModel.Block{}
	block.Miner = localAddr
	block.PreviousHash = blockchainDataModel.GetDeepestLeave(&root).Block.Hash

	queueLock.Lock()
	block.Transactions = pendingTransactions[:blockSize]
	pendingTransactions = pendingTransactions[blockSize:]
	queueLock.Unlock()

	h := sha256.New()
	var notFound = true
	block.Nonce = rand.Intn(10000)
	var initialVal = block.Nonce
	for notFound {
		h.Reset()
		blockStr := fmt.Sprintf("%#v", block)
		h.Write([]byte(blockStr))
		block.Hash = h.Sum(nil)

		if block.Hash[0]&0xFF == 0 {
			// fmt.Printf("After %d trials\n", block.Nonce-initialVal)
			notFound = false
		} else {
			block.Nonce += 1
		}

	}
	var timeStamp = time.Now()
	block.TimeStamp = timeStamp.Format("2006-01-02T15:04:05.999999999Z07:00")
	lastBlock = block
	var stats metrics.Stats
	stats.Attempts = block.Nonce - initialVal
	stats.Times = 1
	metrics.UpdateStats(stats, metricsFile)
	metrics.SaveBlockChain(&root, blockChainFile)
	// printBlock(block)

	rootLock.Lock()
	blockchainDataModel.AppendBlock(&root, &block)
	rootLock.Unlock()

	broadcastNode(block)
	// n := 1 + rand.Intn(4)
	// time.Sleep(time.Duration(n) * time.Second)
}

func broadcastNode(node blockchainDataModel.Block) {
	// Create a UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", blockHandlingMulticast)
	if err != nil {
		fmt.Println("Error resolving UDP address:", err)
		os.Exit(1)
	}

	// Create a UDP connection
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Println("Error creating UDP connection:", err)
		os.Exit(1)
	}
	defer conn.Close()

	blockStr, _ := json.Marshal(node)
	// fmt.Println("Sending block: ")
	// fmt.Println(blockStr)
	_, err = conn.Write(blockStr)
	if err != nil {
		fmt.Println("Error sending data:", err)
		os.Exit(1)
	}
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
		fmt.Println(transaction.TimeStamp + " " + transaction.Amount + " " + transaction.Token)
		// fmt.Printf("Got Transaction of size: %db\n", n)
		fmt.Printf("Got Transaction from: %s\n", transaction.Sender)
		if validateSignature(transaction) {
			queueLock.Lock()
			var onList = false
			for _, trans := range pendingTransactions {
				if trans.TransactionHash == transaction.TransactionHash {
					onList = true
					fmt.Println("not adding trans")
					break
				}
			}
			if !onList {
				pendingTransactions = append(pendingTransactions, transaction)
				// fmt.Println("Adding trans")
			}
			queueLock.Unlock()
			searchHash()
		}
	}
}

func validateSignature(transaction blockchainDataModel.Transaction) bool {
	if transaction.Sender == ammAdress {
		return true
	}
	urlStr := "http://" + transaction.Sender + ":" + sigVerifyPort + "/get-public-key"

	response, err := http.Get(urlStr)

	if err != nil {
		fmt.Println("GET request failed:", err)
		return false
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		fmt.Println("Error:", response.Status)
		return false
	}

	pemBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("pemBytes failed:", err)
		return false
	}
	block, _ := pem.Decode(pemBytes)
	pubKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		fmt.Println("pubKey failed:", err)
		return false
	}

	sig, _ := hex.DecodeString(transaction.SenderSignature)
	hash, _ := hex.DecodeString(transaction.TransactionHash)

	err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hash, sig)

	if err != nil {
		return false
	} else {
		return true
	}
}

func handleBlocks() {
	multicastAddr, _ := net.ResolveUDPAddr("udp", blockHandlingMulticast)
	conn, _ := net.ListenMulticastUDP("udp", nil, multicastAddr)
	defer conn.Close()

	buffer := make([]byte, 1024)

	fmt.Println("Listening for multicast blocks on", blockHandlingMulticast)

	// Infinite loop to listen for multicast messages
	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading from UDP connection:", err)
			continue
		}

		var block = blockchainDataModel.Block{}
		json.Unmarshal(buffer[:n], &block)
		rootLock.Lock()
		blockchainDataModel.AppendBlock(&root, &block)
		rootLock.Unlock()
	}
}

func main() {
	setLocalVariables()
	go handleTransactions()
	handleBlocks()
}
