package engine

var (
	gridCount = 0 //格子数
)

//格子类型
const (
	GRID_TYPE_NORMAL = iota //默认的格子
	GRID_TYPE_TURRET        //炮塔攻击的格子
)

type Grid struct {
	gridType  int   //格子类型
	theirUser *User //所属玩家
	item      *Item //格子中的道具
	user      *User //格子中的玩家
}

func NewGrid(gridType int, theirUser *User) *Grid {
	grid := Grid{}
	grid.gridType = gridType
	grid.theirUser = theirUser
	return &grid
}
