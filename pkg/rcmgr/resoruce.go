package rcmgr

import (
	"fmt"
	"math"
	"math/big"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

// resources tracks the current state of resource consumption
type resources struct {
	limit     Limit
	nconnsIn  int
	nconnsOut int
	nfd       int
	memory    int64
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
		}
		threshold = thresholdBig.Int64()
	} else {
		threshold = threshold / 256
	}

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

func (rc *resources) addConn(dir Direction) error {
	if dir == DirInbound {
		return rc.addConns(1, 0, 1)
	}
	return rc.addConns(0, 1, 1)
}

func (rc *resources) addConns(incount, outcount, fdcount int) error {
	if incount > 0 {
		limit := rc.limit.GetConnLimit(DirInbound)
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
		limit := rc.limit.GetConnLimit(DirOutbound)
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

func (rc *resources) removeConn(dir Direction) {
	if dir == DirInbound {
		rc.removeConns(1, 0, 1)
	}
	rc.removeConns(0, 1, 1)
}

func (rc *resources) removeConns(incount, outcount, fdcount int) {
	rc.nconnsIn -= incount
	rc.nconnsOut -= outcount
	rc.nfd -= fdcount

	if rc.nconnsIn < 0 {
		log.Warn("BUG: too many inbound connections released")
		rc.nconnsIn = 0
	}
	if rc.nconnsOut < 0 {
		log.Warn("BUG: too many outbound connections released")
		rc.nconnsOut = 0
	}
	if rc.nfd < 0 {
		log.Warn("BUG: too many file descriptors released")
		rc.nfd = 0
	}
}

func (rc *resources) stat() ScopeStat {
	return ScopeStat{
		Memory:           rc.memory,
		NumConnsInbound:  rc.nconnsIn,
		NumConnsOutbound: rc.nconnsOut,
		NumFD:            rc.nfd,
	}
}
