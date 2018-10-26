package mysql

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"kelei.com/utils/logger"
)

type SqlDSN struct {
	Name     string
	UserName string
	PassWord string
	Addr     string
	DataBase string
}

func AnalysisFlag2SqlDSN(flag *string) *SqlDSN {
	args := strings.Split(*flag, ",")
	//"game", "root", "111111", "127.0.0.1:3306", "richman_game"
	nickName, userName, passWord, addr, dbName := args[0], args[1], args[2], args[3], args[4]
	return &SqlDSN{nickName, userName, passWord, addr, dbName}
}

func (s *SqlDSN) Remote() string {
	//数据源字符串：用户名:密码@协议(地址:端口)/数据库?参数=参数值
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8", s.UserName, s.PassWord, s.Addr, s.DataBase)
}

type DB struct {
	SqlDSN *SqlDSN
	DB     *sql.DB
}

func NewDB(sqlDSN *SqlDSN) *DB {
	db := &DB{}
	db.SqlDSN = sqlDSN
	db.connect()
	return db
}

func (d *DB) connect() {
	db, err := sql.Open("mysql", d.SqlDSN.Remote())
	db.SetMaxOpenConns(2000)
	db.SetMaxIdleConns(100)
	logger.CheckFatal(err, "创建数据库失败")
	logger.CheckFatal(db.Ping(), "连接数据库失败")
	d.DB = db
}
