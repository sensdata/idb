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
	dnsInfo := &model.DNSInfo{}

	// 检查是否使用 systemd-resolved
	if _, err := os.Stat("/run/systemd/resolve/resolv.conf"); err == nil {
		// 获取 DNS 服务器列表
		out, err := utils.Exec("resolvectl dns")
		if err == nil {
			for _, line := range strings.Split(out, "\n") {
				// 跳过空行
				if len(strings.TrimSpace(line)) == 0 {
					continue
				}
				// 查找包含 eth0 的行
				if strings.Contains(line, "eth0") {
					// 使用更可靠的方式提取DNS服务器信息
					if idx := strings.Index(line, ":"); idx != -1 {
						dnsServers := strings.TrimSpace(line[idx+1:])
						if len(dnsServers) > 0 {
							servers := strings.Fields(dnsServers)
							dnsInfo.Servers = append(dnsInfo.Servers, servers...)
						}
					}
				}
			}
		}

		// 获取超时和重试设置
		if data, err := os.ReadFile("/etc/systemd/resolved.conf"); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "ResolveTimeoutSec=") {
					if val, err := strconv.Atoi(strings.TrimPrefix(line, "ResolveTimeoutSec=")); err == nil {
						dnsInfo.Timeout = val
					}
				} else if strings.HasPrefix(line, "DNSStubRetryCount=") {
					if val, err := strconv.Atoi(strings.TrimPrefix(line, "DNSStubRetryCount=")); err == nil {
						dnsInfo.Retry = val
					}
				}
			}
		}

		if len(dnsInfo.Servers) > 0 {
			return dnsInfo, nil
		}
	}

	// 如果不是 systemd-resolved 或获取失败，使用标准的 resolv.conf
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
					dnsInfo.Retry, _ = strconv.Atoi(strings.TrimPrefix(field, "attempts:"))
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
	var interfaces []model.NetworkInterface

	gateways, _ := getInterfaceGateways()

	// 获取所有网络接口
	links, err := netlink.LinkList()
	if err != nil {
		return interfaces, fmt.Errorf("error fetching links: %v", err)
	}

	for _, link := range links {
		name := link.Attrs().Name
		mac := link.Attrs().HardwareAddr.String()

		status := "down"
		if link.Attrs().Flags&net.FlagUp != 0 {
			status = "up"
		}

		var addrInfos []model.AddressInfo
		addrs, err := netlink.AddrList(link, unix.AF_UNSPEC)
		if err == nil {
			for _, addr := range addrs {
				ip := addr.IP.String()
				mask := net.IP(addr.Mask).String()
				ipType := "IPv6"
				if addr.IP.To4() != nil {
					ipType = "IPv4"
				}

				// 匹配同网段的网关
				var matchedGateways []string
				for _, gw := range gateways[name] {
					if isSameSubnet(ip, mask, gw) {
						matchedGateways = append(matchedGateways, gw)
					}
				}

				addrInfos = append(addrInfos, model.AddressInfo{
					Type: ipType,
					Ip:   ip,
					Mask: mask,
					Gate: matchedGateways,
				})
			}
		}

		stat := link.Attrs().Statistics

		interfaces = append(interfaces, model.NetworkInterface{
			Name:    name,
			Mac:     mac,
			Status:  status,
			Address: addrInfos,
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

// 判断 IP 是否在网段中
func isSameSubnet(ipStr, maskStr, gwStr string) bool {
	ip := net.ParseIP(ipStr)
	gw := net.ParseIP(gwStr)
	if ip == nil || gw == nil {
		return false
	}

	// IPv4
	if ip.To4() != nil && gw.To4() != nil {
		mask := net.IPMask(net.ParseIP(maskStr).To4())
		if mask == nil {
			return false
		}
		ipNet := &net.IPNet{
			IP:   ip.Mask(mask),
			Mask: mask,
		}
		return ipNet.Contains(gw)
	}

	// IPv6
	if ip.To16() != nil && gw.To16() != nil && ip.To4() == nil && gw.To4() == nil {
		ones, _ := net.IPMask(net.ParseIP(maskStr).To16()).Size()
		ipNet := &net.IPNet{
			IP:   ip.Mask(net.CIDRMask(ones, 128)),
			Mask: net.CIDRMask(ones, 128),
		}
		return ipNet.Contains(gw)
	}

	return false
}

// 获取每个接口的网关数组
func getInterfaceGateways() (map[string][]string, error) {
	gateways := make(map[string][]string)
	routes, err := netlink.RouteList(nil, unix.AF_UNSPEC)
	if err != nil {
		return gateways, err
	}

	for _, route := range routes {
		if route.Gw != nil && route.LinkIndex > 0 {
			link, err := netlink.LinkByIndex(route.LinkIndex)
			if err != nil {
				continue
			}
			name := link.Attrs().Name
			gw := route.Gw.String()

			// 去重
			exists := false
			for _, g := range gateways[name] {
				if g == gw {
					exists = true
					break
				}
			}
			if !exists {
				gateways[name] = append(gateways[name], gw)
			}
		}
	}

	return gateways, nil
}
