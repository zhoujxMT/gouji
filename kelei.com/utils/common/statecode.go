package common

//状态码
var (
	STATUS_CODE_SUCCEED = "101" //成功
	STATUS_CODE_NODATA  = "201" //没有数据
	STATUS_CODE_UNKNOWN = "301" //未知错误
	STATUS_CODE_NOBACK  = "401" //无返回值
)

const (
	Procedure_NoReturn = 0
	Procedure_Return   = 1
)
