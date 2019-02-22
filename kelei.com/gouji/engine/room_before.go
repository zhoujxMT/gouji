/*
房间开赛前
*/

package engine

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"

	. "kelei.com/utils/common"
	"kelei.com/utils/frame"
	"kelei.com/utils/logger"
)

var (
	goodCard_KingCount = 5
	goodCard_TwoCount  = 8
)

//获取等待革命的玩家列表
func (r *Room) getTributeUsers() {

}

/*
推送房间的状态信息
push:Matching_Push,64$1$0$5$0|||||#等待席#当前轮次
*/
func (r *Room) matchingPush(user *User) {
	//获取此房间的信息
	userids, statuss := r.getRoomMatchingInfo()
	if user != nil {
		userids = []string{*user.getUserID()}
	}
	pushMessageToUsers("Matching_Push", statuss, userids)
	r.pushJudgment("Matching_Push", statuss[0])
}

/*
推送海选赛匹配信息
push:MatchingHXS_Push,64$1$0$5$0|||||#等待席#当前轮次
*/
func (r *Room) matchingHXSPush(user *User) {
	//获取此房间的信息
	userids, statuss := r.getRoomMatchingInfo()
	if user != nil {
		userids = []string{*user.getUserID()}
	}
	pushMessageToUsers("MatchingHXS_Push", statuss, userids)
}

/*
Go_Push(推送房间的走科信息)
push:userid$ranking|userid$ranking
*/
func (r *Room) goPush(user *User) {
	buff := bytes.Buffer{}
	users := r.getRanking()
	for _, u := range users {
		if u != nil {
			buff.WriteString(fmt.Sprintf("%s$%d|", *u.getUserID(), u.getRanking()))
		}
	}
	str := *RemoveLastChar(buff)
	//没有传入玩家，就给所有玩家推送
	if user == nil {
		pushMessageToUsers("Go_Push", []string{str}, r.getUserIDs())
		r.pushJudgment("Go_Push", str)
	} else {
		pushMessageToUsers("Go_Push", []string{str}, []string{*user.getUserID()})
	}
}

/*
所有人开启准备倒计时
*/
func (r *Room) setOutCountDown() {
	matchid := r.GetMatchID()
	if matchid == Match_HYTW || matchid == Match_HXS {
		return
	}
	users := r.getUsers()
	for _, user := range users {
		user.countDown_setOut(time.Second * 20)
	}
}

//初始化房间中的玩家信息
func (r *Room) initUsersInfo() {
	for i, user := range r.users {
		user.setStatus(UserStatus_NoPass)
		user.setIndex(i)
		user.close_countDown_playCard()
		if i%2 == 0 {
			user.setTeamMark(TeamMark_Red)
		} else {
			user.setTeamMark(TeamMark_Blue)
		}
	}
}

//开局
func (r *Room) opening() {
	//是录制版
	if r.getGameRule() == GameRule_Record {
		return
	}
	//海选赛延迟开赛
	if r.GetMatchID() == Match_HXS {
		go func() {
			defer func() {
				if p := recover(); p != nil {
					errInfo := fmt.Sprintf("opening : { %v }", p)
					logger.Errorf(errInfo)
				}
			}()
			time.Sleep(time.Second * 4)
			//比赛开局
			r.match_Opening()
		}()
	} else {
		//比赛开局
		r.match_Opening()
	}
}

//发牌
func (r *Room) deal() {
	r.match_Opening()
	go func() {
		defer func() {
			if p := recover(); p != nil {
				errInfo := fmt.Sprintf("deal : { %v }", p)
				logger.Errorf(errInfo)
			}
		}()
		pushMessageToUsers("Deal_Push", []string{"1"}, r.getUserIDs())
	}()
}

//是否人齐
func (r *Room) userEnough() bool {
	return r.getUserCount() >= pcount
}

//开局处理
func (r *Room) opening_handle() {
	//进贡处理
	r.jinTributeHandle()
	//革命处理
	r.revolutionHandle()
	//进完贡后,房间重置
	r.reset()
	//检测开赛
	r.check_start()
}

