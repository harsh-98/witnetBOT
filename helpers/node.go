package helpers

import (
	"fmt"
	"time"

	"github.com/harsh-98/witnetBOT/log"
)

type NodeType struct {
	Reputation float64
	NodeID     string
	Blocks     int32
	Active     bool
}

func (d *DataBaseType) AddNodesInTable(nodes map[string]*NodeType) error {
	var query string
	t := time.Now()
	log.Logger.Infof(" number of nodes: %v", len(nodes))
	for _, node := range nodes {
		query = fmt.Sprintf(`%s 
		INSERT INTO tblNodes (NodeID, Active, Reputation, Blocks) VALUES('%v', %t, %v, %v) ON DUPLICATE KEY UPDATE Active=%t, Reputation=%v;
		INSERT INTO reputation (NodeID, Reputation, CreateAt)  VALUES('%v', %v, %s);`,
			query, node.NodeID, node.Active, node.Reputation, node.Blocks, node.Active, node.Reputation, node.NodeID, node.Reputation, t.Format(TIMEFORMAT))
	}
	// log.Logger.Debug(query)
	_, err := sqldb.Exec(query)
	if err != nil {
		log.Logger.Errorf("Error adding nodes to DB: %s\n\r", err)
		return err
	}
	return nil
}

func (d DataBaseType) GetNodes() error {
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
	log.Logger.Infof("Adding %v nodes", len(nodes))
	global.Nodes = nodes
	global.Ranking = nodeRepSort
	return nil
}
