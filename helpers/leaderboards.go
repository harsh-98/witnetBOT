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

func checkUsersNode(v string, list map[string]*string) bool {
	if list[v] != nil {
		return true
	}
	return false
}

var medal = []string{"🥇", "🥈", "🥉"}

func sendReputationBoard(tgID int64) {
	nLen := len(global.ReputationLB)

	str := fmt.Sprintf("🏆 *Reputation 🎖 LeaderBoard* (Nodes with reputation: %v) \n\n", nLen)

	first3 := int(math.Min(3, float64(nLen)))
	for i := 0; i < first3; i++ {

		var isUserNode string
		nodeID := global.ReputationLB[i].NodeID
		if checkUsersNode(nodeID, global.Users[tgID].Nodes) {
			isUserNode = fmt.Sprintf("(Your node) %s", *(global.Users[tgID].Nodes[nodeID]))
		}
		str += fmt.Sprintf("`%s\n\r%s %v - %s\n\rReputation: %v \n\r`\n\n", isUserNode, medal[i], i+1, nodeID, global.ReputationLB[i].Reputation)
	}
	for i := 3; i < nLen; i++ {
		nodeID := global.ReputationLB[i].NodeID
		if checkUsersNode(nodeID, global.Users[tgID].Nodes) {
			isUserNode := fmt.Sprintf("(Your node) %s", *(global.Users[tgID].Nodes[nodeID]))
			str += fmt.Sprintf("`%s\n\r%v - %s\n\rReputation: %v \n\r`\n\n", isUserNode, i+1, nodeID, global.ReputationLB[i].Reputation)
		}
	}
	msg := tgbotapi.NewMessage(int64(tgID), str)
	msg.ParseMode = "markdown"
	TgBot.Send(msg)

}

func sendBlocksBoard(tgID int64) {
	nLen := len(global.BlocksLB)
	var entries string
	log.Logger.Debugf("Block Leaderboard: %v", global.HighestEpoch)

	first3 := int(math.Min(3, float64(nLen)))
	// userBlockCount: Total block minted by all the nodes of user
	// totalBlockCount: Total block minted by the network
	// global.HighestEpoch: Number of epoch passed
	// blockPerNode: blocks minted by specific node
	var totalBlockCount, userBlockCount int64
	for i := 0; i < first3; i++ {
		var isUserNode string
		nodeID := global.BlocksLB[i].NodeID
		blockPerNode := global.BlocksLB[i].Blocks
		if checkUsersNode(nodeID, global.Users[tgID].Nodes) {
			userBlockCount += blockPerNode
			isUserNode = fmt.Sprintf("(Your node) %s", *(global.Users[tgID].Nodes[nodeID]))
		}
		totalBlockCount += blockPerNode
		entries += fmt.Sprintf("`%s\n\r%s %v - %s\n\rBlocks mined: %v \n\r`\n\n", isUserNode, medal[i], i+1, nodeID, blockPerNode)

	}

	for i := 3; i < nLen; i++ {
		blockPerNode := global.BlocksLB[i].Blocks
		totalBlockCount += blockPerNode
		nodeID := global.BlocksLB[i].NodeID
		if checkUsersNode(nodeID, global.Users[tgID].Nodes) {
			userBlockCount += blockPerNode
			isUserNode := fmt.Sprintf("(Your node) %s", *(global.Users[tgID].Nodes[nodeID]))
			entries += fmt.Sprintf("`%s\n\r%v - %s\n\rBlocks mined: %v \n\r`\n\n", isUserNode, i+1, nodeID, blockPerNode)
		}
	}
	header := fmt.Sprintf("🏆*Mining 🔨 LeaderBoard* (Max epoch: %v, Mined block: %v)\n\n\r `(Total Blocks mined by your nodes: %v)`\n\n",
		global.HighestEpoch, totalBlockCount, userBlockCount)

	msg := tgbotapi.NewMessage(int64(tgID), header+entries)
	msg.ParseMode = "markdown"
	TgBot.Send(msg)
}
