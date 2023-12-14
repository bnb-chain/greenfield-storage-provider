package gater

import (
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	payment_types "github.com/bnb-chain/greenfield/x/payment/types"
	storage_types "github.com/bnb-chain/greenfield/x/storage/types"
	virtual_types "github.com/bnb-chain/greenfield/x/virtualgroup/types"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
)

const (
	testAccount = "0xF72aDa8130f934887755492879496b026665FbAB"
)

func mockListObjectsByBucketNameRoute(t *testing.T, g *GateModular) *mux.Router {
	t.Helper()
	router := mux.NewRouter().SkipClean(true)
	var routers []*mux.Router
	routers = append(routers, router.Host("{bucket:.+}."+g.domain).Subrouter())
	routers = append(routers, router.PathPrefix("/{bucket}").Subrouter())
	for _, r := range routers {
		r.NewRoute().Name(listObjectsByBucketRouterName).Methods(http.MethodGet).Path("/").HandlerFunc(g.listObjectsByBucketNameHandler)

	}
	return router
}

func getSampleChecksum() [][]byte {
	checksumsInBase64 := [7]string{"tPsLBcgLxRVKTRJCeYw5FVj0jjqPsqFnbDCr77pf7RA=",
		"7YqCbwK/qC+zaAoJvd971fuJCE0OVQ9ky8bgomUkmRI=",
		"i59qS3vgvN8QIcNKOJggtN4JsZRLYt1ugeGDtP6x7Sk=",
		"tBBu4BPpANbc12SO5TVeQ64DtKwl0F2inE29H9jAw54=",
		"vOw+loeUIXXPEvfYNFmnElTIxj/b0dEEBBF1YbKOoEI=",
		"e0nSN4a5u3EDPaAqemGDZ5gYJ0l6NUjtalmj/BH2uWE=",
		"rRm6iKPMc8gZbw1WKKF2kPXveU2VFEh2izs9e8ovfwk="}
	checksums := make([][]byte, len(checksumsInBase64))
	for j := 0; j < len(checksums); j++ {
		checksums[j], _ = base64.StdEncoding.DecodeString(checksumsInBase64[j])
	}
	return checksums

}
func getTestGroupsInIdMap(length int) map[uint64]*types.Group {
	groupsMap := make(map[uint64]*types.Group)
	groups := getTestGroupsResponse(length)
	for _, g := range groups {
		groupsMap[g.Group.Id.BigInt().Uint64()] = g
	}
	return groupsMap
}

func getTestBucketsInIdMap(length int) map[uint64]*types.Bucket {
	bucketsMap := make(map[uint64]*types.Bucket)
	buckets := getTestBuckets(length)
	for _, b := range buckets {
		bucketsMap[b.BucketInfo.Id.BigInt().Uint64()] = b
	}
	return bucketsMap
}

func getTestBuckets(length int) []*types.Bucket {
	buckets := make([]*types.Bucket, length)

	for i := 0; i < length; i++ {
		bucket := &types.Bucket{
			BucketInfo: &storage_types.BucketInfo{
				Owner:                      testAccount,
				BucketName:                 mockBucketName + strconv.Itoa(i),
				Visibility:                 storage_types.VISIBILITY_TYPE_PUBLIC_READ,
				Id:                         math.NewUintFromBigInt(big.NewInt(int64(i))),
				SourceType:                 storage_types.SOURCE_TYPE_ORIGIN,
				CreateAt:                   1699781080,
				PaymentAddress:             testAccount,
				GlobalVirtualGroupFamilyId: 3,
				ChargedReadQuota:           0,
				BucketStatus:               storage_types.BUCKET_STATUS_CREATED,
			},
			Removed:      false,
			DeleteAt:     0,
			DeleteReason: "",
			Operator:     testAccount,
			CreateTxHash: "0x21c349a869bde1f44378936e2a9a15ed3fb2d54a43eaea8787960bba1134cdc2",
			UpdateTxHash: "0x0cbff0ff3831d61345dbfda5b984e254c4bf87ecf80b45ccbb0635c0547a3b1a",
			UpdateAt:     1279811,
			UpdateTime:   1699781103,
		}
		buckets[i] = bucket
	}
	return buckets
}

func getTestVGFInfoBuckets(length int) []*types.VGFInfoBucket {
	buckets := make([]*types.VGFInfoBucket, length)

	for i := 0; i < length; i++ {
		bucket := &types.VGFInfoBucket{
			BucketInfo: &storage_types.BucketInfo{
				Owner:                      testAccount,
				BucketName:                 mockBucketName + strconv.Itoa(i),
				Visibility:                 storage_types.VISIBILITY_TYPE_PUBLIC_READ,
				Id:                         math.NewUintFromBigInt(big.NewInt(int64(i))),
				SourceType:                 storage_types.SOURCE_TYPE_ORIGIN,
				CreateAt:                   1699781080,
				PaymentAddress:             testAccount,
				GlobalVirtualGroupFamilyId: 3,
				ChargedReadQuota:           0,
				BucketStatus:               storage_types.BUCKET_STATUS_CREATED,
			},
			Removed:      false,
			DeleteAt:     0,
			DeleteReason: "",
			Operator:     testAccount,
			CreateTxHash: "0x21c349a869bde1f44378936e2a9a15ed3fb2d54a43eaea8787960bba1134cdc2",
			UpdateTxHash: "0x0cbff0ff3831d61345dbfda5b984e254c4bf87ecf80b45ccbb0635c0547a3b1a",
			UpdateAt:     1279811,
			UpdateTime:   1699781103,
			Vgf: &virtual_types.GlobalVirtualGroupFamily{
				Id:                    uint32(i),
				PrimarySpId:           uint32(3),
				VirtualPaymentAddress: "0x26281179b8885F21f95b0a246c8AD70957A95A23",
			},
		}
		buckets[i] = bucket
	}
	return buckets
}

func getTestStreamRecord() *payment_types.StreamRecord {
	return &payment_types.StreamRecord{
		Account:           testAccount,
		CrudTimestamp:     1699780994,
		NetflowRate:       math.NewIntFromBigInt(big.NewInt(int64(0))),
		StaticBalance:     math.NewIntFromBigInt(big.NewInt(int64(240000000000000001))),
		BufferBalance:     math.NewIntFromBigInt(big.NewInt(int64(0))),
		LockBalance:       math.NewIntFromBigInt(big.NewInt(int64(0))),
		Status:            payment_types.STREAM_ACCOUNT_STATUS_ACTIVE,
		SettleTimestamp:   0,
		OutFlowCount:      0,
		FrozenNetflowRate: math.NewIntFromBigInt(big.NewInt(int64(0))),
	}
}

func getTestPaymentAccountMeta() []*types.PaymentAccountMeta {
	paymentAccounts := make([]*types.PaymentAccountMeta, 1)

	paymentAccount := &types.PaymentAccountMeta{
		StreamRecord: getTestStreamRecord(),
		PaymentAccount: &types.PaymentAccount{
			Address:    testAccount,
			Owner:      testAccount,
			Refundable: true,
			UpdateAt:   1279659,
			UpdateTime: 1699780707,
		},
	}
	paymentAccounts[0] = paymentAccount
	return paymentAccounts
}

func getTestObjectsInIdMap(length int) map[uint64]*types.Object {
	objectsMap := make(map[uint64]*types.Object)
	objects := getTestObjectsResponse(length)
	for _, o := range objects {
		objectsMap[o.ObjectInfo.Id.BigInt().Uint64()] = o
	}
	return objectsMap
}

func getOneTestObjectResponse() *types.Object {
	owner := testAccount
	object := &types.Object{
		ObjectInfo: &storage_types.ObjectInfo{
			Owner:               owner,
			Creator:             owner,
			BucketName:          mockBucketName,
			ObjectName:          mockObjectName,
			Id:                  math.NewUintFromString("24662"),
			LocalVirtualGroupId: 1,
			PayloadSize:         4802764,
			Visibility:          storage_types.VISIBILITY_TYPE_INHERIT,
			ContentType:         "application/octet-stream",
			CreateAt:            1699781700,
			ObjectStatus:        storage_types.OBJECT_STATUS_SEALED,
			Checksums:           getSampleChecksum(),
		},

		LockedBalance: "0x0000000000000000000000000000000000000000000000000000000000000000",
		UpdateAt:      1280048,
		Operator:      "0x03AbbEe8E426C9887A8ae3C34602AbCA42aeDFa0",
		CreateTxHash:  "0x491227c644bc89f5a058d92167c00d452c63a1dd8d5776c81617a41ec76fcc8c",
		UpdateTxHash:  "0x238737f109a40c675e1bef5ebfb2adef2cac0a723ee20fbd752e78efbf3d579e",
		SealTxHash:    "0x238737f109a40c675e1bef5ebfb2adef2cac0a723ee20fbd752e78efbf3d579e",
	}
	return object
}

