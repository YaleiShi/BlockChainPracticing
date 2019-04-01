package data

import (
	"../../p1"
	"math/rand"
)

type HeartBeatData struct {
	IfNewBlock  bool   `json:"ifNewBlock"`
	Id          int32  `json:"id"`
	BlockJson   string `json:"blockJson"`
	PeerMapJson string `json:"peerMapJson"`
	Addr        string `json:"addr"`
	Hops        int32  `json:"hops"`
}

func NewHeartBeatData(ifNewBlock bool, id int32, blockJson string, peerMapJson string, addr string) HeartBeatData {
	// normal initial function which create an instance
	return HeartBeatData{ifNewBlock, id, blockJson, peerMapJson, addr, 3}
}

func PrepareHeartBeatData(sbc *SyncBlockChain, selfId int32, peerMapJson string, addr string) HeartBeatData {
	// create a new instance of heart beat data
	// decide whether or not create a new block and send to others
	ifNew := decideIfNew()
	var blockJson string
	if ifNew {
		mpt := DefaultMPT()
		newBlock := sbc.GenBlock(mpt)
		blockJson = newBlock.EncodeToJSON()
	}

	return NewHeartBeatData(ifNew, selfId, blockJson, peerMapJson, addr)
}

func decideIfNew() bool {
	return rand.Float32() < 0.5
}

func DefaultMPT() p1.MerklePatriciaTrie {
	mpt := new(p1.MerklePatriciaTrie)
	mpt.Initial()
	mpt.Insert("p", "apple")
	mpt.Insert("aa", "banana")
	mpt.Insert("ap", "orange")
	return *mpt
}
