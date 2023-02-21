package router

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/bnb-chain/greenfield-storage-provider/mock"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata/model"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata/store"
	"github.com/bnb-chain/greenfield-storage-provider/util/https"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
)

func TestRouter_GetUserBuckets(t *testing.T) {
	type fields struct {
		store store.IStore
	}
	type args struct {
		c     *gin.Context
		KeyId string
	}
	type Body struct {
		fields   fields
		args     args
		wantResp interface{}
		wantHerr *https.Error
	}
	tests := []struct {
		name string
		f    func(*testing.T, *gomock.Controller) *Body
	}{
		{
			name: "case 1/GetUserBuckets by uid",
			f: func(t *testing.T, c *gomock.Controller) *Body {
				uid := uuid.Must(uuid.NewV4())
				mockStore := mock.NewMockIStore(c)
				bucket1 := model.Bucket{
					Owner:            "dfldscbc-9skf-xisk-o192-doxl01fcabid",
					BucketName:       "Test",
					IsPublic:         true,
					Id:               "1",
					SourceType:       "SOURCE_TYPE_BSC_CROSS_CHAIN",
					CreateAt:         1676530547,
					PaymentAddress:   "0x000000006b4BD0274e8f943201A922295D13fc28",
					PrimarySpAddress: "0x000000006b4BD0274e8f943201A922295D13fc28",
					ReadQuota:        "1000",
					PaymentPriceTime: 0,
					PaymentOutFlows:  nil,
				}
				buckets := []*model.Bucket{
					&bucket1,
				}
				v := url.Values{}
				v.Set("bucketName", uid.String())

				ctx := &gin.Context{
					Request: &http.Request{
						URL: &url.URL{RawQuery: v.Encode()},
					},
					Params: gin.Params{
						gin.Param{
							Key:   "bucket_name",
							Value: uid.String(),
						},
					},
				}

				mockStore.EXPECT().GetUserBuckets(ctx).Return(buckets, nil).Times(1)

				return &Body{
					fields: fields{
						store: mockStore,
					},
					args: args{
						c: ctx,
					},
					wantResp: buckets,
					wantHerr: nil,
				}
			},
		},
	}
	for _, _tt := range tests {
		t.Run(_tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			tt := _tt.f(t, ctrl)
			r := &Router{
				store: tt.fields.store,
			}
			gotResp, gotHerr := r.GetUserBuckets(tt.args.c)
			if !reflect.DeepEqual(gotResp, tt.wantResp) {
				t.Errorf("Router.GetUserBuckets() gotResp = %v, want %v", gotResp, tt.wantResp)
			}
			if !reflect.DeepEqual(gotHerr, tt.wantHerr) {
				t.Errorf("Router.GetUserBuckets() gotHerr = %v, want %v", gotHerr, tt.wantHerr)
			}
		})
	}
}