func getTestObjectsResponse(objectLength int) []*types.Object {
	length := objectLength
	objects := make([]*types.Object, length)
	owner := testAccount

	for i := 0; i < length; i++ {
		object := &types.Object{
			ObjectInfo: &storage_types.ObjectInfo{
				Owner:               owner,
				Creator:             owner,
				BucketName:          mockBucketName,
				ObjectName:          mockObjectName + strconv.Itoa(i),
				Id:                  math.NewUintFromBigInt(big.NewInt(int64(i))),
				LocalVirtualGroupId: 1,
				PayloadSize:         4802764,
				Visibility:          storage_types.VISIBILITY_TYPE_INHERIT,
				ContentType:         "application/octet-stream",
				CreateAt:            1699781700,
				ObjectStatus:        storage_types.OBJECT_STATUS_SEALED,
				Checksums:           getSampleChecksum(),
			},

			LockedBalance: "0x0000000000000000000000000000000000000000000000000000000000000000",
			UpdateAt:      1280048,
			Operator:      "0x03AbbEe8E426C9887A8ae3C34602AbCA42aeDFa0",
			CreateTxHash:  "0x491227c644bc89f5a058d92167c00d452c63a1dd8d5776c81617a41ec76fcc8c",
			UpdateTxHash:  "0x238737f109a40c675e1bef5ebfb2adef2cac0a723ee20fbd752e78efbf3d579e",
			SealTxHash:    "0x238737f109a40c675e1bef5ebfb2adef2cac0a723ee20fbd752e78efbf3d579e",
		}
		objects[i] = object
	}
	return objects
}

func getTestGroupsResponse(groupLength int) []*types.Group {
	length := groupLength
	groups := make([]*types.Group, length)
	owner := testAccount

	for i := 0; i < length; i++ {
		group := &types.Group{
			Group: &storage_types.GroupInfo{
				Owner:      owner,
				GroupName:  "TestGroupName " + strconv.Itoa(i),
				SourceType: storage_types.SOURCE_TYPE_ORIGIN,
				Id:         math.NewUintFromBigInt(big.NewInt(int64(i))),
				Extra:      "",
			},
			NumberOfMembers: 1,
			Removed:         false,
			UpdateAt:        1280048,
			Operator:        "0x03AbbEe8E426C9887A8ae3C34602AbCA42aeDFa0",
		}
		groups[i] = group
	}
	return groups
}

func getTestGroupMembersResponse(groupLength int) []*types.GroupMember {
	length := groupLength
	groupMembers := make([]*types.GroupMember, length)
	for i := 0; i < length; i++ {
		groupMember := &types.GroupMember{
			Group: &storage_types.GroupInfo{
				Owner:      testAccount,
				GroupName:  "TestGroupName " + strconv.Itoa(i),
				SourceType: storage_types.SOURCE_TYPE_ORIGIN,
				Id:         math.NewUintFromBigInt(big.NewInt(int64(i))),
				Extra:      "",
			},
			AccountId:      testAccount,
			Operator:       testAccount,
			CreateAt:       1376086,
			CreateTime:     1700032197,
			UpdateAt:       1376086,
			UpdateTime:     1700032197,
			Removed:        false,
			ExpirationTime: 253402300799,
		}
		groupMembers[i] = groupMember
	}
	return groupMembers
}

func TestGateModular_ListObjectsByBucketNameHandler(t *testing.T) {
	mockData := getTestObjectsResponse(1000)
	cases := []struct {
		name           string
		fn             func() *GateModular
		request        func() *http.Request
		wantedResult   string
		wantedResultFn func(body string) bool
	}{
		{
			name: "new request context error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().ListObjectsByBucketName(gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Return(nil, uint64(0), uint64(0), false, "", "", "", "", nil, "", mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/", scheme, mockBucketName, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "wrong requestDelimiter",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/?max-keys=1000&delimiter=wrong_char&&continuation-token=NjM5LnBuZw==", scheme, mockBucketName, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "wrong BucketName",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/?max-keys=1000&delimiter=%%2F&&continuation-token=NjM5LnBuZw==", scheme, "aa", testDomain) // aa is an invalid bucket name
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "wrong requestMaxKeys",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/?max-keys=invalidMaxKey&delimiter=%%2F&&continuation-token=NjM5LnBuZw==", scheme, mockBucketName, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "wrong requestStartAfter",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/?max-keys=1000&delimiter=%%2F&&continuation-token=NjM5LnBuZw==&&start-after=%%2F%%2F", scheme, mockBucketName, testDomain) // %%2F%%2F means "//", which is an invalid start-after
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "wrong requestContinuationToken",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/?max-keys=1000&delimiter=%%2F&&continuation-token=NjM5LnBuZw==d", scheme, mockBucketName, testDomain) // NjM5LnBuZw==d is an invalid requestContinuationToken
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "wrong requestContinuationToken2",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				invalidContinuationToken := base64.StdEncoding.EncodeToString([]byte("//"))
				path := fmt.Sprintf("%s%s.%s/?max-keys=1000&delimiter=%%2F&&continuation-token=%s", scheme, mockBucketName, testDomain, invalidContinuationToken)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "wrong requestContinuationToken3",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				invalidContinuationToken := base64.StdEncoding.EncodeToString([]byte("not_start_with_prefix"))
				path := fmt.Sprintf("%s%s.%s/?max-keys=1000&delimiter=%%2F&&continuation-token=%s&&prefix=%s", scheme, mockBucketName, testDomain, invalidContinuationToken, "a_sample_prefix")
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "wrong requestIncludeRemoved",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/?max-keys=1000&delimiter=%%2F&&include-removed=invalid", scheme, mockBucketName, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "wrong prefix",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				invalidPrefix := "%2F%2F" // this is an invalid prefix
				path := fmt.Sprintf("%s%s.%s/?max-keys=1000&delimiter=%%2F&&prefix=%s", scheme, mockBucketName, testDomain, invalidPrefix)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},
		// Disable the following json response unit test and will add it back once we add json response support in metadata_handler soon.
		//{
		//	name: "json response",
		//	fn: func() *GateModular {
		//		g := setup(t)
		//		ctrl := gomock.NewController(t)
		//		clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
		//		clientMock.EXPECT().ListObjectsByBucketName(gomock.Any(), gomock.Any(), gomock.Any(),
		//			gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		//		).Return(mockData, uint64(0), uint64(0), false, "", "", "", "", nil, "", nil).Times(1)
		//		g.baseApp.SetGfSpClient(clientMock)
		//		return g
		//	},
		//	request: func() *http.Request {
		//		path := fmt.Sprintf("%s%s.%s/?max-keys=1000&delimiter=%%2F&format=json", scheme, mockBucketName, testDomain)
		//		req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
		//		return req
		//	},
		//	wantedResultFn: func(body string) bool {
		//		var res types.GfSpListObjectsByBucketNameResponse
		//		err := json.Unmarshal([]byte(body), &res)
		//		if err != nil {
		//			return false
		//		}
		//		assert.Equal(t, len(mockData), len(res.Objects))
		//		assert.Equal(t, mockData[0].ObjectInfo.Id, res.Objects[0].ObjectInfo.Id)
		//		return true
		//	},
		//},
		{
			name: "xml response",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().ListObjectsByBucketName(gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				).Return(mockData, uint64(0), uint64(0), false, "", "", "", "", nil, "", nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/?max-keys=1000&delimiter=%%2F", scheme, mockBucketName, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResultFn: func(body string) bool {
				var res types.GfSpListObjectsByBucketNameResponse
				err := xml.Unmarshal([]byte(body), &res)
				if err != nil {
					return false
				}
				assert.Equal(t, len(mockData), len(res.Objects))
				assert.Equal(t, mockData[0].ObjectInfo.Id, res.Objects[0].ObjectInfo.Id)
				return true
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockListObjectsByBucketNameRoute(t, tt.fn())
			w := httptest.NewRecorder()
			begin := time.Now()
			router.ServeHTTP(w, tt.request())
			end := time.Now()
			assert.Less(t, end.UnixMilli()-begin.UnixMilli(), int64(1000)) // we expected this API can return response in 1 sec after it gets data from DB.
			if tt.wantedResult != "" {
				assert.Contains(t, w.Body.String(), tt.wantedResult)
			}
			if tt.wantedResultFn != nil {
				assert.True(t, tt.wantedResultFn(w.Body.String()))
			}
		})
	}
}

func mockGetObjectMetaRoute(t *testing.T, g *GateModular) *mux.Router {
	t.Helper()
	router := mux.NewRouter().SkipClean(true)
	var routers []*mux.Router
	routers = append(routers, router.Host("{bucket:.+}."+g.domain).Subrouter())
	routers = append(routers, router.PathPrefix("/{bucket}").Subrouter())
	for _, r := range routers {
		r.NewRoute().Name(getObjectMetaRouterName).Methods(http.MethodGet).Path("/{object:.+}").HandlerFunc(g.getObjectMetaHandler).
			Queries(GetObjectMetaQuery, "")
	}
	return router
}

