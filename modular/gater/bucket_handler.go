package gater

import (
	"encoding/xml"
	"errors"
	"net/http"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	metadatatypes "github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"gorm.io/gorm"
)

// getBucketReadQuotaHandler handles the get bucket read quota request.
func (g *GateModular) getBucketReadQuotaHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err                   error
		reqCtx                *RequestContext
		authenticated         bool
		bucketInfo            *storagetypes.BucketInfo
		charge, free, consume uint64
		bucketTraffic         *spdb.BucketTraffic
	)
	startTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			reqCtx.SetHttpCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
			metrics.ReqCounter.WithLabelValues(GatewayTotalFailure).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalFailure).Observe(time.Since(startTime).Seconds())
		} else {
			reqCtx.SetHttpCode(http.StatusOK)
			metrics.ReqCounter.WithLabelValues(GatewayTotalSuccess).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalSuccess).Observe(time.Since(startTime).Seconds())
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
	}()

	reqCtx, err = NewRequestContext(r, g)
	if err != nil {
		return
	}
	authenticated, err = g.baseApp.GfSpClient().VerifyAuthentication(reqCtx.Context(),
		coremodule.AuthOpTypeGetBucketQuota, reqCtx.Account(), reqCtx.bucketName, "")
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to verify authentication", "error", err)
		return
	}
	if !authenticated {
		log.CtxErrorw(reqCtx.Context(), "no permission to operate")
		err = ErrNoPermission
		return
	}

	bucketInfo, err = g.baseApp.Consensus().QueryBucketInfo(reqCtx.Context(), reqCtx.bucketName)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get bucket info from consensus", "error", err)
		err = ErrConsensus
		return
	}

	bucketTraffic, err = g.baseApp.GfSpDB().GetBucketTraffic(
		bucketInfo.Id.Uint64())
	// if the traffic table has not been created and initialized yet, return the chain info
	if errors.Is(err, gorm.ErrRecordNotFound) || bucketTraffic == nil {
		free, err = g.baseApp.Consensus().QuerySPFreeQuota(reqCtx.Context(), g.baseApp.OperatorAddress())
		if err != nil {
			err = ErrConsensus
		}
		charge = bucketInfo.GetChargedReadQuota()
		log.CtxDebugw(reqCtx.Context(), "fail to get traffic info in db, return the chain meta", "free quota", free, "charged quota", charge)
		consume = 0
	} else {
		// if the traffic table has been created, return the db info from meta service
		charge, free, consume, err = g.baseApp.GfSpClient().GetBucketReadQuota(
			reqCtx.Context(), bucketInfo)
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to get bucket read quota", "error", err)
			return
		}
	}
	var xmlInfo = struct {
		XMLName             xml.Name `xml:"GetReadQuotaResult"`
		Version             string   `xml:"version,attr"`
		BucketName          string   `xml:"BucketName"`
		BucketID            string   `xml:"BucketID"`
		ReadQuotaSize       uint64   `xml:"ChargedQuotaSize"`
		SPFreeReadQuotaSize uint64   `xml:"SPFreeReadQuotaSize"`
		ReadConsumedSize    uint64   `xml:"ReadConsumedSize"`
	}{
		Version:             GnfdResponseXMLVersion,
		BucketName:          bucketInfo.GetBucketName(),
		BucketID:            util.Uint64ToString(bucketInfo.Id.Uint64()),
		ReadQuotaSize:       charge,
		SPFreeReadQuotaSize: free,
		ReadConsumedSize:    consume,
	}
	xmlBody, err := xml.Marshal(&xmlInfo)
	if err != nil {
		log.Errorw("failed to marshal xml", "error", err)
		err = ErrEncodeResponse
		return
	}
	w.Header().Set(ContentTypeHeader, ContentTypeXMLHeaderValue)
	if _, err = w.Write(xmlBody); err != nil {
		log.Errorw("failed to write body", "error", err)
		err = ErrEncodeResponse
		return
	}
	log.CtxDebugw(reqCtx.Context(), "succeed to get bucket quota", "xml_info", xmlInfo)
}

