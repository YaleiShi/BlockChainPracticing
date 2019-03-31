package p3

import (
	"../p2"
	"./data"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"time"
)

var TA_SERVER = "http://localhost:6688"
var REGISTER_SERVER = TA_SERVER + "/peer"
var SELF_ADDR string
var SELF_ID int32
var BC_DOWNLOAD_SERVER = FIRST_ADDR + "/upload"
var FIRST_ADDR = "http://localhost:6686"
var FIRST_ID = int32(6686)

var SBC data.SyncBlockChain
var Peers data.PeerList
var ifStarted bool

func Init() {
	// This function will be executed before everything else.
	// Do some initialization here.
	if len(os.Args) > 1 {
		port := os.Args[1]
		portNum, err := strconv.Atoi(port)
		if err == nil {
			SELF_ID = int32(portNum)
			SELF_ADDR = "http://localhost:" + port
		} else {
			SELF_ID = 6686
			SELF_ADDR = "http://localhost:6686"
		}
	} else {
		SELF_ID = 6686
		SELF_ADDR = "http://localhost:6686"
	}
	SBC = data.NewBlockChain()
	Peers = data.NewPeerList(SELF_ID, 32)
	ifStarted = false
}

// Register ID, download BlockChain, start HeartBeat
func Start(w http.ResponseWriter, r *http.Request) {
	if len(os.Args) > 1 {
		fmt.Println("input port: ", os.Args[1])
	}
	Init()
	if SELF_ID == 6686 {
		startFirstNode()
	} else {
		startAfterNode()
	}
}

func startFirstNode() {
	//Register()
	ifStarted = true
	StartHeartBeat()
}

func startAfterNode() {
	//Register()
	Download()
	Peers.Add(FIRST_ADDR, FIRST_ID)
	peerMapJson, _ := Peers.PeerMapToJson()
	firstHB := data.NewHeartBeatData(false, Peers.GetSelfId(), "", peerMapJson, SELF_ADDR)
	ForwardHeartBeat(firstHB)

	ifStarted = true
	StartHeartBeat()
}

// Display peerList and sbc
func Show(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s\n%s", Peers.Show(), SBC.Show())
}

// Register to TA's server, get an ID
func Register() {
	// go to ta server
	// get the id and set up this.id = id
	response, err := http.Get(REGISTER_SERVER)
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	id, _ := strconv.Atoi(string(body))

	Peers.Register(int32(id))
}

// Download blockchain from first server
func Download() {
	response, err := http.Get(BC_DOWNLOAD_SERVER)
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	SBC.UpdateEntireBlockChain(string(body))
}

// Upload blockchain to whoever called this method, return jsonStr
func Upload(w http.ResponseWriter, r *http.Request) {
	blockChainJson, err := SBC.BlockChainToJson()
	if err != nil {
		data.PrintError(err, "Upload")
	}
	fmt.Fprint(w, blockChainJson)
}

// Upload a block to whoever called this method, return jsonStr
func UploadBlock(w http.ResponseWriter, r *http.Request) {
	// called by AskForBlock, w write a response as a json string
	vars := mux.Vars(r)
	height := vars["height"]
	hash := vars["hash"]
	heightInt, _ := strconv.Atoi(height)
	block, ok := SBC.GetBlock(int32(heightInt), hash)
	if !ok {
		fmt.Fprint(w, "No This Block")
	}
	blockJsonStr := block.EncodeToJSON()
	fmt.Fprint(w, blockJsonStr)
}

// Received a heartbeat
func HeartBeatReceive(w http.ResponseWriter, r *http.Request) {
	// receive a heart beat
	var hb data.HeartBeatData
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(body, &hb)
	// check if it has a new block
	ifHave := hb.IfNewBlock
	if ifHave {
		// if have, check if parrent block exist
		blockJson := hb.BlockJson
		block := p2.DecodeFromJson(blockJson)
		ifParent := SBC.CheckParentHash(block)
		// if exist add new block into bc
		// if not exist, call AsfForBlock to add parrent block into bc
		if !ifParent {
			AskForBlock(block.Header.Height-1, block.Header.ParentHash)
		}
		// then add new block into bc
		SBC.Insert(block)
	}
	// add peer list
	Peers.Add(hb.Addr, hb.Id)
	Peers.InjectPeerMapJson(hb.PeerMapJson, SELF_ADDR)
}

// Ask another server to return a block of certain height and hash
func AskForBlock(height int32, hash string) {
	// call http get /height/hash of all the peer's block, which is the function UploadBlock
	// get the block and add into bc
	request := "/block/" + string(height) + "/" + hash
	peerMap := Peers.Copy()

	for addr, _ := range peerMap {
		response, err := http.Get(addr + request)
		if err != nil {
			log.Fatal(err)
		}

		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}

		bodyStr := string(body)
		if reflect.DeepEqual(bodyStr, "No This Block") {
			continue
		}
		block := p2.DecodeFromJson(bodyStr)
		SBC.Insert(block)
		break
	}

}

func ForwardHeartBeat(heartBeatData data.HeartBeatData) {
	request := "/heartbeat/receive"
	peerMap := Peers.Copy()
	heartBeatJson, _ := json.Marshal(heartBeatData)

	for addr, _ := range peerMap {
		_, err := http.Post(addr+request, "application/json", bytes.NewBuffer(heartBeatJson))
		if err != nil {
			log.Fatal(err)
		}

	}
}

func StartHeartBeat() {
	for {
		if !ifStarted {
			break
		}
		time.Sleep(6 * time.Second)
		Peers.Rebalance()
		peerMapJson, _ := Peers.PeerMapToJson()
		selfId := Peers.GetSelfId()
		heartBeatData := data.PrepareHeartBeatData(&SBC, selfId, peerMapJson, SELF_ADDR)
		ForwardHeartBeat(heartBeatData)
	}
}
