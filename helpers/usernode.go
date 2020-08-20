package helpers

import (
	"fmt"
	"strings"

	"errors"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/harsh-98/witnetBOT/log"
)

type UserNode struct {
	NodeID string
	UserID int64
}

func (d DataBaseType) RemoveUserNode(nodeID string, userID int64) error {
	// safe query
	_, err := sqldb.Query("delete from userNodeMap where NodeID=? and UserID=?", nodeID, userID)
	// log.Logger.Debug(str)
	if err != nil {
		log.Logger.Error("Error removing node from DB")
		return err
	}
	// remove node from user's list
	// nodes := global.Users[userID].Nodes
	delete(global.Users[userID].Nodes, nodeID)
	// for j, n := range nodes {
	// 	if n == nodeID {
	// 		log.Logger.Debug("Remove node from user")
	// 		global.Users[userID].Nodes = append(nodes[:j], nodes[j+1:]...)
	// 		break
	// 	}
	// }

	// remove user from node owner
	users := global.NodeUsers[nodeID]
	for j, u := range users {
		if u == userID {
			log.Logger.Debug("Remove user from node owner")
			global.NodeUsers[nodeID] = append(users[:j], users[j+1:]...)
			break
		}
	}

	return nil
}
func (d DataBaseType) AddUserNode(userID int64, nodeIDToName map[string]string) error {
	if len(nodeIDToName) == 0 {
		return errors.New("Usernode list is empty")
	}
	var report string
	userName := global.Users[userID].UserName

	i := 0
	// rows of multiple insert query
	var rows [][]interface{}
	for nodeID, nodeName := range nodeIDToName {
		// the row that will be inserted in the db
		rows = append(rows, []interface{}{userID, nodeID, nodeName})
		// generate report to be sent to admins
		report += fmt.Sprintf("%v: %s (%s)\n", i+1, nodeName, nodeID)
		i += 1
	}

	// exec multiple insert query
	err := multipleInsert("insert into userNodeMap values (?, ?, ?);", rows)
	if err != nil {
		log.Logger.Errorf("Error adding node to DB: %s\n\r", err)
		return err
	}

	// send report that new nodes are needed for user
	ReportToAdmins(fmt.Sprintf("Username: %s (ID: %v) %s", userName, userID, report))

	// add node-user in global.NodeUser for searching users by nodeid
	for nodeID, nodeName := range nodeIDToName {
		global.NodeUsers[nodeID] = append(global.NodeUsers[nodeID], userID)
		// add user's node in global.Users
		global.Users[userID].Nodes[nodeID] = &nodeName
	}
	log.Logger.Trace(fmt.Sprintf("Node of %v:", userID), global.Users[userID].Nodes)
	return nil
}

func addNodes(message *tgbotapi.Message) {
	dbUser, err := GetUserByTelegramID(int64(message.From.ID))
	if err != nil {
		log.Logger.Errorf("Unknown user TG ID = %v trying to add node %s\n\r", message.From.ID, message.Text)
		return
	}
	processKey := strings.Split(message.Text, " ")
	var keys []string
	// using map to handle duplicate key input from user
	// also maps nodeid to nodename
	nodeIDToName := make(map[string]string)
	for _, k := range processKey {
		k = strings.Trim(k, " ")
		if k != "" {
			node := strings.Split(k, ",")
			if len(node) == 1 {
				nodeIDToName[node[0]] = ""
			} else {
				if len(node[1]) > nodeNameLength {
					msg := tgbotapi.NewMessage(int64(message.From.ID), fmt.Sprintf("`⛔️ Invalid length of name %s`", node[1]))
					msg.ParseMode = "markdown"
					TgBot.Send(msg)
					return
				}
				nodeIDToName[node[0]] = node[1]
			}
		}
	}
	for k := range nodeIDToName {
		keys = append(keys, k)
	}
	log.Logger.Debug(keys)

	// handle 0 keys
	if len(keys) == 0 {
		msg := tgbotapi.NewMessage(int64(message.From.ID), fmt.Sprintf("`⛔️ Invalid key(s) %s`", message.Text))
		msg.ParseMode = "markdown"
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
		// TODO decrease complexity
		for n, _ := range dbUser.Nodes {
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
	err = DB.AddUserNode(userID, nodeIDToName)
	extra := " ```   I am watching 🧐 for these nodes , will notify if added to reputation list.\n\n   Meanwhile go have some water 🚰.```"
	if err == nil {
		for _, key := range keys {
			if global.NodeRepMap[key] == nil {
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

func (d DataBaseType) NameNode(message *tgbotapi.Message, nodeID string) {
	userID := int64(message.From.ID)
	userName := global.Users[userID].UserName
	nodeName := message.Text
	if len(nodeName) > nodeNameLength {
		msg := tgbotapi.NewMessage(userID, "Name length more than 50 characters")
		msg.ParseMode = "markdown"
		TgBot.Send(msg)
		return
	}
	// safe query
	_, err := sqldb.Query("update userNodeMap set NodeName=? where NodeID=? and UserID=?;", nodeName, nodeID, userID)
	var msg tgbotapi.MessageConfig
	var failed string
	if err != nil {
		log.Logger.Errorf("Error adding node to DB: %s\n\r", err)
		failed = "⛔️ Failed"
		msg = tgbotapi.NewMessage(userID, fmt.Sprintf("`⛔️ Failed naming Node %s named to %s`", nodeID, nodeName))
	} else {
		msg = tgbotapi.NewMessage(userID, fmt.Sprintf("`✅ Node %s named to %s`", nodeID, nodeName))
	}
	msg.ParseMode = "markdown"
	TgBot.Send(msg)

	// add node name to the user node list
	global.Users[userID].Nodes[nodeID] = &nodeName

	report := fmt.Sprintf("%s Username: %v named node '%s' to '%s'", failed, userName, nodeID, nodeName)
	ReportToAdmins(report)
}
