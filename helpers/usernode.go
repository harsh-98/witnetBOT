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
		fmt.Println("Error removing node from DB")
		return err
	}
	nodes := global.Users[userID].Nodes
	for j, n := range nodes {
		if n == nodeID {
			fmt.Println("remove")
			global.Users[userID].Nodes = append(nodes[:j], nodes[j+1:]...)
			return nil
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

	ReportToAdmins(fmt.Sprintf("Username: %s (ID: %v) %s", userName, userID, report))

	global.Users[userID].Nodes = append(global.Users[userID].Nodes, nodeIDs...)
	return nil
}
