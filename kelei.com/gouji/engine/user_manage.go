/*
玩家管理
*/

package engine

import (
	"net"
	"sync"

	. "kelei.com/utils/delaymsg"
)

type userManage struct {
	judgmentUser   *User
	users          map[string]*User //比赛中的玩家列表
	keepAliveUsers map[string]*User //长连接的玩家列表(功能中的)
	lock           sync.Mutex
}

func userManage_init() *userManage {
	UserManage := userManage{}
	UserManage.users = map[string]*User{}
	UserManage.keepAliveUsers = map[string]*User{}
	return &UserManage
}

var (
	UserManage = *userManage_init()
)

//重置玩家
func (u *userManage) Reset() {
	u.lock.Lock()
	defer u.lock.Unlock()
	for _, user := range u.users {
		user.deleteUserInfo()
		user.reset()
	}
	u.users = map[string]*User{}
}

//获取玩家
func (u *userManage) GetAllUsers() map[string]*User {
	u.lock.Lock()
	defer u.lock.Unlock()
	return u.users
}

//获取玩家
func (u *userManage) GetUser(userid *string) *User {
	u.lock.Lock()
	defer u.lock.Unlock()
	return u.users[*userid]
}

//获取玩家从长连接的玩家列表中
func (u *userManage) GetUserFromKeepAliveUsers(userid *string) *User {
	u.lock.Lock()
	defer u.lock.Unlock()
	return u.keepAliveUsers[*userid]
}

func (u *userManage) GetJudgmentUser() *User {
	return u.judgmentUser
}

func (u *userManage) setJudgmentUser(user *User) {
	u.judgmentUser = user
}

//获取比赛中的玩家数量
func (u *userManage) GetUserCount() int {
	u.lock.Lock()
	defer u.lock.Unlock()
	return len(u.users)
}

//获取长连接的玩家数量
func (u *userManage) GetKeepAliveUsersCount() int {
	u.lock.Lock()
	defer u.lock.Unlock()
	return len(u.keepAliveUsers)
}

//创建一个玩家
func (u *userManage) createUser() *User {
	user := User{}
	user.cards = []Card{}
	return &user
}

//根据userid创建玩家
func (u *userManage) newUser(userid *string) *User {
	user := UserManage.createUser()
	user.setUserID(userid)
	return user
}

//添加玩家(比赛中的玩家,不包含功能中的玩家)
func (u *userManage) AddUser(uid *string, userid *string, conn net.Conn) *User {
	user := u.users[*userid]
	if user == nil {
		user = u.createUser()
		user.uid = uid
		user.userid = userid
		user.conn = conn
		user.ranking = -1
		user.jinLaTributeCards = []Card{}
		user.dm = NewDelayMessage()
		user.dm.Start()
		user.online = true
		user.levelRound = map[int]bool{}
		user.lookedUsers = []*User{}
		u.lock.Lock()
		defer u.lock.Unlock()
		u.users[*userid] = user
	} else {
		user.conn = conn
	}
	return user
}

//添加玩家（往长连接的玩家列表中）
func (u *userManage) AddUserIntoKeepAliveUsers(uid, userid, gaterpc *string) {
	if u.keepAliveUsers[*userid] != nil {
		return
	}
	user := u.createUser()
	user.SetGateRpc(*gaterpc)
	user.uid = uid
	user.userid = userid
	u.lock.Lock()
	defer u.lock.Unlock()
	u.keepAliveUsers[*userid] = user
}

//添加玩家
func (u *userManage) AddTheUser(user *User) {
	userid := *user.getUserID()
	if u.users[userid] != nil {
		return
	}
	u.lock.Lock()
	defer u.lock.Unlock()
	u.users[userid] = user
}

//删除一个玩家（从比赛玩家列表中）
func (u *userManage) RemoveUser(user *User) {
	delete(u.users, *user.getUserID())
}

//删除一个玩家（从长连接玩家列表中）
func (u *userManage) RemoveUserFromKeepAliveUsers(user *User) {
	delete(u.keepAliveUsers, *user.getUserID())
}