func TestGateModular_GetObjectMetaHandler(t *testing.T) {
	cases := []struct {
		name           string
		fn             func() *GateModular
		request        func() *http.Request
		wantedResult   string
		wantedResultFn func(body string) bool
	}{
		{
			name: "new request context error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetObjectMeta(gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s", scheme, mockBucketName, testDomain, mockObjectName, GetObjectMetaQuery)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "invalid bucket name",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetObjectMeta(gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(nil, mockErr).Times(0)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s", scheme, invalidBucketName, testDomain, mockObjectName, GetObjectMetaQuery)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "invalid object name",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetObjectMeta(gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(nil, mockErr).Times(0)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s", scheme, mockBucketName, testDomain, invalidObjectName, GetObjectMetaQuery)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},

		{
			name: "xml response for non existing object",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetObjectMeta(gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(nil, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s", scheme, mockBucketName, testDomain, mockObjectName, GetObjectMetaQuery)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResultFn: func(body string) bool {
				assert.Equal(t, "<GfSpGetObjectMetaResponse></GfSpGetObjectMetaResponse>",
					body)
				return true
			},
		},
		{
			name: "xml response",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetObjectMeta(gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(getOneTestObjectResponse(), nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s", scheme, mockBucketName, testDomain, mockObjectName, GetObjectMetaQuery)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResultFn: func(body string) bool {
				assert.Equal(t, "<GfSpGetObjectMetaResponse><Object><ObjectInfo><Owner>0xF72aDa8130f934887755492879496b026665FbAB</Owner><Creator>0xF72aDa8130f934887755492879496b026665FbAB</Creator><BucketName>mock-bucket-name</BucketName><ObjectName>mock-object-name</ObjectName><Id>24662</Id><LocalVirtualGroupId>1</LocalVirtualGroupId><PayloadSize>4802764</PayloadSize><Visibility>3</Visibility><ContentType>application/octet-stream</ContentType><CreateAt>1699781700</CreateAt><ObjectStatus>1</ObjectStatus><RedundancyType>0</RedundancyType><SourceType>0</SourceType><Checksums>tPsLBcgLxRVKTRJCeYw5FVj0jjqPsqFnbDCr77pf7RA=</Checksums><Checksums>7YqCbwK/qC+zaAoJvd971fuJCE0OVQ9ky8bgomUkmRI=</Checksums><Checksums>i59qS3vgvN8QIcNKOJggtN4JsZRLYt1ugeGDtP6x7Sk=</Checksums><Checksums>tBBu4BPpANbc12SO5TVeQ64DtKwl0F2inE29H9jAw54=</Checksums><Checksums>vOw+loeUIXXPEvfYNFmnElTIxj/b0dEEBBF1YbKOoEI=</Checksums><Checksums>e0nSN4a5u3EDPaAqemGDZ5gYJ0l6NUjtalmj/BH2uWE=</Checksums><Checksums>rRm6iKPMc8gZbw1WKKF2kPXveU2VFEh2izs9e8ovfwk=</Checksums></ObjectInfo><LockedBalance>0x0000000000000000000000000000000000000000000000000000000000000000</LockedBalance><Removed>false</Removed><UpdateAt>1280048</UpdateAt><DeleteAt>0</DeleteAt><DeleteReason></DeleteReason><Operator>0x03AbbEe8E426C9887A8ae3C34602AbCA42aeDFa0</Operator><CreateTxHash>0x491227c644bc89f5a058d92167c00d452c63a1dd8d5776c81617a41ec76fcc8c</CreateTxHash><UpdateTxHash>0x238737f109a40c675e1bef5ebfb2adef2cac0a723ee20fbd752e78efbf3d579e</UpdateTxHash><SealTxHash>0x238737f109a40c675e1bef5ebfb2adef2cac0a723ee20fbd752e78efbf3d579e</SealTxHash></Object></GfSpGetObjectMetaResponse>",
					body)
				return true
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockGetObjectMetaRoute(t, tt.fn())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.request())
			if tt.wantedResult != "" {
				assert.Contains(t, w.Body.String(), tt.wantedResult)
			}
			if tt.wantedResultFn != nil {
				assert.True(t, tt.wantedResultFn(w.Body.String()))
			}
		})
	}
}

func mockListObjectsByIDsHandlerRoute(t *testing.T, g *GateModular) *mux.Router {
	t.Helper()
	router := mux.NewRouter().SkipClean(true)
	router.Path("/").Name(listObjectsByIDsRouterName).Methods(http.MethodGet).Queries(ListObjectsByIDsQuery, "").HandlerFunc(g.listObjectsByIDsHandler)
	return router
}

