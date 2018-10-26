package engine

type Item struct {
	itemid int
}

type ItemImage struct {
	ItemID int
}

func NewItem(itemid int) *Item {
	item := Item{}
	item.itemid = itemid
	return &item
}
