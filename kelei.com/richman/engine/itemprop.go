package engine

func (this *Item) getItemID() int {
	return this.itemid
}

func (this *Item) setItemID(itemid int) {
	this.itemid = itemid
}

func (this *Item) getImage() *ItemImage {
	image := ItemImage{}
	image.ItemID = this.getItemID()
	return &image
}
