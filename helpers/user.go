package helpers

import (
	"errors"
	"fmt"

	"github.com/harsh-98/witnetBOT/log"
)

type UserType struct {
	UserID     int64
	UserName   string
	FirstName  string
	LastName   string
	IsAdmin    bool
	Nodes      []string
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
	str := fmt.Sprintf("insert into tblUsers values (%v, '%s', '%s', '%s', 0)", u.UserID, u.UserName, u.FirstName, u.LastName)
	_, err := sqldb.Exec(str)
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
	rows, err := sqldb.Query("select * from tblUsers")
	if err != nil {
		log.Logger.Errorf("Error fetching users from DB: %s\n\r", err)
		return
	}
	var (
		userID      int64
		tgUserName  string
		tgFirstName string
		tgLastName  string
		isAdmin     bool
	)
	for rows.Next() {
		err := rows.Scan(&userID, &tgUserName, &tgFirstName, &tgLastName, &isAdmin)
		log.Logger.Infof("Adding nodes for userid: %v", userID)
		if err == nil {
			user := UserType{
				UserID:    userID,
				UserName:  tgUserName,
				FirstName: tgFirstName,
				LastName:  tgLastName,
				IsAdmin:   isAdmin,
			}
			user.Nodes = []string{}
			rows2, err2 := sqldb.Query(fmt.Sprintf("SELECT NodeID FROM userNodeMap where userNodeMap.UserID=%v;", userID))
			if err2 != nil {
				log.Logger.Errorf("Error fetching nodes for user ID %v from DB: %s\n\r", userID, err2)
				continue
			}
			var (
				nodeID string
			)

			for rows2.Next() {
				err2 := rows2.Scan(&nodeID)
				log.Logger.Debugf("Add %s for %v", nodeID, userID)
				if err2 == nil {
					user.Nodes = append(user.Nodes, nodeID)
				}
				global.NodeUsers[nodeID] = append(global.NodeUsers[nodeID], userID)
			}
			rows2.Close()
			log.Logger.Debugf("%v", user)
			global.Users[user.UserID] = &user
		}
		log.Logger.Debugf("Len of users %v", len(global.Users))
	}
	rows.Close()
}

func (d DataBaseType) UpdateUser(u *UserType) error {
	str := fmt.Sprintf("update tblUsers set UserName = '%s', FirstName ='%s', LastName = '%s' where UserID = %v",
		u.UserName, u.FirstName, u.LastName, u.UserID)
	_, err := sqldb.Exec(str)
	return err
}
