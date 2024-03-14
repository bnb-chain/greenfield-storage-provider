package gater

import (
	"context"
	"encoding/xml"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	modelgateway "github.com/bnb-chain/greenfield-storage-provider/model/gateway"
	metadatatypes "github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

// getBucketReadQuotaHandler handles the get bucket read quota request.
func (g *GateModular) getBucketReadQuotaHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err                                 error
		authenticated                       bool
		bucketInfo                          *storagetypes.BucketInfo
		charge, free, consume, free_consume uint64
	)
	startTime := time.Now()
	defer func() {
		if err != nil {
			modelgateway.MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
			metrics.ReqCounter.WithLabelValues(GatewayTotalFailure).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalFailure).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(GatewayTotalSuccess).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalSuccess).Observe(time.Since(startTime).Seconds())
		}
	}()

	ctx := context.Background()
	vars := mux.Vars(r)
	bucketName := vars["bucket"]
	yearMonth := vars["year_month"]
	authenticated, err = g.baseApp.GfSpClient().VerifyAuthentication(ctx,
		coremodule.AuthOpTypeGetBucketQuota, "", bucketName, "")
	if err != nil {
		log.CtxErrorw(ctx, "failed to verify authentication", "error", err)
		return
	}
	if !authenticated {
		log.CtxErrorw(ctx, "no permission to operate")
		err = ErrNoPermission
		return
	}

	bucketInfo, err = g.baseApp.Consensus().QueryBucketInfo(ctx, bucketName)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get bucket info from consensus", "error", err)
		err = ErrConsensusWithDetail("failed to get bucket info from consensus, error: " + err.Error())
		return
	}

	charge, free, consume, free_consume, err = g.baseApp.GfSpClient().GetBucketReadQuota(
		ctx, bucketInfo, yearMonth)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get bucket read quota", "error", err)
		return
	}

	var xmlInfo = struct {
		XMLName             xml.Name `xml:"GetReadQuotaResult"`
		Version             string   `xml:"version,attr"`
		BucketName          string   `xml:"BucketName"`
		BucketID            string   `xml:"BucketID"`
		ReadQuotaSize       uint64   `xml:"ReadQuotaSize"`
		SPFreeReadQuotaSize uint64   `xml:"SPFreeReadQuotaSize"`
		ReadConsumedSize    uint64   `xml:"ReadConsumedSize"`
		FreeConsumedSize    uint64   `xml:"FreeConsumedSize"`
	}{
		Version:             GnfdResponseXMLVersion,
		BucketName:          bucketInfo.GetBucketName(),
		BucketID:            util.Uint64ToString(bucketInfo.Id.Uint64()),
		ReadQuotaSize:       charge,
		SPFreeReadQuotaSize: free,
		ReadConsumedSize:    consume,
		FreeConsumedSize:    free_consume,
	}
	xmlBody, err := xml.Marshal(&xmlInfo)
	if err != nil {
		log.Errorw("failed to marshal xml", "error", err)
		err = ErrEncodeResponseWithDetail("failed to marshal xml, error: " + err.Error())
		return
	}
	w.Header().Set(ContentTypeHeader, ContentTypeXMLHeaderValue)
	if _, err = w.Write(xmlBody); err != nil {
		log.Errorw("failed to write body", "error", err)
		err = ErrEncodeResponseWithDetail("failed to write body, error: " + err.Error())
		return
	}
	log.CtxDebugw(ctx, "succeed to get bucket quota", "xml_info", xmlInfo)
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
			reqCtx.SetHTTPCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			modelgateway.MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
			metrics.ReqCounter.WithLabelValues(GatewayTotalFailure).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalFailure).Observe(time.Since(startTime).Seconds())
		} else {
			reqCtx.SetHTTPCode(http.StatusOK)
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
		err = ErrConsensusWithDetail("failed to get bucket info from consensus, error: " + err.Error())
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
		log.CtxErrorw(reqCtx.Context(), "failed to parse max_records num query", "error", err)
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
	for _, record := range records {
		xmlRecords = append(xmlRecords, ReadRecord{
			ObjectName:         record.GetObjectName(),
			ObjectID:           util.Uint64ToString(record.GetObjectId()),
			ReadAccountAddress: record.GetAccountAddress(),
			ReadTimestampUs:    record.GetTimestampUs(),
			ReadSize:           record.GetReadSize(),
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
		err = ErrEncodeResponseWithDetail("failed to marshal xml, error: " + err.Error())
		return
	}

	w.Header().Set(ContentTypeHeader, ContentTypeXMLHeaderValue)
	if _, err = w.Write(xmlBody); err != nil {
		log.Errorw("failed to write body", "error", err)
		err = ErrEncodeResponseWithDetail("failed to write body, error: " + err.Error())
		return
	}
	log.Debugw("succeed to list bucket read records", "xml_info", xmlInfo)
}

// queryBucketMigrationProgressHandler handles the query bucket migration  progress request.
func (g *GateModular) queryBucketMigrationProgressHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err            error
		reqCtx         *RequestContext
		authenticated  bool
		bucketInfo     *storagetypes.BucketInfo
		errDescription string
		migrationState int32
		progressMeta   *gfspserver.MigrateBucketProgressMeta
		migratedBytes  uint64
	)
	startTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			reqCtx.SetHTTPCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			modelgateway.MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
			metrics.ReqCounter.WithLabelValues(GatewayTotalFailure).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalFailure).Observe(time.Since(startTime).Seconds())
		} else {
			reqCtx.SetHTTPCode(http.StatusOK)
			metrics.ReqCounter.WithLabelValues(GatewayTotalSuccess).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalSuccess).Observe(time.Since(startTime).Seconds())
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
	}()

	if reqCtx, err = NewRequestContext(r, g); err != nil {
		return
	}
	authenticated, err = g.baseApp.GfSpClient().VerifyAuthentication(reqCtx.Context(),
		coremodule.AuthOpTypeQueryBucketMigrationProgress, reqCtx.Account(), reqCtx.bucketName, reqCtx.objectName)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to verify authentication", "error", err)
		return
	}
	if !authenticated {
		log.CtxErrorw(reqCtx.Context(), "no permission to operate")
		err = ErrNoPermission
		return
	}

	if bucketInfo, err = g.baseApp.Consensus().QueryBucketInfo(reqCtx.Context(), reqCtx.bucketName); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get bucket info from consensus", "error", err)
		err = ErrConsensusWithDetail("failed to get bucket info from consensus, error: " + err.Error())
		return
	}

	if progressMeta, err = g.baseApp.GfSpClient().GetMigrateBucketProgress(reqCtx.Context(), bucketInfo.Id.Uint64()); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get bucket migration job state", "error", err)
		if !strings.Contains(err.Error(), "record not found") {
			return
		}
		err = nil
	} else {
		migrationState = int32(progressMeta.GetMigrateState())
		migratedBytes = progressMeta.GetMigratedBytes()
	}

	var xmlInfo = struct {
		XMLName          xml.Name `xml:"QueryMigrationProgress"`
		Version          string   `xml:"version,attr"`
		ErrorDescription string   `xml:"ErrorDescription"`
		MigratedBytes    uint64   `xml:"MigratedBytes"`
		MigrationState   uint32   `xml:"MigrationState"`
	}{
		Version:          GnfdResponseXMLVersion,
		ErrorDescription: errDescription,
		MigratedBytes:    migratedBytes,
		MigrationState:   uint32(migrationState),
	}
	xmlBody, err := xml.Marshal(&xmlInfo)
	if err != nil {
		log.Errorw("failed to marshal xml", "error", err)
		err = ErrEncodeResponseWithDetail("failed to marshal xml, error: " + err.Error())
		return
	}
	w.Header().Set(ContentTypeHeader, ContentTypeXMLHeaderValue)
	if _, err = w.Write(xmlBody); err != nil {
		log.Errorw("failed to write body", "error", err)
		err = ErrEncodeResponseWithDetail("failed to write body, error: " + err.Error())
		return
	}
	log.Debugw("succeed to query bucket migration progress", "xml_info", xmlInfo)
}

