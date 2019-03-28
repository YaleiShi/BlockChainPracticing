package p2

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/sha3"
	"sort"
)

type BlockChain struct {
	Chain  map[int32][]Block
	Length int32
}

func (bc *BlockChain) Get(height int32) ([]Block, bool) {
	if blocks, ok := bc.Chain[height]; ok {
		return blocks, ok
	}
	return nil, false
}

func (bc *BlockChain) Insert(b Block) {
	height := b.Header.Height
	if blocks, ok := bc.Chain[height]; ok {
		fmt.Println("append it !!!")
		hash := b.Header.Hash
		for _, v := range blocks {
			if v.Header.Hash == hash {
				fmt.Println("there is a same!!!")
				return
			}
		}
		bc.Chain[height] = append(bc.Chain[height], b)
	} else {
		fmt.Println("new it !!!")
		bc.Chain[height] = []Block{b}
		if height > bc.Length {
			bc.Length = height
		}
	}
}

func (bc *BlockChain) EncodeToJson() (string, error) {
	chain := bc.Chain
	var jsonBlocks []BlockJson
	for _, blocks := range chain {
		for _, block := range blocks {
			jsonBlocks = append(jsonBlocks, block.ToBlockJson())
		}
	}
	res, err := json.Marshal(jsonBlocks)
	return string(res), err
}

func (bc *BlockChain) DecodeFromJSON(js string) (BlockChain, error) {
	var bjs []BlockJson
	err := json.Unmarshal([]byte(js), &bjs)
	fmt.Println("length: ", len(bjs))
	for _, bj := range bjs {
		block := BlockJSONToBlock(bj)
		bc.Insert(block)
	}
	//fmt.Println("height 1: ", len(bc.Chain[1]))
	//fmt.Println("height 2: ", len(bc.Chain[2]))
	//fmt.Println("height 3: ", len(bc.Chain[3]))
	return *bc, err
}

func (bc *BlockChain) Initial() {
	bc.Chain = make(map[int32][]Block)
	bc.Length = 0
}

func (bc *BlockChain) Show() string {
	rs := ""
	var idList []int
	for id := range bc.Chain {
		idList = append(idList, int(id))
	}
	sort.Ints(idList)
	for _, id := range idList {
		var hashs []string
		for _, block := range bc.Chain[int32(id)] {
			hashs = append(hashs, block.Header.Hash+"<="+block.Header.ParentHash)
		}
		sort.Strings(hashs)
		rs += fmt.Sprintf("%v: ", id)
		for _, h := range hashs {
			rs += fmt.Sprintf("%s, ", h)
		}
		rs += "\n"
	}
	sum := sha3.Sum256([]byte(rs))
	rs = fmt.Sprintf("This is the BlockChain: %s\n", hex.EncodeToString(sum[:])) + rs
	return rs
}
