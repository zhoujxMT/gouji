/*
玩家-操作-提示
*/

package engine

import (
	"bytes"
	"fmt"
	"strconv"
)

/*
Hint(提示)
result:
	有能压过的牌:index|index|index
	没有:-2
*/
func Hint(args []string) *string {
	fmt.Print("")
	res := "-2" //压不过
	userid := args[0]
	user := UserManage.GetUser(&userid)
	room := user.getRoom()
	current_cards := room.getCurrentCards()
	playCards := []Card{}
	//不是压牌
	if len(current_cards) == 0 {
		playCards = user.getLeastCards()
	} else {
		//是压牌
		playCards = user.getPressIndexs()
	}
	if len(playCards) > 0 {
		bf := bytes.Buffer{}
		for _, card := range playCards {
			bf.WriteString(strconv.Itoa(card.Index) + "|")
		}
		str := bf.String()
		res = str[:len(str)-1]
	}
	return &res
}

//获取最小牌
func (u *User) getLeastCards() []Card {
	playCards := []Card{}
	cards := u.getCards()
	for i := len(cards) - 1; i >= 0; i-- {
		card := cards[i]
		if len(playCards) > 0 && card.Priority != playCards[0].Priority {
			break
		}
		playCards = append(playCards, card)
	}
	if len(playCards) <= 0 {
		return playCards
	}
	//最小的牌是2,所有的牌一起出
	if playCards[0].Priority >= Priority_Two {
		playCards = cards
	} else {
		kingCount, noHangCount := getKingAndNoHangCount(cards)
		//挂王走
		if alwaysHangKingSucceed(kingCount, noHangCount) {
			kingCards := getKings(cards)
			//如果不能挂的牌的数量<=1,挂上所有的王,挂上所有的2
			if noHangCount <= 1 {
				twoCards := getCardsByPriority(cards, Priority_Two)
				playCards = append(playCards[:0], append(kingCards, playCards[0:]...)...)
				playCards = append(playCards[:0], append(twoCards, playCards[0:]...)...)
			} else {
				//不是最后一套,挂一个王
				playCards = append(playCards, kingCards[0])
			}
		} else {
			//出最小牌
		}
	}
	return playCards
}

//获取大小王集合(倒序)
func getKings(cards []Card) []Card {
	myBKingCards := getCardsByPriority(cards, Priority_BKing)
	mySKingCards := getCardsByPriority(cards, Priority_SKing)
	kingCards := []Card{}
	//将小王倒着放入kingCards
	for i := len(mySKingCards) - 1; i >= 0; i-- {
		kingCards = append(kingCards, mySKingCards[i])
	}
	//将大王倒着放入kingCards
	for i := len(myBKingCards) - 1; i >= 0; i-- {
		kingCards = append(kingCards, myBKingCards[i])
	}
	return kingCards
}

//获取压牌的信息
func (u *User) getPressIndexs() []Card {
	lintCards := []Card{}
	room := u.getRoom()
	currentCards := room.getCurrentCards()
	//有大王
	if isExistByPriority(currentCards, Priority_BKing) {
		return lintCards
	}
	userCards := u.getCards()
	cards := make([]Card, len(userCards))
	copy(cards, userCards)
	SKingCards := getCardsByPriority(currentCards, Priority_SKing)
	//有小王
	if len(SKingCards) > 0 {
		//我的大王
		myBKingCards := getCardsByPriority(cards, Priority_BKing)
		//我的大王数量没有当前牌的小王数量多
		if len(myBKingCards) < len(SKingCards) {
			return lintCards
		}
		//从我的牌列表中取出
		for i := len(SKingCards) - 1; i >= 0; i-- {
			index := myBKingCards[i].Index
			cards = append(cards[:index], cards[index+1:]...)
		}
		//放入提示的牌中
		lintCards = myBKingCards[:len(SKingCards)]
	}
	//获取基牌
	baseCard := currentCards[len(currentCards)-1]
	//基牌是王
	if baseCard.Priority > Priority_Two {
		return lintCards
	}
	//基牌是2
	if baseCard.Priority == Priority_Two {
		twoCards := getCardsByPriority(currentCards, Priority_Two)
		kingCards := getKings(cards)
		//我的王的数量>=2的数量
		if len(kingCards) >= len(twoCards) {
			for i, card := range kingCards {
				if i > len(twoCards)-1 {
					break
				}
				lintCards = append(lintCards, card)
			}
		} else {
			lintCards = []Card{}
		}
	} else {
		//基牌<2
		lintCards = baseCardLessTwo(currentCards, cards, lintCards)
	}
	return lintCards
}

