package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/harsh-98/witnetBOT/log"
)

type RespObj struct {
	JsonRPC string      `json:"jsonrpc" yaml:"jsonrpc"`
	Result  interface{} `json:"result" yaml:"result"`
	Error   interface{} `json:"error" yaml:"error"`
	Id      string      `json:"id" yaml:"id"`
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
	err = json.NewDecoder(conn).Decode(&v)
	if err != nil {
		log.Logger.Error(err)
	}
	return v
}
func (w *WitnetConnector) QueryRPCBlock(msg string) RespObjBlock {
	if !strings.HasSuffix(msg, "\n") {
		msg = fmt.Sprintf("%s\n", msg)
	}
	tcpAddr, err := net.ResolveTCPAddr("tcp", w.Address)
	if err != nil {
		return RespObjBlock{Error: err.Error()}
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return RespObjBlock{Error: err.Error()}
	} else {
		defer conn.Close()
	}

	_, err = conn.Write([]byte(msg))
	if err != nil {
		return RespObjBlock{Error: err.Error()}
	}

	var obj RespObjBlock
	// don't use ioutils.readAll it is blocking call and waits for streaming to end
	err = json.NewDecoder(conn).Decode(&obj)
	if err != nil {
		log.Logger.Error(err)
	}
	return obj
}
func (w *WitnetConnector) ProcessAndUpdateDB(resp RespObj) {
	if resp.Result == nil || resp.Error != nil {
		log.Logger.Errorf("%v", resp)
		return
	}
	result := resp.Result
	var newNodes []string
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
			if global.Nodes[n.NodeID] == nil {
				newNodes = append(newNodes, n.NodeID)
			}
			nodes[n.NodeID] = &n
			nodeRepSort = append(nodeRepSort, n)
		}
		sort.Sort(nodeRepSort)
		global.Nodes = nodes
		global.Ranking = nodeRepSort

		// log.Logger.Debugf("%+v", global.Ranking)
		DB.AddNodesInTable(nodes)
		notifyReputationList(newNodes)
	}
}

var witnet = WitnetConnector{Address: ""}

func QueryWorker() {
	timer := time.NewTimer(time.Duration(Config.GetInt("timer")) * time.Second)
	ticker := time.NewTicker(time.Duration(Config.GetInt("ticker")) * 60 * time.Second)
	done := make(chan bool)
	for {
		select {
		case <-done:
			return
		case _ = <-timer.C:
			log.Logger.Debug("timer")
			queryWitnet()
			timer.Stop()
		case _ = <-ticker.C:
			log.Logger.Debug("ticker")
			queryWitnet()

		}
	}
}

func queryWitnet() {
	witnet.Address = Config.GetString("servAddr")
	resp := witnet.QueryRPC(`{"jsonrpc": "2.0","method": "getReputationAll", "id": "1"}`)
	witnet.ProcessAndUpdateDB(resp)
	queryBlockchain()
}
func queryBlockchain() {
	for {
		dbQuery := "select IFNULL(max(Epoch),0) from blockchain;"
		rows, err := sqldb.Query(dbQuery)
		if err != nil {
			log.Logger.Errorf("Error fetching epoch from blockchain: %s\n\r", err)
			return
		}

		var epoch, limit int
		limit = 10
		for rows.Next() {
			err := rows.Scan(&epoch)
			if err != nil {
				log.Logger.Errorf("Error in reading blockchain: %s\n\r", err)
				return
			}
		}

		epochQuery := fmt.Sprintf(`{"jsonrpc": "2.0","method": "getBlockChain", "params": {"epoch":%v, "limit": %v}, "id": "1"}`, epoch+1, limit)
		log.Logger.Debugf("\n%s\n\n", epochQuery)
		resp := witnet.QueryRPC(epochQuery)
		if resp.Error != nil {
			log.Logger.Error(resp.Error.(string))
		}
		count, err := witnet.ProcessBlocks(resp)
		if err != nil {
			log.Logger.Error(err)
			return
		} else if count < limit {
			return
		}
	}
}

type RespObjBlock struct {
	JsonRPC string      `json:"jsonrpc" yaml:"jsonrpc"`
	Result  B           `json:"result" yaml:"result"`
	Error   interface{} `json:"error" yaml:"error"`
	Id      string      `json:"id" yaml:"id"`
}
type Block struct {
	Epoch float64
	Hash  string
	Miner string
}
type B struct {
	Txs TxTypes `json:"txns" yaml:"txns"`
}
type TxTypes struct {
	Mint Mint `json:"mint" yaml:"mint"`
}
type Mint struct {
	Outputs []Transaction `json:"outputs" yaml:"outputs"`
}
type Transaction struct {
	Pkh string `json:"pkh" yaml:"pkh"`
}

func (witnet *WitnetConnector) ProcessBlocks(resp RespObj) (int, error) {
	result := resp.Result
	epochList := result.([]interface{})

	hashEpoch := make(map[string]float64)
	var blockHash []string
	for _, v := range epochList {
		v := v.([]interface{})
		hash := v[1].(string)
		hashEpoch[hash] = v[0].(float64)
		blockHash = append(blockHash, hash)
	}
	log.Logger.Debugf("block hashes: %v", blockHash)
	var dbQuery string
	for _, hash := range blockHash {
		blockQuery := fmt.Sprintf(`{"jsonrpc": "2.0","method": "getBlock", "params": ["%s"], "id": "1"}`, hash)
		resp := witnet.QueryRPCBlock(blockQuery)
		result = resp.Result
		if resp.Error != nil {
			return 0, errors.New(resp.Error.(string))
		}
		pkh := result.(B).Txs.Mint.Outputs[0].Pkh
		fmt.Printf("%+v", result.(B).Txs.Mint.Outputs[0].Pkh)
		dbQuery += fmt.Sprintf("insert into blockchain (Epoch, Hash, Miner) values (%v, '%s' , '%s'); ", hashEpoch[hash], hash, pkh)
	}
	log.Logger.Debug(dbQuery)
	_, err := sqldb.Exec(dbQuery)
	if err != nil {
		return 0, err
	}
	return len(blockHash), nil
}
