/*
比赛redis初始化
*/

package engine

import (
	"database/sql"

	"kelei.com/utils/frame"
)

var (
	db   *sql.DB
	nmDB *sql.DB
)

func db_Init() {
	db = frame.GetDB("game")
	nmDB = frame.GetDB("member")
}