//比赛开局
func (r *Room) match_Opening() {
	if !r.userEnough() {
		return
	}
	//将玩家比赛的信息同步到数据库
	r.insertUsersInfo()
	//开赛更新玩家数据
	r.updateUserInfo()
	//初始化房间中的玩家信息
	r.initUsersInfo()
	//推送开赛牌局
	r.Opening_Push()
	//设置房间为发牌状态
	r.SetRoomStatus(RoomStatus_Deal)
	go func() {
		defer func() {
			if p := recover(); p != nil {
				errInfo := fmt.Sprintf("match_Opening : { %v }", p)
				logger.Errorf(errInfo)
			}
		}()
		//等待发牌
		if r.GetMatchID() == Match_GJYX {
			time.Sleep(time.Second * 10)
		} else if r.GetMatchID() == Match_HYTW {
			time.Sleep(time.Second * 12)
		} else {
			time.Sleep(time.Second * 10)
		}
		//不是录制版
		if r.getGameRule() != GameRule_Record {
			r.opening_handle()
		}
	}()
}

//将玩家比赛的信息同步到数据库
func (r *Room) insertUsersInfo() {
	users := r.getUsers()
	for _, user := range users {
		user.insertUserInfo()
	}
}

//推送开赛牌局
func (r *Room) Opening_Push() {
	//生成所有人的牌
	cardsList := []CardList{}
	disorderCardsList := []CardList{}
	//	count := 0
	//	CostTime(func() {
	//		for {
	//			count++
	//			cardsList, _ = r.generateCards()
	//			if len(cardsList) > 0 {
	//				break
	//			}
	//		}
	//	}, 100, "发牌")
	//	fmt.Println("100场发牌总次数：", count)
	//	return
	for {
		cardsList, disorderCardsList = r.generateCards()
		if len(cardsList) > 0 {
			break
		}
	}
	users := r.getUsers()
	userids := []string{}
	messages := []string{}
	for i, user := range users {
		//设置玩家牌面
		user.setCards(cardsList[i])
		userids = append(userids, *user.userid)
		buffer := bytes.Buffer{}
		for _, card := range disorderCardsList[i] {
			buffer.WriteString(strconv.Itoa(card.ID))
			buffer.WriteString("|")
		}
		str := buffer.String()
		str = str[0 : len(str)-1]
		messages = append(messages, str)
	}
	//给所有人推送牌局
	pushMessageToUsers("Opening_Push", messages, userids)
	/*
		Opening_Judgment_Push(给裁判推送所有人的牌局)
		push:0|0|3...$
	*/
	r.pushJudgment("Opening_Judgment_Push", strings.Join(messages, "$"))
}

//检测开赛
func (r *Room) check_start() {
	//是否可以开赛
	canStart := true
	//进贡模式
	if r.GetTribute() {
		//检测是否有没进还贡的
		users := r.getUsers()
		for _, user := range users {
			if user.getTributeStatus() != TributeStatus_Done {
				canStart = false
				break
			}
		}
	}
	//革命模式
	if r.GetRevolution() {
		revolutionCount := 0
		//检测是否有“没革命的”
		users := r.getUsers()
		for _, user := range users {
			if user.getStatus() == UserStatus_WaitRevolution {
				canStart = false
				break
			} else if user.isRevolution() {
				revolutionCount += 1
			}
		}
		liuju := func() {
			canStart = false
			r.SetRoomStatus(RoomStatus_Liuju)
			logger.Debugf("流局")
			users[0].setCtlUsers("", "", SetController_Liuju)
			r.setMatchingStatus(MatchingStatus_Over)
			time.Sleep(time.Second * 2)
			r.SetRoomStatus(RoomStatus_Setout)
		}
		//革命人数超过两人,流局
		isLiuju := revolutionCount >= 2 && r.GetRoomStatus() == RoomStatus_Revolution
		if isLiuju {
			liuju()
			if r.getGameRule() != GameRule_Record {
				//重开牌局
				r.match_Opening()
			}
		}
		if canStart {
			r.SetRoomStatus(RoomStatus_Setout)
		}
	}
	if canStart && r.GetRoomStatus() == RoomStatus_Setout {
		//不是录制版
		if r.getGameRule() != GameRule_Record {
			//开赛
			r.start()
		}
	}
}

