package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
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

func TestClient_GetUserBuckets(t *testing.T) {
	type args struct {
		ctx    context.Context
		userID uuid.UUID
	}
	type Body struct {
		client  Client
		args    args
		want    []model.Bucket
		wantErr bool
	}
	tests := []struct {
		name string
		f    func(*testing.T, *gomock.Controller) *Body
	}{
		{
			"case 1/success",
			func(t *testing.T, c *gomock.Controller) *Body {
				user_id := uuid.Must(uuid.NewV4())
				args := args{
					ctx:    context.Background(),
					userID: user_id,
				}
				buckets := []model.Bucket{
					model.Bucket{
						Owner:            "dfldscbc-9skf-xisk-o192-doxl01fcabid",
						BucketName:       "Test",
						IsPublic:         true,
						Id:               "1",
						SourceType:       "SOURCE_TYPE_BSC_CROSS_CHAIN",
						PaymentAddress:   "0x000000006b4BD0274e8f943201A922295D13fc28",
						PrimarySpAddress: "0x000000006b4BD0274e8f943201A922295D13fc28",
						ReadQuota:        "1000",
					},
				}
				mockHTTPCli := mock.NewMockIHTTPClient(c)
				mockHTTPCli.EXPECT().Do(gomock.Any()).DoAndReturn(func(req *http.Request) (*http.Response, error) {
					require.True(t, strings.Contains(req.URL.Path, args.userID.String()))
					hresp := https.Response{
						Data: buckets,
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
					want: []model.Bucket{
						{
							Owner:            buckets[0].Owner,
							BucketName:       buckets[0].BucketName,
							IsPublic:         buckets[0].IsPublic,
							Id:               buckets[0].Id,
							SourceType:       buckets[0].SourceType,
							PaymentAddress:   buckets[0].PaymentAddress,
							PrimarySpAddress: buckets[0].PrimarySpAddress,
							ReadQuota:        buckets[0].ReadQuota,
						},
					},
					wantErr: false,
				}
			},
		},
		{
			"case 2/empty",
			func(t *testing.T, c *gomock.Controller) *Body {
				args := args{
					ctx:    context.Background(),
					userID: uuid.Must(uuid.NewV4()),
				}
				var buckets []model.Bucket
				mockHTTPCli := mock.NewMockIHTTPClient(c)
				mockHTTPCli.EXPECT().Do(gomock.Any()).DoAndReturn(func(req *http.Request) (*http.Response, error) {
					require.True(t, strings.Contains(req.URL.Path, args.userID.String()))
					hresp := https.Response{
						Data: buckets,
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
					args:    args,
					want:    nil,
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
			gotApp, err := c.GetUserBuckets(tt.args.ctx, tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetUserBuckets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotApp, tt.want) {
				t.Errorf("Client.GetUserBuckets() gotApp = %v, want %v", gotApp, tt.want)
				return
			}
		})
	}
}
