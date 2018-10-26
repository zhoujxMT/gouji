/*
房间比赛结束
*/

package engine

import (
	"bytes"
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"

	. "kelei.com/utils/common"
	"kelei.com/utils/logger"
)

//比赛结束的检测
func (r *Room) checkMatchingOver() {
	users := r.getUsers()
	//团队1走科的玩家列表
	redTeamGoUsers := []*User{}
	//团队2走科的玩家列表
	buleTeamGoUsers := []*User{}
	for _, user := range users {
		if !user.isActive() {
			if user.getTeamMark() == TeamMark_Red {
				redTeamGoUsers = append(redTeamGoUsers, user)
			} else {
				buleTeamGoUsers = append(buleTeamGoUsers, user)
			}
		}
	}
	//其中一方全部走科,比赛结束
	if len(redTeamGoUsers) >= pcount/2 || len(buleTeamGoUsers) >= pcount/2 {
		r.matchingOverHandle()
	}
}

/*
比赛结束的处理
*/
func (r *Room) matchingOverHandle() {
	//关闭玩家的
	r.closeUserCountDown()
	//删除所有玩家的负载均衡服务器信息
	r.deleteUsersInfo()
	//局数增加
	r.setInning(r.getInning() + 1)
	//剩余玩家排名处理
	r.userRankingHandle()
	//所有人的排名推送
	r.goPush(nil)
	//展示剩余排面
	r.showSurplusCards()
	//推送比赛结束的信息给所有玩家
	r.pushMatchingEndInfo()
	//设置房间为准备中的状态
	r.SetRoomStatus(RoomStatus_Setout)
	//重置所有的玩家
	r.resetUsers()
	//删除离线和逃跑的玩家
	r.removeUsers()
	//推送房间的状态,所有人变成了未准备的状态
	r.matchingPush(nil)
	//开启准备倒计时
	r.setOutCountDown()
	//上报用户数据
	r.reportedUserInfo()
}

//上报用户数据
func (r *Room) reportedUserInfo() {
	users := r.getUsers()
	for _, user := range users {
		if !user.getIsAI() {
			user.uploadWX()
		}
	}
}

//关闭所有玩家的定时器
func (r *Room) closeUserCountDown() {
	for _, user := range r.getUsers() {
		user.closeCountDown()
	}
}

//删除离线和逃跑的玩家
func (r *Room) removeUsers() {
	users := r.getUsers()
	for _, user := range users {
		//玩家不在线
		if !user.getOnline() || user.getIsAI() {
			//退出房间
			ExitMatch([]string{*user.getUserID()})
		}
	}
}

//玩家排名的处理,通过积分判定胜负
func (r *Room) userRankingHandle() {
	//革命玩家的排名处理
	r.revolutionUserRankingHandle()
	//正常玩家的排名处理
	r.normalUserRankingHandle()
	//根据排名计算基点
	r.calculateBasePoint()

	users := r.getRanking()
	redTeamBasePoint, blueTeamBasePoint := 0, 0
	for _, user := range users {
		if user.getTeamMark() == TeamMark_Red {
			redTeamBasePoint += user.getBasePoint()
		} else {
			blueTeamBasePoint += user.getBasePoint()
		}
	}
	if redTeamBasePoint > blueTeamBasePoint {
		for _, user := range users {
			if user.getTeamMark() == TeamMark_Red {
				user.setMatchResult(MatchResult_Win)
			} else {
				user.setMatchResult(MatchResult_Lose)
			}
		}
	} else if blueTeamBasePoint > redTeamBasePoint {
		for _, user := range users {
			if user.getTeamMark() == TeamMark_Red {
				user.setMatchResult(MatchResult_Lose)
			} else {
				user.setMatchResult(MatchResult_Win)
			}
		}
	}
}

