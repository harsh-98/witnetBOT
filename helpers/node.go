package helpers

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type NodeType struct {
	Reputation float64
	NodeID     string
	Blocks     int32
	Active     bool
}

func (d *DataBaseType) AddNodesInTable(nodes []NodeType) error {
	var query string
	log.Infof(" number of nodes: %v", len(nodes))
	for _, node := range nodes {
		query = fmt.Sprintf("%s INSERT INTO tblNodes (NodeID, Active, Reputation, Blocks) VALUES('%v', %t, %v, %v) ON DUPLICATE KEY UPDATE Active=%t, Reputation=%v;\n",
			query, node.NodeID, node.Active, node.Reputation, node.Blocks, node.Active, node.Reputation)
	}
	log.Debug(query)
	_, err := sqldb.Exec(query)
	if err != nil {
		log.Errorf("Error adding nodes to DB: %s\n\r", err)
		return err
	}
	return nil
}

func (d DataBaseType) GetNodes() error {
	rows, err := sqldb.Query("select * from tblNodes")
	if err != nil {
		log.Errorf("Error fetching nodes from DB: %s\n\r", err)
		return nil
	}
	var (
		nodeID     string
		active     bool
		reputation float64
		blocks     int32
	)
	nodes := []NodeType{}
	for rows.Next() {
		err := rows.Scan(&nodeID, &active, &reputation, &blocks)

		if err != nil {
			log.Errorf("Error reading node row  from  DB: %s\n\r", err)
			continue
		}
		nodes = append(nodes, NodeType{
			NodeID:     nodeID,
			Active:     active,
			Reputation: reputation,
			Blocks:     blocks,
		})
	}
	log.Infof("Adding node with id: %v", len(nodes))
	global.Nodes = nodes
	return nil
}
