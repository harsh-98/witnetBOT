package helpers

import (
	"fmt"
	"math"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/harsh-98/witnetBOT/log"
)

const welcomeMessage = "Hello\n\r\n\r" +
	"I am @elliotsenpai 's unofficial witnet monitor BOT and I am here to help you keep an eye on your nodes."
const addNodeMsg = "Node's public key ? Starts with twit (Testnet: Len 43) or wit(Mainnet: Len 42). You can also enter multiple keys separated by space."
const broadcastMsg = "Reply with the message you want to broadcast"

var TgBot *tgbotapi.BotAPI

func addNodes(message *tgbotapi.Message) {
	dbUser, err := GetUserByTelegramID(int64(message.From.ID))
	if err != nil {
		log.Logger.Errorf("Unknown user TG ID = %v trying to add node %s\n\r", message.From.ID, message.Text)
		return
	}
	processKey := strings.Split(message.Text, " ")
	var keys []string
	// using map to handle duplicate key input from user
	duplicateHandle := make(map[string]bool)
	for _, k := range processKey {
		k = strings.Trim(k, " ")
		if k != "" {
			duplicateHandle[k] = true
		}
	}
	for k := range duplicateHandle {
		keys = append(keys, k)
	}
	log.Logger.Debug(keys)

	// handle 0 keys
	if len(keys) == 0 {
		msg := tgbotapi.NewMessage(int64(message.From.ID), fmt.Sprintf("⛔️ Invalid key(s) %s", message.Text))
		TgBot.Send(msg)
		return
	}
	// handle invalid keys
	for _, key := range keys {
		if !(len(key) == 43 && strings.HasPrefix(key, "twit")) && !(len(key) == 42 && strings.HasPrefix(key, "wit")) {
			msg := tgbotapi.NewMessage(int64(message.From.ID), fmt.Sprintf("⛔️ Invalid key %s \n Either wrong length or prefix ", key))
			TgBot.Send(msg)
			return
		}
	}
	// check if the key is present in userNodeMap
	if dbUser.Nodes != nil {
		for _, n := range dbUser.Nodes {
			for _, key := range keys {
				if n == key {
					msg := tgbotapi.NewMessage(int64(message.From.ID), fmt.Sprintf("⛔️ You have already added this key %s", key))
					TgBot.Send(msg)
					return
				}
			}

		}
	}
	var nodeStatus, repNodeStatus string
	userID := int64(message.From.ID)
	err = DB.AddUserNode(userID, keys)
	extra := " ```   I am watching 🧐 for these nodes , will notify if added to reputation list.\n\n   Meanwhile go have some water 🚰.```"
	if err == nil {
		for _, key := range keys {
			if global.Nodes[key] == nil {
				nodeStatus += fmt.Sprintf("key: %s is not in reputation list \n", key)
			} else {
				repNodeStatus += fmt.Sprintf("key: %s is in reputation list \n", key)
			}
		}
		if repNodeStatus != "" {
			msg := tgbotapi.NewMessage(userID, fmt.Sprintf("✅ Node(s) added!! ```\n\n%s \n```", repNodeStatus))
			msg.ParseMode = "markdown"
			TgBot.Send(msg)
		}
		if nodeStatus != "" {
			msg := tgbotapi.NewMessage(userID, fmt.Sprintf("✅ Node(s) added!! But ```\n\n%s \n``` %s", nodeStatus, extra))
			msg.ParseMode = "markdown"
			TgBot.Send(msg)
		}
	} else {
		log.Logger.Error(err)
		msg := tgbotapi.NewMessage(userID, "Failed adding address.")
		TgBot.Send(msg)
		ReportToAdmins(fmt.Sprintf("Failed adding node %s from user: %v", message.Text, userID))
	}
}
func ReplyReceived(message *tgbotapi.Message) {
	if message.ReplyToMessage.Text == addNodeMsg {
		addNodes(message)
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
	if cb.Data == "RatingGraph" {
		TgBot.AnswerCallbackQuery(tgbotapi.NewCallback(cb.ID, "Downloading. Please Wait !"))
		if !Config.GetBool("disableGraph") {
			GenerateGraph(int64(cb.From.ID))
		} else {
			msg := tgbotapi.NewMessage(int64(cb.From.ID), "`Graphs 📈 are currently disabled ⭕️ !!`")
			msg.ParseMode = "markdown"
			TgBot.Send(msg)
		}

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
		sendLeaderBoard(int64(cb.From.ID))
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
	if dbUser == nil {
		msg := tgbotapi.NewMessage(int64(tgUser.ID), "Enter command `/start` for getting started with bot 🤖")
		msg.ParseMode = "markdown"
		TgBot.Send(msg)
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
				tgbotapi.NewInlineKeyboardButtonData("📈 Get Rating graph", "RatingGraph"),
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
				tgbotapi.NewInlineKeyboardButtonData("📈 Get Rating graph", "RatingGraph"),
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
func checkUsersNode(v string, list []string) bool {
	for _, n := range list {
		if n == v {
			return true
		}
	}
	return false
}
func sendLeaderBoard(tgID int64) {
	nLen := len(global.Ranking)

	str := fmt.Sprintf("🏆 **Leader Board** (Nodes count: %v) \n\n", nLen)

	first3 := int(math.Min(3, float64(nLen)))
	for i := 0; i < first3; i++ {
		medal := []string{"🥇", "🥈", "🥉"}
		var isUserNode string
		if checkUsersNode(global.Ranking[i].NodeID, global.Users[tgID].Nodes) {
			isUserNode = "(Your node)"
		}
		str += fmt.Sprintf("`%s\n\r%s %v - %s\n\rReputation: %v \n\r`\n\n", isUserNode, medal[i], i+1, global.Ranking[i].NodeID, global.Ranking[i].Reputation)
	}
	isUserNode := "(Your node)"
	for i := 3; i < nLen; i++ {
		if checkUsersNode(global.Ranking[i].NodeID, global.Users[tgID].Nodes) {
			str += fmt.Sprintf("`%s\n\r%v - %s\n\rReputation: %v \n\r`\n\n", isUserNode, i+1, global.Ranking[i].NodeID, global.Ranking[i].Reputation)
		}
	}
	msg := tgbotapi.NewMessage(int64(tgID), str)
	msg.ParseMode = "markdown"
	TgBot.Send(msg)

}
