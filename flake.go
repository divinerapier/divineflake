package divineflake

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"
)

type (
	// Flake flake base interface
	Flake interface {
		Generate() uint64
	}

	divineflake struct {
		machineNumber  int64 // 24bits ip+pid                << 40
		timeNumber     int64 // 32bits since start time (ms) << 8
		sequenceNumber int64 //  8bits

		startTime     int64 // 2017/12/16 00:00:00 1513353600000 (ms) 	// flake 起始时间戳 毫秒为单位
		currentTime   int64 // 当前时间 32bits
		timePrecision int64 // 精度

		sync.Mutex
	}

	// Settings flake settings
	Settings struct {
		TimeOffset    time.Time
		TimePrecision time.Duration
		Machine       int64
	}
)

const (
	machinePositionOffset  = 40
	timePositionOffset     = 8
	sequencePositionOffset = 0
)

var (
	defaultflake    Flake
	defaultSettings = Settings{
		TimeOffset:    time.Date(2017, 12, 16, 0, 0, 0, 0, time.Local),
		TimePrecision: time.Millisecond,
	}

	// sets from command line
	machineID int64
)

func init() {
	defaultflake = NewFlake(defaultSettings)

	flag.Int64Var(&machineID, "machine", 0, "local machine id")
}

// NewFlake create a divineflake
func NewFlake(settings Settings) Flake {
	flake := &divineflake{
		startTime:     settings.TimeOffset.UnixNano(),
		timePrecision: int64(settings.TimePrecision),
		machineNumber: settings.Machine,
	}
	if flake.machineNumber == 0 {
		flake.machineNumber = computeMachineID()
	}
	flake.currentTime = flake.toDivineflakeTime(time.Now())
	return flake
}

// Generate generate id with default flake
func Generate() uint64 {
	return defaultflake.Generate()
}

func (flake *divineflake) Generate() uint64 {
	if flake == nil {
		return 0
	}

	flake.Lock()
	defer flake.Unlock()

	if machineID != flake.machineNumber && machineID != 0 {
		flake.machineNumber = machineID
	}

	currentTime := flake.toDivineflakeTime(time.Now())
	if currentTime-flake.currentTime > flake.timePrecision {
		flake.currentTime = currentTime
		flake.timeNumber = currentTime / flake.timePrecision
		flake.sequenceNumber = 0
	} else {
		flake.sequenceNumber = (flake.sequenceNumber + 1) & 0xff
		if flake.sequenceNumber == 0 {
			flake.refreshTime()
			time.Sleep(time.Duration(flake.timePrecision - flake.currentTime%flake.timePrecision))
		}
	}
	return flake.id()
}

func (flake *divineflake) id() uint64 {
	return uint64(flake.machineNumber)<<machinePositionOffset | uint64(flake.timeNumber)<<timePositionOffset | uint64(flake.sequenceNumber)
}

func (flake *divineflake) toDivineflakeTime(tm time.Time) int64 {
	return tm.UnixNano() - flake.startTime
}

func (flake *divineflake) refreshTime() {
	flake.timeNumber++
	flake.currentTime = flake.timeNumber * flake.timePrecision
}

func computeMachineID() int64 {
	ip := localPrivateIPLowerTowBytes()
	pid := os.Getpid()
	return (int64(ip) << 8) | int64(pid&0xff)
}

func localPrivateIPLowerTowBytes() uint16 {
	ips, err := privateIPs()
	if err != nil {
		return 0
	}
	ip := ips[rand.Int()%len(ips)]
	fmt.Println("ip:", ip)
	if len(ip) != 4 {
		panic(fmt.Sprintf("invalid local ip: %v", ip))
	}
	return uint16(ip[2])<<8 | uint16(ip[3])
}

func privateIPs() ([]net.IP, error) {
	var ips []net.IP

	ifaces, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {

		ipnet, ok := iface.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}

		ip := ipnet.IP.To4()
		if ip == nil || !isPrivateIP(ip) {
			continue
		}

		ips = append(ips, ip)
	}

	return ips, nil
}

func localAddrWithPrefix(seg ...byte) net.IP {
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
		if ip == nil || !isPrivateIP(ip) {
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

func isPrivateIP(ip net.IP) bool {
	if len(ip) != 4 {
		return false
	}
	return ip[0] == 10 || (ip[0] == 172 && (ip[1] >= 16 && ip[1] < 32)) || (ip[0] == 192 && ip[1] == 168)
}
