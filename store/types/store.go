package types

type readableUploadProgressType uint32

const (
	UploadProgressMetaCreated              readableUploadProgressType = 0
	UploadProgressDataUploading            readableUploadProgressType = 1
	UploadProgressDataReplicating          readableUploadProgressType = 2
	UploadProgressSealing                  readableUploadProgressType = 3
	UploadProgressSealed                   readableUploadProgressType = 4
	UploadProgressUploadToPrimaryFailed    readableUploadProgressType = 5
	UploadProgressAllocSecondaryFailed     readableUploadProgressType = 6
	UploadProgressReplicateSecondaryFailed readableUploadProgressType = 7
	UploadProgressSignObjectFailed         readableUploadProgressType = 8
	UploadProgressSealObjectFailed         readableUploadProgressType = 9
	UploadProgressObjectDiscontinued       readableUploadProgressType = 10
)

// ToReadableDescription convects readable type to the description string
var ToReadableDescription = map[readableUploadProgressType]string{
	UploadProgressMetaCreated:              "object meta is created onto the chain",
	UploadProgressDataUploading:            "object payload is uploading to the primary SP",
	UploadProgressDataReplicating:          "object payload is replicating to the secondary SPs in the background",
	UploadProgressSealing:                  "object meta is sealing onto the chain in the background",
	UploadProgressSealed:                   "object is succeed to upload and seal onto the chain",
	UploadProgressUploadToPrimaryFailed:    "something is wrong in uploading to primary",
	UploadProgressAllocSecondaryFailed:     "something is wrong in allocating secondary SPs",
	UploadProgressReplicateSecondaryFailed: "something is wrong in replicating secondary SPs",
	UploadProgressSignObjectFailed:         "something is wrong in signing the object",
	UploadProgressSealObjectFailed:         "something is wrong in sealing object",
	UploadProgressObjectDiscontinued:       "object has been discontinued",
}

// StateToProgressType convents inner state to the readable type
var StateToProgressType = map[TaskState]readableUploadProgressType{
	TaskState_TASK_STATE_INIT_UNSPECIFIED:       UploadProgressMetaCreated,
	TaskState_TASK_STATE_UPLOAD_OBJECT_DOING:    UploadProgressDataUploading,
	TaskState_TASK_STATE_UPLOAD_OBJECT_DONE:     UploadProgressDataReplicating,
	TaskState_TASK_STATE_UPLOAD_OBJECT_ERROR:    UploadProgressUploadToPrimaryFailed,
	TaskState_TASK_STATE_ALLOC_SECONDARY_DOING:  UploadProgressDataReplicating,
	TaskState_TASK_STATE_ALLOC_SECONDARY_DONE:   UploadProgressDataReplicating,
	TaskState_TASK_STATE_ALLOC_SECONDARY_ERROR:  UploadProgressAllocSecondaryFailed,
	TaskState_TASK_STATE_REPLICATE_OBJECT_DOING: UploadProgressDataReplicating,
	TaskState_TASK_STATE_REPLICATE_OBJECT_DONE:  UploadProgressDataReplicating,
	TaskState_TASK_STATE_REPLICATE_OBJECT_ERROR: UploadProgressReplicateSecondaryFailed,
	TaskState_TASK_STATE_SIGN_OBJECT_DOING:      UploadProgressSealing,
	TaskState_TASK_STATE_SIGN_OBJECT_DONE:       UploadProgressSealing,
	TaskState_TASK_STATE_SIGN_OBJECT_ERROR:      UploadProgressSignObjectFailed,
	TaskState_TASK_STATE_SEAL_OBJECT_DOING:      UploadProgressSealing,
	TaskState_TASK_STATE_SEAL_OBJECT_DONE:       UploadProgressSealed,
	TaskState_TASK_STATE_SEAL_OBJECT_ERROR:      UploadProgressSealObjectFailed,
	TaskState_TASK_STATE_OBJECT_DISCONTINUED:    UploadProgressObjectDiscontinued,
}

// StateToDescription convents state to description.
func StateToDescription(state TaskState) string {
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
