package divineflake

import (
	"net"
)

func LocalAddrWithPrefix(seg ...byte) net.IP {
	ifaces, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}
	if len(seg) > 4 {
		seg = seg[:4]
	}

	for _, iface := range ifaces {

		ipnet, ok := iface.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}

		ip := ipnet.IP.To4()
		if ip == nil {
			continue
		}

		ok = true

		for i, v := range seg {
			if ip[i] != v {
				ok = false
				break
			}
		}
		if ok {
			return ip
		}
	}

	return nil
}

func InetAddr(addr net.IP) uint32 {
	if len(addr) != 4 {
		return 0
	}
	return uint32(addr[0]) | uint32(addr[1]) | uint32(addr[2]) | uint32(addr[3])
}
