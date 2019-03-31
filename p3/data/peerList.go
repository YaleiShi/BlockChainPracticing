package data

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"sync"
)

type PeerList struct {
	selfId    int32
	peerMap   map[string]int32
	maxLength int32
	mux       sync.Mutex
}

type PeerJson struct {
	SelfId    int32            `json:"selfId"`
	MaxLength int32            `json:"maxLength"`
	PeerMap   map[string]int32 `json:"peerMap"`
}

func NewPeerList(id int32, maxLength int32) PeerList {
	//initial function
	peerMap := make(map[string]int32)
	return PeerList{selfId: id, peerMap: peerMap, maxLength: maxLength}
}

func (peers *PeerList) Add(addr string, id int32) {
	peers.mux.Lock()
	if id != peers.selfId {
		peers.peerMap[addr] = id
	}
	peers.mux.Unlock()
}

func (peers *PeerList) Delete(addr string) {
	peers.mux.Lock()
	delete(peers.peerMap, addr)
	peers.mux.Unlock()
}

func (peers *PeerList) Rebalance() {
	peers.mux.Lock()
	var closest []int

	for _, v := range peers.peerMap {
		closest = append(closest, int(v))
	}
	closest = append(closest, int(peers.selfId))
	sort.Ints(closest)

	idx := -1
	for index, value := range closest {
		if value == int(peers.selfId) {
			idx = index
		}
	}

	distance := make(map[int]int)

	for index, value := range closest {
		mindis := min(Abs(index-idx), Abs(len(closest)-Abs(index-idx)))

		distance[value] = mindis
	}

	for k, v := range peers.peerMap {
		if distance[int(v)] > int(peers.maxLength/2) {
			delete(peers.peerMap, k)
		}
	}
	peers.mux.Unlock()
}

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (peers *PeerList) Show() string {
	peers.mux.Lock()
	defer peers.mux.Unlock()
	peerJson := PeerJson{peers.selfId, peers.maxLength, peers.peerMap}
	s, _ := json.Marshal(peerJson)
	return string(s)
}

func (peers *PeerList) Register(id int32) {
	//set id
	peers.mux.Lock()
	peers.selfId = id
	fmt.Printf("SelfId=%v\n", id)
	peers.mux.Unlock()
}

func (peers *PeerList) Copy() map[string]int32 {
	// deep copy of the map
	peers.mux.Lock()
	defer peers.mux.Unlock()
	return peers.peerMap
}

func (peers *PeerList) GetSelfId() int32 {
	peers.mux.Lock()
	defer peers.mux.Unlock()
	return peers.selfId
}

func (peers *PeerList) PeerMapToJson() (string, error) {
	peers.mux.Lock()
	defer peers.mux.Unlock()
	s, err := json.Marshal(peers.peerMap)
	return string(s), err
}

func (peers *PeerList) InjectPeerMapJson(peerMapJsonStr string, selfAddr string) {
	// insert everything in peerMapJsonStr into the peer map
	// remove the selfAddr, selfAddr is the current node address
	peers.mux.Lock()

	var peerMap map[string]int32
	json.Unmarshal([]byte(peerMapJsonStr), &peerMap)

	for addr, id := range peerMap {
		peers.peerMap[addr] = id
	}
	delete(peers.peerMap, selfAddr)
	peers.mux.Unlock()
}

func TestPeerListRebalance() {
	peers := NewPeerList(5, 4)
	peers.Add("1111", 1)
	peers.Add("4444", 4)
	peers.Add("-1-1", -1)
	peers.Add("0000", 0)
	peers.Add("2121", 21)
	peers.Rebalance()
	expected := NewPeerList(5, 4)
	expected.Add("1111", 1)
	expected.Add("4444", 4)
	expected.Add("2121", 21)
	expected.Add("-1-1", -1)
	fmt.Println(reflect.DeepEqual(peers, expected))

	peers = NewPeerList(5, 2)
	peers.Add("1111", 1)
	peers.Add("4444", 4)
	peers.Add("-1-1", -1)
	peers.Add("0000", 0)
	peers.Add("2121", 21)
	peers.Rebalance()
	expected = NewPeerList(5, 2)
	expected.Add("4444", 4)
	expected.Add("2121", 21)
	fmt.Println(reflect.DeepEqual(peers, expected))

	peers = NewPeerList(5, 4)
	peers.Add("1111", 1)
	peers.Add("7777", 7)
	peers.Add("9999", 9)
	peers.Add("11111111", 11)
	peers.Add("2020", 20)
	peers.Rebalance()
	expected = NewPeerList(5, 4)
	expected.Add("1111", 1)
	expected.Add("7777", 7)
	expected.Add("9999", 9)
	expected.Add("2020", 20)
	fmt.Println(reflect.DeepEqual(peers, expected))
}
