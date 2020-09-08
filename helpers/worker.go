package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
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

func (w *WitnetConnector) updateReputationLB(resp RespObj) {
	if resp.Result == nil || resp.Error != nil {
		log.Logger.Errorf("%v", resp)
		return
	}
	result := resp.Result
	switch result.(type) {
	case map[string]interface{}:
		nodeRepMap := make(map[string]*NodeRepDetails)
		var nodeRepSort NodeRepSort
    stats := result.(map[string]interface{})["stats"].(map[string]interface{})

		for nodeID, v := range stats { // use type assertion to loop over map[string]interface{}
      nodeStats := v.(map[string]interface{})
			n := NodeRepDetails{
				NodeID:     nodeID,
				Active:     nodeStats["is_active"].(bool),
				Reputation: nodeStats["reputation"].(float64),
			}
			if global.NodeRepMap[nodeID] == nil {
				notifyNodeHasReputation(nodeID)
			}
			nodeRepMap[nodeID] = &n
			nodeRepSort = append(nodeRepSort, n)
		}
		sort.Sort(nodeRepSort)
		global.NodeRepMap = nodeRepMap
		global.ReputationLB = nodeRepSort

		// log.Logger.Debugf("%+v", global.ReputationLB)
		DB.updateReputationDB(nodeRepMap)
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
  witnet.updateReputationLB(resp)
	queryBlockchain()
	GetNodeBlk()

}

func GetNodeBlk() error {
	// safe query
	query := "select blockCount , Miner, lastXEpochs, reward from lightBlockchain  order by blockCount desc;"
	rows, err := sqldb.Query(query)
	if err != nil {
		log.Logger.Error("Err in querying lightblockchain", err)
		return err
	}
	var blockCount int64
	var reward float64
	var miner, lastXEpochs string
	var nodeBlkSort NodeBlkSort
	nodeBlkMap := make(map[string]NodeBlkDetails)
	for rows.Next() {
		err = rows.Scan(&blockCount, &miner, &lastXEpochs, &reward)
		if err != nil {
			log.Logger.Errorf("Error reading row  from  lightblockchain: %s\n\r", err)
			continue
		}
		n := NodeBlkDetails{
			Blocks:      blockCount,
			NodeID:      miner,
			LastXEpochs: lastXEpochs,
			Reward:      reward,
		}
		nodeBlkMap[miner] = n
		nodeBlkSort = append(nodeBlkSort, n)
	}
	log.Logger.Debug(len(nodeBlkSort))
	log.Logger.Trace(nodeBlkSort)
	global.NodeBlkMap = nodeBlkMap
	global.BlocksLB = nodeBlkSort
	return nil
}

func queryBlockchain() {
	for {
		// safe query
		dbQuery := "select IFNULL(max(latestEpoch),0) from lightBlockchain;"
		rows, err := sqldb.Query(dbQuery)
		if err != nil {
			log.Logger.Errorf("Error fetching epoch from blockchain: %s\n\r", err)
			return
		}

		var epoch, limit int
		limit = Config.GetInt("blockchainLimitPerQuery")
		for rows.Next() {
			err := rows.Scan(&epoch)
			if err != nil {
				log.Logger.Errorf("Error in reading blockchain: %s\n\r", err)
				rows.Close()
				return
			}
		}
		rows.Close()
		global.HighestEpoch = epoch

		epochQuery := fmt.Sprintf(`{"jsonrpc": "2.0","method": "getBlockChain", "params": {"epoch":%v, "limit": %v}, "id": "1"}`, epoch+1, limit)
		log.Logger.Debug(epochQuery)
		resp := witnet.QueryRPC(epochQuery)
		if resp.Error != nil {
			log.Logger.Error(resp.Error.(string))
			return
		}
		count, err := witnet.ProcessBlocks(resp)
		if err != nil {
			log.Logger.Error(err)
			return
		}
		// if number of entries is less than limit means that we have queried the latest epoch
		if count < limit {
			return
		}
	}
}

type RespObjBlock struct {
	JsonRPC string      `json:"jsonrpc" yaml:"jsonrpc"`
	Result  Block       `json:"result" yaml:"result"`
	Error   interface{} `json:"error" yaml:"error"`
	Id      string      `json:"id" yaml:"id"`
}

type Block struct {
	Txs TxTypes `json:"txns" yaml:"txns"`
}
type TxTypes struct {
	Mint Mint `json:"mint" yaml:"mint"`
}
type Mint struct {
	Outputs []Transaction `json:"outputs" yaml:"outputs"`
}
type Transaction struct {
	Pkh   string  `json:"pkh" yaml:"pkh"`
	Value float64 `json:"value" yaml:"value"`
}
type MinerDetails struct {
	Reward float64
	Epochs []int64
}

// returns numberofminers, error
func (witnet *WitnetConnector) ProcessBlocks(resp RespObj) (int, error) {
	result := resp.Result
	epochList := result.([]interface{})

	// hashEpoch := make(map[string]float64)
	minerArray := make(map[string]*MinerDetails)
	// var blockHashes []string
	for _, v := range epochList {
		v := v.([]interface{})
		// hash and epoch from each entry
		hash := v[1].(string)
		epoch := int64(v[0].(float64))
		// rpc query string
		blockQuery := fmt.Sprintf(`{"jsonrpc": "2.0","method": "getBlock", "params": ["%s"], "id": "1"}`, hash)
		//query the block with blockHash rpc query
		resp := witnet.QueryRPCBlock(blockQuery)
		// if querying block resulted in error
		if resp.Error != nil {
			return 0, errors.New(resp.Error.(string))
		}
		// TODO handle multiple miners of the block
		// minerTxns := resp.Result.Txs.Mint.Outputs
		// for _, minerTxn := range minerTxns {
		// 	pkh := transaction.Pkh
		// 	value := transaction.Value
		// }
		transaction := resp.Result.Txs.Mint.Outputs[0]
		pkh := transaction.Pkh
		reward := transaction.Value
		minerPtr := minerArray[pkh]
		if minerPtr != nil {
			minerPtr.Epochs = append(minerPtr.Epochs, epoch)
			minerPtr.Reward += reward
		} else {
			minerDetails := MinerDetails{
				Epochs: []int64{epoch},
				Reward: reward,
			}
			minerArray[pkh] = &minerDetails
		}
	}

	// It might happen that the minerArray is empty
	if len(minerArray) == 0 {
		return 0, errors.New("minerArray is of 0 len")
	}

	err := saveMinerDetails(minerArray)
	if err != nil {
		return 0, err
	}
	// notifyBlockMined
	if !Config.GetBool("disableBlockMinedNotify") {
		for minerPkh, minerDetails := range minerArray {
			for _, user := range global.NodeUsers[minerPkh] {
				nodeName := global.Users[user].Nodes[minerPkh]
				for _, epoch := range minerDetails.Epochs {
					msg := tgbotapi.NewMessage(int64(user), fmt.Sprintf("`👌 #%v block was mined by your node: %s[%s]`", epoch, *nodeName, minerPkh))
					msg.ParseMode = "markdown"
					TgBot.Send(msg)
				}
			}
		}
	}
	return len(epochList), nil
}

func saveMinerDetails(minerArray map[string]*MinerDetails) error {
	var rows [][]interface{}
	for minerPkh, miner := range minerArray {
		var startIndex int
		lastX := 5
		epochsLen := len(miner.Epochs)
		if epochsLen-lastX > 0 {
			startIndex = epochsLen - lastX
		}
		// lastFiveEpochs := miner.FiveEpochs[startIndex:]
		var highestEpoch int64
		var lastFiveEpochs string
		for i := startIndex; i < epochsLen; i++ {
			if i != startIndex {
				lastFiveEpochs += ","
			}
			if miner.Epochs[i] > highestEpoch {
				highestEpoch = miner.Epochs[i]
					if int(highestEpoch) > global.HighestEpoch {
						global.HighestEpoch = int(highestEpoch)
					}
			}
			//strconv.Itoa(123)
			lastFiveEpochs += fmt.Sprintf("%v", miner.Epochs[i])
		}

		previousEpoch := epochsLen - startIndex - 5
		totalReward := miner.Reward / 1000000000
		rows = append(rows, []interface{}{minerPkh, highestEpoch, totalReward, epochsLen, lastFiveEpochs,
			highestEpoch, totalReward, epochsLen, previousEpoch, lastFiveEpochs})
	}
	query := `INSERT INTO lightBlockchain (Miner, latestEpoch, reward, blockCount, lastXEpochs) VALUES(?, ?, ?, ?, ?)
	ON DUPLICATE KEY
	UPDATE latestEpoch=?, reward=reward+?, blockCount =  blockCount + ?, lastXEpochs = CONCAT(SUBSTRING_INDEX(lastXEpochs, ',', ?), ',', ?);`
	return multipleInsert(query, rows)
}
