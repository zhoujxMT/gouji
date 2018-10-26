package engine

import (
	"encoding/json"

	"kelei.com/utils/common"
	"kelei.com/utils/logger"
)

type GridSystem struct {
	gridCount       int
	grids           []*Grid
	turretPosition  []int //炮塔位置
	turretTheirUser []int //炮塔所属玩家
	room            *Room
}

type GridSystemImage struct {
	Grids []GridImage
}

func NewGridSystem(room *Room) *GridSystem {
	gridSystem := GridSystem{}
	gridSystem.gridCount = 16
	gridSystem.turretPosition = []int{2, 6, 10, 14}
	gridSystem.turretTheirUser = []int{0, 1, 0, 1}
	gridSystem.room = room
	gridSystem.build()
	return &gridSystem
}

func (this *GridSystem) build() {
	for i := 0; i < this.gridCount; i++ {
		gridType := GRID_TYPE_NORMAL
		if common.IndexIntOf(this.turretPosition, i) >= 0 {
			gridType = GRID_TYPE_TURRET
		}
		grid := this.AddGrid()
		grid.SetGridType(gridType)
		grid.setIndex(i)
	}
}

func (this *GridSystem) AddGrid() *Grid {
	grid := NewGrid()
	grid.setRoom(this.getRoom())
	this.grids = append(this.grids, grid)
	return grid
}

func (this *GridSystem) getGridCount() int {
	return this.gridCount
}

func (this *GridSystem) getTurretPosition() []int {
	return this.turretPosition
}

func (this *GridSystem) getGrid(index int) *Grid {
	return this.getGrids()[index]
}

func (this *GridSystem) getGrids() []*Grid {
	return this.grids
}

func (this *GridSystem) getEmptyGridsIndex() []int {
	emptyGrids := []int{}
	grids := this.getRoom().GetGridSystem().getGrids()
	for i, grid := range grids {
		if grid.getGridType() == GRID_TYPE_NORMAL && grid.getUser() == nil && grid.getItem() == nil {
			emptyGrids = append(emptyGrids, i)
		}
	}
	return emptyGrids
}

func (this *GridSystem) getRoom() *Room {
	return this.room
}

func (this *GridSystem) fill() {
	this.fillTheirUser()
	this.fillUser()
}

//网格所属玩家
func (this *GridSystem) fillTheirUser() {
	room := this.getRoom()
	users := room.GetUsers()
	grids := this.getGrids()
	index := 0
	for i, grid := range grids {
		if common.IndexIntOf(this.turretPosition, i) >= 0 {
			userIndex := this.turretTheirUser[index]
			grid.SetTheirUser(users[userIndex])
			index++
		}
	}
}

//网格中的玩家
func (this *GridSystem) fillUser() {
	emptyGrids := this.getEmptyGridsIndex()
	indexs := common.Perm(len(emptyGrids))
	users := this.getRoom().GetUsers()
	for i, user := range users {
		this.getGrids()[emptyGrids[indexs[i]]].setUser(user)
		user.setGrid(this.getGrids()[emptyGrids[indexs[i]]])
	}
}

func (this *GridSystem) getImage() *string {
	msg := this.getImageChecked(this.getGrids())
	return msg
}

func (this *GridSystem) getImageChecked(grids []*Grid) *string {
	image := GridSystemImage{[]GridImage{}}
	for _, grid := range grids {
		gridImage := *grid.getImage()
		image.Grids = append(image.Grids, gridImage)
	}
	b, err := json.Marshal(image)
	logger.CheckFatal(err)
	msg := string(b)
	return &msg
}
