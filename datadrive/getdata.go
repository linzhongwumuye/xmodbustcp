package datadrive


//根据一批测点ID获取值
type DataGeter interface {
	/*
		主动获取的方式:根据测点，使用URL请求、Redis等方式进行数据拉取
	*/
	GetData(rids []string) (values []string,err error)
	/*
		首先主动获取一次需要订阅的值,然后通过对值变化的情况进行回调,例如：
		从Redis获取初始值后，订阅RMQ的变化值，如果值变化，根据网络交互协议，视情况而定多少个测点组一个包发出去
	*/
	//BookData([]string) map[string]interface{}
}