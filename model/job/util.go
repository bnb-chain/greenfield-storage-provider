package job

import (
	servicetypes "github.com/bnb-chain/greenfield-storage-provider/service/types"
)

type readableUploadProgressType uint32

const (
	UploadProgressMetaCreated     readableUploadProgressType = 0
	UploadProgressDataUploading   readableUploadProgressType = 1
	UploadProgressDataReplicating readableUploadProgressType = 2
	UploadProgressSealing         readableUploadProgressType = 3
	UploadProgressCompleted       readableUploadProgressType = 4
	UploadProgressFailed          readableUploadProgressType = 5
)

// ToReadableDescription convects readable type to the description string
var ToReadableDescription = map[readableUploadProgressType]string{
	UploadProgressMetaCreated:     "object meta is created in the chain, but the upload has been not started yet",
	UploadProgressDataUploading:   "object payload is uploading to the primary SP",
	UploadProgressDataReplicating: "object payload is replicating to the secondary SPs in the background",
	UploadProgressSealing:         "object meta is sealing onto the chain in the background",
	UploadProgressCompleted:       "object is succeed to upload",
	UploadProgressFailed:          "something is wrong in the upload process",
}

// StateToProgressType convents inner state to the readable type
var StateToProgressType = map[servicetypes.JobState]readableUploadProgressType{
	servicetypes.JobState_JOB_STATE_INIT_UNSPECIFIED:       UploadProgressDataUploading,
	servicetypes.JobState_JOB_STATE_UPLOAD_OBJECT_DOING:    UploadProgressDataUploading,
	servicetypes.JobState_JOB_STATE_UPLOAD_OBJECT_DONE:     UploadProgressDataReplicating,
	servicetypes.JobState_JOB_STATE_UPLOAD_OBJECT_ERROR:    UploadProgressFailed,
	servicetypes.JobState_JOB_STATE_ALLOC_SECONDARY_DOING:  UploadProgressDataReplicating,
	servicetypes.JobState_JOB_STATE_ALLOC_SECONDARY_DONE:   UploadProgressDataReplicating,
	servicetypes.JobState_JOB_STATE_ALLOC_SECONDARY_ERROR:  UploadProgressFailed,
	servicetypes.JobState_JOB_STATE_REPLICATE_OBJECT_DOING: UploadProgressDataReplicating,
	servicetypes.JobState_JOB_STATE_REPLICATE_OBJECT_DONE:  UploadProgressDataReplicating,
	servicetypes.JobState_JOB_STATE_REPLICATE_OBJECT_ERROR: UploadProgressFailed,
	servicetypes.JobState_JOB_STATE_SIGN_OBJECT_DOING:      UploadProgressSealing,
	servicetypes.JobState_JOB_STATE_SIGN_OBJECT_DONE:       UploadProgressSealing,
	servicetypes.JobState_JOB_STATE_SIGN_OBJECT_ERROR:      UploadProgressFailed,
	servicetypes.JobState_JOB_STATE_SEAL_OBJECT_DOING:      UploadProgressSealing,
	servicetypes.JobState_JOB_STATE_SEAL_OBJECT_DONE:       UploadProgressCompleted,
	servicetypes.JobState_JOB_STATE_SEAL_OBJECT_ERROR:      UploadProgressFailed,
}

// StateToDescription convents state to description.
func StateToDescription(state servicetypes.JobState) string {
	uploadProgressType, ok := StateToProgressType[state]
	if !ok {
		return state.String()
	}
	description, ok := ToReadableDescription[uploadProgressType]
	if !ok {
		return state.String()
	}
	return description
}
