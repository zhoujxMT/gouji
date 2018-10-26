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
	index     int
	gridType  int     //格子类型
	theirUser *User   //所属玩家
	item      *Item   //格子中的道具
	user      *User   //格子中的玩家
	turret    *Turret //被哪个炮塔攻击
	room      *Room
}

type GridImage struct {
	I         int
	GridType  int
	TheirUser int //所属玩家
	Item      int //格子中的道具
	User      int //格子中的玩家
}

func NewGrid() *Grid {
	grid := Grid{}
	return &grid
}

/*
	获取下一个玩家index
*/
func (this *Grid) getNextOneGrid(step int) *Grid {
	gridSystem := this.getRoom().GetGridSystem()
	gridCount := gridSystem.getGridCount()
	index := this.getIndex()
	index = index + step
	if index > gridCount-1 {
		index = index - (gridCount - 1)
		index -= 1
	}
	grid := gridSystem.getGrid(index)
	return grid
}

/*
	玩家离开
*/
func (this *Grid) userLeave() {
	this.setUser(nil)
}

/*
	玩家进入
*/
func (this *Grid) userEnter(user *User) {
	this.showImage()
	this.crowdUser(user)
	this.setUser(user)
	this.eatItem(user)
	this.turretAttack()
}

/*
	挤人
*/
func (this *Grid) crowdUser(user *User) {
	currUser := this.getUser()
	if currUser == nil {
		return
	}
	currUser.move(1)
	println("挤人")
}

/*
	吃道具
*/
func (this *Grid) eatItem(user *User) {
	if this.getGridType() != GRID_TYPE_NORMAL {
		return
	}
	item := this.getItem()
	if item == nil {
		return
	}
	println("吃道具")
	user.eatItem(item)
	this.setItem(nil)
	this.getRoom().GetItemSystem().generateItem(1)
}

/*
	生成一个新道具
*/
func (this *Grid) newItem() {
	itemSystem := this.getRoom().GetItemSystem()
	item := itemSystem.newItem(10001)
	this.setItem(item)
}

/*
	炮塔攻击
*/
func (this *Grid) turretAttack() {
	if this.getGridType() != GRID_TYPE_TURRET {
		return
	}
	user := this.getUser()
	theirUser := this.getTheirUser()
	if user == theirUser {
		return
	}
	turret := this.getTurret()
	println("攻击")
	turret.attack()
}
