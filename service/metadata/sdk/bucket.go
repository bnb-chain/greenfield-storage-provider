package sdk

import (
	"context"
	"fmt"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata/model"
	"github.com/gofrs/uuid"
)

func (c *Client) GetUserBuckets(ctx context.Context, userID uuid.UUID) (ret []model.Bucket, err error) {
	url := fmt.Sprintf("http://%s/accounts/%s/buckets", "localhost:9733", userID)
	err = c.get(ctx, url, &ret, withCustomMetricsHandlerName("GetUserBuckets"))
	if err != nil {
		return nil, err
	}

	return ret, nil
}
