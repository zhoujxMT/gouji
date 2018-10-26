/*
玩家道具效果
*/

package engine

import (
	. "kelei.com/utils/common"
)

//获取道具的效果值
func (u *User) getEffectVal(effectType int) int {
	effectVal := 0
	effectType = IndexMapOf(u.getEffectTypes(), effectType)
	if effectType != -1 {
		effectVal = u.getEffectTypes()[effectType]
	}
	return effectVal
}

//获取是否有此道具效果
func (u *User) getEffectType(effectType int) int {
	if IndexMapOf(u.getEffectTypes(), effectType) == -1 {
		//没有道具效果,返回-1
		effectType = -1
	}
	//有道具效果,返回effectType
	return effectType
}

//王的记忆
func (u *User) getKingMemory() int {
	return u.getEffectType(101)
}

//2的记忆
func (u *User) getTwoMemory() int {
	return u.getEffectType(102)
}

//积分杠杆
func (u *User) getIntegralLever() int {
	return u.getEffectType(103)
}

//积分免损
func (u *User) getIntegralInv() int {
	return u.getEffectType(104)
}

//胜率锁定
func (u *User) getWinRateLock() int {
	return u.getEffectType(105)
}

//元宝杠杆
func (u *User) getIngotLever() int {
	return u.getEffectType(106)
}

//元宝免损
func (u *User) getIngotInv() int {
	return u.getEffectType(107)
}

//魅力飞天
func (u *User) getCharmFlying() int {
	return u.getEffectType(108)
}

//经验加成
func (u *User) getExpAddition() int {
	return u.getEffectType(109)
}

//积分道具效果
func (u *User) effect_Integral(val int) int {
	//积分杠杆
	if effectType := u.getIntegralLever(); effectType != -1 {
		effectVal := u.getEffectVal(effectType)
		val = val * effectVal
	}
	//积分免损卡
	if u.getIntegralInv() != -1 {
		if val < 0 {
			val = 0
		}
	}
	return val
}

//元宝道具效果
func (u *User) effect_Ingot(val int) int {
	//元宝杠杆
	if effectType := u.getIngotLever(); effectType != -1 {
		effectVal := u.getEffectVal(effectType)
		val = val * effectVal
	}
	//元宝免损卡
	if u.getIngotInv() != -1 {
		if val < 0 {
			val = 0
		}
	}
	return val
}

//经验道具效果
func (u *User) effect_Exp(val int) int {
	//经验加成
	if effectType := u.getExpAddition(); effectType != -1 {
		effectVal := u.getEffectVal(effectType)
		val = int(float64(val) * (1 + float64(effectVal)/100))
	}
	return val
}

//魅力道具效果
func (u *User) effect_Charm(val int) int {
	//经验加成
	if effectType := u.getCharmFlying(); effectType != -1 {
		effectVal := u.getEffectVal(effectType)
		val = val + effectVal
	}
	return val
}