//计算基点
func (r *Room) calculateBasePoint() {
	threeFamily := r.threeFamily()
	users := r.getRanking()
	for ranking, user := range users {
		x := 1
		if threeFamily {
			ranking = 6
			if user.getRanking() >= Ranking_Four {
				x = -1
			}
		}
		bp := basePoint[ranking]
		bp = bp * x
		user.setBasePoint(bp)
	}
}

/*
ShowSurplusCards_Push(展示剩余牌面)
push:44|,46|1$1$1$1$16$2
*/
func (r *Room) showSurplusCards() {
	buff := bytes.Buffer{}
	users := r.getRanking()
	for _, user := range users {
		buff.WriteString(fmt.Sprintf("%s|%s,", *user.getUserID(), *user.getCardsID()))
	}
	message := buff.String()
	message = message[:len(message)-1]
	time.Sleep(time.Millisecond * 10)
	time.Sleep(time.Second * 1)
	pushMessageToUsers("ShowSurplusCards_Push", []string{message}, r.getUserIDs())
	time.Sleep(time.Second * 2)
}

//获取红蓝队头二科的数量
func (r *Room) getTop2UserCount() (int, int) {
	redTeamTop2Count, blueTeamTop2Count := 0, 0
	for i, user := range r.getRanking() {
		if user == nil {
			break
		}
		if i < 2 {
			if user.getTeamMark() == TeamMark_Red {
				redTeamTop2Count += 1
			} else {
				blueTeamTop2Count += 1
			}
		}
	}
	return redTeamTop2Count, blueTeamTop2Count
}

//获取头二科是否都是自己团队的
func (r *Room) revolutionUserRankingHandle() {
	//获取红蓝队头二科的数量
	redTeamTop2Count, blueTeamTop2Count := r.getTop2UserCount()
	//获取革命的玩家
	revolutionUser := r.getRevolutionUser()
	//有革命的玩家
	if revolutionUser != nil {
		top2Count := 0
		if revolutionUser.getTeamMark() == TeamMark_Red {
			top2Count = redTeamTop2Count
		} else {
			top2Count = blueTeamTop2Count
		}
		if top2Count >= 2 { //头二科都是自己团队的,革命的这个人是三科
			r.setRankingByIndex(revolutionUser, Ranking_Three)
		} else { //革命的这个人是四科
			r.setRankingByIndex(revolutionUser, Ranking_Four)
		}
	}
}

//正常玩家的排名处理
func (r *Room) normalUserRankingHandle() {
	users := r.getUsers()
	for _, user := range users {
		user.close_countDown_playCard()
		r.setRanking(user, 1)
	}
}

//是否三户
func (r *Room) threeFamily() bool {
	users := r.getRanking()
	//前三科的团队是一样的
	if users[Ranking_One].getTeamMark() == users[Ranking_Two].getTeamMark() && users[Ranking_One].getTeamMark() == users[Ranking_Three].getTeamMark() {
		return true
	}
	return false
}

//结算积分信息
var IntegralData_HYTW = []int{40, 20, 0, 0, -20, -40, 40}

/*
获取团队结算数据
out:团队胜利元宝数,团队失败元宝数,玩家对应的积分数
*/
func (r *Room) getTeamBalanceInfo() (int, int, int, int) {
	users := r.getRanking()
	winIngot, loseIngot := 0, 0
	winIntegral, loseIntegral := 0, 0
	threeFamily := r.threeFamily()
	for ranking, user := range users {
		if user == nil {
			break
		}
		matchResult := user.getMatchResult()
		x := 1
		if threeFamily {
			ranking = 6
			if matchResult == MatchResult_Lose {
				x = -1
			}
		}
		ingot, integral := 0, 0
		if r.GetMatchID() == Match_GJYX {
			balanceData, err := redis.Ints(r.GetBalanceData(ranking, "ingot", "integral"))
			logger.CheckError(err)
			ingot, integral = balanceData[0], balanceData[1]
		} else if r.GetMatchID() == Match_HYTW {
			ingot, integral = 0, IntegralData_HYTW[ranking]
		}
		ingot = ingot * x
		integral = integral * x
		if matchResult == MatchResult_Win {
			winIngot += ingot
			winIntegral += integral
		} else if matchResult == MatchResult_Lose {
			loseIngot += ingot
			loseIntegral += integral
		}
	}
	return winIngot, loseIngot, winIntegral, loseIntegral
}

