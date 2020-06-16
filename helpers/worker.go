package helpers

import (
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/harsh-98/witnetBOT/log"
	"github.com/spf13/viper"
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

// {} with type
type NodeRepSort []NodeType

func (s NodeRepSort) Len() int {
	return len(s)
}
func (s NodeRepSort) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s NodeRepSort) Less(i, j int) bool {
	return s[i].Reputation > s[j].Reputation
}

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

func (w *WitnetConnector) ProcessAndUpdateDB(resp RespObj) {
	if resp.Result == nil || resp.Error != nil {
		log.Logger.Errorf("%v", resp)
		return
	}
	result := resp.Result
	switch result.(type) {
	case map[string]interface{}:
		nodes := make(map[string]*NodeType)
		var nodeRepSort NodeRepSort
		for k, v := range result.(map[string]interface{}) { // use type assertion to loop over map[string]interface{}
			n := NodeType{
				NodeID:     k,
				Active:     v.([]interface{})[1].(bool),
				Reputation: v.([]interface{})[0].(float64),
			}
			nodes[n.NodeID] = &n
			nodeRepSort = append(nodeRepSort, n)
		}
		sort.Sort(nodeRepSort)
		global.Nodes = nodes
		global.Ranking = nodeRepSort
		// log.Logger.Debugf("%+v", global.Ranking)
		DB.AddNodesInTable(nodes)
	}
}

func QueryWorker(vip *viper.Viper) {
	witnet := WitnetConnector{Address: vip.GetString("servAddr")}
	timer := time.NewTimer(time.Duration(vip.GetInt("timer")) * time.Second)
	ticker := time.NewTicker(time.Duration(vip.GetInt("ticker")) * 60 * time.Second)
	done := make(chan bool)
	for {
		select {
		case <-done:
			return
		case _ = <-timer.C:
			log.Logger.Debug("timer")
			resp := witnet.QueryRPC(`{"jsonrpc": "2.0","method": "getReputationAll", "id": "1"}`)
			witnet.ProcessAndUpdateDB(resp)
			timer.Stop()
		case _ = <-ticker.C:
			log.Logger.Debug("ticker")
			resp := witnet.QueryRPC(`{"jsonrpc": "2.0","method": "getReputationAll", "id": "1"}`)
			witnet.ProcessAndUpdateDB(resp)

		}
	}
}