// listBucketReadQuotaHandler handles the lost bucket read quota request.
func (g *GateModular) listBucketReadQuotaHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err           error
		offset, limit uint64
		result        map[string]*metadatatypes.GfSpGetBucketReadQuotaResponse
	)
	startTime := time.Now()
	defer func() {
		if err != nil {
			modelgateway.MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
			metrics.ReqCounter.WithLabelValues(GatewayTotalFailure).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalFailure).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(GatewayTotalSuccess).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalSuccess).Observe(time.Since(startTime).Seconds())
		}
	}()

	ctx := context.Background()
	vars := mux.Vars(r)
	yearMonth := vars["year_month"]
	offsetStr := vars["offset"]
	offset, err = strconv.ParseUint(offsetStr, 10, 32)
	if err != nil {
		log.Errorw("failed to ParseUint offset", "error", err)
		err = ErrInvalidQuery
		return
	}
	limitStr := vars["limit"]
	limit, err = strconv.ParseUint(limitStr, 10, 32)
	if err != nil {
		log.Errorw("failed to ParseUint limit", "error", err)
		err = ErrInvalidQuery
		return
	}
	if limit > 500 || limit == 0 {
		log.Errorw("limit is too large or limit equals 0")
		err = ErrInvalidQuery
		return
	}

	result, err = g.baseApp.GfSpClient().ListBucketReadQuota(
		ctx, yearMonth, uint32(offset), uint32(limit))
	if err != nil {
		log.CtxErrorw(ctx, "failed to get bucket read quota", "error", err)
		return
	}

	var xmlInfo = struct {
		XMLName xml.Name `xml:"GetReadQuotaResult"`
		Version string   `xml:"version,attr"`
		Result  map[string]*metadatatypes.GfSpGetBucketReadQuotaResponse
	}{
		Version: GnfdResponseXMLVersion,
		Result:  result,
	}

	xmlBody, err := xml.Marshal(&xmlInfo)
	if err != nil {
		log.Errorw("failed to marshal xml", "error", err)
		err = ErrEncodeResponseWithDetail("failed to marshal xml, error: " + err.Error())
		return
	}
	w.Header().Set(ContentTypeHeader, ContentTypeXMLHeaderValue)
	if _, err = w.Write(xmlBody); err != nil {
		log.Errorw("failed to write body", "error", err)
		err = ErrEncodeResponseWithDetail("failed to write body, error: " + err.Error())
		return
	}
	log.CtxDebugw(ctx, "succeed to get bucket quota", "xml_info", xmlInfo)
}
