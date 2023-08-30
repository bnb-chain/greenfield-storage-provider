package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	golog "log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var LatestHeight int

type ChainCase struct {
	Name         string `json:"name"`
	Block        string `json:"block"`
	BlockResults string `json:"block_results"`
}

var (
	StatusRes          string
	MockBLockRes       []string
	MockBLockResultRes []string
)

func initMockRes() {
	jsonFilePath := filepath.Join("./test", "case.json")

	file, err := os.Open(jsonFilePath)
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return
	}

	decoder := json.NewDecoder(file)

	var cases []ChainCase
	err = decoder.Decode(&cases)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	LatestHeight = len(cases)
	StatusRes = "{\"sync_info\":{\"latest_block_height\":\"1\"}}"

	for _, c := range cases {
		MockBLockRes = append(MockBLockRes, c.Block)
		MockBLockResultRes = append(MockBLockResultRes, c.BlockResults)
	}
}

func Block(height int64) []byte {
	res := MockBLockRes[height-1]
	return []byte(res)
}

func BlockResult(height int64) []byte {
	res := MockBLockResultRes[height-1]
	golog.Println(res)
	return []byte(res)
}

func GetStatus() []byte {
	return []byte(StatusRes)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type RPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Id      int             `json:"id"`
	Params  json.RawMessage `json:"params"` // must be map[string]interface{} or []interface{}
}

type RPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Id      int             `json:"id"`
	Result  json.RawMessage `json:"result"` // must be map[string]interface{} or []interface{}
}

func reader(conn *websocket.Conn) {
	tick := time.NewTicker(time.Millisecond * 20)
	for range tick.C {
		// read in a message
		_, p, err := conn.ReadMessage()
		if err != nil {
			golog.Printf("failed to read error:%v", err)
			return
		}
		// print out that message for clarity
		golog.Printf("info :%v", p)
	}
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	// upgrade this connection to a WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		golog.Println(err)
	}

	golog.Println("Client Connected")
	err = ws.WriteMessage(1, []byte("Hi Client!"))
	if err != nil {
		golog.Println(err)
	}

	reader(ws)
}

func homePage(w http.ResponseWriter, r *http.Request) {
	body := r.Body
	buf := new(bytes.Buffer)
	buf.ReadFrom(body)
	var req RPCRequest
	err := json.Unmarshal(buf.Bytes(), &req)
	if err != nil {
		golog.Println("unmarshal failed")
		return
	}
	switch req.Method {
	case "status":
		resp := GetStatus()
		ret := &RPCResponse{
			JSONRPC: req.JSONRPC,
			Id:      req.Id,
			Result:  resp,
		}
		result, _ := json.Marshal(ret)
		if _, err = w.Write(result); err != nil {
			golog.Printf("failed to write  error:%v", err)
			return
		}
	case "block":
		var m map[string]json.RawMessage
		err = json.Unmarshal(req.Params, &m)
		if err != nil {
			golog.Printf("failed to unmarshal error:%v", err)
			return
		}
		hs := string(m["height"])
		height, _ := strconv.ParseInt(strings.Trim(hs, "\""), 10, 64)
		golog.Println(height)
		resp := Block(height)
		ret := &RPCResponse{
			JSONRPC: req.JSONRPC,
			Id:      req.Id,
			Result:  resp,
		}
		result, _ := json.Marshal(ret)
		if _, err = w.Write(result); err != nil {
			golog.Printf("failed to write error:%v", err)
			return
		}
	case "block_results":
		var m map[string]json.RawMessage
		err = json.Unmarshal(req.Params, &m)
		if err != nil {
			golog.Printf("failed to unmarshal error:%v", err)
			return
		}
		hs := string(m["height"])
		height, _ := strconv.ParseInt(strings.Trim(hs, "\""), 10, 64)
		resp := BlockResult(height)
		ret := &RPCResponse{
			JSONRPC: req.JSONRPC,
			Id:      req.Id,
			Result:  resp,
		}
		result, _ := json.Marshal(ret)
		if _, err = w.Write(result); err != nil {
			golog.Printf("failed to write error:%v", err)
			return
		}
	}
}

func setupRoutes() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", homePage)
	mux.HandleFunc("/websocket", wsEndpoint)

	http.ListenAndServe(":8080", mux)
}

func MockChainRPCServer() {
	initMockRes()
	setupRoutes()
}
