package helpers

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/harsh-98/witnetBOT/log"
)

type NodeType struct {
	Reputation float64
	NodeID     string
	Blocks     int32
	Active     bool
}

func (d *DataBaseType) AddNodesInTable(nodes map[string]*NodeType) error {
	loc, _ := time.LoadLocation("UTC")
	t := time.Now().In(loc)
	log.Logger.Infof(" number of nodes: %v", len(nodes))

	// truncate reputation
	_, err := sqldb.Query("truncate tblNodes;")
	if err != nil {
		log.Logger.Errorf("DB: Failed truncating tblNodes: %s\n\r", err)
		return err
	}

	// insert rows in reputation and tblNodes
	var tblNodeRows, reputationRows [][]interface{}
	for _, node := range nodes {
		tblNodeRows = append(tblNodeRows, []interface{}{node.NodeID, node.Active, node.Reputation, node.Blocks, node.Active, node.Reputation})
		reputationRows = append(reputationRows, []interface{}{node.NodeID, node.Reputation, t.Format(TIMEFORMAT)})
	}

	err = multipleInsert("INSERT INTO tblNodes (NodeID, Active, Reputation, Blocks) VALUES(?, ?, ?, ?) ON DUPLICATE KEY UPDATE Active=?, Reputation=?;", tblNodeRows)
	if err != nil {
		log.Logger.Errorf("DB: Error adding tblNodes: %s\n\r", err)
		return err
	}
	err = multipleInsert("INSERT INTO reputation (NodeID, Reputation, CreateAt)  VALUES(?, ?, ?);", reputationRows)
	if err != nil {
		log.Logger.Errorf("DB: Error adding Reputation : %s\n\r", err)
		return err
	}
	return nil
}

func (d DataBaseType) GetNodes() error {
	// safe query
	rows, err := sqldb.Query("select * from tblNodes order by Reputation desc")
	if err != nil {
		log.Logger.Errorf("Error fetching nodes from DB: %s\n\r", err)
		return nil
	}
	var (
		nodeID     string
		active     bool
		reputation float64
		blocks     int32
	)
	// with := {} is appended, is var is used then var nodes map[string]*NodeType
	nodes := map[string]*NodeType{}
	var nodeRepSort NodeRepSort
	for rows.Next() {
		err := rows.Scan(&nodeID, &active, &reputation, &blocks)

		if err != nil {
			log.Logger.Errorf("Error reading node row  from  DB: %s\n\r", err)
			continue
		}
		n := NodeType{
			NodeID:     nodeID,
			Active:     active,
			Reputation: reputation,
			Blocks:     blocks,
		}
		nodes[nodeID] = &n
		nodeRepSort = append(nodeRepSort, n)
	}
	rows.Close()
	log.Logger.Infof("Adding %v nodes", len(nodes))
	global.Nodes = nodes
	global.ReputationLB = nodeRepSort
	return nil
}

// this function notifies owner's of node if
// this node is added in reputation list
func notifyNodeHasReputation(nodeID string) {
	userIDs := global.NodeUsers[nodeID]

	for userID := range userIDs {
		nodeName := global.Users[int64(userID)].Nodes[nodeID]
		msg := tgbotapi.NewMessage(int64(userID), fmt.Sprintf("`🥂Your node %s[%s] is added in reputation list.`", *nodeName, nodeID))
		msg.ParseMode = "markdown"
		TgBot.Send(msg)
	}
}