func TestGateModular_ListObjectsByIDsHandler(t *testing.T) {

	cases := []struct {
		name           string
		fn             func() *GateModular
		request        func() *http.Request
		wantedResult   string
		wantedResultFn func(body string) bool
	}{
		{
			name: "new request context error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().ListObjectsByIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?%s&ids=%s", scheme, testDomain, ListObjectsByIDsQuery, "1,2,3,4")
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "invalid id",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?%s&ids=%s", scheme, testDomain, ListObjectsByIDsQuery, "a,2,3,4")
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "invalid id number",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				ids := "0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,100"
				path := fmt.Sprintf("%s%s/?%s&ids=%s", scheme, testDomain, ListObjectsByIDsQuery, ids)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "repeated id number",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				ids := "1,1"
				path := fmt.Sprintf("%s%s/?%s&ids=%s", scheme, testDomain, ListObjectsByIDsQuery, ids)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "xml response for 'not found object'",
			fn: func() *GateModular {
				testResponseData := getTestObjectsInIdMap(1)
				testResponseData[4] = nil
				testResponseData[5] = &types.Object{
					ObjectInfo: nil,
				}
				testResponseData[6] = &types.Object{
					ObjectInfo: &storage_types.ObjectInfo{
						Owner:               testAccount,
						Creator:             testAccount,
						BucketName:          mockBucketName,
						ObjectName:          mockObjectName + strconv.Itoa(6),
						Id:                  math.NewUintFromBigInt(big.NewInt(int64(6))),
						LocalVirtualGroupId: 1,
						PayloadSize:         4802764,
						Visibility:          storage_types.VISIBILITY_TYPE_INHERIT,
						ContentType:         "application/octet-stream",
						CreateAt:            1699781700,
						ObjectStatus:        storage_types.OBJECT_STATUS_SEALED,
						Checksums:           nil,
					},

					LockedBalance: "0x0000000000000000000000000000000000000000000000000000000000000000",
					UpdateAt:      1280048,
					Operator:      "0x03AbbEe8E426C9887A8ae3C34602AbCA42aeDFa0",
					CreateTxHash:  "0x491227c644bc89f5a058d92167c00d452c63a1dd8d5776c81617a41ec76fcc8c",
					UpdateTxHash:  "0x238737f109a40c675e1bef5ebfb2adef2cac0a723ee20fbd752e78efbf3d579e",
					SealTxHash:    "0x238737f109a40c675e1bef5ebfb2adef2cac0a723ee20fbd752e78efbf3d579e",
				}

				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().ListObjectsByIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(testResponseData, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				ids := "1"
				path := fmt.Sprintf("%s%s/?%s&ids=%s", scheme, testDomain, ListObjectsByIDsQuery, ids)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResultFn: func(body string) bool {
				assert.Equal(t, "<GfSpListObjectsByIDsResponse><ObjectEntry><Id>4</Id></ObjectEntry><ObjectEntry><Id>5</Id><Value><LockedBalance></LockedBalance><Removed>false</Removed><UpdateAt>0</UpdateAt><DeleteAt>0</DeleteAt><DeleteReason></DeleteReason><Operator></Operator><CreateTxHash></CreateTxHash><UpdateTxHash></UpdateTxHash><SealTxHash></SealTxHash></Value></ObjectEntry><ObjectEntry><Id>6</Id><Value><ObjectInfo><Owner>0xF72aDa8130f934887755492879496b026665FbAB</Owner><Creator>0xF72aDa8130f934887755492879496b026665FbAB</Creator><BucketName>mock-bucket-name</BucketName><ObjectName>mock-object-name6</ObjectName><Id>6</Id><LocalVirtualGroupId>1</LocalVirtualGroupId><PayloadSize>4802764</PayloadSize><Visibility>3</Visibility><ContentType>application/octet-stream</ContentType><CreateAt>1699781700</CreateAt><ObjectStatus>1</ObjectStatus><RedundancyType>0</RedundancyType><SourceType>0</SourceType></ObjectInfo><LockedBalance>0x0000000000000000000000000000000000000000000000000000000000000000</LockedBalance><Removed>false</Removed><UpdateAt>1280048</UpdateAt><DeleteAt>0</DeleteAt><DeleteReason></DeleteReason><Operator>0x03AbbEe8E426C9887A8ae3C34602AbCA42aeDFa0</Operator><CreateTxHash>0x491227c644bc89f5a058d92167c00d452c63a1dd8d5776c81617a41ec76fcc8c</CreateTxHash><UpdateTxHash>0x238737f109a40c675e1bef5ebfb2adef2cac0a723ee20fbd752e78efbf3d579e</UpdateTxHash><SealTxHash>0x238737f109a40c675e1bef5ebfb2adef2cac0a723ee20fbd752e78efbf3d579e</SealTxHash></Value></ObjectEntry><ObjectEntry><Id>0</Id><Value><ObjectInfo><Owner>0xF72aDa8130f934887755492879496b026665FbAB</Owner><Creator>0xF72aDa8130f934887755492879496b026665FbAB</Creator><BucketName>mock-bucket-name</BucketName><ObjectName>mock-object-name0</ObjectName><Id>0</Id><LocalVirtualGroupId>1</LocalVirtualGroupId><PayloadSize>4802764</PayloadSize><Visibility>3</Visibility><ContentType>application/octet-stream</ContentType><CreateAt>1699781700</CreateAt><ObjectStatus>1</ObjectStatus><RedundancyType>0</RedundancyType><SourceType>0</SourceType><Checksums>tPsLBcgLxRVKTRJCeYw5FVj0jjqPsqFnbDCr77pf7RA=</Checksums><Checksums>7YqCbwK/qC+zaAoJvd971fuJCE0OVQ9ky8bgomUkmRI=</Checksums><Checksums>i59qS3vgvN8QIcNKOJggtN4JsZRLYt1ugeGDtP6x7Sk=</Checksums><Checksums>tBBu4BPpANbc12SO5TVeQ64DtKwl0F2inE29H9jAw54=</Checksums><Checksums>vOw+loeUIXXPEvfYNFmnElTIxj/b0dEEBBF1YbKOoEI=</Checksums><Checksums>e0nSN4a5u3EDPaAqemGDZ5gYJ0l6NUjtalmj/BH2uWE=</Checksums><Checksums>rRm6iKPMc8gZbw1WKKF2kPXveU2VFEh2izs9e8ovfwk=</Checksums></ObjectInfo><LockedBalance>0x0000000000000000000000000000000000000000000000000000000000000000</LockedBalance><Removed>false</Removed><UpdateAt>1280048</UpdateAt><DeleteAt>0</DeleteAt><DeleteReason></DeleteReason><Operator>0x03AbbEe8E426C9887A8ae3C34602AbCA42aeDFa0</Operator><CreateTxHash>0x491227c644bc89f5a058d92167c00d452c63a1dd8d5776c81617a41ec76fcc8c</CreateTxHash><UpdateTxHash>0x238737f109a40c675e1bef5ebfb2adef2cac0a723ee20fbd752e78efbf3d579e</UpdateTxHash><SealTxHash>0x238737f109a40c675e1bef5ebfb2adef2cac0a723ee20fbd752e78efbf3d579e</SealTxHash></Value></ObjectEntry></GfSpListObjectsByIDsResponse>",
					body)
				return true
			},
		},
		{
			name: "xml response",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().ListObjectsByIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(getTestObjectsInIdMap(1), nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				ids := "1"
				path := fmt.Sprintf("%s%s/?%s&ids=%s", scheme, testDomain, ListObjectsByIDsQuery, ids)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResultFn: func(body string) bool {
				assert.Equal(t, "<GfSpListObjectsByIDsResponse><ObjectEntry><Id>0</Id><Value><ObjectInfo><Owner>0xF72aDa8130f934887755492879496b026665FbAB</Owner><Creator>0xF72aDa8130f934887755492879496b026665FbAB</Creator><BucketName>mock-bucket-name</BucketName><ObjectName>mock-object-name0</ObjectName><Id>0</Id><LocalVirtualGroupId>1</LocalVirtualGroupId><PayloadSize>4802764</PayloadSize><Visibility>3</Visibility><ContentType>application/octet-stream</ContentType><CreateAt>1699781700</CreateAt><ObjectStatus>1</ObjectStatus><RedundancyType>0</RedundancyType><SourceType>0</SourceType><Checksums>tPsLBcgLxRVKTRJCeYw5FVj0jjqPsqFnbDCr77pf7RA=</Checksums><Checksums>7YqCbwK/qC+zaAoJvd971fuJCE0OVQ9ky8bgomUkmRI=</Checksums><Checksums>i59qS3vgvN8QIcNKOJggtN4JsZRLYt1ugeGDtP6x7Sk=</Checksums><Checksums>tBBu4BPpANbc12SO5TVeQ64DtKwl0F2inE29H9jAw54=</Checksums><Checksums>vOw+loeUIXXPEvfYNFmnElTIxj/b0dEEBBF1YbKOoEI=</Checksums><Checksums>e0nSN4a5u3EDPaAqemGDZ5gYJ0l6NUjtalmj/BH2uWE=</Checksums><Checksums>rRm6iKPMc8gZbw1WKKF2kPXveU2VFEh2izs9e8ovfwk=</Checksums></ObjectInfo><LockedBalance>0x0000000000000000000000000000000000000000000000000000000000000000</LockedBalance><Removed>false</Removed><UpdateAt>1280048</UpdateAt><DeleteAt>0</DeleteAt><DeleteReason></DeleteReason><Operator>0x03AbbEe8E426C9887A8ae3C34602AbCA42aeDFa0</Operator><CreateTxHash>0x491227c644bc89f5a058d92167c00d452c63a1dd8d5776c81617a41ec76fcc8c</CreateTxHash><UpdateTxHash>0x238737f109a40c675e1bef5ebfb2adef2cac0a723ee20fbd752e78efbf3d579e</UpdateTxHash><SealTxHash>0x238737f109a40c675e1bef5ebfb2adef2cac0a723ee20fbd752e78efbf3d579e</SealTxHash></Value></ObjectEntry></GfSpListObjectsByIDsResponse>",
					body)
				return true
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockListObjectsByIDsHandlerRoute(t, tt.fn())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.request())
			if tt.wantedResult != "" {
				assert.Contains(t, w.Body.String(), tt.wantedResult)
			}
			if tt.wantedResultFn != nil {
				assert.True(t, tt.wantedResultFn(w.Body.String()))
			}
		})
	}
}

func mockListGroupsByIDsHandlerRoute(t *testing.T, g *GateModular) *mux.Router {
	t.Helper()
	router := mux.NewRouter().SkipClean(true)
	router.Path("/").Name(listGroupsByIDsRouterName).Methods(http.MethodGet).Queries(ListGroupsByIDsQuery, "").HandlerFunc(g.listGroupsByIDsHandler)
	return router
}

func TestGateModular_ListGroupsByIDsHandler(t *testing.T) {

	cases := []struct {
		name           string
		fn             func() *GateModular
		request        func() *http.Request
		wantedResult   string
		wantedResultFn func(body string) bool
	}{
		{
			name: "new request context error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().ListGroupsByIDs(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?%s&ids=%s", scheme, testDomain, ListGroupsByIDsQuery, "1,2,3,4")
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "invalid id",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?%s&ids=%s", scheme, testDomain, ListGroupsByIDsQuery, "a,2,3,4")
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "invalid id number",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				ids := "0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,100"
				path := fmt.Sprintf("%s%s/?%s&ids=%s", scheme, testDomain, ListGroupsByIDsQuery, ids)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "repeated id number",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				ids := "1,1"
				path := fmt.Sprintf("%s%s/?%s&ids=%s", scheme, testDomain, ListGroupsByIDsQuery, ids)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "xml response",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().ListGroupsByIDs(gomock.Any(), gomock.Any()).Return(getTestGroupsInIdMap(1), nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				ids := "1"
				path := fmt.Sprintf("%s%s/?%s&ids=%s", scheme, testDomain, ListGroupsByIDsQuery, ids)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResultFn: func(body string) bool {
				assert.Equal(t, "<GfSpListGroupsByIDsResponse><GroupEntry><Id>0</Id><Value><Group><Owner>0xF72aDa8130f934887755492879496b026665FbAB</Owner><GroupName>TestGroupName 0</GroupName><SourceType>0</SourceType><Id>0</Id><Extra></Extra></Group><Operator>0x03AbbEe8E426C9887A8ae3C34602AbCA42aeDFa0</Operator><CreateAt>0</CreateAt><CreateTime>0</CreateTime><UpdateAt>1280048</UpdateAt><UpdateTime>0</UpdateTime><NumberOfMembers>1</NumberOfMembers><Removed>false</Removed></Value></GroupEntry></GfSpListGroupsByIDsResponse>",
					body)
				return true
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockListGroupsByIDsHandlerRoute(t, tt.fn())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.request())
			if tt.wantedResult != "" {
				assert.Contains(t, w.Body.String(), tt.wantedResult)
			}
			if tt.wantedResultFn != nil {
				assert.True(t, tt.wantedResultFn(w.Body.String()))
			}
		})
	}
}

