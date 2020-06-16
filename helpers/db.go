package helpers

import (
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/harsh-98/witnetBOT/log"
	"github.com/spf13/viper"
)

var sqldb *sql.DB = nil

type dataBaseInterface interface {
	Init(vip *viper.Viper) error
	Close()
	GetUsers() error
	AddUser(u *UserType) error
	UpdateUser(u *UserType) error
	RemoveUserNode(nodeID string, userID int64) error
	AddUserNode(n UserNode) (NodeType, error)
}

type DataBaseType struct {
	_ dataBaseInterface
}

var DB DataBaseType

func (d DataBaseType) Init(vip *viper.Viper) error {
	var err error = nil
	// https://github.com/go-sql-driver/mysql/blob/v1.5.0/dsn.go#L68
	var config = mysql.NewConfig()

	// https://pkg.go.dev/github.com/go-sql-driver/mysql?tab=doc#Config
	config.User = vip.GetString("user")
	config.Passwd = vip.GetString("passwd")
	config.DBName = vip.GetString("dbName")

	// MultiStatement for handling multiple query batch
	config.MultiStatements = vip.GetBool("multipleStatement")
	config.Net = vip.GetString("net")
	config.Addr = vip.GetString("addr")

	// https://pkg.go.dev/github.com/go-sql-driver/mysql?tab=doc#NewConnector
	connector, err := mysql.NewConnector(config)
	if err != nil {
		log.Logger.Error(err)
	}
	// earlier sql.Open was used
	// https://github.com/golang/go/blob/go1.14.4/src/database/sql/sql.go#L745
	sqldb = sql.OpenDB(connector)
	if err != nil {
		fmt.Println(err)
		return err
	}
	d.GetUsers()
	d.GetNodes()
	return nil
}

func (d DataBaseType) Close() {
	sqldb.Close()
}
