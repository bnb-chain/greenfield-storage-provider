package metadata

//
//import (
//	"context"
//
//	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
//	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
//	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
//)
//
//// GfSpGetEndpointBySpAddress get endpoint by sp address
//func (r *MetadataModular) GfSpGetEndpointBySpAddress(
//	ctx context.Context,
//	req *types.GfSpGetEndpointBySpAddressRequest) (
//	resp *types.GfSpGetEndpointBySpAddressResponse, err error) {
//	ctx = log.Context(ctx, req)
//
//	sp, err := r.baseApp.GfSpDB().GetSpByAddress(req.SpAddress, spdb.OperatorAddressType)
//	if err != nil {
//		log.CtxErrorw(ctx, "failed to get sp", "error", err)
//		return
//	}
//
//	resp = &types.GfSpGetEndpointBySpAddressResponse{Endpoint: sp.Endpoint}
//	log.CtxInfow(ctx, "succeed to get endpoint by a sp address")
//	return resp, nil
//}