//获取所有玩家结算数据
func (r *Room) getUserBalanceInfo() (*string, map[string][]int) {
	winIngot, loseIngot, winIntegral, loseIntegral := r.getTeamBalanceInfo()
	mapInfo := map[string][]int{}
	buff := bytes.Buffer{}
	users := r.getRanking()
	for _, user := range users {
		userid := *user.getUserID()
		mapInfo[userid] = make([]int, 2)
		matchResult := user.getMatchResult()
		//通用的积分
		if matchResult == MatchResult_Win {
			mapInfo[userid][0] = winIntegral
		} else if matchResult == MatchResult_Lose {
			mapInfo[userid][0] = loseIntegral
		}
		if r.GetMatchID() == Match_GJYX {
			ingot := 0
			if matchResult == MatchResult_Win {
				ingot = winIngot
			} else if matchResult == MatchResult_Lose {
				ingot = loseIngot
			}
			//元宝,等级,是否破产
			userIngot, level, bankrupt := 0, 1, 0
			itemEffect := Res_NoData
			//不是AI
			if !user.getIsAI() {
				ingot = user.effect_Ingot(ingot)
				userIngot = user.updateIngot(ingot, 4)
				if userIngot <= 0 {
					bankrupt = 1
				}
				level = user.getLevel()
				itemEffect = user.getItemEffect()
			}
			mapInfo[userid][1] = userIngot
			buff.WriteString(fmt.Sprintf("%s$%d$%d$%s$%d$%d|", userid, ingot, bankrupt, itemEffect, user.getTeamMark(), level))
		} else if r.GetMatchID() == Match_HYTW {
			integral := mapInfo[userid][0]
			level := 1
			//不是AI
			if !user.getIsAI() {
				level = user.getLevel()
			}
			buff.WriteString(fmt.Sprintf("%s$%d$%d$%d$%d|", userid, integral, matchResult, user.getTeamMark(), level))
		}
	}
	usersInfo := buff.String()
	usersInfo = usersInfo[:len(usersInfo)-1]
	return &usersInfo, mapInfo
}

//推送比赛结束的信息-够级英雄
func (r *Room) pushMatchingEndInfo_GJYX() {
	usersInfo, mapInfo := r.getUserBalanceInfo()
	roomData, err := redis.Ints(r.GetRoomData("winExp", "loseExp", "expendIngot"))
	logger.CheckFatal(err, "pushMatchingEndInfo:1")
	winExp, loseExp, expendIngot := roomData[0], roomData[1], roomData[2]
	messages := make([]string, pcount)
	userids := r.getUserIDs()
	for i, userid := range userids {
		user := UserManage.GetUser(&userid)
		//获取玩家经验、胜平负
		getExp, matchResult := 0, user.getMatchResult()
		if matchResult == MatchResult_Win {
			getExp = winExp
		} else if matchResult == MatchResult_Flat {
			getExp = loseExp
		} else if matchResult == MatchResult_Lose {
			getExp = loseExp
		}
		isUpgrade, level, lvlExp, upExp, upIngotAward := 0, 0, 0, 0, 0
		userIntegral := 0
		if !user.getIsAI() {
			//经验的道具效果
			getExp = user.effect_Exp(getExp)
			//更新玩家经验
			isUpgrade, level, lvlExp, upExp, upIngotAward = user.updateExp(getExp)
			//更新玩家积分
			userIntegral = mapInfo[userid][0]
			//积分的道具效果
			userIntegral = user.effect_Integral(userIntegral)
			//比赛结束更新玩家信息
			user.endUpdateUserInfo(map[string]int{"integral": userIntegral, "matchResult": matchResult})
		}
		//三户
		if r.threeFamily() {
			if user.getRanking() >= Ranking_Four {
				matchResult = -2
			} else {
				matchResult = 2
			}
		}
		//获取玩家元宝
		userIngot := mapInfo[userid][1]
		//等级|积分|经验|当前级别经验|当前级别升级经验|房费|比赛结果(2圈三户1胜0平-1负-2被圈三户)|是否升级|金币|升级金币奖励|房间底分,userid$获得元宝$是否破产|
		messages[i] = fmt.Sprintf("%d|%d|%d|%d|%d|%d|%d|%d|%d|%d|%d,%s", level, userIntegral, getExp, lvlExp, upExp, expendIngot, matchResult, isUpgrade, userIngot, upIngotAward, r.getMultiple(), *usersInfo)
	}
	pushMessageToUsers("MatchingEnd_Push", messages, userids)
	r.pushJudgment("MatchingEnd_Push", "1")
}

