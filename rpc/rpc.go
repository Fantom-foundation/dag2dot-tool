package rpc

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// An RPC client requests event data from RPC service on a Lachasis node
type RPC struct {
	url string
}

func NewRPC(host string, port int) *RPC {
	return &RPC{
		url: "http://" + host + ":" + strconv.FormatInt(int64(port), 10) + "/",
	}
}

type TopHeads struct {
	JsonRPC string   `json:"jsonrpc"`
	Id      int64    `json:"id"`
	Result  []string `json:"result"`
}

type EventHeader struct {
	ClaimedTime      int64    `json:"claimedTime"`
	Creator          int64   `json:"creator"`
	Epoch            int64    `json:"epoch"`
	ExtraData        string   `json:"extraData"`
	Frame            int64    `json:"frame"`
	GasPowerLeft 	GasPowerLeft`json:"gasPowerLeft"`
	GasPowerUsed     int64    `json:"gasPowerUsed"`
	Hash             string   `json:"hash"`
	IsRoot           bool     `json:"isRoot"`
	Lamport          int64    `json:"lamport"`
	MedianTime       int64    `json:"medianTime"`
	Parents          []string `json:"parents"`
	PrevEpochHash    string   `json:"prevEpochHash"`
	Seq              int64    `json:"seq"`
	TransactionsRoot string   `json:"transactionsRoot"`
	Version          int      `json:"version"`
}

type GasPowerLeft struct {
        Gas [2]uint64
}

type EventHeaderResponse struct {
	JsonRPC string      `json:"jsonrpc"`
	Id      int64       `json:"id"`
	Result  EventHeader `json:"result"`
}

type Event struct {
	EventHeader
	Transactions	[]string	`json:"transactions"`
}

type EventResponse struct {
	JsonRPC string      `json:"jsonrpc"`
	Id      int64       `json:"id"`
	Result  Event 		`json:"result"`
}

func (rpc *RPC) GetTopHeads() (*TopHeads, error) {
	req := `{"jsonrpc":"2.0","method":"ftm_getHeads","params":["pending"],"id":1}`

	body, err := rpc.call(req)
	if err != nil {
		log.Printf("Call RPC error: %s\n", err)
		return nil, err
	}

	top := TopHeads{}
	err = json.Unmarshal(body, &top)
	if err != nil {
		log.Printf("Json parse response ftm_getHeads body error: %s\n", err)
		return nil, err
	}

	return &top, nil
}

func (rpc *RPC) GetEventHeader(hash string) (*EventHeader, error) {
	req := `{"jsonrpc":"2.0","method":"ftm_getEventHeader","params":["` + hash + `"],"id":1}`

	body, err := rpc.call(req)
	if err != nil {
		log.Printf("Call RPC error: %s\n", err)
		return nil, err
	}

	head := EventHeaderResponse{}
	err = json.Unmarshal(body, &head)
	if err != nil {
		log.Printf("Json parse response ftm_getEventHeader body error: %s\n", err)
		return nil, err
	}

	return &head.Result, nil
}

func (rpc *RPC) GetEvent(hash string) (*Event, error) {
	req := `{"jsonrpc":"2.0","method":"ftm_getEvent","params":["` + hash + `", true],"id":1}`

	body, err := rpc.call(req)
	if err != nil {
		log.Printf("Call RPC error: %s\n", err)
		return nil, err
	}

	head := EventResponse{}
	err = json.Unmarshal(body, &head)
	if err != nil {
		log.Printf("Json parse response ftm_getEvent body error: %s\n", err)
		return nil, err
	}

	return &head.Result, nil
}

func (rpc *RPC) call(reqBody string) ([]byte, error) {
	reqIO := strings.NewReader(reqBody)

	resp, err := http.Post(rpc.url, "application/json", reqIO)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 	log.Printf("REQUEST: %s\nRESPONSE: %s\n", reqBody, body)

	return body, nil
}
