package common

//状态码
var (
	SC_OK      = "200" //成功
	SC_NOBACK  = "201" //无返回值
	SC_ERR     = "300" //未知错误
	SC_NODATA  = "301" //没有数据
	SC_GAMEERR = "302" //调用游戏服失败
	SC_GATEERR = "303" //调用网关失败
	SC_LBERR   = "304" //调用负载均衡失败
)

//兼容旧版本
var (
	Res_Succeed  = "1"
	Res_NoData   = "-1"   //没有数据
	Res_Unknown  = "-101" //未知数据
	Res_ArgsDiff = "-102" //参数数量不符
	Res_NoPerm   = "-103" //玩家没有操作权限
	Res_CCNoSelf = "-104" //当前牌权不是自己
	Res_NoBack   = "-c"   //没有返回值
)

//自定义常量
const (
	CONST_EMPTY = " "
)

const (
	Procedure_NoReturn = 0
	Procedure_Return   = 1
)