//推送比赛结束的信息-好友同玩
func (r *Room) pushMatchingEndInfo_HYTW() {
	usersInfo, mapInfo := r.getUserBalanceInfo()
	redIntegral, blueIntegral := 0, 0
	//增加玩家好友同玩积分
	for _, user := range r.getUsers() {
		theIntegral := mapInfo[*user.getUserID()][0]
		user.setHYTWIntegral(user.getHYTWIntegral() + theIntegral)
		if user.getTeamMark() == TeamMark_Red {
			redIntegral = theIntegral
		} else {
			blueIntegral = theIntegral
		}
	}
	if redIntegral > 0 {
		r.redIntegral = r.redIntegral + redIntegral
	}
	if blueIntegral > 0 {
		r.blueIntegral = r.blueIntegral + blueIntegral
	}
	messages := make([]string, pcount)
	userids := r.getUserIDs()
	inning := r.getInning()
	islast := 0
	if inning == r.getInnings() {
		islast = 1
	}
	for i, userid := range userids {
		user := UserManage.GetUser(&userid)
		level := user.getLevel()
		integral := user.getHYTWIntegral()
		matchResult := user.getMatchResult()
		//三户
		if r.threeFamily() {
			if user.getRanking() >= Ranking_Four {
				matchResult = -2
			} else {
				matchResult = 2
			}
		}
		//等级|积分|第几局|是否是最后一局|红队积分|蓝队积分|比赛结果(2圈三户1胜0平-1负-2被圈三户),userid$获得积分$是否胜利|
		messages[i] = fmt.Sprintf("%d|%d|%d|%d|%d|%d|%d,%s", level, integral, inning, islast, r.redIntegral, r.blueIntegral, matchResult, *usersInfo)
	}
	if islast == 1 {
		r.resetHYTW()
	}
	pushMessageToUsers("MatchingEnd_Push", messages, userids)
}

//好友同玩-重置房间
func (r *Room) resetHYTW() {
	//设置房间为第一局
	r.setInning(0)
	//清空房间积分
	r.redIntegral = 0
	r.blueIntegral = 0
	//清空玩家积分
	users := r.getUsers()
	for _, user := range users {
		if user != nil {
			user.setHYTWIntegral(0)
		}
	}
}

/*
MatchingEnd_Push(推送比赛结束的信息)
out:
	够级英雄: 等级|积分|经验|当前级别经验|当前级别升级经验|房费|比赛结果(1胜0平-1负)|是否升级|金币,userid$获得元宝$是否破产$buff道具列表$红蓝队$玩家等级|
	好友同玩: 等级|积分|第几局|是否是最后一局|红队积分|蓝队积分,userid$获得积分$是否胜利$红蓝队$玩家等级|
*/
func (r *Room) pushMatchingEndInfo() {
	if r.GetMatchID() == Match_GJYX {
		r.pushMatchingEndInfo_GJYX()
	} else if r.GetMatchID() == Match_HYTW {
		r.pushMatchingEndInfo_HYTW()
	}
}
