package blockchainDataModel

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

type Metadata struct {
	ExchangeToken string
	MaxSlippage   string
	ExchangeRate  string
}

type Transaction struct {
	TimeStamp       string
	Sender          string
	Reciever        string
	Amount          string
	Token           string
	Metadata        Metadata
	TransactionHash string
	SenderSignature string
}

type Block struct {
	TimeStamp    string
	Miner        string
	Hash         []byte
	PreviousHash []byte
	Nonce        int
	Transactions []Transaction
}

type TreeNode struct {
	Block    Block
	Length   int
	Children []*TreeNode
}

func (b *Block) Compare(block Block) bool {
	return bytes.Equal(b.Hash, block.Hash)
}

func (n *TreeNode) Print(lvl int) {
	fmt.Print(strings.Repeat(" ", lvl))
	fmt.Println(n.Block.Miner)
	for _, child := range n.Children {
		fmt.Print(strings.Repeat(" ", lvl))
		child.Print(lvl + 2)
	}
}

func (n *TreeNode) SaveToFile(f *os.File, lvl int) {
	f.WriteString(strings.Repeat(" ", lvl))
	str := ""
	if len(n.Block.Transactions) == 0 {
		str = "root"
	} else {
		str = NodeToString(n.Block)
	}
	f.WriteString(str + "\n")

	for _, child := range n.Children {
		f.WriteString(strings.Repeat(" ", lvl))
		child.SaveToFile(f, lvl+2)
	}
}

func NodeToString(block Block) string {
	outstr := fmt.Sprintf("%#v", fmt.Sprintf("%x", block.Hash))

	for i, s := range block.Transactions {
		outstr += fmt.Sprintf("%d : %x", i, s)
	}

	return outstr
}

func NewRoot() *TreeNode {
	var treeNode = TreeNode{}
	treeNode.Block.Miner = "root"
	treeNode.Length = 0
	return &treeNode
}

func (n *TreeNode) AddChild(child *TreeNode) {
	n.Children = append(n.Children, child)
}

func AppendBlock(root *TreeNode, child *Block) bool {
	if bytes.Equal(root.Block.Hash, child.Hash) {
		return true
	} else if bytes.Equal(root.Block.Hash, child.PreviousHash) {

		hasThisNode := false
		for _, rootChild := range root.Children {
			if bytes.Equal(rootChild.Block.Hash, child.Hash) {
				hasThisNode = true
			}
		}

		if !hasThisNode {
			var newNode = TreeNode{}
			newNode.Block = *child
			newNode.Length = root.Length + 1
			root.AddChild(&newNode)
			return true
		}
	} else {
		for _, rootChild := range root.Children {
			if AppendBlock(rootChild, child) {
				return true
			}
		}
	}
	return false
}

func GetDeepestLeave(n *TreeNode) *TreeNode {

	var maxChild = n

	for _, child := range n.Children {

		var maxLocal = GetDeepestLeave(child)

		if maxLocal.Length > maxChild.Length {
			maxChild = maxLocal
		}
	}

	return maxChild
}

// func GetHash(block Block) []byte {
// 	h := sha256.New()
// 	block_str, _ := json.Marshal(block)

// 	h.Write([]byte(block_str))
// 	return h.Sum(nil)
// }

// func ParseTransaction(input []byte) Transaction {
// 	data := Transaction{}
// 	json.Unmarshal(input, &data)
// 	return data
// }

// func ParseBlock(input []byte) Block {
// 	data := Block{}
// 	json.Unmarshal(input, &data)
// 	return data
// }
