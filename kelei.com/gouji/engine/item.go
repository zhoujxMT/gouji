/*
道具
*/

package engine

type Item struct {
	baginfoid int
	itemid    int
	count     int
}

func (i *Item) getBagInfoID() int {
	return i.baginfoid
}

func (i *Item) getItemID() int {
	return i.itemid
}

func (i *Item) getCount() int {
	return i.count
}
