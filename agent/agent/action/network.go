package action

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

func GetNetwork() (*model.NetworkInfo, error) {
	var network model.NetworkInfo

	dnsInfo, err := getDNSInfo()
	if err != nil || dnsInfo == nil {
		fmt.Printf("get dns failed")
	} else {
		network.DNS = *dnsInfo
	}

	interfaces, err := getNetwork()
	if err != nil || interfaces == nil {
		fmt.Printf("get interfaces failed")
	} else {
		network.Networks = interfaces
	}

	return &network, nil
}

func getDNSInfo() (*model.DNSInfo, error) {
	// 检查是否使用 systemd-resolved
	if _, err := os.Stat("/run/systemd/resolve/resolv.conf"); err == nil {
		// 使用 systemd-resolved 的配置文件
		file, err := os.Open("/run/systemd/resolve/resolv.conf")
		if err != nil {
			return nil, err
		}
		defer file.Close()

		return parseDNSConfig(file)
	}

	// 如果不是 systemd-resolved，使用标准的 resolv.conf
	file, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return parseDNSConfig(file)
}

// 解析 DNS 配置文件
func parseDNSConfig(file *os.File) (*model.DNSInfo, error) {
	dnsInfo := &model.DNSInfo{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		// 跳过注释和空行
		if strings.HasPrefix(line, "#") || len(strings.TrimSpace(line)) == 0 {
			continue
		}

		if strings.HasPrefix(line, "nameserver") {
			fields := strings.Fields(line)
			if len(fields) > 1 {
				dnsInfo.Servers = append(dnsInfo.Servers, fields[1])
			}
		}
		if strings.HasPrefix(line, "options") {
			fields := strings.Fields(line)
			for _, field := range fields {
				if strings.HasPrefix(field, "timeout:") {
					dnsInfo.Timeout, _ = strconv.Atoi(strings.TrimPrefix(field, "timeout:"))
				} else if strings.HasPrefix(field, "attempts:") {
					dnsInfo.RetryTimes, _ = strconv.Atoi(strings.TrimPrefix(field, "attempts:"))
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(dnsInfo.Servers) == 0 {
		return nil, fmt.Errorf("没有找到 DNS 服务器信息")
	}

	return dnsInfo, nil
}

func getNetwork() ([]model.NetworkInterface, error) {
	var interfaces = []model.NetworkInterface{}

	// 获取所有网络接口
	links, err := netlink.LinkList()
	if err != nil {
		fmt.Println("Error fetching links:", err)
		return interfaces, err
	}

	for _, link := range links {
		// 获取接口的名字
		name := link.Attrs().Name
		mac := link.Attrs().HardwareAddr.String()

		// 获取接口的IP地址和子网掩码
		address := ""
		mask := ""
		ipType := ""
		addrs, err := netlink.AddrList(link, unix.AF_UNSPEC)
		if err != nil {
			fmt.Println("Error fetching addresses:", err)
		} else {
			if len(addrs) > 0 {
				addr := addrs[0]
				address = addr.IP.String()
				mask = net.IP(addr.Mask).String()

				// 判断IP类型
				if addr.IP.To4() != nil {
					ipType = "IPv4"
				} else {
					ipType = "IPv6"
				}
			}
		}

		// 获取默认网关（对于所有接口，实际上需要遍历路由表找到带有默认路由的接口）
		gateway := ""
		routes, err := netlink.RouteList(link, unix.AF_UNSPEC)
		if err != nil {
			fmt.Println("Error fetching routes:", err)
		} else {
			if len(routes) > 0 {
				route := routes[0]
				if route.Dst == nil {
					gateway = route.Gw.String()
				}
			}
		}

		stat := link.Attrs().Statistics
		fmt.Printf("Tx: %d, Rx: %d", stat.TxBytes, stat.RxBytes)

		interfaces = append(interfaces, model.NetworkInterface{
			Name: name,
			Mac:  mac,
			Address: model.AddressInfo{
				Type: ipType,
				Ip:   address,
				Mask: mask,
				Gate: gateway,
			},
			Traffic: model.TrafficInfo{
				Rx:      utils.FormatMemorySize(stat.RxBytes),
				RxBytes: int(stat.RxBytes),
				Tx:      utils.FormatMemorySize(stat.TxBytes),
				TxBytes: int(stat.TxBytes),
			},
		})
	}
	return interfaces, nil
}
