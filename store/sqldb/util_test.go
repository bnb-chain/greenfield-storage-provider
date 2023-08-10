package sqldb

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetCurrentYearMonth(t *testing.T) {
	result := GetCurrentYearMonth()
	assert.NotEmpty(t, result)
}

func TestGetCurrentUnixTime(t *testing.T) {
	result := GetCurrentUnixTime()
	assert.NotEmpty(t, result)
}

func TestGetCurrentTimestampUs(t *testing.T) {
	result := GetCurrentTimestampUs()
	assert.NotEmpty(t, result)
}

func TestTimestampUsToTime(t *testing.T) {
	result := TimestampUsToTime(1691565230556069)
	assert.Equal(t, result.Year(), 2023)
}

func TestTimestampSecToTime(t *testing.T) {
	result := TimestampSecToTime(1691565230)
	assert.Equal(t, result.Year(), 2023)
}

func TestTimeToYearMonth(t *testing.T) {
	result := TimeToYearMonth(time.Date(2023, time.August, 9, 15, 13, 50, 556069000, time.Local))
	assert.Equal(t, "2023-08", result)
}

func Test_isAlreadyExists(t *testing.T) {
	err := errors.New("Error 1050 (42S01)")
	result := isAlreadyExists(err)
	assert.Equal(t, true, result)
}
