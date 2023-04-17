package job

import servicetypes "github.com/bnb-chain/greenfield-storage-provider/service/types"

var stateToDescription = map[servicetypes.JobState]string{
	servicetypes.JobState_JOB_STATE_INIT_UNSPECIFIED:       "the job is in the init state",
	servicetypes.JobState_JOB_STATE_UPLOAD_OBJECT_DOING:    "upload job is running in the foreground process",
	servicetypes.JobState_JOB_STATE_UPLOAD_OBJECT_DONE:     "upload job has been completed in the foreground process",
	servicetypes.JobState_JOB_STATE_UPLOAD_OBJECT_ERROR:    "something is failed in the foreground upload process",
	servicetypes.JobState_JOB_STATE_ALLOC_SECONDARY_DOING:  "allocate sp is running in the foreground process",
	servicetypes.JobState_JOB_STATE_ALLOC_SECONDARY_DONE:   "allocate job has been completed in the foreground process",
	servicetypes.JobState_JOB_STATE_ALLOC_SECONDARY_ERROR:  "something is failed in the foreground allocating process",
	servicetypes.JobState_JOB_STATE_REPLICATE_OBJECT_DOING: "replicate job is running in the background process",
	servicetypes.JobState_JOB_STATE_REPLICATE_OBJECT_DONE:  "replicate job has been completed in the background process",
	servicetypes.JobState_JOB_STATE_REPLICATE_OBJECT_ERROR: "something is failed in the background replicating process",
	servicetypes.JobState_JOB_STATE_SIGN_OBJECT_DOING:      "sign is running in the background process",
	servicetypes.JobState_JOB_STATE_SIGN_OBJECT_DONE:       "sign has been completed in the background process",
	servicetypes.JobState_JOB_STATE_SIGN_OBJECT_ERROR:      "something is failed in background signing process",
	servicetypes.JobState_JOB_STATE_SEAL_OBJECT_DOING:      "seal is running in the background process",
	servicetypes.JobState_JOB_STATE_SEAL_OBJECT_DONE:       "seal has been completed in the background process, the object is succeed to upload",
	servicetypes.JobState_JOB_STATE_SEAL_OBJECT_ERROR:      "something is failed in the background sealing process",
}

// JobStateToDescription convents state to description.
func JobStateToDescription(state servicetypes.JobState) string {
	description, ok := stateToDescription[state]
	if !ok {
		return state.String()
	}
	return description
}
