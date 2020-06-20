package helpers

import "github.com/spf13/viper"

type Global struct {
	Nodes     map[string]*NodeType
	Users     map[int64]*UserType
	Ranking   NodeRepSort
	NodeUsers map[string][]int64
}

// map is nil by default not like array like is initialized as empty
// trying to query or update nil map will error
var global = Global{
	Users:     make(map[int64]*UserType),
	Nodes:     make(map[string]*NodeType),
	Ranking:   NodeRepSort{},
	NodeUsers: make(map[string][]int64),
}

var TIMEFORMAT = "2006-01-02 15:04:05"

var Config *viper.Viper
