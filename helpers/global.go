package helpers

import "github.com/spf13/viper"

type Global struct {
	// nodes detials
	NodeRepMap map[string]*NodeRepDetails
	// node has reputation
	NodeBlkMap map[string]NodeBlkDetails
	// users registered with the bot
	Users map[int64]*UserType
	// stores the nodeid with their reputation
	ReputationLB NodeRepSort
	// stores the nodeid with block minted by that id
	BlocksLB NodeBlkSort
	// stores the list of users subscripted for a node
	NodeUsers map[string][]int64
	// stores the admin list
	Admin []*UserType
	// highest epoch registered by the bot
	HighestEpoch int
	// genesis reward
	Genesis map[string]*TestnetReward
	// genesis unlock period map
	GenesisUnlock map[int64][]UnlockReward
}

// map is nil by default, not like array which is initialized as empty
// trying to query or update nil map will error
var global = Global{
	Users:        make(map[int64]*UserType),
	NodeRepMap:   make(map[string]*NodeRepDetails),
	NodeBlkMap:   make(map[string]NodeBlkDetails),
	ReputationLB: NodeRepSort{},
	BlocksLB:     NodeBlkSort{},
	Admin:        []*UserType{},
	NodeUsers:    make(map[string][]int64),
	Genesis:      make(map[string]*TestnetReward),
}

var TIMEFORMAT = "2006-01-02 15:04:05"
var RFC822Z     = "02 Jan 2006 15:04"

var Config *viper.Viper
