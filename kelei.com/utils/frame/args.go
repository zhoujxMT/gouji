package frame

//初始化参数列表
func initArgs(args_ Args) {
	args = args_
	//服务名称默认值
	if args.ServerName == "" {
		args.ServerName = "ServerName"
	}
	//加载框架完成
	if args.Loaded == nil {
		args.Loaded = func() {
		}
	}
}
