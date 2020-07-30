package helpers

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/harsh-98/witnetBOT/log"
)

func (d *DataBaseType) updateReputationDB(nodes map[string]*NodeRepDetails) error {
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
		tblNodeRows = append(tblNodeRows, []interface{}{node.NodeID, node.Active, node.Reputation, node.Active, node.Reputation})
		reputationRows = append(reputationRows, []interface{}{node.NodeID, node.Reputation, t.Format(TIMEFORMAT)})
	}

	err = multipleInsert("INSERT INTO tblNodes (NodeID, Active, Reputation) VALUES(?, ?, ?) ON DUPLICATE KEY UPDATE Active=?, Reputation=?;", tblNodeRows)
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

func (d DataBaseType) GetNodeRep() error {
	// safe query
	rows, err := sqldb.Query("select NodeID, Active, Reputation from tblNodes order by Reputation desc")
	if err != nil {
		log.Logger.Errorf("Error fetching nodes from DB: %s\n\r", err)
		return nil
	}
	var (
		nodeID     string
		active     bool
		reputation float64
	)
	// when := {} syntax is used, other way var nodes map[string]*NodeRepDetails
	nodeRepMap := map[string]*NodeRepDetails{}
	var nodeRepSort NodeRepSort
	for rows.Next() {
		err := rows.Scan(&nodeID, &active, &reputation)

		if err != nil {
			log.Logger.Errorf("Error reading node row  from  DB: %s\n\r", err)
			continue
		}
		n := NodeRepDetails{
			NodeID:     nodeID,
			Active:     active,
			Reputation: reputation,
		}
		nodeRepMap[nodeID] = &n
		nodeRepSort = append(nodeRepSort, n)
	}
	rows.Close()
	log.Logger.Infof("Adding Rep of %v nodes", len(nodeRepMap))
	global.NodeRepMap = nodeRepMap
	global.ReputationLB = nodeRepSort
	return nil
}

// this function notifies owner's of node if
// this node is added in reputation list
func notifyNodeHasReputation(nodeID string) {
	userIDs := global.NodeUsers[nodeID]
	if userIDs != nil {
		for _, userID := range userIDs {
			nodeName := global.Users[int64(userID)].Nodes[nodeID]
			msg := tgbotapi.NewMessage(int64(userID), fmt.Sprintf("`🥂Your node %s[%s] is added in reputation list.`", *nodeName, nodeID))
			msg.ParseMode = "markdown"
			TgBot.Send(msg)
		}
	}
}