//开始
func (r *Room) begin() {
	if frame.GetMode() == frame.MODE_RELEASE {
		defer func() {
			if p := recover(); p != nil {
				logger.Errorf("[recovery] begin : %v", p)
			}
		}()
	}
	//没发牌，不能开牌
	if r.GetRoomStatus() != RoomStatus_Deal {
		return
	}
	fmt.Println("xxxxx")
	r.opening_handle()
	r.SetRoomStatus(RoomStatus_Revolution)
	r.setMatchingStatus(MatchingStatus_Run)
	pushMessageToUsers("Begin_Push", []string{"1"}, r.getUserIDs())
	go func() {
		defer func() {
			if p := recover(); p != nil {
				errInfo := fmt.Sprintf("begin : { %v }", p)
				logger.Errorf(errInfo)
			}
		}()
		i := 0
		for {
			time.Sleep(time.Second)
			if r.getMatchingStatus() == MatchingStatus_Run {
				i++
				if i >= 10 {
					break
				}
			} else if r.getMatchingStatus() == MatchingStatus_Over {
				return
			}
		}
		r.start()
	}()
}

//开赛
func (r *Room) start() {
	if !r.userEnough() {
		return
	}
	if r.GetRoomStatus() == RoomStatus_Deal {
		return
	}
	//重置所有玩家的排名
	r.resetUsersRanking()
	//设置比赛开始
	r.SetRoomStatus(RoomStatus_Match)
	//设置第一个出牌人
	r.setFirstController()
	//设置成-1之后说明比赛已经开始,为了前台牌面布置
	r.gtwUserThreeInfo = "-1"
}

//检测有一个人网络掉线了,比赛暂停
func (r *Room) checkMatchPause(user *User) {
	//裁判端
	if r.getGameRule() == GameRule_Record {
		r.pause()
		/*
			push:[Online_Push] userid|status
			des:status=0离线 status=1回来
		*/
		user.setOnline(false)
		r.pushJudgment("Online_Push", fmt.Sprintf("%s|%d", *user.getUserID(), 0))
	}
}

//暂停
func (r *Room) pause() {
	if !r.isMatching() {
		return
	}
	pushMessageToUsers("Pause_Push", []string{"1"}, r.getUserIDs())
	r.pushJudgment("Pause_Push", "1")
	r.pause_revolution()
	r.setMatchingStatus(MatchingStatus_Pause)
	logger.Debugf("暂停")
	for _, user := range r.getUsers() {
		if user != nil {
			user.pause()
		}
	}
}

//暂停革命
func (r *Room) pause_revolution() {
	if r.GetRoomStatus() == RoomStatus_Revolution {
		for _, user := range r.getUsers() {
			if user.getStatus() == UserStatus_WaitRevolution {
				user.pause()
			}
		}
	}
}

//恢复
func (r *Room) resume() {
	pushMessageToUsers("Resume_Push", []string{"1"}, r.getUserIDs())
	r.pushJudgment("Resume_Push", "1")
	r.resume_revolution()
	r.setMatchingStatus(MatchingStatus_Run)
	logger.Debugf("恢复")
	for _, user := range r.getUsers() {
		if user != nil {
			user.resume()
		}
	}
}

//恢复革命
func (r *Room) resume_revolution() {
	if r.GetRoomStatus() == RoomStatus_Revolution {
		for _, user := range r.getUsers() {
			if user.getStatus() == UserStatus_WaitRevolution {
				user.resume()
			}
		}
	}
}

