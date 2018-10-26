package engine

import (
	"encoding/json"

	"kelei.com/utils/common"
	"kelei.com/utils/logger"
)

type ItemSystem struct {
	itemCount int
	items     []*Item
	room      *Room
}

type ItemSystemImage struct {
	Items []ItemImage
}

func NewItemSystem(room *Room) *ItemSystem {
	itemSystem := ItemSystem{}
	itemSystem.itemCount = 6
	itemSystem.items = []*Item{}
	itemSystem.room = room
	return &itemSystem
}

func (this *ItemSystem) newItem(itemid int) *Item {
	item := NewItem(itemid)
	return item
}

func (this *ItemSystem) getItemCount() int {
	return this.itemCount
}

func (this *ItemSystem) getItems() []*Item {
	return this.items
}

func (this *ItemSystem) getRoom() *Room {
	return this.room
}

/*
	在任意n个空位生成道具
*/
func (this *ItemSystem) generateItem(n int) {
	println("aaaa:", n)
	gridSystem := this.getRoom().GetGridSystem()
	emptyGrids := gridSystem.getEmptyGridsIndex()
	indexs := common.Perm(len(emptyGrids))
	for i := 0; i < n; i++ {
		gridSystem.getGrids()[emptyGrids[indexs[i]]].newItem()
	}
}

func (this *ItemSystem) fill() {
	this.generateItem(this.getItemCount())
}

func (this *ItemSystem) getImage() *string {
	image := ItemSystemImage{[]ItemImage{}}
	items := this.getItems()
	for _, item := range items {
		itemImage := *item.getImage()
		image.Items = append(image.Items, itemImage)
	}
	b, err := json.Marshal(image)
	logger.CheckFatal(err)
	msg := string(b)
	return &msg
}
