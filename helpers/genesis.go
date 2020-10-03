package helpers

import (
	"fmt"
	"math"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/harsh-98/witnetBOT/log"
)

type TestnetReward struct {
	Timelock int64
	Reward   float64
}

func GenesisReward(id int64) {
	user := global.Users[id]
	response := " *Testnet rewards*\n\n"
	for nodeID, nodeName := range user.Nodes {
		reward := global.Genesis[nodeID]
		if reward == nil {
			continue
		}
		response = response + fmt.Sprintf("NodeID: %s\nNodeName: %s\nRewards: %v wits\nLocked till: %v\n\n",
			nodeID, *nodeName, reward.Reward, reward.Timelock)
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
		reward.Reward = reward.Reward / math.Pow(10, 9)
		global.Genesis[nodeID] = &reward
	}
	return nil
}