//解散牌局
func (r *Room) dissolve() {
	r.reset()
	users := r.getUsers()
	for _, user := range users {
		user.reset()
		user.setStatus(UserStatus_Setout)
		user.push("Dissolve_Push", &Res_Succeed)
	}
	r.deleteUsersInfo()
	r.SetRoomStatus(RoomStatus_Setout)
	r.setMatchingStatus(MatchingStatus_Over)
	logger.Debugf("解散牌局")
}

//重置玩家排名
func (r *Room) resetUsersRanking() {
	for _, user := range r.getUsers() {
		user.ranking = -1
	}
}

//获取王的数量
func (r *Room) getCountWithKing(cards CardList) int {
	count := 0
	for _, card := range cards {
		if card.Priority == Priority_SKing || card.Priority == Priority_BKing {
			count++
		} else {
			break
		}
	}
	return count
}

//获取2的数量
func (r *Room) getCountWithTwo(cards CardList) int {
	count := 0
	for _, card := range cards {
		if card.Priority >= Priority_Two {
			if card.Priority == Priority_Two {
				count++
			}
		} else {
			break
		}
	}
	return count
}

//获取大于10的牌的数量
func (r *Room) getCountWithGreaterThanTen(cards CardList) int {
	count := 0
	for _, card := range cards {
		if card.Priority > Priority_Ten {
			count++
		} else {
			break
		}
	}
	return count
}

//生成所有人的牌
func (r *Room) generateCards_back() ([]CardList, []CardList) {
	rr := rand.New(rand.NewSource(time.Now().UnixNano()))
	indexs := rr.Perm(cardPoolSize)
	cardLists := make([]CardList, pcount)
	disorderCardLists := make([]CardList, pcount)
	cardids := make([]int, 6)
	gtws := make([][]int, 6)
	for i, _ := range gtws {
		gtws[i] = make([]int, 5)
	}
	mapGtwIndexs := make(map[int]int)
	var index = 0
	for i := 0; i < pcount; i++ {
		cardLists[i] = CardList(make(CardList, perCapitaCardCount))
		disorderCardLists[i] = CardList(make(CardList, perCapitaCardCount))
		for j := 0; j < perCapitaCardCount; j++ {
			card := cardPool[indexs[index]]
			//上班3记录
			if card.Value == 3 {
				gtws[i][card.Suit] += 1
				if mapGtwIndexs[i] == 0 && gtws[i][card.Suit] >= 2 {
					mapGtwIndexs[i] = j
					cardids[i] = card.ID
				}
			}
			cardLists[i][j] = card
			disorderCardLists[i][j] = card
			index = index + 1
		}
		sort.Sort(cardLists[i])
		if r.getCountWithGreaterThanTen(cardLists[i]) < 11 {
			return []CardList{}, []CardList{}
		}
		for k := 0; k < len(cardLists[i]); k++ {
			cardLists[i][k].Index = k
			disorderCardLists[i][k].Index = k
		}
	}
	if r.isGtwMode() { //是上班模式
		if !r.haveGtw(mapGtwIndexs, cardids) {
			return []CardList{}, []CardList{}
		}
	} else { //不是上班模式
		r.randomFirstController()
	}
	cardLists = r.handleGenerateCards(cardLists)
	disorderCardLists = r.handleGenerateCards(disorderCardLists)
	return cardLists, disorderCardLists
}

// Clone 完整复制数据
func (r *Room) clone(a, b interface{}) error {
	buff := new(bytes.Buffer)
	enc := gob.NewEncoder(buff)
	dec := gob.NewDecoder(buff)
	if err := enc.Encode(a); err != nil {
		return err
	}
	if err := dec.Decode(b); err != nil {
		return err
	}
	return nil
}

