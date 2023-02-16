package sdk

import (
	"context"
	"fmt"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata/model"
)

func (c *Client) ListObjectsByBucketName(ctx context.Context, bucketName string) (ret []*model.Object, err error) {
	url := fmt.Sprintf("http://%s/buckets/%s/objects", "localhost:9733", bucketName)
	err = c.get(ctx, url, &ret, withCustomMetricsHandlerName("ListObjectsByBucketName"))
	if err != nil {
		return nil, err
	}

	return ret, nil
}
