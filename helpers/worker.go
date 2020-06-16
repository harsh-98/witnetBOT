package helpers

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type RespObj struct {
	JsonRPC string      `json:"jsonrpc" yaml:"jsonrpc"`
	Result  interface{} `json:"result" yaml:"result"`
	Error   interface{} `json:"error" yaml:"error"`
	Id      int         `json:"id" yaml:"id"`
}
type WitnetConnector struct {
	Address string
}

var service = os.ExpandEnv("$SERVADDR")

func (w *WitnetConnector) QueryRPC(msg string) RespObj {
	if !strings.HasSuffix(msg, "\n") {
		msg = fmt.Sprintf("%s\n", msg)
	}
	tcpAddr, err := net.ResolveTCPAddr("tcp", w.Address)
	if err != nil {
		return RespObj{Error: err.Error()}
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return RespObj{Error: err.Error()}
	} else {
		defer conn.Close()
	}

	_, err = conn.Write([]byte(msg))
	if err != nil {
		return RespObj{Error: err.Error()}
	}

	var v RespObj
	// don't use ioutils.readAll it is blocking call and waits for streaming to end
	json.NewDecoder(conn).Decode(&v)
	return v
}

var Nodes = []NodeType{}

func (w *WitnetConnector) ProcessAndUpdateDB(resp RespObj) {
	if resp.Result == nil || resp.Error != nil {
		log.Errorf("%v", resp)
		return
	}
	result := resp.Result
	switch result.(type) {
	case map[string]interface{}:
		nodes := []NodeType{}
		for k, v := range result.(map[string]interface{}) { // use type assertion to loop over map[string]interface{}
			n := NodeType{
				NodeID:     k,
				Active:     v.([]interface{})[1].(bool),
				Reputation: v.([]interface{})[0].(float64),
			}
			nodes = append(nodes, n)
		}
		Nodes = nodes
		DB.AddNodesInTable(nodes)
	}
}

func QueryWorker() {
	witnet := WitnetConnector{Address: service}
	timer := time.NewTimer(5000 * time.Second)
	ticker := time.NewTicker(60 * 10 * time.Second)
	done := make(chan bool)
	for {
		select {
		case <-done:
			return
		case _ = <-timer.C:
			log.Debug("timer")
			resp := witnet.QueryRPC(`{"jsonrpc": "2.0","method": "getReputationAll", "id": "1"}`)
			witnet.ProcessAndUpdateDB(resp)
			timer.Stop()
		case _ = <-ticker.C:
			log.Debug("ticker")
			resp := witnet.QueryRPC(`{"jsonrpc": "2.0","method": "getReputationAll", "id": "1"}`)
			witnet.ProcessAndUpdateDB(resp)

		}
	}
}