// listBucketReadRecordHandler handles list bucket read record request.
func (g *GateModular) listBucketReadRecordHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err              error
		reqCtx           *RequestContext
		authenticated    bool
		startTimestampUs int64
		endTimestampUs   int64
		maxRecordNum     int64
		records          []*metadatatypes.ReadRecord
		nextTimestampUs  int64
	)
	startTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			reqCtx.SetHttpCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
			metrics.ReqCounter.WithLabelValues(GatewayTotalFailure).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalFailure).Observe(time.Since(startTime).Seconds())
		} else {
			reqCtx.SetHttpCode(http.StatusOK)
			metrics.ReqCounter.WithLabelValues(GatewayTotalSuccess).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalSuccess).Observe(time.Since(startTime).Seconds())
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
	}()

	reqCtx, err = NewRequestContext(r, g)
	if err != nil {
		return
	}
	authenticated, err = g.baseApp.GfSpClient().VerifyAuthentication(reqCtx.Context(),
		coremodule.AuthOpTypeListBucketReadRecord, reqCtx.Account(), reqCtx.bucketName, "")
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to verify authentication", "error", err)
		return
	}
	if !authenticated {
		log.CtxErrorw(reqCtx.Context(), "no permission to operate")
		err = ErrNoPermission
		return
	}

	bucketInfo, err := g.baseApp.Consensus().QueryBucketInfo(reqCtx.Context(), reqCtx.bucketName)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get bucket info from consensus", "error", err)
		err = ErrConsensus
		return
	}

	startTimestampUs, err = util.StringToInt64(reqCtx.vars["start_ts"])
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse start_ts query", "error", err)
		err = ErrInvalidQuery
		return
	}
	endTimestampUs, err = util.StringToInt64(reqCtx.vars["end_ts"])
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse end_ts query", "error", err)
		err = ErrInvalidQuery
		return
	}
	maxRecordNum, err = util.StringToInt64(reqCtx.vars["max_records"])
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse max record num query", "error", err)
		err = ErrInvalidQuery
		return
	}
	if maxRecordNum > g.maxListReadQuota || maxRecordNum < 0 {
		maxRecordNum = g.maxListReadQuota
	}

	records, nextTimestampUs, err = g.baseApp.GfSpClient().ListBucketReadRecord(
		reqCtx.Context(), bucketInfo, startTimestampUs, endTimestampUs, maxRecordNum)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to list bucket read record", "error", err)
		return
	}

	type ReadRecord struct {
		XMLName            xml.Name `xml:"ReadRecord"`
		ObjectName         string   `xml:"ObjectName"`
		ObjectID           string   `xml:"ObjectID"`
		ReadAccountAddress string   `xml:"ReadAccountAddress"`
		ReadTimestampUs    int64    `xml:"ReadTimestampUs"`
		ReadSize           uint64   `xml:"ReadSize"`
	}
	xmlRecords := make([]ReadRecord, 0)
	for _, r := range records {
		xmlRecords = append(xmlRecords, ReadRecord{
			ObjectName:         r.GetObjectName(),
			ObjectID:           util.Uint64ToString(r.GetObjectId()),
			ReadAccountAddress: r.GetAccountAddress(),
			ReadTimestampUs:    r.GetTimestampUs(),
			ReadSize:           r.GetReadSize(),
		})
	}
	var xmlInfo = struct {
		XMLName              xml.Name     `xml:"GetBucketReadQuotaResult"`
		Version              string       `xml:"version,attr"`
		NextStartTimestampUs int64        `xml:"NextStartTimestampUs"`
		ReadRecords          []ReadRecord `xml:"ReadRecord"`
	}{
		Version:              GnfdResponseXMLVersion,
		NextStartTimestampUs: nextTimestampUs,
		ReadRecords:          xmlRecords,
	}
	xmlBody, err := xml.Marshal(&xmlInfo)
	if err != nil {
		log.Errorw("failed to marshal xml", "error", err)
		err = ErrEncodeResponse
		return
	}

	w.Header().Set(ContentTypeHeader, ContentTypeXMLHeaderValue)
	if _, err = w.Write(xmlBody); err != nil {
		log.Errorw("failed to write body", "error", err)
		err = ErrEncodeResponse
		return
	}
	log.Debugw("succeed to list bucket read records", "xml_info", xmlInfo)
}