func mockGetGroupListHandlerRoute(t *testing.T, g *GateModular) *mux.Router {
	t.Helper()
	router := mux.NewRouter().SkipClean(true)
	router.Path("/").Name(getGroupListRouterName).Methods(http.MethodGet).Queries(GetGroupListGroupQuery, "").HandlerFunc(g.getGroupListHandler)
	return router
}

func TestGateModular_GetGroupListHandler(t *testing.T) {

	cases := []struct {
		name           string
		fn             func() *GateModular
		request        func() *http.Request
		wantedResult   string
		wantedResultFn func(body string) bool
	}{
		{
			name: "new request context error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetGroupList(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, int64(0), mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?group-query&source-type=SOURCE_TYPE_ORIGIN&limit=10&offset=0&name=e&prefix=t", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "blank name",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?group-query&source-type=SOURCE_TYPE_ORIGIN&limit=10&offset=0&name=&prefix=t", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "blank prefix",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?group-query&source-type=SOURCE_TYPE_ORIGIN&limit=10&offset=0&name=n&prefix=", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "bad offset",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?group-query&source-type=SOURCE_TYPE_ORIGIN&limit=&offset=a&name=n&prefix=t", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "bad limit",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?group-query&source-type=SOURCE_TYPE_ORIGIN&limit=a&offset=&name=n&prefix=t", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "limit is negative",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?group-query&source-type=SOURCE_TYPE_ORIGIN&limit=-1&offset=&name=n&prefix=t", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "bad sourceType",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?group-query&source-type=SOURCE_TYPE_UNKNOWN&limit=1&offset=0&name=n&prefix=t", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "xml response",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetGroupList(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(getTestGroupsResponse(1), int64(1), nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?group-query&source-type=SOURCE_TYPE_ORIGIN&limit=1&offset=0&name=n&prefix=t", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResultFn: func(body string) bool {
				assert.Equal(t, "<GfSpGetGroupListResponse><Groups><Group><Owner>0xF72aDa8130f934887755492879496b026665FbAB</Owner><GroupName>TestGroupName 0</GroupName><SourceType>0</SourceType><Id>0</Id><Extra></Extra></Group><Operator>0x03AbbEe8E426C9887A8ae3C34602AbCA42aeDFa0</Operator><CreateAt>0</CreateAt><CreateTime>0</CreateTime><UpdateAt>1280048</UpdateAt><UpdateTime>0</UpdateTime><NumberOfMembers>1</NumberOfMembers><Removed>false</Removed></Groups><Count>1</Count></GfSpGetGroupListResponse>",
					body)
				return true
			},
		},

		{
			name: "xml response for large limit", // return 1000 records when limit is more than 1000
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetGroupList(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(int64(1000)), gomock.Any(), gomock.Any()).Return(getTestGroupsResponse(1000), int64(1), nil).AnyTimes()
				clientMock.EXPECT().GetGroupList(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(int64(1001)), gomock.Any(), gomock.Any()).Return(getTestGroupsResponse(1001), int64(1), nil).AnyTimes()
				clientMock.EXPECT().GetGroupList(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(int64(50)), gomock.Any(), gomock.Any()).Return(getTestGroupsResponse(50), int64(1), nil).AnyTimes()
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?group-query&source-type=SOURCE_TYPE_ORIGIN&limit=1001&offset=0&name=n&prefix=t", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResultFn: func(body string) bool {
				var res types.GfSpGetGroupListResponse
				err := xml.Unmarshal([]byte(body), &res)
				if err != nil {
					return false
				}
				assert.Equal(t, 1000, len(res.Groups))
				return true
			},
		},
		{
			name: "xml response for default limit", // return 50 records when limit is not set
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetGroupList(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(int64(1000)), gomock.Any(), gomock.Any()).Return(getTestGroupsResponse(1000), int64(1), nil).AnyTimes()
				clientMock.EXPECT().GetGroupList(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(int64(1001)), gomock.Any(), gomock.Any()).Return(getTestGroupsResponse(1001), int64(1), nil).AnyTimes()
				clientMock.EXPECT().GetGroupList(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(int64(50)), gomock.Any(), gomock.Any()).Return(getTestGroupsResponse(50), int64(1), nil).AnyTimes()
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?group-query&source-type=SOURCE_TYPE_ORIGIN&offset=0&name=n&prefix=t", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResultFn: func(body string) bool {
				var res types.GfSpGetGroupListResponse
				err := xml.Unmarshal([]byte(body), &res)
				if err != nil {
					return false
				}
				assert.Equal(t, 50, len(res.Groups))
				return true
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockGetGroupListHandlerRoute(t, tt.fn())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.request())
			if tt.wantedResult != "" {
				assert.Contains(t, w.Body.String(), tt.wantedResult)
			}
			if tt.wantedResultFn != nil {
				assert.True(t, tt.wantedResultFn(w.Body.String()))
			}
		})
	}
}

func mockListPaymentAccountStreamsHandlerRoute(t *testing.T, g *GateModular) *mux.Router {
	t.Helper()
	router := mux.NewRouter().SkipClean(true)
	router.Path("/").Name(listPaymentAccountStreamsRouterName).Methods(http.MethodGet).Queries(ListPaymentAccountStreamsQuery, "").HandlerFunc(g.listPaymentAccountStreamsHandler)
	return router
}

func TestGateModular_ListPaymentAccountStreamsHandler(t *testing.T) {

	cases := []struct {
		name           string
		fn             func() *GateModular
		request        func() *http.Request
		wantedResult   string
		wantedResultFn func(body string) bool
	}{
		{
			name: "new request context error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().ListPaymentAccountStreams(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?%s&%s=%s", scheme, testDomain, ListPaymentAccountStreamsQuery, PaymentAccountQuery, testAccount)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "invalid payment account",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?%s&%s=%s", scheme, testDomain, ListPaymentAccountStreamsQuery, PaymentAccountQuery, "invalid_payment_account")
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "xml response",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().ListPaymentAccountStreams(gomock.Any(), gomock.Any()).Return(getTestBuckets(1), nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?%s&%s=%s", scheme, testDomain, ListPaymentAccountStreamsQuery, PaymentAccountQuery, testAccount)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResultFn: func(body string) bool {
				assert.Equal(t, "<GfSpListPaymentAccountStreamsResponse><Buckets><BucketInfo><Owner>0xF72aDa8130f934887755492879496b026665FbAB</Owner><BucketName>mock-bucket-name0</BucketName><Visibility>1</Visibility><Id>0</Id><SourceType>0</SourceType><CreateAt>1699781080</CreateAt><PaymentAddress>0xF72aDa8130f934887755492879496b026665FbAB</PaymentAddress><GlobalVirtualGroupFamilyId>3</GlobalVirtualGroupFamilyId><ChargedReadQuota>0</ChargedReadQuota><BucketStatus>0</BucketStatus></BucketInfo><Removed>false</Removed><DeleteAt>0</DeleteAt><DeleteReason></DeleteReason><Operator>0xF72aDa8130f934887755492879496b026665FbAB</Operator><CreateTxHash>0x21c349a869bde1f44378936e2a9a15ed3fb2d54a43eaea8787960bba1134cdc2</CreateTxHash><UpdateTxHash>0x0cbff0ff3831d61345dbfda5b984e254c4bf87ecf80b45ccbb0635c0547a3b1a</UpdateTxHash><UpdateAt>1279811</UpdateAt><UpdateTime>1699781103</UpdateTime><StorageSize></StorageSize></Buckets></GfSpListPaymentAccountStreamsResponse>",
					body)
				return true
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockListPaymentAccountStreamsHandlerRoute(t, tt.fn())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.request())
			if tt.wantedResult != "" {
				assert.Contains(t, w.Body.String(), tt.wantedResult)
			}
			if tt.wantedResultFn != nil {
				assert.True(t, tt.wantedResultFn(w.Body.String()))
			}
		})
	}
}

func mockListUserPaymentAccountsHandlerRoute(t *testing.T, g *GateModular) *mux.Router {
	t.Helper()
	router := mux.NewRouter().SkipClean(true)
	router.Path("/").Name(listUserPaymentAccountsRouterName).Methods(http.MethodGet).Queries(ListUserPaymentAccountsQuery, "").HandlerFunc(g.listUserPaymentAccountsHandler)
	return router
}

