package frame

import (
	"database/sql"

	"kelei.com/utils/logger"
	"kelei.com/utils/mysql"
)

var (
	dbs = make(map[string]*mysql.DB)
)

type Sql struct {
	SqlDSNs []*mysql.SqlDSN
}

func loadSql() {
	if args.Sql == nil {
		return
	}
	sqlDSNs := args.Sql.SqlDSNs
	for _, sqlDSN := range sqlDSNs {
		logger.Infof("[连接Sql %s]", sqlDSN.Addr)
		dbs[sqlDSN.Name] = mysql.NewDB(sqlDSN)
	}
}

//获得DB
func GetDB(args ...string) *sql.DB {
	var db *mysql.DB
	if len(args) > 0 {
		dbName := args[0]
		for _, db_ := range dbs {
			if db_.SqlDSN.Name == dbName {
				db = db_
				break
			}
		}
	} else {
		for _, db_ := range dbs {
			db = db_
			break
		}
	}
	return db.DB
}
