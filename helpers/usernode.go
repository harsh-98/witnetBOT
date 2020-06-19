package helpers

import (
	"fmt"

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
func (d DataBaseType) AddUserNode(n UserNode) error {
	str := fmt.Sprintf("insert into userNodeMap values (%v, '%s')", n.UserID, n.NodeID)
	_, err := sqldb.Exec(str)
	if err != nil {
		log.Logger.Errorf("Error adding user's %v node to DB: %s\n\r", n.NodeID, err)
		return err
	}
	// rows, err := sqldb.Query(fmt.Sprintf("select * from tblNodes where NodeID='%s'", n.NodeID))
	// if err != nil {
	// 	log.Logger.Error(err)
	// 	return err
	// }
	// var (
	// 	node       NodeType
	// 	active     bool
	// 	reputation float64
	// 	blocks     int32
	// 	nodeID     string
	// )
	// for rows.Next() {
	// 	rows.Scan(&nodeID, &active, &reputation, &blocks)
	// 	node = NodeType{
	// 		NodeID:     nodeID,
	// 		Blocks:     blocks,
	// 		Reputation: reputation,
	// 		Active:     active,
	// 	}
	// }
	// rows.Close()
	global.Users[n.UserID].Nodes = append(global.Users[n.UserID].Nodes, n.NodeID)
	return nil
}
