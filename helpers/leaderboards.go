package helpers

import (
	"fmt"
	"math"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/harsh-98/witnetBOT/log"
)

type NodeRepDetails struct {
	Reputation float64
	NodeID     string
	Active     bool
}

// {} with type not need
// reputation leaderboard list
type NodeRepSort []NodeRepDetails

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
type NodeBlkDetails struct {
	Blocks      int64
	NodeID      string
	LastXEpochs string
	Reward      float64
}
type NodeBlkSort []NodeBlkDetails

func (s NodeBlkSort) Len() int {
	return len(s)
}
func (s NodeBlkSort) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s NodeBlkSort) Less(i, j int) bool {
	return s[i].Blocks > s[j].Blocks
}

func checkUsersNode(v string, list map[string]*string) bool {
	if list[v] != nil {
		return true
	}
	return false
}

var medal = []string{"🥇", "🥈", "🥉"}

func sendReputationBoard(tgID, rcvrID int64) {
	nLen := len(global.ReputationLB)

	entry := fmt.Sprintf("🏆 *Reputation 🎖 LeaderBoard* (Nodes with reputation: %v) \n\n", nLen)

	first3 := int(math.Min(3, float64(nLen)))

	// user nodes
	userNodes := make(map[string]*string)
	if global.Users[tgID] != nil {
		userNodes = global.Users[tgID].Nodes
	}

	// msg array , with each message length less than 4096
	var entries []string
	for i := 0; i < first3; i++ {

		var isUserNode string
		nodeID := global.ReputationLB[i].NodeID
		if checkUsersNode(nodeID, userNodes) {
			isUserNode = fmt.Sprintf("(Your node) %s", *(userNodes[nodeID]))
		}
		entry += fmt.Sprintf("`%s\n\r%s %v - %s\n\rReputation: %v \n\r`\n\n", isUserNode, medal[i], i+1, nodeID, global.ReputationLB[i].Reputation)
		if len(entry) > 3800 {
			entries = append(entries, entry)
			entry = ""
		}
	}
	for i := 3; i < nLen; i++ {
		nodeID := global.ReputationLB[i].NodeID
		if checkUsersNode(nodeID, userNodes) {
			isUserNode := fmt.Sprintf("(Your node) %s", *(userNodes[nodeID]))
			entry += fmt.Sprintf("`%s\n\r%v - %s\n\rReputation: %v \n\r`\n\n", isUserNode, i+1, nodeID, global.ReputationLB[i].Reputation)
			if len(entry) > 3800 {
				entries = append(entries, entry)
				entry = ""
			}
		}
	}
	if len(entry) > 0 {
		entries = append(entries, entry)
	}
	for _, _entry := range entries {
		msg := tgbotapi.NewMessage(rcvrID, _entry)
		msg.ParseMode = "markdown"
		TgBot.Send(msg)
	}
}

func sendBlocksBoard(tgID, rcvrID int64) {
	nLen := len(global.BlocksLB)
	var entry string
	log.Logger.Debugf("Block Leaderboard: %v", global.HighestEpoch)

	first3 := int(math.Min(3, float64(nLen)))

	// user nodes
	userNodes := make(map[string]*string)
	if global.Users[tgID] != nil {
		userNodes = global.Users[tgID].Nodes
	}
	var entries []string

	// userBlockCount: Total block minted by all the nodes of user
	// totalBlockCount: Total block minted by the network
	// global.HighestEpoch: Number of epoch passed
	// blockPerNode: blocks minted by specific node
	var totalBlockCount, userBlockCount int64
	for i := 0; i < first3; i++ {
		var isUserNode string
		nodeID := global.BlocksLB[i].NodeID
		blockPerNode := global.BlocksLB[i].Blocks
		if checkUsersNode(nodeID, userNodes) {
			userBlockCount += blockPerNode
			isUserNode = fmt.Sprintf("(Your node) %s", *(userNodes[nodeID]))
		}
		totalBlockCount += blockPerNode
		entry += fmt.Sprintf("`%s\n\r%s %v - %s\n\rBlocks mined: %v \n\r`\n\n", isUserNode, medal[i], i+1, nodeID, blockPerNode)
		if len(entry) > 3800 {
			entries = append(entries, entry)
			entry = ""
		}
	}

	for i := 3; i < nLen; i++ {
		blockPerNode := global.BlocksLB[i].Blocks
		totalBlockCount += blockPerNode
		nodeID := global.BlocksLB[i].NodeID
		if checkUsersNode(nodeID, userNodes) {
			userBlockCount += blockPerNode
			isUserNode := fmt.Sprintf("(Your node) %s", *(userNodes[nodeID]))
			entry += fmt.Sprintf("`%s\n\r%v - %s\n\rBlocks mined: %v \n\r`\n\n", isUserNode, i+1, nodeID, blockPerNode)
			if len(entry) > 3800 {
				entries = append(entries, entry)
				entry = ""
			}
		}
	}
	header := fmt.Sprintf("🏆*Mining 🔨 LeaderBoard* (Max epoch: %v, Mined block: %v)\n\n\r `(Total Blocks mined by your nodes: %v)`\n\n",
		global.HighestEpoch, totalBlockCount, userBlockCount)
	if len(entry) > 0 {
		entries = append(entries, entry)
	}
	// iterate over msgs and send individual messages
	for i, _entry := range entries {
		if i == 0 {
			_entry = header + _entry
		}
		msg := tgbotapi.NewMessage(rcvrID, _entry)
		msg.ParseMode = "markdown"
		TgBot.Send(msg)
	}
}
