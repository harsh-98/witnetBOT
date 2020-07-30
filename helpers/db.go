package helpers

import (
	"database/sql"

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
	AddUserNode(n UserNode) (NodeRepDetails, error)
}

type DataBaseType struct {
	_ dataBaseInterface
}

var DB DataBaseType

func (d DataBaseType) Init() error {
	var err error = nil
	// https://github.com/go-sql-driver/mysql/blob/v1.5.0/dsn.go#L68
	var config = mysql.NewConfig()

	// https://pkg.go.dev/github.com/go-sql-driver/mysql?tab=doc#Config
	config.User = Config.GetString("user")
	config.Passwd = Config.GetString("passwd")
	config.DBName = Config.GetString("dbName")

	// MultiStatement for handling multiple query batch
	config.MultiStatements = Config.GetBool("multipleStatement")
	config.Net = Config.GetString("net")
	config.Addr = Config.GetString("addr")

	// https://pkg.go.dev/github.com/go-sql-driver/mysql?tab=doc#NewConnector
	connector, err := mysql.NewConnector(config)
	if err != nil {
		log.Logger.Error(err)
	}
	// earlier sql.Open was used
	// https://github.com/golang/go/blob/go1.14.4/src/database/sql/sql.go#L745
	sqldb = sql.OpenDB(connector)
	if err != nil {
		log.Logger.Error(err)
		return err
	}
	d.GetUsers()
	d.GetNodeRep()
	GetNodeBlk()
	return nil
}

func (d DataBaseType) Close() {
	sqldb.Close()
}

func multipleInsert(prepareQueryStr string, rows [][]interface{}) error {
	// BEGIN: initialise prepare query
	// https://stackoverflow.com/questions/26345318/how-can-i-prevent-sql-injection-attacks-in-go-while-using-database-sql
	tx, err := sqldb.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(prepareQueryStr)
	defer stmt.Close()
	if err != nil {
		tx.Rollback()
		return err
	}
	// END: initialise prepare query

	for _, row := range rows {
		_, err = stmt.Exec(row...)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// BEGIN: sql tx commit
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}
	// END: sql tx commit
	return nil
}
