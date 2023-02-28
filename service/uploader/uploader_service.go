package uploader

import (
	"context"
	"io"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
	servicetypes "github.com/bnb-chain/greenfield-storage-provider/service/types"
	types "github.com/bnb-chain/greenfield-storage-provider/service/uploader/types"
	"github.com/bnb-chain/greenfield-storage-provider/util/hash"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var _ types.UploaderServiceServer = &Uploader{}

// UploadObject upload an object payload data with object info.
func (uploader *Uploader) UploadObject(
	stream types.UploaderService_UploadObjectServer) error {

	var (
		objectInfo             = &storagetypes.ObjectInfo{}
		segmentInfo            = &servicetypes.SegmentInfo{}
		readSteam, writeStream = io.Pipe()
		errCh                  = make(chan error)
		resCh                  = make(chan uint32, 10)
	)

	defer func() {
		close(errCh)
		close(resCh)
		writeStream.Close()
		readSteam.Close()
	}()

	readPipeline := func() []byte {
		data := make([]byte, model.SegmentSize)
		curr := 0
		for {
			readN, err := readSteam.Read(data[curr:])
			if err != nil || curr == model.SegmentSize {
				break
			}
			curr += readN
		}
		return data[:curr]
	}
	uploadPayloadFun := func(segIdx uint32, data []byte) {
		key := piecestore.EncodeSegmentPieceKey(objectInfo.Id.Uint64(), segIdx)
		if err := uploader.pieceStore.PutSegment(key, data); err != nil {
			errCh <- err
		}
		if segIdx < 0 || int(segIdx) >= len(segmentInfo.GetCheckSum()) {
			errCh <- merrors.ErrIndexOutOfBounds
			return
		}
		segmentInfo.GetCheckSum()[segIdx] = hash.GenerateChecksum(data)
		resCh <- segIdx
	}

	//TODO:: query the upload object process in spdb, continue from breakpoint

	// read payload from gRPC stream and write pipeline
	go func() {
		//init := true
		for {
			req, err := stream.Recv()
			if err != nil || err != io.EOF {
				errCh <- err
				return
			}
			//if init {
			//	init = false
			//	objectInfo = req.GetObjectInfo()
			//	segments := util.ComputeSegmentCount(objectInfo.GetPayloadSize())
			//	segmentInfo.ObjectInfo = objectInfo
			//	segmentInfo.CheckSum = make([][]byte, segments)
			//	uploader.cache.Add(objectInfo.Id, segmentInfo)
			//}
			writeStream.Write(req.GetPayload())
			if err == io.EOF {
				return
			}
		}
	}()

	// read segments from pipeline and parallel upload to piece store
	go func() {
		var segmentIdx uint32 = 0
		for {
			data := readPipeline()
			if len(data) == 0 {
				return
			}
			go uploadPayloadFun(segmentIdx, data)
			segmentIdx++
		}
	}()

	// watch the result of upload segments or the error
	for {
		select {
		case err := <-errCh:
			// TODO:: achieve to sp db
			return err
		case <-resCh:
			segmentInfo.Completed++
			uploader.cache.Add(objectInfo.Id, segmentInfo)
			if int(segmentInfo.Completed) == len(segmentInfo.GetCheckSum()) {
				integrityHash, signature, err := uploader.signer.SignIntegrityHash(context.Background(), segmentInfo.GetCheckSum())
				if err != nil {
					go func() { errCh <- err }()
					continue
				}
				segmentInfo.IntegrityHash, segmentInfo.Signature = integrityHash, signature
				// TODO:: store to sp db for sealing object
				uploader.cache.Add(objectInfo.Id, segmentInfo)
				// TODO:: notify stone node to replicate object to other SPs
				goto finish
			}
		}
	}
finish:
	return nil
}

// QueryUploadingObject query an uploading object with object id from cache
func (uploader *Uploader) QueryUploadingObject(
	ctx context.Context,
	req *types.QueryUploadingObjectRequest) (
	resp *types.QueryUploadingObjectResponse, err error) {
	ctx = log.Context(ctx, req)
	objectId := req.GetObjectId()
	val, ok := uploader.cache.Get(objectId)
	if !ok {
		err = merrors.ErrCacheMiss
		return
	}
	resp.SegmentInfo = val.(*servicetypes.SegmentInfo)
	return
}