//生成所有人的牌
func (r *Room) generateCards() ([]CardList, []CardList) {
	rr := rand.New(rand.NewSource(time.Now().UnixNano()))
	indexs := rr.Perm(cardPoolSize)
	cardLists := make([]CardList, pcount)
	disorderCardLists := make([]CardList, pcount)
	cardids := make([]int, 6)
	gtws := make([][]int, 6)
	for i, _ := range gtws {
		gtws[i] = make([]int, 5)
	}
	mapGtwIndexs := make(map[int]int)
	var index = 0
	for i := 0; i < pcount; i++ {
		cardLists[i] = CardList(make(CardList, perCapitaCardCount))
		disorderCardLists[i] = CardList(make(CardList, perCapitaCardCount))
	}
	for j := 0; j < perCapitaCardCount; j++ {
		for i := 0; i < pcount; i++ {
			card := cardPool[indexs[index]]
			//上班3记录
			if card.Value == 3 {
				gtws[i][card.Suit] += 1
				if len(mapGtwIndexs) == 0 && mapGtwIndexs[i] == 0 && gtws[i][card.Suit] >= 2 {
					mapGtwIndexs[i] = j
					cardids[i] = card.ID
				}
			}
			disorderCardLists[i][j] = card
			index = index + 1
		}
	}
	//初始牌不完整
	if !r.initCardCountIsIntegrity() {
		for i := 0; i < pcount; i++ {
			sort.Sort(disorderCardLists[i])
		}
	}
	disorderCardLists = r.handleGenerateCards(disorderCardLists)
	for i := 0; i < pcount; i++ {
		disorderCardList := disorderCardLists[i]
		for j, card := range disorderCardList {
			cardLists[i][j] = card
		}
	}
	kingCount := 0
	twoCount := 0
	for i := 0; i < pcount; i++ {
		sort.Sort(cardLists[i])
		//玩家初始牌是完整的情况下，判断大于11的牌的数量
		if r.initCardCountIsIntegrity() {
			if r.getCountWithGreaterThanTen(cardLists[i]) < 11 {
				return []CardList{}, []CardList{}
			}
			if r.getDealMode() == DEAL_RED {
				if r.getUserByIndex(i).getTeamMark() == TeamMark_Red {
					kingCount += r.getCountWithKing(cardLists[i])
					twoCount += r.getCountWithTwo(cardLists[i])
				}
			} else if r.getDealMode() == DEAL_BLUE {
				if r.getUserByIndex(i).getTeamMark() == TeamMark_Blue {
					kingCount += r.getCountWithKing(cardLists[i])
					twoCount += r.getCountWithTwo(cardLists[i])
				}
			}
		}
		for k := 0; k < len(cardLists[i]); k++ {
			cardLists[i][k].Index = k
		}
	}
	if r.getDealMode() != DEAL_NORMAL {
		if kingCount < goodCard_KingCount || twoCount < goodCard_TwoCount {
			return []CardList{}, []CardList{}
		}
	}
	if r.isGtwMode() { //是上班模式
		if !r.haveGtw(mapGtwIndexs, cardids) {
			return []CardList{}, []CardList{}
		}
	} else { //不是上班模式
		r.randomFirstController()
	}
	return cardLists, disorderCardLists
}

//随机获取第一个出牌人
func (r *Room) randomFirstController() {
	activeUsers := []*User{}
	for _, user := range r.getUsers() {
		if user.isActive() {
			activeUsers = append(activeUsers, user)
		}
	}
	rd := Random(0, len(activeUsers))
	r.firstController = activeUsers[rd]
}

//是否是上班模式
func (r *Room) isGtwMode() bool {
	if cardCount < 36 {
		return false
	}
	return true
}

//是否有上班的
func (r *Room) haveGtw(mapGtwIndexs map[int]int, cardids []int) bool {
	//没有上班的人，处理一下
	if len(mapGtwIndexs) == 0 {
		fmt.Println("没有上班的人，处理一下")
		return false
	}
	userIndex := 0
	secondThreeIndex := 0
	for userIndex_, secondThreeIndex_ := range mapGtwIndexs {
		userIndex = userIndex_
		secondThreeIndex = secondThreeIndex_
		break
	}
	user := r.getUsers()[userIndex]
	//userid,cardid,第二张3出现的位置
	r.gtwUserThreeInfo = fmt.Sprintf("%s,%d,%d", *user.getUserID(), cardids[userIndex], secondThreeIndex)
	r.firstController = user
	logger.Debugf("上班的人是：%s", *user.getUserID())
	return true
}

