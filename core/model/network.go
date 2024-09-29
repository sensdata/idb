package model

// 网络信息
type NetworkInfo struct {
	DNS      DNSInfo            `json:"dns"`      //Dns
	Networks []NetworkInterface `json:"networks"` //网络接口
}

type DNSInfo struct {
	Servers    []string
	Timeout    int
	RetryTimes int
}

// 网络接口
type NetworkInterface struct {
	Name    string      `json:"name"`    //名称
	Status  string      `json:"status"`  //状态
	Mac     string      `json:"mac"`     //mac地址
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
	Rx      string `json:"rx"`       //接收数据量
	RxBytes int    `json:"rx_bytes"` //接收数据量
	RxSpeed string `json:"rx_speed"` //接收实时速率
	Tx      string `json:"tx"`       //发送数据量
	TxBytes int    `json:"tx_bytes"` //发送数据量
	TxSpeed string `json:"tx_speed"` //发送实时速率
}
