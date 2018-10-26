/*
测试类
*/

package engine

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func Test_main(t *testing.T) {
	convey.Convey("main,启动", t, func() {
	})
}

func Test_excel(t *testing.T) {
	convey.Convey("excel,检测牌是否能压过", t, func() {
		//	//ID(0-53)
		//	ID int
		//	//牌值(1-15)
		//	Value int
		//	//花色(0:王 1:红桃 2:方块 3:黑桃 4:梅花)
		//	Suit int
		//	//优先级（数越大的值约大 1-15）
		//	Priority int
		//	//一套牌中的索引
		//	Index int
		a := []Card{Card{0, 15, 0, 15, 0}, Card{2, 1, 3, 12, 0}}
		b := []Card{Card{2, 1, 3, 12, 0}, Card{2, 1, 3, 12, 0}}
		convey.So(excel(a, b), convey.ShouldEqual, false)
	})
}
