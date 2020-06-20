package helpers

import (
	"fmt"

	"errors"

	"github.com/harsh-98/witnetBOT/log"
)

type UserNode struct {
	NodeID string
	UserID int64
}

func (d DataBaseType) RemoveUserNode(nodeID string, userID int64) error {
	str := fmt.Sprintf("delete from UserNodeMap where NodeID = '%s' and UserID = %v", nodeID, userID)
	log.Logger.Debug(str)
	_, err := sqldb.Exec(str)
	if err != nil {
		log.Logger.Error("Error removing node from DB")
		return err
	}
	// remove node from user's list
	nodes := global.Users[userID].Nodes
	for j, n := range nodes {
		if n == nodeID {
			log.Logger.Debug("Remove node from user")
			global.Users[userID].Nodes = append(nodes[:j], nodes[j+1:]...)
			break
		}
	}
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
func (d DataBaseType) AddUserNode(userID int64, nodeIDs []string) error {
	if len(nodeIDs) == 0 {
		return errors.New("Usernode list is empty")
	}
	var str string
	var report string
	userName := global.Users[userID].UserName

	for i, nodeID := range nodeIDs {
		str = fmt.Sprintf("%s insert into userNodeMap values (%v, '%s');", str, userID, nodeID)
		report = fmt.Sprintf("%s %v: %s\n", report, i+1, nodeID)
	}
	// str = fmt.Sprintf("%s insert into userNodeMap values (%v, '%s');", str, n.UserID, n.NodeID)
	_, err := sqldb.Exec(str)
	if err != nil {
		log.Logger.Errorf("Error adding node to DB: %s\n\r", err)
		return err
	}
	// send report that new nodes are needed for user
	ReportToAdmins(fmt.Sprintf("Username: %s (ID: %v) %s", userName, userID, report))

	// add node-user in global.NodeUser for searching users by nodeid
	for _, nodeID := range nodeIDs {
		global.NodeUsers[nodeID] = append(global.NodeUsers[nodeID], userID)
	}
	// add user's node in global.Users
	global.Users[userID].Nodes = append(global.Users[userID].Nodes, nodeIDs...)
	return nil
}
