/*
牌类
*/

package engine

type Card struct {
	//ID(0-53)
	ID int
	//牌值(1-15)
	Value int
	//花色(0:王 1:红桃 2:方块 3:黑桃 4:梅花)
	Suit int
	//优先级（数越大的值约大 1-15）
	Priority int
	//一套牌中的索引
	Index int
}
