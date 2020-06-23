package helpers

import "github.com/spf13/viper"

type Global struct {
	// nodes having reputation
	Nodes map[string]*NodeType
	// users registered with the bot
	Users map[int64]*UserType
	// stores the nodeid with their reputation
	ReputationLB NodeRepSort
	// stores the nodeid with block minted by that id
	BlocksLB NodeBlockSort
	// stores the list of users subscripted for a node
	NodeUsers map[string][]int64
	// stores the admin list
	Admin []*UserType
}

// map is nil by default, not like array which is initialized as empty
// trying to query or update nil map will error
var global = Global{
	Users:        make(map[int64]*UserType),
	Nodes:        make(map[string]*NodeType),
	ReputationLB: NodeRepSort{},
	BlocksLB:     NodeBlockSort{},
	Admin:        []*UserType{},
	NodeUsers:    make(map[string][]int64),
}

var TIMEFORMAT = "2006-01-02 15:04:05"

var Config *viper.Viper
