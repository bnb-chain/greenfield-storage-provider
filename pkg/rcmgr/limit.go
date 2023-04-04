package rcmgr

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	"github.com/shirou/gopsutil/v3/docker"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/util"
)

// func PodLimitMemory() error {
// 	config, err := rest.InClusterConfig()
// 	if err != nil {
// 		log.Errorw("failed to get kubernetes cluster config", "error", err)
// 		return err
// 	}
// 	clientSet, err := kubernetes.NewForConfig(config)
// 	if err != nil {
// 		log.Errorw("failed to create a new kubernetes client set", "error", err)
// 	}
//
// 	pod, err := clientSet.CoreV1().Pods("").Get(context.Background(), "", metav1.GetOptions{})
//
// 	var container *v1.Container
// 	for _, c := range pod.Spec.Containers {
// 		if c.Name == "containerName" {
// 			container = &c
// 			break
// 		}
// 	}
//
// 	if container == nil {
// 		log.Error("failed to get container")
// 	}
// 	memoryLimit := container.Resources.Limits.Memory().String()
// }

const (
	LimitFactor              = 0.85
	DefaultMemorySize uint64 = 8 * 1024 * 1024
)

// Limit is an interface that that specifies basic resource limits.
type Limit interface {
	// GetMemoryLimit returns the (current) memory limit.
	GetMemoryLimit() int64
	// GetConnLimit returns the connection limit, for inbound or outbound connections.
	GetConnLimit(Direction) int
	// GetConnTotalLimit returns the total connection limit
	GetConnTotalLimit() int
	// GetFDLimit returns the file descriptor limit.
	GetFDLimit() int
	// String returns the Limit state string
	String() string
}

// Limiter is an interface for providing limits to the resource manager.
type Limiter interface {
	GetSystemLimits() Limit
	GetTransientLimits() Limit
	GetServiceLimits(svc string) Limit
	String() string
}

var _ Limit = &BaseLimit{}

// BaseLimit is a mixin type for basic resource limits.
type BaseLimit struct {
	Conns         int   `json:",omitempty"`
	ConnsInbound  int   `json:",omitempty"`
	ConnsOutbound int   `json:",omitempty"`
	FD            int   `json:",omitempty"`
	Memory        int64 `json:",omitempty"`
}

// GetMemoryLimit returns the (current) memory limit.
func (limit *BaseLimit) GetMemoryLimit() int64 {
	return limit.Memory
}

// GetConnLimit returns the connection limit, for inbound or outbound connections.
func (limit *BaseLimit) GetConnLimit(direction Direction) int {
	if direction == DirInbound {
		return limit.ConnsInbound
	}
	return limit.ConnsOutbound
}

// GetConnTotalLimit returns the total connection limit
func (limit *BaseLimit) GetConnTotalLimit() int {
	return limit.Conns
}

// GetFDLimit returns the file descriptor limit.
func (limit *BaseLimit) GetFDLimit() int {
	return limit.FD
}

// String returns the Limit state string
// TODO:: supports connection and fd field
func (limit *BaseLimit) String() string {
	return fmt.Sprintf("memory limits %d", limit.Memory)
}

// InfiniteBaseLimit are a limiter configuration that uses unlimited limits, thus effectively not limiting anything.
// Keep in mind that the operating system limits the number of file descriptors that an application can use.
var InfiniteBaseLimit = BaseLimit{
	Conns:         math.MaxInt,
	ConnsInbound:  math.MaxInt,
	ConnsOutbound: math.MaxInt,
	FD:            math.MaxInt,
	Memory:        math.MaxInt64,
}

// DynamicLimits generate limits by os resource
func DynamicLimits() *BaseLimit {
	availableMem := DefaultMemorySize
	virtualMem, err := mem.VirtualMemory()
	if err != nil {
		log.Errorw("failed to get os memory states", "error", err)
	} else {
		availableMem = virtualMem.Available
	}

	dockerIDs, err := docker.GetDockerIDList()
	if err != nil {
		log.Errorw("failed to get docker id", "error", err)
	}
	log.Infow("docker id list", "length", len(dockerIDs), "list", dockerIDs)
	containerMem, err := docker.CgroupMemDocker(dockerIDs[0])
	log.Infow("use gopsutil to get container mem", "memory", containerMem)
	podLimitMem, err := getPodLimitMem()
	if err != nil {
		log.Errorw("failed to get pod limit memory", "error", err)
	}
	log.Infow("print limit memory", "availableMem", availableMem, "podLimitMem", podLimitMem)
	limits := &BaseLimit{}
	limits.Memory = int64(float64(availableMem) * LimitFactor)
	// TODO:: get from os and compatible with a variety of os
	limits.FD = math.MaxInt
	limits.Conns = math.MaxInt
	limits.ConnsInbound = math.MaxInt
	limits.ConnsOutbound = math.MaxInt
	return limits
}

const memLimitInBytesPath = "/sys/fs/cgroup/memory/memory.limit_in_bytes"

func getPodLimitMem() (uint64, error) {
	log.Info("enter getPodLimitMem")
	data, err := ReadLines(memLimitInBytesPath)
	if err != nil {
		log.Errorw("failed to open memory limit in bytes file", "error", err)
		return 0, err
	}

	if len(data) != 1 {
		log.Errorw("wrong limit pod memory data", "file path", memLimitInBytesPath)
	}
	limitMem, err := util.StringToUint64(data[0])
	if err != nil {
		log.Errorw("failed to convert string to uint64", "error", err)
		return 0, err
	}
	log.Infow("end", "limitMem", limitMem)
	return limitMem, nil
}

// ReadLines reads contents from a file and splits them by new lines.
// A convenience wrapper to ReadLinesOffsetN(filename, 0, -1).
func ReadLines(filename string) ([]string, error) {
	return ReadLinesOffsetN(filename, 0, -1)
}

// ReadLinesOffsetN reads contents from file and splits them by new line.
// The offset tells at which line number to start.
// The count determines the number of lines to read (starting from offset):
// n >= 0: at most n lines
// n < 0: whole file
func ReadLinesOffsetN(filename string, offset uint, n int) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return []string{""}, err
	}
	defer f.Close()

	var ret []string

	r := bufio.NewReader(f)
	for i := 0; i < n+int(offset) || n < 0; i++ {
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF && len(line) > 0 {
				ret = append(ret, strings.Trim(line, "\n"))
			}
			break
		}
		if i < int(offset) {
			continue
		}
		ret = append(ret, strings.Trim(line, "\n"))
	}

	return ret, nil
}