//获取上班玩家的3信息
func (r *Room) getGTW() string {
	return r.gtwUserThreeInfo
}

//处理生成的牌
func (r *Room) handleGenerateCards(cardLists []CardList) []CardList {
	userStrs := make([]string, pcount)
	for i := 0; i < revolutionCount; i++ {
		//		userStrs[i] = "0-15-0-15|0-15-0-15|1-14-0-14|1-14-0-14|3-2-1-13|3-2-1-13|3-2-1-13|2-1-1-12|2-1-1-12|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|13-12-1-10|13-12-1-10|11-10-1-8|11-10-1-8|11-10-1-8|9-8-1-6|5-4-1-2|5-4-1-2|5-4-1-2|5-4-1-2|5-4-1-2|5-4-1-2|5-4-1-2|5-4-1-2|5-4-1-2|5-4-1-2|5-4-1-2|5-4-1-2|5-4-1-2|5-4-1-2|5-4-1-2"
		userStrs[i] = "14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11|14-13-1-11"
	}
	//	for i := 0; i < 6; i++ {
	//		if i%2 == 0 {
	//			userStrs[i] = "2-1-1-12|2-1-1-12|2-1-1-12"
	//		} else {
	//			userStrs[i] = "10-9-1-7|11-10-1-8|12-11-1-9|13-12-1-10|14-13-1-11"
	//		}
	//	}
	for i := 0; i < pcount; i++ {
		if userStrs[i] == "" {
			cardLists[i] = cardLists[i][:cardCount]
			continue
		}
		arr := strings.Split(userStrs[i], "|")
		cards := CardList{}
		for i, cardInfo := range arr {
			cardInfo = cardInfo
			arr2 := strings.Split(cardInfo, "-")
			id_, value, suit, priority := arr2[0], arr2[1], arr2[2], arr2[3]
			id, _ := strconv.Atoi(id_)
			v, _ := strconv.Atoi(value)
			s, _ := strconv.Atoi(suit)
			p, _ := strconv.Atoi(priority)
			cards = append(cards, Card{id, v, s, p, i})
		}
		cardLists[i] = cards
	}
	return cardLists
}

//获取第一个出牌的人
func (r *Room) getFirstController() (user *User) {
	return r.firstController
}

//设置第一个出牌人
func (r *Room) setFirstController() {
	//获取第一个出牌的人
	user := r.getFirstController()
	//设置第一个出牌人
	r.setController(user, SetController_NewCycle)
}

//开赛更新玩家数据
func (r *Room) updateUserInfo() {
	if r.GetMatchID() == Match_GJYX {
		roomData, err := redis.Ints(r.GetRoomData("expendIngot", "integral", "charm"))
		logger.CheckFatal(err, "updateUserInfo:1")
		expendIngot, integral, charm := roomData[0], roomData[1], roomData[2]
		users := r.users
		for _, user := range users {
			user.updateIngot(-expendIngot, 3)
			user.beginGetItemEffect()
			user.beginUpdateUserInfo(map[string]int{"integral": integral, "charm": charm})
		}
	}
}

/*
还贡处理
in:牌index列表
out:-1第一局不需要进贡 -2不是头二科不需要还贡 -3没有选择要还的牌 -4还牌数量不符
	1还贡成功
*/
func (r *Room) huanTributeHandle(args []string) *string {
	fmt.Print("")
	res := "1"
	userid := args[0]
	user := UserManage.GetUser(&userid)
	//贡模式
	if r.GetTribute() {
		//第一轮不需要进贡
		if r.getInning() == 1 {
			res = "-1"
			return &res
		}
		//玩家要还的牌
		card_indexs_str := args[1]
		if card_indexs_str == "" {
			res = "-3"
			return &res
		}
		//将字符串的index数组转化成数字index数组
		card_indexs := StrArrToIntArr(strings.Split(card_indexs_str, "|"))
		sort.Ints(card_indexs)
		//还落贡
		res = r.huanLaTribute(user, card_indexs)
	}
	return &res
}

