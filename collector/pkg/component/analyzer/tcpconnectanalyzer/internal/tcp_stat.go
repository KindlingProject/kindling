package internal

import (
	"path"
	"strconv"
)

func NewPidTcpStat(hostProc string, pid int) (NetSocketStateMap, error) {
	tcpFilePath := path.Join(hostProc, strconv.Itoa(pid), "net/tcp")
	return newNetIPSocket(tcpFilePath)
}