func TestGateModular_ListUserPaymentAccountsHandler(t *testing.T) {

	cases := []struct {
		name           string
		fn             func() *GateModular
		request        func() *http.Request
		wantedResult   string
		wantedResultFn func(body string) bool
	}{
		{
			name: "new request context error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().ListUserPaymentAccounts(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?%s", scheme, testDomain, ListUserPaymentAccountsQuery)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set(GnfdUserAddressHeader, testAccount)

				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "invalid payment account",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?%s", scheme, testDomain, ListUserPaymentAccountsQuery)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set(GnfdUserAddressHeader, "invalid_payment_account")

				return req
			},
			wantedResult: "invalid request header",
		},
		{
			name: "xml response",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().ListUserPaymentAccounts(gomock.Any(), gomock.Any()).Return(getTestPaymentAccountMeta(), nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?%s", scheme, testDomain, ListUserPaymentAccountsQuery)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set(GnfdUserAddressHeader, testAccount)

				return req
			},
			wantedResultFn: func(body string) bool {
				assert.Equal(t, "<GfSpListUserPaymentAccountsResponse><PaymentAccounts><PaymentAccount><Address>0xF72aDa8130f934887755492879496b026665FbAB</Address><Owner>0xF72aDa8130f934887755492879496b026665FbAB</Owner><Refundable>true</Refundable><UpdateAt>1279659</UpdateAt><UpdateTime>1699780707</UpdateTime></PaymentAccount><StreamRecord><Account>0xF72aDa8130f934887755492879496b026665FbAB</Account><CrudTimestamp>1699780994</CrudTimestamp><NetflowRate>0</NetflowRate><StaticBalance>240000000000000001</StaticBalance><BufferBalance>0</BufferBalance><LockBalance>0</LockBalance><Status>0</Status><SettleTimestamp>0</SettleTimestamp><OutFlowCount>0</OutFlowCount><FrozenNetflowRate>0</FrozenNetflowRate></StreamRecord></PaymentAccounts></GfSpListUserPaymentAccountsResponse>",
					body)
				return true
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockListUserPaymentAccountsHandlerRoute(t, tt.fn())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.request())
			if tt.wantedResult != "" {
				assert.Contains(t, w.Body.String(), tt.wantedResult)
			}
			if tt.wantedResultFn != nil {
				assert.True(t, tt.wantedResultFn(w.Body.String()))
			}
		})
	}
}

func mockListBucketsByIDsHandlerRoute(t *testing.T, g *GateModular) *mux.Router {
	t.Helper()
	router := mux.NewRouter().SkipClean(true)
	router.Path("/").Name(listBucketsByIDsRouterName).Methods(http.MethodGet).Queries(ListBucketsByIDsQuery, "").HandlerFunc(g.listBucketsByIDsHandler)
	return router
}

func TestGateModular_ListBucketsByIDsHandler(t *testing.T) {

	cases := []struct {
		name           string
		fn             func() *GateModular
		request        func() *http.Request
		wantedResult   string
		wantedResultFn func(body string) bool
	}{
		{
			name: "new request context error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().ListBucketsByIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?%s&ids=%s", scheme, testDomain, ListBucketsByIDsQuery, "1,2,3,4")
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "invalid id",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?%s&ids=%s", scheme, testDomain, ListBucketsByIDsQuery, "a,2,3,4")
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "invalid id number",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				ids := "0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,100"
				path := fmt.Sprintf("%s%s/?%s&ids=%s", scheme, testDomain, ListBucketsByIDsQuery, ids)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "repeated id number",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				ids := "1,1"
				path := fmt.Sprintf("%s%s/?%s&ids=%s", scheme, testDomain, ListBucketsByIDsQuery, ids)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "xml response",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().ListBucketsByIDs(gomock.Any(), gomock.Any(), gomock.Any()).Return(getTestBucketsInIdMap(1), nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				ids := "1"
				path := fmt.Sprintf("%s%s/?%s&ids=%s", scheme, testDomain, ListBucketsByIDsQuery, ids)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResultFn: func(body string) bool {
				assert.Equal(t, "<GfSpListBucketsByIDsResponse><BucketEntry><Id>0</Id><Value><BucketInfo><Owner>0xF72aDa8130f934887755492879496b026665FbAB</Owner><BucketName>mock-bucket-name0</BucketName><Visibility>1</Visibility><Id>0</Id><SourceType>0</SourceType><CreateAt>1699781080</CreateAt><PaymentAddress>0xF72aDa8130f934887755492879496b026665FbAB</PaymentAddress><GlobalVirtualGroupFamilyId>3</GlobalVirtualGroupFamilyId><ChargedReadQuota>0</ChargedReadQuota><BucketStatus>0</BucketStatus></BucketInfo><Removed>false</Removed><DeleteAt>0</DeleteAt><DeleteReason></DeleteReason><Operator>0xF72aDa8130f934887755492879496b026665FbAB</Operator><CreateTxHash>0x21c349a869bde1f44378936e2a9a15ed3fb2d54a43eaea8787960bba1134cdc2</CreateTxHash><UpdateTxHash>0x0cbff0ff3831d61345dbfda5b984e254c4bf87ecf80b45ccbb0635c0547a3b1a</UpdateTxHash><UpdateAt>1279811</UpdateAt><UpdateTime>1699781103</UpdateTime><StorageSize></StorageSize></Value></BucketEntry></GfSpListBucketsByIDsResponse>",
					body)
				return true
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockListBucketsByIDsHandlerRoute(t, tt.fn())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.request())
			if tt.wantedResult != "" {
				assert.Contains(t, w.Body.String(), tt.wantedResult)
			}
			if tt.wantedResultFn != nil {
				assert.True(t, tt.wantedResultFn(w.Body.String()))
			}
		})
	}
}

func mockGetUserGroupsHandlerRoute(t *testing.T, g *GateModular) *mux.Router {
	t.Helper()
	router := mux.NewRouter().SkipClean(true)
	router.Path("/").Name(getUserGroupsRouterName).Methods(http.MethodGet).Queries(GetUserGroupsQuery, "").HandlerFunc(g.getUserGroupsHandler)
	return router
}

func TestGateModular_GetUserGroupsHandler(t *testing.T) {

	cases := []struct {
		name           string
		fn             func() *GateModular
		request        func() *http.Request
		wantedResult   string
		wantedResultFn func(body string) bool
	}{
		{
			name: "new request context error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetUserGroups(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?user-groups&start-after=0&limit=1000", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set(GnfdUserAddressHeader, testAccount)
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "invalid user address",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?user-groups&start-after=0&limit=1000", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set(GnfdUserAddressHeader, "invalid_account")
				return req
			},
			wantedResult: "invalid request header",
		},
		{
			name: "invalid requestStartAfter",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?user-groups&start-after=a&limit=1000", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set(GnfdUserAddressHeader, testAccount)
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "invalid limit",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?user-groups&limit=a", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set(GnfdUserAddressHeader, testAccount)
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "xmlResponse",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetUserGroups(gomock.Any(), gomock.Eq(testAccount), gomock.Eq(uint64(2)), gomock.Eq(uint32(1))).Return(getTestGroupMembersResponse(1), nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?user-groups&start-after=2&limit=1", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set(GnfdUserAddressHeader, testAccount)
				return req
			},
			wantedResult: "<GfSpGetUserGroupsResponse><Groups><Group><Owner>0xF72aDa8130f934887755492879496b026665FbAB</Owner><GroupName>TestGroupName 0</GroupName><SourceType>0</SourceType><Id>0</Id><Extra></Extra></Group><AccountId>0xF72aDa8130f934887755492879496b026665FbAB</AccountId><Operator>0xF72aDa8130f934887755492879496b026665FbAB</Operator><CreateAt>1376086</CreateAt><CreateTime>1700032197</CreateTime><UpdateAt>1376086</UpdateAt><UpdateTime>1700032197</UpdateTime><Removed>false</Removed><ExpirationTime>253402300799</ExpirationTime></Groups></GfSpGetUserGroupsResponse>",
			wantedResultFn: func(body string) bool {
				var res types.GfSpGetUserGroupsResponse
				err := xml.Unmarshal([]byte(body), &res)
				if err != nil {
					return false
				}
				assert.Equal(t, 1, len(res.Groups))
				return true
			},
		},
		{
			name: "large xmlResponse",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetUserGroups(gomock.Any(), gomock.Eq(testAccount), gomock.Eq(uint64(2)), gomock.Eq(uint32(1000))).Return(getTestGroupMembersResponse(1000), nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?user-groups&start-after=2&limit=1000", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set(GnfdUserAddressHeader, testAccount)
				return req
			},
			wantedResultFn: func(body string) bool {
				var res types.GfSpGetUserGroupsResponse
				err := xml.Unmarshal([]byte(body), &res)
				if err != nil {
					return false
				}
				assert.Equal(t, 1000, len(res.Groups))
				return true
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockGetUserGroupsHandlerRoute(t, tt.fn())
			w := httptest.NewRecorder()
			begin := time.Now()
			router.ServeHTTP(w, tt.request())
			end := time.Now()
			log.Printf("GetUserGroupsHandler takes %d to finish", end.UnixMilli()-begin.UnixMilli())

			assert.Less(t, end.UnixMilli()-begin.UnixMilli(), int64(300)) // we expected this API can return response in 0.3 sec after it gets data from DB.
			if tt.wantedResult != "" {
				assert.Contains(t, w.Body.String(), tt.wantedResult)
			}
			if tt.wantedResultFn != nil {
				assert.True(t, tt.wantedResultFn(w.Body.String()))
			}
		})
	}
}

