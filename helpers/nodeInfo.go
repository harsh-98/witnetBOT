package helpers

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/harsh-98/witnetBOT/log"
)

func sendNodesStats(tgID int, dbUser *UserType) {
	nLen := len(dbUser.Nodes)
	log.Logger.Debug(*dbUser)
	if nLen == 0 {
		msg := tgbotapi.NewMessage(int64(tgID), "⛔️ No nodes added yet")
		TgBot.Send(msg)
		return
	}
	for i, v := range dbUser.Nodes {
		log.Logger.Debug(v)
		n := global.Nodes[v]
		log.Logger.Debug(n)
		var status string
		if n == nil {
			if !Config.GetBool("allowReputationNilNodeStats") {
				continue
			}
			n = &NodeType{NodeID: v}
		}
		if n.Active {
			status = "Active ✅"
		} else {
			status = "Not Active ⭕️"
		}
		str := fmt.Sprintf("`Node %v/%v - %s\n\r\n\r"+
			"Name: %s\n\r"+
			"Reputation: %v`", i+1, nLen, status, n.NodeID, n.Reputation)
		var keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📖 Details", fmt.Sprintf(":NodeDetails_%v", n.NodeID)),
				tgbotapi.NewInlineKeyboardButtonData("⛔️ Remove", fmt.Sprintf(":RemoveNode_%v", n.NodeID)),
			),
		)
		msg := tgbotapi.NewMessage(int64(tgID), str)
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = keyboard
		TgBot.Send(msg)
	}
}

func sendNodeDetails(tgID int, nodeID string) {
	n := global.Nodes[nodeID]
	if n == nil {
		n = &NodeType{
			NodeID: nodeID,
		}
	}
	str := fmt.Sprintf("`NodeID: %s\n\r\n\rActive: %t\n\rReputation: %v\n\r`",
		n.NodeID, n.Active, n.Reputation)
	if !Config.GetBool("disableComplexQuery") {
		query := fmt.Sprintf(`
			select * from 
				(select count(epoch) as blockCount from blockchain where Miner =  "%s") as T1 
				inner join  
				(select group_concat(epoch) as epochs  from 
					(select * from blockchain where Miner =  "%s" order by   Epoch desc limit 5) as T) as T2 on true ;`, nodeID, nodeID)
		// log.Logger.Debug(query)
		rows, err := sqldb.Query(query)
		if err != nil {
			// log.Logger.Debug(err)
			msg := tgbotapi.NewMessage(int64(tgID), "⛔️ Fetching details for Node resulted in error")
			TgBot.Send(msg)
		}
		var (
			blockCount int
			epoch      string
		)
		for rows.Next() {
			rows.Scan(&blockCount, &epoch)
		}
		rows.Close()
		str += fmt.Sprintf("`BlockMinted: %v\n\rBlock submitted last 5 Epochs: %s\n\rBlock rewards : %v\n\r`",
			blockCount, epoch, 500*blockCount)
	}
	msg := tgbotapi.NewMessage(int64(tgID), str)
	msg.ParseMode = "markdown"
	TgBot.Send(msg)
}
