﻿package helpers

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/harsh-98/witnetBOT/log"
)

const welcomeMessage = "Hello\n\r\n\r" +
	"I am @elliotsenpai 's unofficial witnet monitor BOT and I am here to help you keep an eye on your nodes."
const nodeNameLength = 50
const addNodeMsg = "Node's public key: Starts with wit(Mainnet: Len 42). \n\n" +
	"You can also enter multiple keys separated by space.\n\n" +
	"For naming your node provide name separated by comma(name length should be less than 50).\n\n" +
	" Example: nodeID1,nodeName1 nodeID2,nodeName2"
const broadcastMsg = "Reply with the message you want to broadcast"
const nameNodeMsg = "Provide name for node: "

var TgBot *tgbotapi.BotAPI

func ReplyReceived(message *tgbotapi.Message) {
	if message.ReplyToMessage.Text == addNodeMsg {
		addNodes(message)
	}
	if message.ReplyToMessage.Text == broadcastMsg {
		for _, u := range global.Users {
			if u.UserID != int64(message.From.ID) {
				msg := tgbotapi.NewMessage(int64(u.UserID), fmt.Sprintf("📢 <b>Important message from harshjain</b>\n\r\n\r%s", message.Text))
				msg.ParseMode = "HTML"
				TgBot.Send(msg)
			}
		}
	}
	if strings.HasPrefix(message.ReplyToMessage.Text, nameNodeMsg) {
		nodeID := strings.TrimPrefix(message.ReplyToMessage.Text, nameNodeMsg)
		DB.NameNode(message, nodeID)
	}
	mainMenu(message.From)
}

func CallbackQueryReceived(cb *tgbotapi.CallbackQuery) {
	if cb.Data == "NodesStats" {
		TgBot.AnswerCallbackQuery(tgbotapi.NewCallback(cb.ID, "Nodes Stats"))
		dbUser, err := GetUserByTelegramID(int64(cb.From.ID))
		if err == nil {
			sendNodesStats(int64(cb.From.ID), dbUser)
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
	if cb.Data == "GenesisReward" {
		TgBot.AnswerCallbackQuery(tgbotapi.NewCallback(cb.ID, "Fetching result"))
		GenesisReward(int64(cb.From.ID))
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
	if cb.Data == "ReputationBoard" {
		TgBot.AnswerCallbackQuery(tgbotapi.NewCallback(cb.ID, "Reputation Board"))
		sendReputationBoard(int64(cb.From.ID), int64(cb.From.ID))
	}
	if cb.Data == "BlockBoard" {
		TgBot.AnswerCallbackQuery(tgbotapi.NewCallback(cb.ID, "Block Board"))
		sendBlocksBoard(int64(cb.From.ID), int64(cb.From.ID))
	}
	if cb.Data[0] == ':' {
		TgBot.AnswerCallbackQuery(tgbotapi.NewCallback(cb.ID, "Ok"))
		var nodeID, text string
		if strings.HasPrefix(cb.Data, ":NodeDetails") {
			nodeID = strings.TrimPrefix(cb.Data, ":NodeDetails_")
			sendNodeDetails(int64(cb.From.ID), nodeID)
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
		if strings.HasPrefix(cb.Data, ":NodeNameIt") {
			nodeID = strings.TrimPrefix(cb.Data, ":NodeNameIt_")
			TgBot.AnswerCallbackQuery(tgbotapi.NewCallback(cb.ID, "Name you node"))
			msg := tgbotapi.NewMessage(int64(cb.From.ID), nameNodeMsg+nodeID)
			msg.ReplyMarkup = tgbotapi.ForceReply{
				ForceReply: true,
				Selective:  false,
			}
			TgBot.Send(msg)
		}
	}
	mainMenu(cb.From)
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
			ReportToAdmins(fmt.Sprintf("🧝‍♂️ Updated user: '%v' '%s' '%s' '%s'\n\r",
				update.Message.From.ID, update.Message.From.UserName, update.Message.From.FirstName, update.Message.From.LastName))
			if err != nil {
				ReportToAdmins(fmt.Sprintf("⛔️ An error occured while updating user ID %v in DB: %s", dbUser.UserID, err))
				return
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
			Nodes:     make(map[string]*string),
		}
		err = DB.AddUser(dbUser)
		if err != nil {
			ReportToAdmins("⛔️ An error occured while saving him in DB")
			return
		}
		global.Users[dbUser.UserID] = dbUser
	}

	log.Logger.Debug(update.Message.CommandArguments())
	if update.Message.Command() == "start" { // && update.Message.Chat.IsPrivate() {
		startCommandReceived(update.Message.From)
	} else if update.Message.Command() == "admin" {
		err = adminActionCmd(update.Message)
		if err != nil {
			msg := tgbotapi.NewMessage(int64(update.Message.From.ID), welcomeMessage)
			msg.ParseMode = "HTML"
			TgBot.Send(msg)
		}
	}
}

func adminActionCmd(msg *tgbotapi.Message) error {
	tgUser := msg.From
	for _, u := range global.Admin {
		// check if user is admin
		if u.UserID == int64(tgUser.ID) {
			argStr := msg.CommandArguments()
			args := strings.Split(argStr, " ")
			// check length of arguments to command
			if len(args) < 2 {
				return fmt.Errorf("Args len is %v", len(args))
			}
			// function name and mimicUserID
			funcName := args[0]
			id, err := strconv.Atoi(args[1])
			if err != nil {
				return err
			}
			mimicUserID := int64(id)

			// switch for function calls
			switch funcName {
			case "repLB":
				sendReputationBoard(mimicUserID, int64(tgUser.ID))
			case "blkLB":
				sendBlocksBoard(mimicUserID, int64(tgUser.ID))
			default:
				return fmt.Errorf("Function not defined %s", funcName)
			}
		}
	}
	mainMenu(tgUser)
	return nil
}

func startCommandReceived(tgUser *tgbotapi.User) {
	msg := tgbotapi.NewMessage(int64(tgUser.ID), welcomeMessage)
	msg.ParseMode = "HTML"
	TgBot.Send(msg)
	mainMenu(tgUser)
}

func ReportToAdmins(message string) {
	for _, u := range global.Admin {
		msg := tgbotapi.NewMessage(int64(u.UserID), message)
		TgBot.Send(msg)
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
				tgbotapi.NewInlineKeyboardButtonData("🎖 Reputation  Leader Board 🏆", "ReputationBoard"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔨 Mining Leader Board 🏆", "BlockBoard"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🤑 Testnet Rewards", "GenesisReward"),
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
				tgbotapi.NewInlineKeyboardButtonData("🎖 Reputation  Leader Board 🏆", "ReputationBoard"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔨 Mining Leader Board 🏆", "BlockBoard"),
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