func mockGetGroupMembersHandlerRoute(t *testing.T, g *GateModular) *mux.Router {
	t.Helper()
	router := mux.NewRouter().SkipClean(true)
	router.Path("/").Name(getGroupMembersRouterName).Methods(http.MethodGet).Queries(GetGroupMembersQuery, "").HandlerFunc(g.getGroupMembersHandler)
	return router
}

func TestGateModular_GetGroupMembersHandler(t *testing.T) {

	cases := []struct {
		name           string
		fn             func() *GateModular
		request        func() *http.Request
		wantedResult   string
		wantedResultFn func(body string) bool
	}{
		{
			name: "new request context error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetGroupMembers(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?group-members&group-id=2&limit=10", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "invalid requestStartAfter",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?group-members&group-id=2&limit=10&start-after=a", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "invalid group id",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?group-members&group-id=a&limit=10", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "invalid limit",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?group-members&group-id=2&limit=a", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "xmlResponse",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetGroupMembers(gomock.Any(), gomock.Eq(uint64(73)), gomock.Eq("0x76d32704A1f415a0a8139997Bb40978b9EEf031f"), gomock.Eq(uint32(1))).Return(getTestGroupMembersResponse(1), nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?group-members&group-id=73&limit=1&start-after=0x76d32704A1f415a0a8139997Bb40978b9EEf031f", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "<GfSpGetGroupMembersResponse><Groups><Group><Owner>0xF72aDa8130f934887755492879496b026665FbAB</Owner><GroupName>TestGroupName 0</GroupName><SourceType>0</SourceType><Id>0</Id><Extra></Extra></Group><AccountId>0xF72aDa8130f934887755492879496b026665FbAB</AccountId><Operator>0xF72aDa8130f934887755492879496b026665FbAB</Operator><CreateAt>1376086</CreateAt><CreateTime>1700032197</CreateTime><UpdateAt>1376086</UpdateAt><UpdateTime>1700032197</UpdateTime><Removed>false</Removed><ExpirationTime>253402300799</ExpirationTime></Groups></GfSpGetGroupMembersResponse>",
			wantedResultFn: func(body string) bool {
				var res types.GfSpGetGroupMembersResponse
				err := xml.Unmarshal([]byte(body), &res)
				if err != nil {
					return false
				}
				assert.Equal(t, 1, len(res.Groups))
				return true
			},
		},
		{
			name: "large xmlResponse",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetGroupMembers(gomock.Any(), gomock.Eq(uint64(73)), gomock.Eq("0x76d32704A1f415a0a8139997Bb40978b9EEf031f"), gomock.Eq(uint32(1000))).Return(getTestGroupMembersResponse(1000), nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?group-members&group-id=73&limit=1000&start-after=0x76d32704A1f415a0a8139997Bb40978b9EEf031f", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResultFn: func(body string) bool {
				var res types.GfSpGetGroupMembersResponse
				err := xml.Unmarshal([]byte(body), &res)
				if err != nil {
					return false
				}
				assert.Equal(t, 1000, len(res.Groups))
				return true
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockGetGroupMembersHandlerRoute(t, tt.fn())
			w := httptest.NewRecorder()
			begin := time.Now()
			router.ServeHTTP(w, tt.request())
			end := time.Now()
			log.Printf("GetGroupMembersHandler takes %d to finish", end.UnixMilli()-begin.UnixMilli())
			assert.Less(t, end.UnixMilli()-begin.UnixMilli(), int64(300)) // we expected this API can return response in 0.3 sec after it gets data from DB.
			if tt.wantedResult != "" {
				assert.Contains(t, w.Body.String(), tt.wantedResult)
			}
			if tt.wantedResultFn != nil {
				assert.True(t, tt.wantedResultFn(w.Body.String()))
			}
		})
	}
}

func mockGetUserOwnedGroupsHandlerRoute(t *testing.T, g *GateModular) *mux.Router {
	t.Helper()
	router := mux.NewRouter().SkipClean(true)
	router.Path("/").Name(getUserOwnedGroupsRouterName).Methods(http.MethodGet).Queries(GetUserOwnedGroupsQuery, "").HandlerFunc(g.getUserOwnedGroupsHandler)
	return router
}

func TestGateModular_GetUserOwnedGroupsHandler(t *testing.T) {

	cases := []struct {
		name           string
		fn             func() *GateModular
		request        func() *http.Request
		wantedResult   string
		wantedResultFn func(body string) bool
	}{
		{
			name: "new request context error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetUserOwnedGroups(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?owned-groups&start-after=15&limit=1", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set(GnfdUserAddressHeader, testAccount)
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "invalid user address",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?owned-groups&start-after=15&limit=1", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set(GnfdUserAddressHeader, "invalid_account")
				return req
			},
			wantedResult: "invalid request header",
		},

		{
			name: "invalid requestStartAfter",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?owned-groups&start-after=a&limit=1", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set(GnfdUserAddressHeader, testAccount)
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "invalid limit",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?owned-groups&start-after=15&limit=a", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set(GnfdUserAddressHeader, testAccount)
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "xmlResponse",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetUserOwnedGroups(gomock.Any(), gomock.Eq(testAccount), gomock.Eq(uint64(15)), gomock.Eq(uint32(1))).Return(getTestGroupMembersResponse(1), nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?owned-groups&start-after=15&limit=1", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set(GnfdUserAddressHeader, testAccount)
				return req
			},
			wantedResult: "<GfSpGetUserOwnedGroupsResponse><Groups><Group><Owner>0xF72aDa8130f934887755492879496b026665FbAB</Owner><GroupName>TestGroupName 0</GroupName><SourceType>0</SourceType><Id>0</Id><Extra></Extra></Group><AccountId>0xF72aDa8130f934887755492879496b026665FbAB</AccountId><Operator>0xF72aDa8130f934887755492879496b026665FbAB</Operator><CreateAt>1376086</CreateAt><CreateTime>1700032197</CreateTime><UpdateAt>1376086</UpdateAt><UpdateTime>1700032197</UpdateTime><Removed>false</Removed><ExpirationTime>253402300799</ExpirationTime></Groups></GfSpGetUserOwnedGroupsResponse>",
			wantedResultFn: func(body string) bool {
				var res types.GfSpGetUserOwnedGroupsResponse
				err := xml.Unmarshal([]byte(body), &res)
				if err != nil {
					return false
				}
				assert.Equal(t, 1, len(res.Groups))
				return true
			},
		},
		{
			name: "large xmlResponse",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetUserOwnedGroups(gomock.Any(), gomock.Eq(testAccount), gomock.Eq(uint64(15)), gomock.Eq(uint32(1000))).Return(getTestGroupMembersResponse(1000), nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?owned-groups&start-after=15&limit=1000", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set(GnfdUserAddressHeader, testAccount)
				return req
			},
			wantedResultFn: func(body string) bool {
				var res types.GfSpGetUserOwnedGroupsResponse
				err := xml.Unmarshal([]byte(body), &res)
				if err != nil {
					return false
				}
				assert.Equal(t, 1000, len(res.Groups))
				return true
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockGetUserOwnedGroupsHandlerRoute(t, tt.fn())
			w := httptest.NewRecorder()
			begin := time.Now()
			router.ServeHTTP(w, tt.request())
			end := time.Now()
			log.Printf("GetUserOwnedGroupsHandler takes %d to finish", end.UnixMilli()-begin.UnixMilli())
			assert.Less(t, end.UnixMilli()-begin.UnixMilli(), int64(300)) // we expected this API can return response in 0.3 sec after it gets data from DB.
			if tt.wantedResult != "" {
				assert.Contains(t, w.Body.String(), tt.wantedResult)
			}
			if tt.wantedResultFn != nil {
				assert.True(t, tt.wantedResultFn(w.Body.String()))
			}
		})
	}
}

func mockGetBucketMetaHandlerRoute(t *testing.T, g *GateModular) *mux.Router {
	t.Helper()
	router := mux.NewRouter().SkipClean(true)
	var routers []*mux.Router
	routers = append(routers, router.Host("{bucket:.+}."+g.domain).Subrouter())
	routers = append(routers, router.PathPrefix("/{bucket}").Subrouter())
	for _, r := range routers {
		r.NewRoute().Name(getBucketMetaRouterName).Methods(http.MethodGet).Queries(GetBucketMetaQuery, "").HandlerFunc(g.getBucketMetaHandler)
	}
	return router
}

