package helpers

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/harsh-98/witnetBOT/log"
)

func sendNodesStats(tgID int64, dbUser *UserType) {
	nLen := len(dbUser.Nodes)
	log.Logger.Debug(*dbUser)
	if nLen == 0 {
		msg := tgbotapi.NewMessage(tgID, "⛔️ No nodes added yet")
		TgBot.Send(msg)
		return
	}
	i := 0
	for nodeID, nodeName := range dbUser.Nodes {
		n := global.Nodes[nodeID]
		log.Logger.Debug("Nodeid ", nodeID, " Node: ", n, " Name: ", *nodeName)
		var status string
		if n == nil {
			if !Config.GetBool("allowReputationNilNodeStats") {
				continue
			}
			n = &NodeType{NodeID: nodeID}
		}
		if n.Active {
			status = "Active ✅"
		} else {
			status = "Not Active ⭕️"
		}

		// nodeStats string
		str := fmt.Sprintf("`Node %v/%v - %s\n\r\n\r"+
			"ID: %s\n\r"+
			"Name: %s \n\r"+
			"Reputation: %v`", i+1, nLen, status, n.NodeID, *nodeName, n.Reputation)

		// add buttons for node
		var buttons []tgbotapi.InlineKeyboardButton
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("📖 Details", fmt.Sprintf(":NodeDetails_%v", n.NodeID)))
		if *nodeName == "" {
			buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("✍️ Name It", fmt.Sprintf(":NodeNameIt_%v", n.NodeID)))
		}
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("⛔️ Remove", fmt.Sprintf(":RemoveNode_%v", n.NodeID)))
		// keyboard for node
		var keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				buttons...,
			),
		)
		msg := tgbotapi.NewMessage(tgID, str)
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = keyboard
		TgBot.Send(msg)
		i++
	}
}

func sendNodeDetails(tgID int64, nodeID string) {
	n := global.Nodes[nodeID]
	nodeName := global.Users[tgID].Nodes[nodeID]
	if n == nil {
		n = &NodeType{
			NodeID: nodeID,
		}
	}
	str := fmt.Sprintf("`NodeID: %s\n\rNodeName: %s\n\r\n\rActive: %t\n\rReputation: %v\n\r`",
		n.NodeID, *nodeName, n.Active, n.Reputation)

	// Preventing sql injection
	rows, err := sqldb.Query("select blockCount, reward, lastXEpochs from lightBlockchain where Miner=?;", nodeID)
	// `select * from
	// 	(select count(epoch), sum(reward) as blockCount from blockchain where Miner=?) as T1
	// 	inner join
	// 	(select group_concat(epoch) as epochs  from
	// 		(select * from blockchain where Miner=? order by   Epoch desc limit 5) as T) as T2 on true ;`
	// log.Logger.Debug(query)
	if err != nil {
		// log.Logger.Debug(err)
		msg := tgbotapi.NewMessage(tgID, "⛔️ Fetching details for Node resulted in error")
		TgBot.Send(msg)
	}
	var (
		blockCount  int
		lastXEpochs string
		reward      int
	)
	for rows.Next() {
		rows.Scan(&blockCount, &reward, &lastXEpochs)
	}
	rows.Close()
	str += fmt.Sprintf("`BlockMinted: %v\n\rBlock submitted last 5 Epochs: %s\n\rBlock rewards : %v\n\r`",
		blockCount, strings.Trim(lastXEpochs, ","), reward)
	msg := tgbotapi.NewMessage(tgID, str)
	msg.ParseMode = "markdown"
	TgBot.Send(msg)
}
