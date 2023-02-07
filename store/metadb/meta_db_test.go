package metadb

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
)

func Test_IntegrityMeta_Json_Marshal_Unmarshal(t *testing.T) {
	meta := &IntegrityMeta{
		ObjectID:       1,
		PieceIdx:       1,
		RedundancyType: ptypes.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE,
	}
	data, err := json.Marshal(meta)
	assert.Equal(t, nil, err)
	fmt.Println(string(data))

	var meta1 IntegrityMeta
	err = json.Unmarshal(data, &meta1)
	assert.Equal(t, nil, err)
	assert.Equal(t, meta.ObjectID, meta1.ObjectID)
	assert.Equal(t, meta.RedundancyType, meta1.RedundancyType)
}