func TestGateModular_GetBucketMetaHandler(t *testing.T) {

	cases := []struct {
		name           string
		fn             func() *GateModular
		request        func() *http.Request
		wantedResult   string
		wantedResultFn func(body string) bool
	}{
		{
			name: "new request context error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetBucketMeta(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/%s?%s", scheme, testDomain, mockBucketName, GetBucketMetaQuery)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "wrong BucketName",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/%s?%s", scheme, testDomain, "aa", GetBucketMetaQuery) // aa is an invalid bucket name
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "invalid request params for query",
		},
		{
			name: "xml response",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetBucketMeta(gomock.Any(), gomock.Any(), gomock.Any()).Return(getTestVGFInfoBuckets(1)[0], getTestStreamRecord(), nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/%s?%s", scheme, testDomain, mockBucketName, GetBucketMetaQuery)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedResult: "<GfSpGetBucketMetaResponse><Bucket><BucketInfo><Owner>0xF72aDa8130f934887755492879496b026665FbAB</Owner><BucketName>mock-bucket-name0</BucketName><Visibility>1</Visibility><Id>0</Id><SourceType>0</SourceType><CreateAt>1699781080</CreateAt><PaymentAddress>0xF72aDa8130f934887755492879496b026665FbAB</PaymentAddress><GlobalVirtualGroupFamilyId>3</GlobalVirtualGroupFamilyId><ChargedReadQuota>0</ChargedReadQuota><BucketStatus>0</BucketStatus></BucketInfo><Removed>false</Removed><DeleteAt>0</DeleteAt><DeleteReason></DeleteReason><Operator>0xF72aDa8130f934887755492879496b026665FbAB</Operator><CreateTxHash>0x21c349a869bde1f44378936e2a9a15ed3fb2d54a43eaea8787960bba1134cdc2</CreateTxHash><UpdateTxHash>0x0cbff0ff3831d61345dbfda5b984e254c4bf87ecf80b45ccbb0635c0547a3b1a</UpdateTxHash><UpdateAt>1279811</UpdateAt><UpdateTime>1699781103</UpdateTime><Vgf><Id>0</Id><PrimarySpId>3</PrimarySpId><VirtualPaymentAddress>0x26281179b8885F21f95b0a246c8AD70957A95A23</VirtualPaymentAddress></Vgf><StorageSize></StorageSize></Bucket><StreamRecord><Account>0xF72aDa8130f934887755492879496b026665FbAB</Account><CrudTimestamp>1699780994</CrudTimestamp><NetflowRate>0</NetflowRate><StaticBalance>240000000000000001</StaticBalance><BufferBalance>0</BufferBalance><LockBalance>0</LockBalance><Status>0</Status><SettleTimestamp>0</SettleTimestamp><OutFlowCount>0</OutFlowCount><FrozenNetflowRate>0</FrozenNetflowRate></StreamRecord></GfSpGetBucketMetaResponse>",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockGetBucketMetaHandlerRoute(t, tt.fn())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.request())
			if tt.wantedResult != "" {
				assert.Contains(t, w.Body.String(), tt.wantedResult)
			}
			if tt.wantedResultFn != nil {
				assert.True(t, tt.wantedResultFn(w.Body.String()))
			}
		})
	}
}

func mockGetUserBucketsHandlerRoute(t *testing.T, g *GateModular) *mux.Router {
	t.Helper()
	router := mux.NewRouter().SkipClean(true)
	router.Path("/").Name(getUserBucketsRouterName).Methods(http.MethodGet).HandlerFunc(g.getUserBucketsHandler)

	return router
}

func TestGateModular_GetUserBucketsHandler(t *testing.T) {

	cases := []struct {
		name           string
		fn             func() *GateModular
		request        func() *http.Request
		wantedResult   string
		wantedResultFn func(body string) bool
	}{
		{
			name: "new request context error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetUserBuckets(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set(GnfdUserAddressHeader, testAccount)
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "invalid user address",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set(GnfdUserAddressHeader, "invalid_account")
				return req
			},
			wantedResult: "invalid request header",
		},
		{
			name: "wrong requestIncludeRemoved",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/?&&include-removed=invalid", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set(GnfdUserAddressHeader, testAccount)
				return req
			},
			wantedResult: "invalid request params for query",
		},

		{
			name: "xml response",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetUserBuckets(gomock.Any(), testAccount, gomock.Any()).Return(getTestVGFInfoBuckets(1), nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/", scheme, testDomain)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set(GnfdUserAddressHeader, testAccount)
				return req
			},
			wantedResult: "<GfSpGetUserBucketsResponse><Buckets><BucketInfo><Owner>0xF72aDa8130f934887755492879496b026665FbAB</Owner><BucketName>mock-bucket-name0</BucketName><Visibility>1</Visibility><Id>0</Id><SourceType>0</SourceType><CreateAt>1699781080</CreateAt><PaymentAddress>0xF72aDa8130f934887755492879496b026665FbAB</PaymentAddress><GlobalVirtualGroupFamilyId>3</GlobalVirtualGroupFamilyId><ChargedReadQuota>0</ChargedReadQuota><BucketStatus>0</BucketStatus></BucketInfo><Removed>false</Removed><DeleteAt>0</DeleteAt><DeleteReason></DeleteReason><Operator>0xF72aDa8130f934887755492879496b026665FbAB</Operator><CreateTxHash>0x21c349a869bde1f44378936e2a9a15ed3fb2d54a43eaea8787960bba1134cdc2</CreateTxHash><UpdateTxHash>0x0cbff0ff3831d61345dbfda5b984e254c4bf87ecf80b45ccbb0635c0547a3b1a</UpdateTxHash><UpdateAt>1279811</UpdateAt><UpdateTime>1699781103</UpdateTime><Vgf><Id>0</Id><PrimarySpId>3</PrimarySpId><VirtualPaymentAddress>0x26281179b8885F21f95b0a246c8AD70957A95A23</VirtualPaymentAddress></Vgf><StorageSize></StorageSize></Buckets></GfSpGetUserBucketsResponse>",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockGetUserBucketsHandlerRoute(t, tt.fn())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.request())
			if tt.wantedResult != "" {
				assert.Contains(t, w.Body.String(), tt.wantedResult)
			}
			if tt.wantedResultFn != nil {
				assert.True(t, tt.wantedResultFn(w.Body.String()))
			}
		})
	}
}

func mockListObjectPoliciesRoute(t *testing.T, g *GateModular) *mux.Router {
	t.Helper()
	router := mux.NewRouter().SkipClean(true)
	var routers []*mux.Router
	routers = append(routers, router.Host("{bucket:.+}."+g.domain).Subrouter())
	routers = append(routers, router.PathPrefix("/{bucket}").Subrouter())
	for _, r := range routers {
		r.NewRoute().Name(listObjectPoliciesRouterName).Methods(http.MethodGet).Path("/{object:.+}").Queries(ListObjectPoliciesQuery, "").
			HandlerFunc(g.listObjectPoliciesHandler)
	}
	return router
}

func TestGateModular_listObjectPoliciesHandler(t *testing.T) {
	cases := []struct {
		name       string
		fn         func() *GateModular
		request    func() *http.Request
		wantedCode int
	}{
		{
			name: "bukcet name error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s&%s", scheme, invalidBucketName, testDomain, invalidObjectName, ListObjectPoliciesQuery,
					"limit=10&action-type=6")
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedCode: 400,
		},
		{
			name: "object name error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s&%s", scheme, mockBucketName, testDomain, invalidObjectName, ListObjectPoliciesQuery,
					"limit=10&action-type=6")
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedCode: 400,
		},
		{
			name: "limit value error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s&%s", scheme, mockBucketName, testDomain, mockObjectName, ListObjectPoliciesQuery,
					"limit=a&action-type=6")
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedCode: 400,
		},
		{
			name: "start-after value error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s&%s", scheme, mockBucketName, testDomain, mockObjectName, ListObjectPoliciesQuery,
					"limit=10&action-type=6&start-after=a")
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedCode: 400,
		},
		{
			name: "action type value error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s&%s", scheme, mockBucketName, testDomain, mockObjectName, ListObjectPoliciesQuery,
					"limit=10&action-type=100")
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedCode: 400,
		},
		{
			name: "action type value error 2",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s&%s", scheme, mockBucketName, testDomain, mockObjectName, ListObjectPoliciesQuery,
					"limit=10&action-type=a")
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedCode: 400,
		},
		{
			name: "rpc error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().ListObjectPolicies(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any()).Return(nil, errors.New(`rpc error: code = Unknown desc = {"code_space":"metadata","http_status_code":404,"inner_code":90008,"description":"the specified bucket does not exist"}`)).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s&%s", scheme, mockBucketName, testDomain, mockObjectName, ListObjectPoliciesQuery,
					"limit=10&action-type=6")
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedCode: 404,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockListObjectPoliciesRoute(t, tt.fn())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.request())
			assert.Equal(t, w.Code, tt.wantedCode)
		})
	}
}
