package sdk

import (
	"context"
	"fmt"

	"github.com/bnb-chain/greenfield-storage-provider/service/metadata/model"
	"github.com/gofrs/uuid"
)

func (c *Client) ListObjectsByBucketName(ctx context.Context, userID uuid.UUID, bucketName string) (ret []*model.Object, err error) {
	url := fmt.Sprintf("http://%s/account/%s/buckets/%s/objects", "localhost:9733", userID, bucketName)
	err = c.get(ctx, url, &ret, withCustomMetricsHandlerName("ListObjectsByBucketName"))
	if err != nil {
		return nil, err
	}

	return ret, nil
}