//还落贡
func (r *Room) huanLaTribute(user *User, card_indexs []int) string {
	res := "1"
	userRanking := user.getRanking()
	//不是头二科,不需要还贡
	if userRanking > 1 {
		res = "-2"
		return res
	}
	//还牌数量不符(还贡的落贡数量和吃到的落贡数不符:对方进不起了)
	if len(card_indexs) != len(user.getJinLaTributeCards()) {
		res = "-4"
		return res
	}
	return res
}

/*
进贡处理(后台处理)
*/
func (r *Room) jinTributeHandle() {
	//进贡模式
	if !r.GetTribute() {
		return
	}
	//第一轮不需要进贡
	if r.getInning() == 1 {
		users := r.getUsers()
		for _, user := range users {
			user.setTributeStatus(TributeStatus_Done)
		}
		return
	}
	//进落贡
	r.jinLaTribute()
}

//进落贡
func (r *Room) jinLaTribute() {
	users := r.getRanking()
	for _, user := range users {
		//头科和二科不需要进贡,等玩家还贡
		if user.getRanking() <= 1 {
			continue
		}
		//玩家已进过贡
		if user.getTributeStatus() == TributeStatus_Done {
			continue
		}
		cards := []Card{}
		if user.getRanking() == 4 { //二落进二科,1个最大的
			cards = user.getCards()[:1]
		} else if user.getRanking() == 5 { //大落进头科,2个最大的
			cards = user.getCards()[:2]
		} else {
			//3科4科直接设置为已进贡
			user.setTributeStatus(TributeStatus_Done)
			continue
		}
		//设置玩家已进贡
		user.setTributeStatus(TributeStatus_Done)
		//储存进的贡
		user.setJinLaTributeCards(cards)
		//从自己的牌中删除
		indexs := []int{}
		for _, card := range cards {
			indexs = append(indexs, card.Index)
		}
		user.updateUserCards(indexs)
		//将贡进给吃贡人
		var eatTributeUser *User
		if user.getRanking() == 4 { //二科吃二落,1个最大的
			eatTributeUser = r.getRanking()[1]
		} else if user.getRanking() == 5 { //头科吃大落,2个最大的
			eatTributeUser = r.getRanking()[0]
		}
		//存储进的贡
		eatTributeUser.setJinLaTributeCards(cards)
		eatTributeUser.addCards(cards)
	}
}

//获取革命的玩家
func (r *Room) getRevolutionUser() *User {
	var revolutionUser *User
	for _, user := range r.getUsers() {
		if user.isRevolution() {
			revolutionUser = user
			break
		}
	}
	return revolutionUser
}

//是否有革命的人
func (r *Room) haveRevolutionUser() bool {
	return r.getRevolutionUser() != nil
}

//革命处理
func (r *Room) revolutionHandle() {
	//不是革命模式
	if !r.GetRevolution() {
		return
	}
	users := r.getUsers()
	revolutionUsers := []*User{}
	for _, user := range users {
		maxCard := user.getCards()[0]
		//最大的牌小于2
		if maxCard.Priority < Priority_Two {
			user.setStatus(UserStatus_WaitRevolution)
			revolutionUsers = append(revolutionUsers, user)
		}
	}
	//有可以革命的人
	if len(revolutionUsers) > 0 {
		//房间正在等待革命
		r.SetRoomStatus(RoomStatus_Revolution)
		for _, user := range users {
			status := UserStatus_WaitOtherRevolution
			if user.getStatus() == UserStatus_WaitRevolution {
				status = UserStatus_WaitRevolution
				//开启革命计时器
				user.countDown_revolution()
			}
			message := fmt.Sprintf(",%s,%d,%d", *user.getUserID(), status, 10)
			user.setController(message)
		}
	}
}
