﻿package helpers

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/harsh-98/witnetBOT/log"
)

const welcomeMessage = "Hello\n\r\n\r" +
	"I am @elliotsenpai 's unofficial witnet monitor BOT and I am here to help you keep an eye on your nodes."
const addNodeMsg = "Node's public key ? Starts with twit (Testnet: Len 43) or wit(Mainnet: Len 42)"
const broadcastMsg = "Reply with the message you want to broadcast"

var TgBot *tgbotapi.BotAPI

func ReplyReceived(message *tgbotapi.Message) {
	if message.ReplyToMessage.Text == addNodeMsg {
		dbUser, err := GetUserByTelegramID(int64(message.From.ID))
		if err != nil {
			fmt.Printf("Unknown user TG ID = %v trying to add node %s\n\r", message.From.ID, message.Text)
			return
		}
		key := message.Text
		if !(len(key) == 43 && strings.HasPrefix(key, "twit")) && !(len(key) == 42 && strings.HasPrefix(key, "wit")) {
			msg := tgbotapi.NewMessage(int64(message.From.ID), "⛔️ Invalid key length")
			TgBot.Send(msg)
			return
		}
		// var nKey, bKey string
		// for _, a := range NetworkConfig.InitialNodes {
		// 	if a.PubKey == key || a.Address == key {
		// 		bKey = a.Address
		// 		nKey = a.PubKey
		// 	}
		// }
		// if bKey == "" || nKey == "" {
		// 	msg := tgbotapi.NewMessage(int64(message.From.ID), "⛔️ Not a Validator key")
		// 	TgBot.Send(msg)
		// 	return
		// }
		if dbUser.Nodes != nil {
			for _, n := range dbUser.Nodes {
				if n == key {
					msg := tgbotapi.NewMessage(int64(message.From.ID), "⛔️ You have already added this key")
					TgBot.Send(msg)
					return
				}
			}
		}
		var userNode = UserNode{
			UserID: int64(dbUser.UserID),
			NodeID: key,
		}
		node, err := DB.AddUserNode(userNode)
		if err == nil {
			dbUser.Nodes = append(dbUser.Nodes, node.NodeID)
			msg := tgbotapi.NewMessage(int64(message.From.ID), "✅ Node added")
			TgBot.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(int64(message.From.ID), fmt.Sprintf("Failed node with  %s doesn't exist", key))
			TgBot.Send(msg)
		}
	}
	if message.ReplyToMessage.Text == broadcastMsg {
		for _, u := range global.Users {
			if !u.IsAdmin {
				msg := tgbotapi.NewMessage(int64(u.UserID), fmt.Sprintf("📢 <b>Important message from harshjain</b>\n\r\n\r%s", message.Text))
				msg.ParseMode = "HTML"
				TgBot.Send(msg)
			}
		}
	}
	mainMenu(message.From)
}

func CallbackQueryReceived(cb *tgbotapi.CallbackQuery) {
	if cb.Data == "NodesStats" {
		TgBot.AnswerCallbackQuery(tgbotapi.NewCallback(cb.ID, "Nodes Stats"))
		dbUser, err := GetUserByTelegramID(int64(cb.From.ID))
		if err == nil {
			sendNodesStats(cb.From.ID, dbUser)
		}
	}
	if cb.Data == "AddUserNode" {
		TgBot.AnswerCallbackQuery(tgbotapi.NewCallback(cb.ID, "Add node"))
		msg := tgbotapi.NewMessage(int64(cb.From.ID), addNodeMsg)
		msg.ReplyMarkup = tgbotapi.ForceReply{
			ForceReply: true,
			Selective:  false,
		}
		TgBot.Send(msg)
		return
	}
	if cb.Data == "Download" {
		TgBot.AnswerCallbackQuery(tgbotapi.NewCallback(cb.ID, "Download. Please Wait !"))
		fileable := tgbotapi.NewDocumentUpload(int64(cb.From.ID), "node.zip")
		TgBot.Send(fileable)
	}
	if cb.Data == "Broadcast" {
		TgBot.AnswerCallbackQuery(tgbotapi.NewCallback(cb.ID, "Broadcast message"))
		msg := tgbotapi.NewMessage(int64(cb.From.ID), broadcastMsg)
		msg.ReplyMarkup = tgbotapi.ForceReply{
			ForceReply: true,
			Selective:  false,
		}
		TgBot.Send(msg)
		return
	}
	if cb.Data == "LeaderBoard" {
		TgBot.AnswerCallbackQuery(tgbotapi.NewCallback(cb.ID, "Leader Board"))
		sendLeaderBoard(cb.From.ID)
	}
	if cb.Data[0] == ':' {
		TgBot.AnswerCallbackQuery(tgbotapi.NewCallback(cb.ID, "Ok"))
		var nodeID, text string
		if strings.HasPrefix(cb.Data, ":NodeDetails") {
			nodeID = strings.TrimPrefix(cb.Data, ":NodeDetails_")
			sendNodeDetails(cb.From.ID, nodeID)
		}
		if strings.HasPrefix(cb.Data, ":RemoveNode") {
			nodeID = strings.TrimPrefix(cb.Data, ":RemoveNode_")
			text = "✅ Node removed"
			err := DB.RemoveUserNode(nodeID, int64(cb.From.ID))
			if err != nil {
				text = fmt.Sprintf("⛔️ %s", err)
			}
			msg := tgbotapi.NewMessage(int64(cb.From.ID), text)
			TgBot.Send(msg)
		}
	}
	mainMenu(cb.From)
}

