package gfsprcmgr

import (
	"fmt"
	"math"
	"math/big"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsplimit"
	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

// resources tracks the current state of resource consumption
type resources struct {
	limit        corercmgr.Limit
	nconnsIn     int
	nconnsOut    int
	nfd          int
	ntasksHigh   int
	ntasksMedium int
	ntasksLow    int
	memory       int64
}

func (rc *resources) remaining() corercmgr.Limit {
	l := &gfsplimit.GfSpLimit{}
	if rc.limit.GetMemoryLimit() > rc.memory {
		l.Memory = rc.limit.GetMemoryLimit() - rc.memory
	} else {
		l.Memory = 0
	}
	if rc.limit.GetTaskTotalLimit() > rc.ntasksHigh+rc.ntasksHigh+rc.ntasksHigh {
		l.Tasks = int32(rc.limit.GetTaskTotalLimit() - (rc.ntasksHigh + rc.ntasksHigh + rc.ntasksHigh))
	} else {
		l.Tasks = 0
	}
	if rc.limit.GetTaskLimit(corercmgr.ReserveTaskPriorityHigh) > rc.ntasksLow {
		l.TasksHighPriority = int32(rc.limit.GetTaskLimit(corercmgr.ReserveTaskPriorityHigh) - rc.ntasksLow)
	} else {
		l.TasksHighPriority = 0
	}
	if rc.limit.GetTaskLimit(corercmgr.ReserveTaskPriorityMedium) > rc.ntasksLow {
		l.TasksMediumPriority = int32(rc.limit.GetTaskLimit(corercmgr.ReserveTaskPriorityMedium) - rc.ntasksLow)
	} else {
		l.TasksMediumPriority = 0
	}
	if rc.limit.GetTaskLimit(corercmgr.ReserveTaskPriorityLow) > rc.ntasksLow {
		l.TasksLowPriority = int32(rc.limit.GetTaskLimit(corercmgr.ReserveTaskPriorityLow) - rc.ntasksLow)
	} else {
		l.TasksLowPriority = 0
	}
	if rc.limit.GetTaskLimit(corercmgr.ReserveTaskPriorityLow) > rc.ntasksLow {
		l.TasksLowPriority = int32(rc.limit.GetTaskLimit(corercmgr.ReserveTaskPriorityLow) - rc.ntasksLow)
	} else {
		l.TasksLowPriority = 0
	}
	return l
}

func addInt64WithOverflow(a int64, b int64) (c int64, ok bool) {
	c = a + b
	return c, (c > a) == (b > 0)
}

// mulInt64WithOverflow checks for overflow in multiplying two int64s. See
// https://groups.google.com/g/golang-nuts/c/h5oSN5t3Au4/m/KaNQREhZh0QJ
func mulInt64WithOverflow(a, b int64) (c int64, ok bool) {
	const mostPositive = 1<<63 - 1
	const mostNegative = -(mostPositive + 1)
	c = a * b
	if a == 0 || b == 0 || a == 1 || b == 1 {
		return c, true
	}
	if a == mostNegative || b == mostNegative {
		return c, false
	}
	return c, c/b == a
}

func (rc *resources) checkMemory(rsvp int64, prio uint8) error {
	if rsvp < 0 {
		return fmt.Errorf("can't reserve negative memory. rsvp=%v", rsvp)
	}
	limit := rc.limit.GetMemoryLimit()
	if limit == math.MaxInt64 {
		// Special case where we've set max limits.
		return nil
	}
	newmem, addOk := addInt64WithOverflow(rc.memory, rsvp)
	threshold, mulOk := mulInt64WithOverflow(1+int64(prio), limit)
	if !mulOk {
		thresholdBig := big.NewInt(limit)
		thresholdBig = thresholdBig.Mul(thresholdBig, big.NewInt(1+int64(prio)))
		thresholdBig.Rsh(thresholdBig, 8) // Divide 256
		if !thresholdBig.IsInt64() {
			// Shouldn't happen since the threshold can only be <= limit
			threshold = limit
			goto finish
		}
		threshold = thresholdBig.Int64()
	} else {
		threshold = threshold / 256
	}
finish:
	if !addOk || newmem > threshold {
		return &ErrMemoryLimitExceeded{
			current:   rc.memory,
			attempted: rsvp,
			limit:     limit,
			priority:  prio,
			err:       ErrResourceLimitExceeded,
		}
	}
	return nil
}

func (rc *resources) reserveMemory(size int64, prio uint8) error {
	if err := rc.checkMemory(size, prio); err != nil {
		return err
	}
	rc.memory += size
	return nil
}

func (rc *resources) releaseMemory(size int64) {
	rc.memory -= size
	// sanity check for bugs upstream
	if rc.memory < 0 {
		log.Warn("BUG: too much memory released")
		rc.memory = 0
	}
}

func (rc *resources) addTask(num int, prio corercmgr.ReserveTaskPriority) error {
	if prio == corercmgr.ReserveTaskPriorityHigh {
		return rc.addTasks(num, 0, 0)
	} else if prio == corercmgr.ReserveTaskPriorityMedium {
		return rc.addTasks(0, num, 0)
	} else {
		return rc.addTasks(0, 0, num)
	}
}

func (rc *resources) addTasks(high, medium, low int) error {
	if rc.ntasksHigh+rc.ntasksMedium+rc.ntasksLow+high+medium+low > rc.limit.GetTaskTotalLimit() {
		return &ErrTaskLimitExceeded{
			current:   rc.ntasksHigh + rc.ntasksMedium + rc.ntasksLow,
			attempted: high + medium + low,
			limit:     rc.limit.GetTaskTotalLimit(),
			err:       fmt.Errorf("total task limit exceeded: %w", ErrResourceLimitExceeded),
		}
	}
	if high+rc.ntasksHigh > rc.limit.GetTaskLimit(corercmgr.ReserveTaskPriorityHigh) {
		return &ErrTaskLimitExceeded{
			current:   rc.ntasksHigh,
			attempted: high,
			limit:     rc.limit.GetTaskLimit(corercmgr.ReserveTaskPriorityHigh),
			err:       fmt.Errorf("high priority task limit exceeded: %w", ErrResourceLimitExceeded),
		}
	}
	if medium+rc.ntasksMedium > rc.limit.GetTaskLimit(corercmgr.ReserveTaskPriorityMedium) {
		return &ErrTaskLimitExceeded{
			current:   rc.ntasksMedium,
			attempted: medium,
			limit:     rc.limit.GetTaskLimit(corercmgr.ReserveTaskPriorityMedium),
			err:       fmt.Errorf("medium priority task limit exceeded: %w", ErrResourceLimitExceeded),
		}
	}
	if low+rc.ntasksLow > rc.limit.GetTaskLimit(corercmgr.ReserveTaskPriorityLow) {
		return &ErrTaskLimitExceeded{
			current:   rc.ntasksLow,
			attempted: low,
			limit:     rc.limit.GetTaskLimit(corercmgr.ReserveTaskPriorityLow),
			err:       fmt.Errorf("low priority task limit exceeded: %w", ErrResourceLimitExceeded),
		}
	}
	rc.ntasksHigh += high
	rc.ntasksMedium += medium
	rc.ntasksLow += low
	return nil
}

func (rc *resources) removeTask(num int, prio corercmgr.ReserveTaskPriority) {
	if prio == corercmgr.ReserveTaskPriorityHigh {
		rc.removeTasks(num, 0, 0)
	} else if prio == corercmgr.ReserveTaskPriorityMedium {
		rc.removeTasks(0, num, 0)
	} else if prio == corercmgr.ReserveTaskPriorityLow {
		rc.removeTasks(0, 0, num)
	}
}

func (rc *resources) removeTasks(high, medium, low int) {
	rc.ntasksHigh -= high
	rc.ntasksMedium -= medium
	rc.ntasksLow -= low

	if rc.ntasksHigh < 0 {
		log.Error("BUG: too many high priority task released")
		rc.ntasksHigh = 0
	}
	if rc.ntasksMedium < 0 {
		log.Error("BUG: too many medium priority task released")
		rc.ntasksMedium = 0
	}
	if rc.ntasksLow < 0 {
		log.Error("BUG:  too many low priority task released")
		rc.ntasksLow = 0
	}
}

func (rc *resources) addConn(dir corercmgr.Direction) error {
	if dir == corercmgr.DirInbound {
		return rc.addConns(1, 0, 1)
	}
	return rc.addConns(0, 1, 1)
}

func (rc *resources) addConns(incount, outcount, fdcount int) error {
	if incount > 0 {
		limit := rc.limit.GetConnLimit(corercmgr.DirInbound)
		if rc.nconnsIn+incount > limit {
			return &ErrConnLimitExceeded{
				current:   rc.nconnsIn,
				attempted: incount,
				limit:     limit,
				err:       fmt.Errorf("cannot reserve inbound connection: %w", ErrResourceLimitExceeded),
			}
		}
	}
	if outcount > 0 {
		limit := rc.limit.GetConnLimit(corercmgr.DirOutbound)
		if rc.nconnsOut+outcount > limit {
			return &ErrConnLimitExceeded{
				current:   rc.nconnsOut,
				attempted: outcount,
				limit:     limit,
				err:       fmt.Errorf("cannot reserve outbound connection: %w", ErrResourceLimitExceeded),
			}
		}
	}
	if connLimit := rc.limit.GetConnTotalLimit(); rc.nconnsIn+incount+rc.nconnsOut+outcount > connLimit {
		return &ErrConnLimitExceeded{
			current:   rc.nconnsIn + rc.nconnsOut,
			attempted: incount + outcount,
			limit:     connLimit,
			err:       fmt.Errorf("cannot reserve connection: %w", ErrResourceLimitExceeded),
		}
	}
	if fdcount > 0 {
		limit := rc.limit.GetFDLimit()
		if rc.nfd+fdcount > limit {
			return &ErrConnLimitExceeded{
				current:   rc.nfd,
				attempted: fdcount,
				limit:     limit,
				err:       fmt.Errorf("cannot reserve file descriptor: %w", ErrResourceLimitExceeded),
			}
		}
	}

	rc.nconnsIn += incount
	rc.nconnsOut += outcount
	rc.nfd += fdcount
	return nil
}

func (rc *resources) removeConn(dir corercmgr.Direction) {
	if dir == corercmgr.DirInbound {
		rc.removeConns(1, 0, 1)
	}
	rc.removeConns(0, 1, 1)
}

func (rc *resources) removeConns(incount, outcount, fdcount int) {
	rc.nconnsIn -= incount
	rc.nconnsOut -= outcount
	rc.nfd -= fdcount

	if rc.nconnsIn < 0 {
		log.Error("BUG: too many inbound connections released")
		rc.nconnsIn = 0
	}
	if rc.nconnsOut < 0 {
		log.Error("BUG: too many outbound connections released")
		rc.nconnsOut = 0
	}
	if rc.nfd < 0 {
		log.Error("BUG: too many file descriptors released")
		rc.nfd = 0
	}
}

func (rc *resources) stat() corercmgr.ScopeStat {
	return corercmgr.ScopeStat{
		Memory:           rc.memory,
		NumTasksHigh:     int64(rc.ntasksHigh),
		NumTasksMedium:   int64(rc.ntasksMedium),
		NumTasksLow:      int64(rc.ntasksLow),
		NumConnsInbound:  int64(rc.nconnsIn),
		NumConnsOutbound: int64(rc.nconnsOut),
		NumFD:            int64(rc.nfd),
	}
}
