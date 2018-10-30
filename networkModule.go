package main

import (
	"log"
	"net"
	"strings"
	"runtime"
	"strconv"
	"fmt"
)

const Port = ":3000"

func GetIFAddr(iface net.Interface, ifAdds *map[string]string)  {
	//var ipadd net.Addr
	tempAddsRef := *ifAdds

	//fmt.Printf("==> network interface: %s \n", iface.Name)
	adds, err := iface.Addrs()
	if err != nil {
		log.Println(err.Error())
	}

	for idx, add := range adds {
		//fmt.Printf("ip: %s\n", add)
		ipInfo := strings.Split(add.String(), "/")
		ifName := strings.Replace(iface.Name, " ", "", -1)
		tempAddsRef[strings.ToLower(ifName)  + "_" + strconv.Itoa(idx)] = ipInfo[0]
	}
	//fmt.Printf("ipv4: %s\n", ipInfo[0])
}

// get all ip addr and fill follow 'map[ip version]ip address' format
func GetIPv4Addr(ifAdds map[string]string, ipAdds *map[string]string)  {
	tempAdds := *ipAdds
	for k, v := range ifAdds {
		if (strings.HasPrefix(k, "local") ||  strings.HasPrefix(k, "en") ||  strings.HasPrefix(k, "eth")) && strings.HasSuffix(k, "_1") {
			tempAdds["eth_ipv4"] = v
		} else if strings.HasPrefix(k, "wireless") && strings.HasSuffix(k, "_1") {
			tempAdds["wireless_ipv4"] = v
		}
	}
}

// get all local ip v4 addresses
func GetLocalIpAddrs() map[string]string {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Println(err.Error())
	}

	ifAdds := make(map[string]string)
	ipAdds := make(map[string]string)

	fmt.Printf("OS type: %s\n", runtime.GOOS)
	//FoundLocalIPV4Addr := false // found local ipv4 address or not, default false
	for _, iface := range ifaces {
		//fmt.Printf("-------> %s \n", iface.Name)
		if strings.ToLower(runtime.GOOS) == "darwin" { // Mac system
			//log.Printf("This is '%s' environment.", runtime.GOOS)
			if strings.HasPrefix(strings.ToLower(iface.Name), "wireless") ||
				strings.HasPrefix(strings.ToLower(iface.Name), "en")  {

				GetIFAddr(iface, &ifAdds)
			}

		} else if strings.ToLower(runtime.GOOS) == "windows" {
			if strings.HasPrefix(strings.ToLower(iface.Name), "wireless") ||
				strings.HasPrefix(strings.ToLower(iface.Name), "local") {

				GetIFAddr(iface, &ifAdds)
			}
		}  else if strings.ToLower(runtime.GOOS) == "linux" {
			if strings.HasPrefix(strings.ToLower(iface.Name), "wireless") ||
				strings.HasPrefix(strings.ToLower(iface.Name), "eth") {

				GetIFAddr(iface, &ifAdds)
			}
		}
	}

	GetIPv4Addr(ifAdds, &ipAdds)

	return ipAdds
}

//func LocalIpAddrGet() string {
//	// 需要通过代码获取ip地址 ？？？
//
//
//	return "192.168.1.99" + Port
//}
