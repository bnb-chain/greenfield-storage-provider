package gateway

import (
	"context"
	"encoding/xml"
	"errors"
	"net/http"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/model/job"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	servicetypes "github.com/bnb-chain/greenfield-storage-provider/service/types"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/bnb-chain/greenfield/types/s3util"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

// getBucketReadQuota handles the get bucket read quota request
func (g *Gateway) getBucketReadQuotaHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err            error
		errDescription *errorDescription
		reqContext     *requestContext
		addr           sdktypes.AccAddress
	)

	reqContext = newRequestContext(r)
	defer func() {
		if errDescription != nil {
			_ = errDescription.errorResponse(w, reqContext)
		}
		if errDescription != nil && errDescription.statusCode != http.StatusOK {
			log.Errorf("action(%v) statusCode(%v) %v", getBucketReadQuotaRouterName, errDescription.statusCode, reqContext.generateRequestDetail())
		} else {
			log.Infof("action(%v) statusCode(200) %v", getBucketReadQuotaRouterName, reqContext.generateRequestDetail())
		}
	}()

	if g.downloader == nil {
		log.Error("failed to get bucket read quota due to not config downloader")
		errDescription = NotExistComponentError
		return
	}

	if addr, err = g.verifySignature(reqContext); err != nil {
		log.Errorw("failed to verify signature", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}
	if err = g.checkAuthorization(reqContext, addr); err != nil {
		log.Errorw("failed to check authorization", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}
	resp, err := g.downloader.GetBucketReadQuota(context.Background(), reqContext.bucketInfo, reqContext.vars["year_month"])
	if err != nil {
		log.Errorw("failed to get bucket read quota", "error", err)
		errDescription = makeErrorDescription(err)
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
	}{
		Version:             model.GnfdResponseXMLVersion,
		BucketName:          reqContext.bucketInfo.GetBucketName(),
		BucketID:            util.Uint64ToString(reqContext.bucketInfo.Id.Uint64()),
		ReadQuotaSize:       resp.GetChargedQuotaSize(),
		SPFreeReadQuotaSize: resp.GetSpFreeQuotaSize(),
		ReadConsumedSize:    resp.GetConsumedSize(),
	}
	xmlBody, err := xml.Marshal(&xmlInfo)
	if err != nil {
		log.Errorw("failed to marshal xml", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}
	w.Header().Set(model.ContentTypeHeader, model.ContentTypeXMLHeaderValue)
	w.Header().Set(model.GnfdRequestIDHeader, reqContext.requestID)
	if _, err = w.Write(xmlBody); err != nil {
		log.Errorw("failed to write body", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}
	log.Debugw("get bucket quota", "xml_info", xmlInfo)
}

// listBucketReadRecord handles the list bucket read records request
func (g *Gateway) listBucketReadRecordHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err              error
		errDescription   *errorDescription
		reqContext       *requestContext
		addr             sdktypes.AccAddress
		startTimestampUs int64
		endTimestampUs   int64
		maxRecordNum     int64
	)

	reqContext = newRequestContext(r)
	defer func() {
		if errDescription != nil {
			_ = errDescription.errorResponse(w, reqContext)
		}
		if errDescription != nil && errDescription.statusCode != http.StatusOK {
			log.Errorf("action(%v) statusCode(%v) %v", listBucketReadRecordRouterName, errDescription.statusCode, reqContext.generateRequestDetail())
		} else {
			log.Infof("action(%v) statusCode(200) %v", listBucketReadRecordRouterName, reqContext.generateRequestDetail())
		}
	}()

	if g.downloader == nil {
		log.Error("failed to list bucket read record due to not config downloader")
		errDescription = NotExistComponentError
		return
	}

	if addr, err = g.verifySignature(reqContext); err != nil {
		log.Errorw("failed to verify signature", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}
	if err = g.checkAuthorization(reqContext, addr); err != nil {
		log.Errorw("failed to check authorization", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}

	if startTimestampUs, err = util.StringToInt64(reqContext.vars["start_ts"]); err != nil {
		log.Errorw("failed to parse start_ts query", "error", err)
		errDescription = InvalidQuery
		return
	}
	if endTimestampUs, err = util.StringToInt64(reqContext.vars["end_ts"]); err != nil {
		log.Errorw("failed to parse end_ts query", "error", err)
		errDescription = InvalidQuery
		return
	}
	if maxRecordNum, err = util.StringToInt64(reqContext.vars["max_records"]); err != nil {
		log.Errorw("failed to parse max record num query", "error", err)
		errDescription = InvalidQuery
		return
	}
	if maxRecordNum > model.DefaultMaxListLimit || maxRecordNum < 0 {
		maxRecordNum = model.DefaultMaxListLimit
	}
	resp, err := g.downloader.ListBucketReadRecord(context.Background(), reqContext.bucketInfo, startTimestampUs, endTimestampUs, maxRecordNum)
	if err != nil {
		log.Errorw("failed to list bucket read record", "error", err)
		errDescription = makeErrorDescription(err)
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
	for _, r := range resp.ReadRecords {
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
		Version:              model.GnfdResponseXMLVersion,
		NextStartTimestampUs: resp.GetNextStartTimestampUs(),
		ReadRecords:          xmlRecords,
	}
	xmlBody, err := xml.Marshal(&xmlInfo)
	if err != nil {
		log.Errorw("failed to marshal xml", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}

	w.Header().Set(model.ContentTypeHeader, model.ContentTypeXMLHeaderValue)
	w.Header().Set(model.GnfdRequestIDHeader, reqContext.requestID)
	if _, err = w.Write(xmlBody); err != nil {
		log.Errorw("failed to write body", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}
	log.Debugw("list bucket read records", "xml_info", xmlInfo)
}

// queryUploadProgressHandler handles the query upload object progress request
func (g *Gateway) queryUploadProgressHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err                 error
		errDescription      *errorDescription
		reqContext          *requestContext
		addr                sdktypes.AccAddress
		jobState            servicetypes.JobState
		jobStateDescription string
	)

	reqContext = newRequestContext(r)
	defer func() {
		if errDescription != nil {
			_ = errDescription.errorResponse(w, reqContext)
		}
		if errDescription != nil && errDescription.statusCode != http.StatusOK {
			log.Errorf("action(%v) statusCode(%v) %v", queryUploadProgressRouterName, errDescription.statusCode, reqContext.generateRequestDetail())
		} else {
			log.Infof("action(%v) statusCode(200) %v", queryUploadProgressRouterName, reqContext.generateRequestDetail())
		}
	}()

	if g.uploader == nil {
		log.Error("failed to query upload progress due to not config uploader")
		errDescription = NotExistComponentError
		return
	}

	if err = s3util.CheckValidBucketName(reqContext.bucketName); err != nil {
		log.Errorw("failed to check bucket name", "bucket_name", reqContext.bucketName, "error", err)
		errDescription = InvalidBucketName
		return
	}
	if err = s3util.CheckValidObjectName(reqContext.objectName); err != nil {
		log.Errorw("failed to check object name", "object_name", reqContext.objectName, "error", err)
		errDescription = InvalidKey
		return
	}

	if addr, err = reqContext.verifySignature(); err != nil {
		log.Errorw("failed to verify signature", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}
	if err = g.checkAuthorization(reqContext, addr); err != nil {
		log.Errorw("failed to check authorization", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}
	if reqContext.objectInfo.GetObjectStatus() == storagetypes.OBJECT_STATUS_SEALED {
		jobStateDescription = job.StateToDescription(servicetypes.JobState_JOB_STATE_SEAL_OBJECT_DONE)
	} else {
		jobState, err = g.uploader.QueryUploadProgress(context.Background(), reqContext.objectInfo.Id.Uint64())
		if err != nil {
			err = merrors.GRPCErrorToInnerError(err)
			if errors.Is(err, merrors.ErrNoSuchObject) {
				jobStateDescription = job.ToReadableDescription[job.UploadProgressMetaCreated]
			} else {
				log.Errorw("failed to query upload progress", "error", err)
				errDescription = makeErrorDescription(err)
				return
			}
		} else {
			jobStateDescription = job.StateToDescription(jobState)
		}
	}
	var xmlInfo = struct {
		XMLName             xml.Name `xml:"QueryUploadProgress"`
		Version             string   `xml:"version,attr"`
		ProgressDescription string   `xml:"ProgressDescription"`
	}{
		Version:             model.GnfdResponseXMLVersion,
		ProgressDescription: jobStateDescription,
	}
	xmlBody, err := xml.Marshal(&xmlInfo)
	if err != nil {
		log.Errorw("failed to marshal xml", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}
	w.Header().Set(model.ContentTypeHeader, model.ContentTypeXMLHeaderValue)
	w.Header().Set(model.GnfdRequestIDHeader, reqContext.requestID)
	if _, err = w.Write(xmlBody); err != nil {
		log.Errorw("failed to write body", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}
	log.Debugw("query upload progress", "xml_info", xmlInfo)
}
