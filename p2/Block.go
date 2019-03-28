package p2

import (
	"../p1"
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"golang.org/x/crypto/sha3"
	"time"
)

type Block struct {
	Header Header
	Value  p1.MerklePatriciaTrie `json:"value"`
}

type Header struct {
	Height     int32
	TimeStamp  int64
	Hash       string
	ParentHash string
	Size       int32
}

type BlockJson struct {
	Height     int32             `json:"height"`
	Timestamp  int64             `json:"timeStamp"`
	Hash       string            `json:"hash"`
	ParentHash string            `json:"parentHash"`
	Size       int32             `json:"size"`
	MPT        map[string]string `json:"mpt"`
}

func (b *Block) Initial(height int32, parentHash string, value p1.MerklePatriciaTrie) {
	b.Header.Height = height
	b.Header.TimeStamp = time.Now().Unix()
	b.Header.ParentHash = parentHash
	b.Value = value

	b.Header.Size = getMPTLength(value)
	str := string(b.Header.Height) + string(b.Header.TimeStamp) + b.Header.ParentHash + b.Value.GetRoot() + string(b.Header.Size)
	b.Header.Hash = HashBlock(str)
}

func HashBlock(str string) string {
	sum := sha3.Sum256([]byte(str))
	return hex.EncodeToString(sum[:])
}

func DecodeFromJson(js string) Block {
	var bj BlockJson
	json.Unmarshal([]byte(js), &bj)
	return BlockJSONToBlock(bj)
}

func BlockJSONToBlock(bj BlockJson) Block {
	var block Block
	var mpt p1.MerklePatriciaTrie
	mpt.Initial()
	mptKV := bj.MPT
	for key, value := range mptKV {
		mpt.Insert(key, value)
	}

	block.Value = mpt
	block.Header.Hash = bj.Hash
	block.Header.Size = bj.Size
	block.Header.TimeStamp = bj.Timestamp
	block.Header.ParentHash = bj.ParentHash
	block.Header.Height = bj.Height
	return block
}

func (b *Block) EncodeToJSON() string {
	bj := b.ToBlockJson()
	j, _ := json.Marshal(bj)
	return string(j)
}

func (b *Block) ToBlockJson() BlockJson {
	var bj BlockJson
	bj.Height = b.Header.Height
	bj.Timestamp = b.Header.TimeStamp
	bj.Hash = b.Header.Hash
	bj.ParentHash = b.Header.ParentHash
	bj.Size = b.Header.Size
	bj.MPT = b.Value.GetLeafMap()
	return bj
}

func getMPTLength(data p1.MerklePatriciaTrie) int32 {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(data)
	if err != nil {
		return 0
	}
	return int32(len(buf.Bytes()))
}