func sendNodeDetails(tgID int, nodeID string) {
	for _, n := range global.Nodes {
		if n.NodeID == nodeID {
			str := fmt.Sprintf("`NodeID: %s\n\rActive: %t\n\rReputation: %v\n\rBlock: %v\n\r`",
				n.NodeID, n.Active, n.Reputation, n.Blocks)
			msg := tgbotapi.NewMessage(int64(tgID), str)
			msg.ParseMode = "markdown"
			TgBot.Send(msg)
			return
		}
	}
	msg := tgbotapi.NewMessage(int64(tgID), "⛔️ Node not found")
	TgBot.Send(msg)
}

func CommandReceived(update tgbotapi.Update) {
	var dbUser *UserType
	var err error
	dbUser, err = GetUserByTelegramID(int64(update.Message.From.ID))
	if err == nil {
		if dbUser.UserName != update.Message.From.UserName ||
			dbUser.FirstName != update.Message.From.FirstName ||
			dbUser.LastName != update.Message.From.LastName {
			dbUser.UserName = update.Message.From.UserName
			dbUser.FirstName = update.Message.From.FirstName
			dbUser.LastName = update.Message.From.LastName
			err := DB.UpdateUser(dbUser)
			if err != nil {
				ReportToAdmins(fmt.Sprintf("⛔️ An error occured while updating user ID %v in DB: %s", dbUser.UserID, err))
			}
		}
	} else {
		ReportToAdmins(fmt.Sprintf("🧝‍♂️ New registered user: '%v' '%s' '%s' '%s'\n\r",
			update.Message.From.ID, update.Message.From.UserName, update.Message.From.FirstName, update.Message.From.LastName))
		dbUser = &UserType{
			UserID:    int64(update.Message.From.ID),
			UserName:  update.Message.From.UserName,
			FirstName: update.Message.From.FirstName,
			LastName:  update.Message.From.LastName,
			IsAdmin:   false,
		}
		err = DB.AddUser(dbUser)
		if err != nil {
			ReportToAdmins("⛔️ An error occured while saving him in DB")
			return
		}
		global.Users[dbUser.UserID] = dbUser
	}
	if update.Message.Command() == "start" { // && update.Message.Chat.IsPrivate() {
		startCommandReceived(update.Message.From)
	}
}

func startCommandReceived(tgUser *tgbotapi.User) {
	msg := tgbotapi.NewMessage(int64(tgUser.ID), welcomeMessage)
	msg.ParseMode = "HTML"
	TgBot.Send(msg)
	mainMenu(tgUser)
}

func ReportToAdmins(message string) {
	for _, u := range global.Users {
		if u.IsAdmin {
			msg := tgbotapi.NewMessage(int64(u.UserID), message)
			TgBot.Send(msg)
		}
	}
}

func mainMenu(tgUser *tgbotapi.User) {
	dbUser, _ := GetUserByTelegramID(int64(tgUser.ID)) //uint32
	if dbUser.UserID == 0 {
		return
	}
	if dbUser.LastMenuID > 0 {
		TgBot.DeleteMessage(tgbotapi.DeleteMessageConfig{
			ChatID:    int64(tgUser.ID),
			MessageID: dbUser.LastMenuID,
		})
	}
	var keyboard tgbotapi.InlineKeyboardMarkup
	msg := tgbotapi.NewMessage(int64(tgUser.ID), "How may I assist you ?")
	if dbUser.IsAdmin {
		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📊 Nodes Stats", "NodesStats"),
				tgbotapi.NewInlineKeyboardButtonData("➕ Add node", "AddUserNode"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏆 Leader Board", "LeaderBoard"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📥 Download last version", "Download"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📣 Broadcast message", "Broadcast"),
			),
		)
	} else {
		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📊 Nodes Stats", "NodesStats"),
				tgbotapi.NewInlineKeyboardButtonData("➕ Add node", "AddUserNode"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏆 Leader Board", "LeaderBoard"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📥 Download last version", "Download"),
			),
		)
	}
	msg.ReplyMarkup = keyboard
	resp, _ := TgBot.Send(msg)
	dbUser.LastMenuID = resp.MessageID
}

