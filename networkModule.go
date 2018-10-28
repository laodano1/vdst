package main

import (
	"log"
	"net"
	"strings"
	"runtime"
)

const Port = ":3000"

//func GetLocalIPAddr()  {
func GetLocalIpAddr() string {
	localIpv4Addr := ""
	ifaces, err := net.Interfaces()
	//ifaces, err := net.InterfaceAddrs()
	if err != nil {
		log.Println(err.Error())
	}

	//fmt.Printf("OS type: %s\n", runtime.GOOS)
	FoundLocalIPV4Addr := false // found local ipv4 address or not, default false
	for _, iface := range ifaces {
		//fmt.Printf("-------> %s \n", iface.Name)
		if strings.ToLower(runtime.GOOS) == "darwin" { // Mac system
			//log.Printf("This is '%s' environment.", runtime.GOOS)
			if iface.Name == "en0" {
				adds, err := iface.Addrs()
				if err != nil {
					log.Println(err.Error())
				}

				for _, add := range adds {
					if strings.HasPrefix(add.String(), "192") {
						//fmt.Printf("%s\n", add)
						strs := strings.Split(add.String(), "/")
						log.Printf("interface: '%s', ip: '%s'", iface.Name, strs[0])
						localIpv4Addr = strs[0]
						FoundLocalIPV4Addr = true
						break
					}
				}

				if FoundLocalIPV4Addr {
					break
				}
			}
		}
	}

	return localIpv4Addr
}

//func LocalIpAddrGet() string {
//	// 需要通过代码获取ip地址 ？？？
//
//
//	return "192.168.1.99" + Port
//}
