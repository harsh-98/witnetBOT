package helpers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/harsh-98/witnetBOT/log"
)

type TestnetReward struct {
	Timelock string
	Reward   string
}
type UnlockReward struct {
	Addr  string `json:"addr" yaml:"jsonrpc"`
	Value int64  `json:"val" yaml:"jsonrpc"`
}

func GenesisReward(id int64) {
	user := global.Users[id]
	response := " *Testnet rewards*"
	fmt.Println(len(user.Nodes))
	for nodeID, nodeName := range user.Nodes {
		genesisNodeData := global.Genesis[nodeID]
		if genesisNodeData == nil {
			continue
		}
		rewards := strings.Split(genesisNodeData.Reward, ",")
		response = response + fmt.Sprintf("\n\nNodeID: %s\nNodeName: %s\n",
			nodeID, *nodeName)
		for i, v := range strings.Split(genesisNodeData.Timelock, ",") {
			reward, err := strconv.Atoi(rewards[i])
			log.Logger.Errorf("Reward converting error %s %s", nodeID, err)
			timelock, err := strconv.ParseInt(v, 10, 64)
			log.Logger.Errorf("Timelock converting error %s %s", nodeID, err)
			response = response + fmt.Sprintf("```Rewards: %.3f wits locked till: %v```",
				float64(reward)/math.Pow(10, 9), time.Unix(timelock, 0).Format(RFC822Z))
		}
	}
	msg := tgbotapi.NewMessage(id, response)
	msg.ParseMode = "markdown"
	TgBot.Send(msg)
}

func GetGenesisRewards() error {
	// safe query
	rows, err := sqldb.Query("select NodeID, Timelock, Wits from genesis")
	if err != nil {
		log.Logger.Errorf("Error rewards for genesis: %s\n\r", err)
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var nodeID string
		var reward TestnetReward
		err = rows.Scan(&nodeID, &reward.Timelock, &reward.Reward)
		if err != nil {
			log.Logger.Errorf("Row fetching for genesis failed: %s\n\r", err)
			continue
		}
		global.Genesis[nodeID] = &reward
	}
	return nil
}
func loadGenesisUnlock() {
	jsonFile, err := os.Open("genesisUnlock.json")
	defer jsonFile.Close()
	if err != nil {
		os.Exit(1)
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &global.GenesisUnlock)
}
func genesisUnlockReward() {
	rows, err := sqldb.Query("select notifyTimeStamp, notified from genesisUnlockNotify where notified=false order by notifyTimeStamp asc  limit 1;")
	if err != nil {
		log.Logger.Errorf("Error querying genesisUnlockNotify: %s\n\r", err)
	}
	defer rows.Close()
	var timeStamp int64
	var notify bool
	for rows.Next() {
		err := rows.Scan(&timeStamp, notify)
		log.Logger.Errorf("Error scanning genesisUnlockNotify: %s\n\r", err)
		if !notify && time.Now().UTC().Unix() > timeStamp {
			for _, v := range global.GenesisUnlock[timeStamp] {
				notifyUsersGensisUnlock(
					fmt.Sprintf("Your reward %.2f unlocked at %s", float64(v.Value)/math.Pow(10,9), time.Unix(timeStamp,0).Format(RFC822Z)),
					v.Addr)
			}
			sqldb.Exec("update gensisUnlockNotify set notified=? where notifyTimeStamp=?", true, timeStamp)
		}
	}
}