//基牌小于2
func baseCardLessTwo(currentCards []Card, cards []Card, lintCards []Card) []Card {
	//获取基牌
	baseCard := currentCards[len(currentCards)-1]
	baseCardCount := 0
	for _, card := range currentCards {
		if card.Priority <= Priority_Two {
			baseCardCount += 1
		}
	}
	cardsInfo := getCardsInfo(baseCard, baseCardCount, cards)
	//设置提示的牌
	setLintCards := func(cards []Card) {
		for _, card := range cards {
			lintCards = append(lintCards, card)
		}
	}
	/*
		非数量相等的牌
		result:找到了牌true
	*/
	notEqual := func() bool {
		lessCards := cardsInfo["less"]
		//有数量少的
		if lessCards != nil {
			gteTwoCards := cardsInfo["gteTwo"]
			needBandCount := baseCardCount - len(lessCards)
			kingCount := 0
			for i := 0; i < needBandCount; i++ {
				card := gteTwoCards[i]
				if card.Priority > Priority_Two {
					kingCount += 1
				}
			}
			bandCard := func() {
				setLintCards(gteTwoCards[0:needBandCount])
				setLintCards(lessCards)
			}
			//挂王的数量<=1
			if kingCount <= 1 {
				//提示
				bandCard()
			} else {
				greaterCards := cardsInfo["greater"]
				//有数量多的
				if greaterCards != nil {
					//拆牌
					setLintCards(greaterCards[:baseCardCount])
				} else {
					//挂王的数量>1
					bandCard()
				}
			}
		} else {
			greaterCards := cardsInfo["greater"]
			//有数量多的
			if greaterCards != nil {
				//拆牌
				setLintCards(greaterCards[:baseCardCount])
			} else {
				gteTwoCards := cardsInfo["gteTwo"]
				//有>=2的牌列表
				if gteTwoCards != nil {
					if len(gteTwoCards) >= baseCardCount {
						setLintCards(gteTwoCards[:baseCardCount])
					} else {
						//没有合适的牌
						return false
					}
				} else {
					//没有合适的牌
					return false
				}
			}
		}
		return true
	}

	equalCards := cardsInfo["equal"]
	//有数量相等的
	if equalCards != nil {
		//如果是>=2的牌,查找挂拆有没有合适的牌
		if equalCards[0].Priority >= Priority_Two && len(equalCards) > 1 {
			//挂拆没有合适的牌,还是用>=2的牌
			if notEqual() == false {
				setLintCards(equalCards)
			}
		} else {
			setLintCards(equalCards)
		}
	} else {
		if notEqual() == false {
			lintCards = []Card{}
		}
	}
	return lintCards
}

//统计b中大于基牌的牌的信息
func getCardsInfo(baseCard Card, baseCardCount int, cards []Card) map[string][]Card {
	cardsInfo := map[string][]Card{} //equal,less,greater,gteTwo
	cardsList := make(map[int][]Card)
	prioritys := []int{}
	for _, card := range cards {
		if card.Priority <= baseCard.Priority {
			continue
		}
		if cardsList[card.Priority] == nil {
			cardsList[card.Priority] = []Card{}
			prioritys = append(prioritys, card.Priority)
		}
		cardsList[card.Priority] = append(cardsList[card.Priority], card)
	}
	for i := len(prioritys) - 1; i >= 0; i-- {
		priority := prioritys[i]
		cards := cardsList[priority]
		if len(cards) == baseCardCount {
			cardsInfo["equal"] = cards
			if cards[0].Priority < Priority_Two {
				break
			}
		}
		if cards[0].Priority < Priority_Two {
			if len(cards) < baseCardCount {
				if cardsInfo["less"] == nil {
					cardsInfo["less"] = cards
				}
				if len(cardsInfo["less"]) < len(cards) {
					cardsInfo["less"] = cards
				}
			}
			if len(cards) > baseCardCount {
				if cardsInfo["greater"] == nil {
					cardsInfo["greater"] = cards
				}
			}
		}
		if cards[0].Priority >= Priority_Two {
			if cardsInfo["gteTwo"] == nil {
				cardsInfo["gteTwo"] = cards
			} else {
				for _, card := range cards {
					cardsInfo["gteTwo"] = append(cardsInfo["gteTwo"], card)
				}
			}
		}
	}
	//少于基牌数量的牌，挂上>=2的牌还打不了，就删除
	if cardsInfo["less"] != nil {
		if len(cardsInfo["less"])+len(cardsInfo["gteTwo"]) < baseCardCount {
			delete(cardsInfo, "less")
		}
	}
	//	for k, v := range cardsInfo {
	//		fmt.Println(k, ":", len(v))
	//	}
	return cardsInfo
}
