package model

// 网络信息
type NetworkInfo struct {
	DNS      []string           `json:"dns"`      //Dns
	Timeout  string             `json:"timeout"`  //查询超时
	Attempts string             `json:"attempts"` //重试次数
	Networks []NetworkInterface `json:"networks"` //网络接口
}

// 网络接口
type NetworkInterface struct {
	Name    string      `json:"name"`    //名称
	Status  string      `json:"status"`  //状态
	Mac     string      `json:"Mac"`     //mac地址
	Proto   string      `json:"proto"`   //分配方式
	Address AddressInfo `json:"address"` //地址信息
	Traffic TrafficInfo `json:"traffic"` //流量信息
}

// 地址信息
type AddressInfo struct {
	Type string `json:"type"` //类型
	Ip   string `json:"ip"`   //ip地址
	Mask string `json:"mask"` //子网掩码
	Gate string `json:"gate"` //网关
}

// 流量信息
type TrafficInfo struct {
	Rx      string `json:"rx"`      //接收数据量
	RxBytes int    `json:"rxBytes"` //接收数据量
	RxSpeed string `json:"rxSpeed"` //接收实时速率
	Tx      string `json:"tx"`      //发送数据量
	TxBytes int    `json:"txBytes"` //发送数据量
	TxSpeed string `json:"txSpeed"` //发送实时速率
}
