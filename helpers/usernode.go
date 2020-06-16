package helpers

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type UserNode struct {
	NodeID string
	UserID int64
}

func (d DataBaseType) RemoveUserNode(nodeID string, userID int64) error {
	str := fmt.Sprintf("delete from UserNodeMap where NodeID = '%s' and UserID = %v", nodeID, userID)
	log.Debug(str)
	_, err := sqldb.Exec(str)
	if err != nil {
		fmt.Println("Error removing node from DB")
		return err
	}
	for i, u := range global.Users {
		if u.UserID == userID {
			for j, n := range global.Users[i].Nodes {
				if n == nodeID {
					global.Users[i].Nodes = append(global.Users[i].Nodes[:j], global.Users[i].Nodes[j+1:]...)
				}
			}
		}
	}
	return nil
}
func (d DataBaseType) AddUserNode(n UserNode) (NodeType, error) {
	str := fmt.Sprintf("insert into userNodeMap values (%v, '%s')", n.UserID, n.NodeID)
	_, err := sqldb.Exec(str)
	if err != nil {
		log.Errorf("Error adding user's %v node to DB: %s\n\r", n.NodeID, err)
		return NodeType{}, err
	}
	rows, err := sqldb.Query(fmt.Sprintf("select * from tblNodes where NodeID='%s'", n.NodeID))
	if err != nil {
		log.Error(err)
		return NodeType{}, err
	}
	var (
		node       NodeType
		active     bool
		reputation float64
		blocks     int32
		nodeID     string
	)
	for rows.Next() {
		rows.Scan(&nodeID, &active, &reputation, &blocks)
		node = NodeType{
			NodeID:     nodeID,
			Blocks:     blocks,
			Reputation: reputation,
			Active:     active,
		}
	}
	rows.Close()
	for i, u := range global.Users {
		if u.UserID == n.UserID {
			global.Users[i].Nodes = append(u.Nodes, nodeID)
		}
	}
	return node, nil
}
