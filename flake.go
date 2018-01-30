package divineflake

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type Flake interface {
	Generate() uint64
}

type divineflake struct {
	machineNumber  int64 // 24bits ip+pid                << 40
	timeNumber     int64 // 32bits since start time (ms) << 8
	sequenceNumber int64 //  8bits

	startTime     int64 // 2017/12/16 00:00:00 1513353600000 (ms) 	// flake 起始时间戳 毫秒为单位
	currentTime   int64 // 当前时间 32bits
	timePrecision int64 // 精度

	sync.Mutex
}

var (
	defaultflake     Flake
	defaultStartTime = time.Date(2017, 12, 16, 0, 0, 0, 0, time.Local)
)

const ()

func init() {
	defaultflake = NewFlake()
}

func NewFlake() Flake {
	flake := &divineflake{}
	flake.machineNumber = machineID()
	flake.startTime = time.Date(2017, 12, 16, 0, 0, 0, 0, time.Local).UnixNano()
	flake.timePrecision = int64(time.Millisecond)
	flake.currentTime = flake.toDivineflakeTime(time.Now())
	return flake
}

func Generate() uint64 {
	return defaultflake.Generate()
}

func (flake *divineflake) Generate() uint64 {
	if flake == nil {
		return 0
	}

	flake.Lock()
	defer flake.Unlock()

	currentTime := flake.toDivineflakeTime(time.Now())
	if currentTime-flake.currentTime > flake.timePrecision {
		flake.currentTime = currentTime
		flake.timeNumber = currentTime / flake.timePrecision
		flake.sequenceNumber = 0
	} else {
		flake.sequenceNumber = (flake.sequenceNumber + 1) & 0xff
		if flake.sequenceNumber == 0 {
			flake.moveCurrentTime()
			time.Sleep(time.Duration(flake.timePrecision - flake.currentTime%flake.timePrecision))
		}
	}
	return flake.toID()
}

func (flake *divineflake) toID() uint64 {
	return uint64(flake.machineNumber)<<40 | uint64(flake.timeNumber)<<8 | uint64(flake.sequenceNumber)
}

func machineID() int64 {
	ip := localPrivateIPLowerTowBytes()
	pid := os.Getpid()
	return (int64(ip) << 8) | int64(pid&0xff)
}

func (flake *divineflake) toDivineflakeTime(tm time.Time) int64 {
	return tm.UnixNano() - flake.startTime
}

func (flake *divineflake) moveCurrentTime() {
	flake.timeNumber++
	flake.currentTime = flake.timeNumber * flake.timePrecision
}

func localPrivateIPLowerTowBytes() uint16 {
	localIP := LocalAddrWithPrefix(192, 168)
	if len(localIP) != 4 {
		panic(fmt.Sprintf("invalid local ip: %v", localIP))
	}
	return uint16(localIP[2])<<8 | uint16(localIP[3])
}
