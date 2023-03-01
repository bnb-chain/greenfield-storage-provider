package router

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/bnb-chain/greenfield-storage-provider/service/metadata/mock"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata/model"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata/store"
	"github.com/bnb-chain/greenfield-storage-provider/util/https"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
)

func TestRouter_ListObjectsByBucketName(t *testing.T) {
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
			name: "case 1/ListObjectsByBucketName by bucketName",
			f: func(t *testing.T, c *gomock.Controller) *Body {
				bucketName := "Test"
				mockStore := mock.NewMockIStore(c)
				object1 := model.Object{
					Owner:                "63e6aefe-0df6-48cf-8e84-031bc2a9169d\n\n",
					BucketName:           "Test",
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
				}
				objects := []*model.Object{
					&object1,
				}

				ctx := &gin.Context{
					Request: &http.Request{
						URL: &url.URL{},
					},
					Params: gin.Params{},
				}

				mockStore.EXPECT().ListObjectsByBucketName(ctx, bucketName).Return(objects, nil).Times(1)

				return &Body{
					fields: fields{
						store: mockStore,
					},
					args: args{
						c: ctx,
					},
					wantResp: objects,
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
			gotResp, gotHerr := r.ListObjectsByBucketName(tt.args.c, "Test")
			if !reflect.DeepEqual(gotResp, tt.wantResp) {
				t.Errorf("Router.ListObjectsByBucketName() gotResp = %v, want %v", gotResp, tt.wantResp)
			}
			if !reflect.DeepEqual(gotHerr, tt.wantHerr) {
				t.Errorf("Router.ListObjectsByBucketName() gotHerr = %v, want %v", gotHerr, tt.wantHerr)
			}
		})
	}
}
