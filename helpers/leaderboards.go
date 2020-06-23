package helpers

import (
	"fmt"
	"math"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/harsh-98/witnetBOT/log"
)

// {} with type not need
// reputation leaderboard list
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

// blocks count leaderboard list
type NodeBlock struct {
	Blocks int64
	NodeID string
}
type NodeBlockSort []NodeBlock

func (s NodeBlockSort) Len() int {
	return len(s)
}
func (s NodeBlockSort) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s NodeBlockSort) Less(i, j int) bool {
	return s[i].Blocks > s[j].Blocks
}

func checkUsersNode(v string, list []string) bool {
	for _, n := range list {
		if n == v {
			return true
		}
	}
	return false
}

var medal = []string{"🥇", "🥈", "🥉"}

func sendReputationBoard(tgID int64) {
	nLen := len(global.ReputationLB)

	str := fmt.Sprintf("🏆 **Reputation 🎖 Leader Board** (Nodes count: %v) \n\n", nLen)

	first3 := int(math.Min(3, float64(nLen)))
	for i := 0; i < first3; i++ {

		var isUserNode string
		if checkUsersNode(global.ReputationLB[i].NodeID, global.Users[tgID].Nodes) {
			isUserNode = "(Your node)"
		}
		str += fmt.Sprintf("`%s\n\r%s %v - %s\n\rReputation: %v \n\r`\n\n", isUserNode, medal[i], i+1, global.ReputationLB[i].NodeID, global.ReputationLB[i].Reputation)
	}
	isUserNode := "(Your node)"
	for i := 3; i < nLen; i++ {
		if checkUsersNode(global.ReputationLB[i].NodeID, global.Users[tgID].Nodes) {
			str += fmt.Sprintf("`%s\n\r%v - %s\n\rReputation: %v \n\r`\n\n", isUserNode, i+1, global.ReputationLB[i].NodeID, global.ReputationLB[i].Reputation)
		}
	}
	msg := tgbotapi.NewMessage(int64(tgID), str)
	msg.ParseMode = "markdown"
	TgBot.Send(msg)

}

func sendBlocksBoard(tgID int64) {
	nLen := len(global.BlocksLB)
	var str string
	log.Logger.Debugf("Length of block LB: %v", nLen)

	first3 := int(math.Min(3, float64(nLen)))
	var blockCount int64
	for i := 0; i < first3; i++ {

		var isUserNode string
		if checkUsersNode(global.BlocksLB[i].NodeID, global.Users[tgID].Nodes) {
			isUserNode = "(Your node)"
		}
		blockCount += global.BlocksLB[i].Blocks
		str += fmt.Sprintf("`%s\n\r%s %v - %s\n\rBlocks mined: %v \n\r`\n\n", isUserNode, medal[i], i+1, global.BlocksLB[i].NodeID, global.BlocksLB[i].Blocks)

	}

	isUserNode := "(Your node)"
	for i := 3; i < nLen; i++ {
		blockCount += global.BlocksLB[i].Blocks
		if checkUsersNode(global.BlocksLB[i].NodeID, global.Users[tgID].Nodes) {
			str += fmt.Sprintf("`%s\n\r%v - %s\n\rBlocks mined: %v \n\r`\n\n", isUserNode, i+1, global.BlocksLB[i].NodeID, global.BlocksLB[i].Blocks)
		}
	}
	header := fmt.Sprintf("🏆 **Mining 🔨 Leader Board** (Total Blocks count: %v) \n\n", blockCount)
	msg := tgbotapi.NewMessage(int64(tgID), header+str)
	msg.ParseMode = "markdown"
	TgBot.Send(msg)
}
