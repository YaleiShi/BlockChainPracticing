package data

import (
	"../../p1"
	"../../p2"
	"fmt"
	"sync"
)

type SyncBlockChain struct {
	bc  p2.BlockChain
	mux sync.Mutex
}

func NewBlockChain() SyncBlockChain {
	var blockChain p2.BlockChain
	blockChain.Initial()
	return SyncBlockChain{bc: blockChain}
}

func (sbc *SyncBlockChain) Get(height int32) ([]p2.Block, bool) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.Get(height)
}

func (sbc *SyncBlockChain) GetBlock(height int32, hash string) (p2.Block, bool) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()

	var block p2.Block
	blocks, ok := sbc.bc.Get(height)
	if !ok {
		return block, ok
	}

	for _, v := range blocks {
		if v.Header.Hash == hash {
			return v, true
		}
	}
	return block, false
}

func (sbc *SyncBlockChain) Insert(block p2.Block) {
	sbc.mux.Lock()
	sbc.bc.Insert(block)
	sbc.mux.Unlock()
}

func (sbc *SyncBlockChain) CheckParentHash(insertBlock p2.Block) bool {
	// check the insertBlock's parent block is existed
	sbc.mux.Lock()
	defer sbc.mux.Unlock()

	height := insertBlock.Header.Height
	pHeight := height - 1
	pHash := insertBlock.Header.ParentHash

	blocks, ok := sbc.bc.Get(pHeight)
	if !ok {
		return false
	}

	for _, v := range blocks {
		if v.Header.Hash == pHash {
			return true
		}
	}

	return false
}

func (sbc *SyncBlockChain) UpdateEntireBlockChain(blockChainJson string) {
	sbc.mux.Lock()

	sbc.bc.DecodeFromJSON(blockChainJson)

	sbc.mux.Unlock()
}

func (sbc *SyncBlockChain) BlockChainToJson() (string, error) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.EncodeToJson()
}

func (sbc *SyncBlockChain) GenBlock(mpt p1.MerklePatriciaTrie) p2.Block {
	// generate next block
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	bc := sbc.bc
	length := bc.Length
	var block p2.Block
	if length == 0 {
		block.Initial(1, "genesis", mpt)
		sbc.bc.Insert(block)
		return block
	}
	bcs, _ := bc.Get(length)
	parentHash := bcs[0].Header.Hash
	block.Initial(length+1, parentHash, mpt)
	sbc.bc.Insert(block)
	return block
}

func PrintError(e error, s string) {
	fmt.Print(s, " : ", e)
}

func (sbc *SyncBlockChain) Show() string {
	return sbc.bc.Show()
}
