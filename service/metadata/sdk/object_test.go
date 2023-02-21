package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/mock"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata/model"
	"github.com/bnb-chain/greenfield-storage-provider/util/https"
	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestClient_ListObjectsByBucketName(t *testing.T) {
	type args struct {
		ctx        context.Context
		userID     uuid.UUID
		bucketName string
	}
	type Body struct {
		client  Client
		args    args
		want    []model.Object
		wantErr bool
	}
	tests := []struct {
		name string
		f    func(*testing.T, *gomock.Controller) *Body
	}{
		{
			"case 1/success",
			func(t *testing.T, c *gomock.Controller) *Body {
				args := args{
					ctx:        context.Background(),
					userID:     uuid.Must(uuid.NewV4()),
					bucketName: "test",
				}
				objects := []model.Object{
					model.Object{
						Owner:                "46765cbc-d30c-4f4a-a814-b68181fcab12",
						BucketName:           "test",
						ObjectName:           "test-object",
						Id:                   "1000",
						PayloadSize:          100,
						IsPublic:             false,
						ContentType:          "video",
						CreateAt:             0,
						ObjectStatus:         "OBJECT_STATUS_INIT",
						RedundancyType:       "REDUNDANCY_REPLICA_TYPE",
						SourceType:           "SOURCE_TYPE_ORIGIN",
						Checksums:            nil,
						SecondarySpAddresses: nil,
						LockedBalance:        "1000",
					},
				}
				mockHTTPCli := mock.NewMockIHTTPClient(c)
				mockHTTPCli.EXPECT().Do(gomock.Any()).DoAndReturn(func(req *http.Request) (*http.Response, error) {
					require.True(t, strings.Contains(req.URL.Path, args.userID.String()))
					hresp := https.Response{
						Data: objects,
					}
					bts, _ := json.Marshal(hresp)
					buffer := bytes.NewBuffer(bts)
					resp := &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(buffer),
					}
					return resp, nil
				}).Times(1)
				return &Body{
					client: Client{
						name:    "ut",
						httpCli: mockHTTPCli,
						option: &Option{
							HTTPTimeout: time.Second,
						},
					},
					args: args,
					want: []model.Object{
						{
							Owner:                objects[0].Owner,
							BucketName:           objects[0].BucketName,
							ObjectName:           objects[0].ObjectName,
							Id:                   objects[0].Id,
							PayloadSize:          objects[0].PayloadSize,
							IsPublic:             objects[0].IsPublic,
							ContentType:          objects[0].ContentType,
							CreateAt:             objects[0].CreateAt,
							ObjectStatus:         objects[0].ObjectStatus,
							RedundancyType:       objects[0].RedundancyType,
							SourceType:           objects[0].SourceType,
							Checksums:            objects[0].Checksums,
							SecondarySpAddresses: objects[0].SecondarySpAddresses,
							LockedBalance:        objects[0].LockedBalance,
						},
					},
					wantErr: false,
				}
			},
		},
	}
	for _, _tt := range tests {
		t.Run(_tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			tt := _tt.f(t, ctrl)
			c := &tt.client
			_, err := c.ListObjectsByBucketName(tt.args.ctx, tt.args.userID, tt.args.bucketName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetApiPackageByIndexName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
