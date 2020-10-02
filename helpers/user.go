package helpers

import (
	"errors"

	"github.com/harsh-98/witnetBOT/log"
)

type UserType struct {
	UserID     int64
	UserName   string
	FirstName  string
	LastName   string
	IsAdmin    bool
	Nodes      map[string]*string
	LastMenuID int
}

// earlier var users := UserType{} wasused
// if := the []UserType{}
// else var Users []UserType
func GetUserByTelegramID(tgID int64) (*UserType, error) {
	for _, u := range global.Users {
		if tgID == u.UserID {
			return u, nil
		}
	}
	return nil, errors.New("User not found")
}

func (d DataBaseType) AddUser(u *UserType) error {
	// Too many connections is query returns rows which must be closed or use Exec
	//https://github.com/go-sql-driver/mysql/issues/111
	// safe query
	_, err := sqldb.Exec("insert into tblUsers values (?, ?, ?, ?, 0)", u.UserID, u.UserName, u.FirstName, u.LastName)
	if err != nil {
		log.Logger.Errorf("Error adding user to DB: %s\n\r", err)
		return err
	}
	global.Users[u.UserID] = u
	// rows, err := sqldb.Query("select * from tblUsers order by UserID desc limit 1")
	// if err != nil {
	// 	log.Error(err)
	// 	return err
	// }
	// nodes := make([], 0)
	// u.Nodes
	// var userID uint32
	// for rows.Next() {
	// 	err := rows.Scan(&userID)
	// 	if err == nil {
	// 		u.UserID = userID
	// 	}
	// }
	// rows.Close()
	return nil
}

func (d DataBaseType) GetUsers() {
	// safe query
	rows, err := sqldb.Query("select * from tblUsers")
	if err != nil {
		log.Logger.Errorf("Error fetching users from DB: %s\n\r", err)
		return
	}
	defer rows.Close()
	var (
		userID      int64
		tgUserName  string
		tgFirstName string
		tgLastName  string
		isAdmin     bool
	)
	for rows.Next() {
		err := rows.Scan(&userID, &tgUserName, &tgFirstName, &tgLastName, &isAdmin)
		// log.Logger.Infof("Adding nodes for userid: %v", userID)
		if err == nil {
			user := UserType{
				UserID:    userID,
				UserName:  tgUserName,
				FirstName: tgFirstName,
				LastName:  tgLastName,
				IsAdmin:   isAdmin,
			}
			user.Nodes = make(map[string]*string)
			// Preventing sql injection
			// dont't use fmt.Sprintf and prefer db.Query or db.Prepare
			// https://www.calhoun.io/what-is-sql-injection-and-how-do-i-avoid-it-in-go/
			rows2, err2 := sqldb.Query("SELECT NodeID, NodeName FROM userNodeMap where userNodeMap.UserID=?;", userID)
			if err2 != nil {
				log.Logger.Errorf("Error fetching nodes for user ID %v from DB: %s\n\r", userID, err2)
				continue
			}
			var (
				nodeID string
			)

			for rows2.Next() {
				// using a global nodeName is not correct as the point will always point to same location for different values
				// hence overwriting with last value
				var nodeName string
				err2 := rows2.Scan(&nodeID, &nodeName)
				log.Logger.Tracef("Add %s for %v: Name %s", nodeID, userID, nodeName)
				if err2 == nil {
					user.Nodes[nodeID] = &nodeName
					global.NodeUsers[nodeID] = append(global.NodeUsers[nodeID], userID)
				}

			}
			defer rows2.Close()
			log.Logger.Tracef("%v", user)
			if isAdmin {
				global.Admin = append(global.Admin, &user)
			}
			global.Users[user.UserID] = &user
		}
	}
	log.Logger.Debugf("Len of users %v", len(global.Users))
}

func (d DataBaseType) UpdateUser(u *UserType) error {
	// safe query
	_, err := sqldb.Exec("update tblUsers set UserName=?, FirstName=?, LastName=? where UserID=?",
		u.UserName, u.FirstName, u.LastName, u.UserID)
	return err
}
