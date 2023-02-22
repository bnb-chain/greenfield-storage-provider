package libs

import (
	"testing"

	"github.com/stretchr/testify/assert"

	p2p "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
)

func TestValueToMetricsLabel(t *testing.T) {
	lc := newMetricsLabelCache()
	r := &p2p.PexResponse{}
	str := lc.ValueToMetricLabel(r)
	assert.Equal(t, "v1_PexResponse", str)

	// subsequent calls to the function should produce the same result
	str = lc.ValueToMetricLabel(r)
	assert.Equal(t, "v1_PexResponse", str)
}
