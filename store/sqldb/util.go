package sqldb

import (
	"time"
)

// GetCurrentYearMonth get current year and month
func GetCurrentYearMonth() string {
	return TimeToYearMonth(time.Now())
}

// GetCurrentUnixTime return a second timestamp
func GetCurrentUnixTime() int64 {
	return time.Now().Unix()
}

// GetCurrentTimestampUs return a microsecond timestamp
func GetCurrentTimestampUs() int64 {
	return time.Now().UnixMicro()
}

// TimestampUsToTime convert a microsecond timestamp to time.Time
func TimestampUsToTime(ts int64) time.Time {
	tUnix := ts / int64(time.Millisecond)
	tUnixNanoRemainder := (ts % int64(time.Millisecond)) * int64(time.Microsecond)
	return time.Unix(tUnix, tUnixNanoRemainder)
}

// TimestampSecToTime convert a second timestamp to time.Time
func TimestampSecToTime(timeUnix int64) time.Time {
	return time.Unix(timeUnix, 0)
}

// TimeToYearMonth convent time.Time to YYYY-MM string
func TimeToYearMonth(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")[0:7]
}
