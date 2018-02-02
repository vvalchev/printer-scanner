package main
import (
    "net"
    "log"
)

func dupIP(ip net.IP) net.IP {
    dup := make(net.IP, len(ip))
    copy(dup, ip)
    return dup
}

func inc(ip net.IP) {
    for j := len(ip) - 1; j >= 0; j-- {
        ip[j]++
        if ip[j] > 0 {
            break
        }
    }
}

func NetGetNetworkIPs() (ip []net.IP, err error) {
    var ipList []net.IP

    ifaces,err := net.Interfaces()
    if err != nil {
        return nil, err
    }

    // scan all interfaces
    for _,iface := range ifaces {
        if iface.Flags & net.FlagUp != 0 && iface.Flags & net.FlagBroadcast != 0 {
            addrs,err := iface.Addrs()
            if err != nil {
                log.Printf("Error reading interface address.", iface , err)
            } else {
                for _,addr := range addrs {
                    ip,ipnet,err := net.ParseCIDR(addr.String())
                    if err != nil {
                        log.Printf("Error parsing CIDR.", addr.String(), err)
                    } else {
                        ip = ip.To4() // make sure that the IP is IPv4
                        if ip != nil { // will be nill if IPv6
                            log.Printf("Scanning network interface '%v' with CIDR = %v, IP = %v\n", iface.Name, addr.String(), ip)
                            for xip := ip.Mask(ipnet.Mask); ipnet.Contains(xip); inc(xip) {
                                ipList = append(ipList, dupIP(xip))
                            }
                        }
                    }
                }
            }
        }
    }

    return ipList,nil

}
