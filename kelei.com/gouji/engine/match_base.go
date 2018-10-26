/*
比赛基类
*/

package engine

const (
	deckCount          = 4
	perCapitaCardCount = 36 //平均每人多少张牌
)

const (
	Priority_BKing = 15
	Priority_SKing = 14
	Priority_Two   = 13
	Priority_Ten   = 8
)

var (
	pcount          = 6  //一桌的人数(正常6)
	cardCount       = 36 //每个人的初始牌数
	chaosPCount     = 4  //乱打的人数(正常4)
	revolutionCount = 0  //革命的人数
)

const (
	MatchingMode_Normal         = iota //匹配
	MatchingMode_EnterAndSetout        //匹配>进入>准备
	MatchingMode_Enter                 //匹配>进入
)

//获取革命的人数
func GetRevolutionCount() int {
	return revolutionCount
}

//设置革命的人数
func SetRevolutionCount(count int) {
	revolutionCount = count
}

//获取平均牌数
func GetCardCount() int {
	return cardCount
}

//设置平均牌数
func SetCardCount(count int) {
	cardCount = count
}

//获取人数
func GetPCount() int {
	return pcount
}

//设置人数
func SetPCount(count int) {
	pcount = count
	if pcount == 4 {
		chaosPCount = 3
	} else if pcount == 6 {
		chaosPCount = 4
	}
	Reset()
}

func InsertCardSlice(slice, insertion []Card, index int) []Card {
	return append(slice[:index], append(insertion, slice[index:]...)...)
}

type Deck struct {
	Cards []Card
}

func (d *Deck) getCard(cardid int) *Card {
	return &d.Cards[cardid]
}

var (
	generateDeck = func() (deck Deck) {
		deck = Deck{make([]Card, 54)}
		deck.Cards[0] = Card{0, 15, 0, 15, 0}
		deck.Cards[1] = Card{1, 14, 0, 14, 0}
		priority := []int{12, 13, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
		index := 2
		for i := 0; i < 4; i++ {
			for j := 0; j < 13; j++ {
				deck.Cards[index] = Card{index, j + 1, i + 1, priority[j], 0}
				index = index + 1
			}
		}
		//描述
		//		arr := []string{"", "黑桃", "红桃", "梅花", "方块"}
		//		for _, card := range deck.Cards {
		//			strSuit := arr[card.Suit]
		//			fmt.Printf("%s%d(%d-%d-%d-%d)\n", strSuit, card.Value, card.ID, card.Value, card.Suit, card.Priority)
		//		}
		return deck
	}
	//整付牌
	deck             = generateDeck()
	generateCardPool = func(deckCount int) (cardPool []Card) {
		for i := 0; i < deckCount; i++ {
			cardPool = InsertCardSlice(cardPool, deck.Cards, len(cardPool))
		}
		return cardPool
	}
	cardPool     = generateCardPool(deckCount)
	cardPoolSize = len(cardPool)
)

type CardList []Card

func (list CardList) Len() int {
	return len(list)
}

func (list CardList) Less(i, j int) bool {
	iPriority := list[i].Priority*10 - list[i].Suit
	jPriority := list[j].Priority*10 - list[j].Suit
	if iPriority > jPriority {
		return true
	} else {
		return false
	}
}

func (list CardList) Swap(i, j int) {
	var temp Card = list[i]
	list[i] = list[j]
	list[j] = temp
}
