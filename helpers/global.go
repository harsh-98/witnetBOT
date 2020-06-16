package helpers

type Global struct {
	Nodes   map[string]*NodeType
	Users   map[int64]*UserType
	Ranking NodeRepSort
}

// map is nil by default not like array like is initialized as empty
// trying to query or update nil map will error
var global = Global{
	Users:   make(map[int64]*UserType),
	Nodes:   make(map[string]*NodeType),
	Ranking: NodeRepSort{},
}