func sendNodesStats(tgID int, dbUser *UserType) {
	nLen := len(dbUser.Nodes)
	log.Logger.Debug(*dbUser)
	if dbUser == nil {
		msg := tgbotapi.NewMessage(int64(tgID), "⛔️ No nodes added yet")
		TgBot.Send(msg)
		return
	}
	for i, v := range dbUser.Nodes {
		for _, n := range global.Nodes {
			if n.NodeID == v {
				var status string
				if n.Active {
					status = "Online ✅"
				} else {
					status = "Offline ⭕️"
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
	}
}

func sendLeaderBoard(tgID int) {
	msg := tgbotapi.NewMessage(int64(tgID), "")
	msg.ParseMode = "markdown"
	TgBot.Send(msg)
	// top := make([]int, len(NetworkConfig.InitialNodes))
	// for i, _ := range NetworkConfig.InitialNodes {
	// 	top[i] = i
	// }
	// w := true
	// idx := 0
	// var tmp int
	// for {
	// 	var sum1, val1, sum2, val2 float64 = 0, 0, 0, 0
	// 	sum1 = float64(NetworkConfig.InitialNodes[top[idx]].Hb.TotalUpTimeSec) + float64(NetworkConfig.InitialNodes[top[idx]].Hb.TotalDownTimeSec)
	// 	if sum1 != 0 {
	// 		val1 = float64(NetworkConfig.InitialNodes[top[idx]].Hb.TotalUpTimeSec) / sum1
	// 	}
	// 	sum2 = float64(NetworkConfig.InitialNodes[top[idx+1]].Hb.TotalUpTimeSec) + float64(NetworkConfig.InitialNodes[top[idx+1]].Hb.TotalDownTimeSec)
	// 	if sum2 != 0 {
	// 		val2 = float64(NetworkConfig.InitialNodes[top[idx+1]].Hb.TotalUpTimeSec) / sum2
	// 	}
	// 	if val1 < val2 {
	// 		tmp = top[idx]
	// 		top[idx] = top[idx+1]
	// 		top[idx+1] = tmp
	// 		w = false
	// 	}
	// 	idx++
	// 	if idx >= len(top)-2 {
	// 		if w {
	// 			break
	// 		} else {
	// 			w = true
	// 			idx = 0
	// 		}
	// 	}
	// }
	// msg := tgbotapi.NewMessage(int64(tgID), "<b>🏆 Leader Board</b>")
	// msg.ParseMode = "HTML"
	// TgBot.Send(msg)
	// for i := 0; i < 5; i++ {
	// 	shard := fmt.Sprintf("%v", NetworkConfig.InitialNodes[top[i]].Hb.ReceivedShard)
	// 	if NetworkConfig.InitialNodes[top[i]].Hb.ReceivedShard == 4294967295 {
	// 		shard = "meta"
	// 	}
	// 	str := fmt.Sprintf("`#%v %s - shard %s\n\r%s\n\rUp:%v Down:%v`", i+1,
	// 		NetworkConfig.InitialNodes[top[i]].Hb.NodeDisplayName,
	// 		shard,
	// 		NetworkConfig.InitialNodes[top[i]].Address,
	// 		NetworkConfig.InitialNodes[top[i]].Hb.TotalUpTimeSec,
	// 		NetworkConfig.InitialNodes[top[i]].Hb.TotalDownTimeSec)
	// 	msg := tgbotapi.NewMessage(int64(tgID), str)
	// 	msg.ParseMode = "markdown"
	// 	TgBot.Send(msg)
	// }
	// dbUser, err := GetUserByTelegramID(uint32(tgID))
	// if err != nil {
	// 	return
	// }
	// for i := 5; i < len(top); i++ {
	// 	for _, n := range *dbUser.Nodes {
	// 		if NetworkConfig.InitialNodes[top[i]].Address == n.BalancesKey {
	// 			shard := fmt.Sprintf("%v", NetworkConfig.InitialNodes[top[i]].Hb.ReceivedShard)
	// 			if NetworkConfig.InitialNodes[top[i]].Hb.ReceivedShard == 4294967295 {
	// 				shard = "meta"
	// 			}
	// 			str := fmt.Sprintf("`#%v %s - shard %s\n\r%s\n\rUp:%v Down:%v`", i+1,
	// 				NetworkConfig.InitialNodes[top[i]].Hb.NodeDisplayName,
	// 				shard,
	// 				NetworkConfig.InitialNodes[top[i]].Address,
	// 				NetworkConfig.InitialNodes[top[i]].Hb.TotalUpTimeSec,
	// 				NetworkConfig.InitialNodes[top[i]].Hb.TotalDownTimeSec)
	// 			msg := tgbotapi.NewMessage(int64(tgID), str)
	// 			msg.ParseMode = "markdown"
	// 			TgBot.Send(msg)
	// 		}
	// 	}
	// }
}
