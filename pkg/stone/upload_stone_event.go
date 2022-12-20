package stone

import (
	types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	"github.com/looplab/fsm"
)

/*
 UploadPayLoadStoneFSM state transitions:
		"JOB_STATE_CREATE_OBJECT_DONE" -> "JOB_STATE_UPLOAD_PAYLOAD_INIT" [ label = "UploadPayLoadInitEvent" ];
		"JOB_STATE_UPLOAD_PAYLOAD_INIT" -> "JOB_STATE_UPLOAD_PAYLOAD_DOING" [ label = "UploadPayloadDoingEvent" ];
		"JOB_STATE_UPLOAD_PAYLOAD_DOING" -> "JOB_STATE_UPLOAD_PAYLOAD_DONE" [ label = "UploadPayloadDoneEvent" ];
    	"JOB_STATE_UPLOAD_PAYLOAD_DONE" -> "JOB_STATE_SEAL_OBJECT_INIT" [ label = "SealObjectInitEvent" ];
    	"JOB_STATE_SEAL_OBJECT_INIT" -> "JOB_STATE_SEAL_OBJECT_SIGNATURE_DONE" [ label = "SealObjectSignatureDoneEvent" ];
    	"JOB_STATE_SEAL_OBJECT_SIGNATURE_DONE" -> "JOB_STATE_SEAL_OBJECT_TX_DOING" [ label = "SealObjectDoingEvent" ];
    	"JOB_STATE_SEAL_OBJECT_TX_DOING" -> "JOB_STATE_SEAL_OBJECT_DONE" [ label = "SealObjectDoneEvent" ];
 UploadPrimaryStoneFSM state transitions(UploadPayLoadStoneFSM state is JOB_STATE_UPLOAD_PAYLOAD_DOING):
		"JOB_STATE_INIT_UNSPECIFIED" -> "JOB_STATE_UPLOAD_PAYLOAD_INIT" [ label = "UploadPrimaryInitEvent" ];
		"JOB_STATE_UPLOAD_PAYLOAD_INIT" -> "JOB_STATE_UPLOAD_PAYLOAD_DOING" [ label = "UploadPrimaryDoingEvent" ];
		"JOB_STATE_UPLOAD_PAYLOAD_DOING" -> "JOB_STATE_UPLOAD_PAYLOAD_DOING" [ label = "UploadPieceDoneEvent" ];
		"JOB_STATE_UPLOAD_PAYLOAD_DOING" -> "JOB_STATE_UPLOAD_PAYLOAD_DONE" [ label = "UploadPrimaryDoneEvent" ];
 UploadSecondaryStoneFSM state transitions(UploadPayLoadStoneFSM state is JOB_STATE_UPLOAD_PAYLOAD_DOING):
		"JOB_STATE_INIT_UNSPECIFIED" -> "JOB_STATE_ALLOC_SECONDARY_INIT" [ label = "AllocSecondaryInitEvent" ];
		"JOB_STATE_ALLOC_SECONDARY_INIT" -> "JOB_STATE_ALLOC_SECONDARY_DOING" [ label = "AllocSecondaryDoingEvent" ];
		"JOB_STATE_ALLOC_SECONDARY_DOING" -> "JOB_STATE_ALLOC_SECONDARY_DONE" [ label = "AllocSecondaryDoneEvent" ];
    	"JOB_STATE_ALLOC_SECONDARY_DONE" -> "JOB_STATE_UPLOAD_PAYLOAD_INIT" [ label = "UploadSecondaryInitEvent" ];
    	"JOB_STATE_UPLOAD_PAYLOAD_INIT" -> "JOB_STATE_UPLOAD_PAYLOAD_DOING" [ label = "UploadSecondaryDoingEvent" ];
		"JOB_STATE_UPLOAD_PAYLOAD_INIT" -> "JOB_STATE_UPLOAD_PAYLOAD_DOING" [ label = "UploadPieceNotifyEvent" ];
    	"JOB_STATE_UPLOAD_PAYLOAD_DOING" -> "JOB_STATE_UPLOAD_PAYLOAD_DOING" [ label = "UploadPieceDoneEvent" ];
    	"JOB_STATE_UPLOAD_PAYLOAD_DOING" -> "JOB_STATE_UPLOAD_PAYLOAD_DONE" [ label = "UploadSecondaryDoneEvent" ];
 UploadPayLoadStoneFSM state transfer JOB_STATE_UPLOAD_PAYLOAD_DONE
*/

// define FSM Event Name
var (
	UploadPayloadInitEvent       string = "UploadPayLoadInitEvent"
	UploadPayloadDoingEvent      string = "UploadPayloadDoingEvent"
	UploadPayloadDoneEvent       string = "UploadPayloadDoneEvent"
	SealObjectInitEvent          string = "SealObjectInitEvent"
	SealObjectSignatureDoneEvent string = "SealObjectSignatureDoneEvent"
	SealObjectDoingEvent         string = "SealObjectDoingEvent"
	SealObjectDoneEvent          string = "SealObjectDoneEvent"
	InterruptEvent               string = "InterruptEvent"

	UploadPrimaryInitEvent  string = "UploadPrimaryInitEvent"
	UploadPrimaryDoingEvent string = "UploadPrimaryDoingEvent"
	UploadPrimaryDoneEvent  string = "UploadPrimaryDoneEvent"
	UploadPieceDoneEvent    string = "UploadPieceDoneEvent"

	AllocSecondaryInitEvent   string = "AllocSecondaryInitEvent"
	AllocSecondaryDoingEvent  string = "AllocSecondaryDoingEvent"
	AllocSecondaryDoneEvent   string = "AllocSecondaryDoneEvent"
	UploadSecondaryInitEvent  string = "UploadSecondaryInitEvent"
	UploadSecondaryDoingEvent string = "UploadSecondaryDoingEvent"
	UploadPieceNotifyEvent    string = "UploadPieceNotifyEvent"
	UploadSecondaryDoneEvent  string = "UploadSecondaryDoneEvent"
)

// define FSM event transitions

var UploadPayloadFsmEvent = fsm.Events{
	{Name: UploadPayloadInitEvent, Src: []string{types.JOB_STATE_CREATE_OBJECT_DONE}, Dst: types.JOB_STATE_UPLOAD_PAYLOAD_INIT},
	{Name: UploadPayloadDoingEvent, Src: []string{types.JOB_STATE_UPLOAD_PAYLOAD_INIT}, Dst: types.JOB_STATE_UPLOAD_PAYLOAD_DOING},
	{Name: UploadPayloadDoneEvent, Src: []string{types.JOB_STATE_UPLOAD_PAYLOAD_DOING}, Dst: types.JOB_STATE_UPLOAD_PAYLOAD_DONE},
	{Name: SealObjectInitEvent, Src: []string{types.JOB_STATE_UPLOAD_PAYLOAD_DONE}, Dst: types.JOB_STATE_SEAL_OBJECT_INIT},
	{Name: SealObjectSignatureDoneEvent, Src: []string{types.JOB_STATE_SEAL_OBJECT_INIT}, Dst: types.JOB_STATE_SEAL_OBJECT_SIGNATURE_DONE},
	{Name: SealObjectDoingEvent, Src: []string{types.JOB_STATE_SEAL_OBJECT_SIGNATURE_DONE}, Dst: types.JOB_STATE_SEAL_OBJECT_TX_DOING},
	{Name: SealObjectDoneEvent, Src: []string{types.JOB_STATE_SEAL_OBJECT_TX_DOING}, Dst: types.JOB_STATE_SEAL_OBJECT_DONE},
	{Name: InterruptEvent, Src: []string{}, Dst: types.JOB_STATE_ERROR},
}
var UploadPrimaryFsmEvent = fsm.Events{
	{Name: UploadPrimaryInitEvent, Src: []string{types.JOB_STATE_INIT_UNSPECIFIED}, Dst: types.JOB_STATE_UPLOAD_PAYLOAD_INIT},
	{Name: UploadPrimaryDoingEvent, Src: []string{types.JOB_STATE_UPLOAD_PAYLOAD_INIT}, Dst: types.JOB_STATE_UPLOAD_PAYLOAD_DOING},
	{Name: UploadPieceDoneEvent, Src: []string{types.JOB_STATE_UPLOAD_PAYLOAD_DOING}, Dst: types.JOB_STATE_UPLOAD_PAYLOAD_DOING},
	{Name: UploadPrimaryDoneEvent, Src: []string{types.JOB_STATE_UPLOAD_PAYLOAD_DOING}, Dst: types.JOB_STATE_UPLOAD_PAYLOAD_DONE},
	{Name: InterruptEvent, Src: []string{}, Dst: types.JOB_STATE_ERROR},
}
var UploadSecondaryFsmEvent = fsm.Events{
	{Name: AllocSecondaryInitEvent, Src: []string{types.JOB_STATE_INIT_UNSPECIFIED}, Dst: types.JOB_STATE_ALLOC_SECONDARY_INIT},
	{Name: AllocSecondaryDoingEvent, Src: []string{types.JOB_STATE_ALLOC_SECONDARY_INIT}, Dst: types.JOB_STATE_ALLOC_SECONDARY_DOING},
	{Name: AllocSecondaryDoneEvent, Src: []string{types.JOB_STATE_ALLOC_SECONDARY_DOING}, Dst: types.JOB_STATE_ALLOC_SECONDARY_DONE},
	{Name: UploadSecondaryInitEvent, Src: []string{types.JOB_STATE_ALLOC_SECONDARY_DONE}, Dst: types.JOB_STATE_UPLOAD_PAYLOAD_INIT},
	{Name: UploadSecondaryDoingEvent, Src: []string{types.JOB_STATE_UPLOAD_PAYLOAD_INIT}, Dst: types.JOB_STATE_UPLOAD_PAYLOAD_DOING},
	{Name: UploadPieceDoneEvent, Src: []string{types.JOB_STATE_UPLOAD_PAYLOAD_DOING}, Dst: types.JOB_STATE_UPLOAD_PAYLOAD_DOING},
	{Name: UploadPieceNotifyEvent, Src: []string{types.JOB_STATE_UPLOAD_PAYLOAD_DOING}, Dst: types.JOB_STATE_UPLOAD_PAYLOAD_DOING},
	{Name: UploadSecondaryDoneEvent, Src: []string{types.JOB_STATE_UPLOAD_PAYLOAD_DOING}, Dst: types.JOB_STATE_UPLOAD_PAYLOAD_DONE},
	{Name: InterruptEvent, Src: []string{}, Dst: types.JOB_STATE_ERROR},
}

// define FSM action, the Action associated with callback
var (
	ActionBeforeEvent                       string = "before_"
	ActionLeaveState                        string = "leave_"
	ActionEnterState                        string = "enter_"
	ActionAfterEvent                        string = "after_"
	ActionEnterStateUploadPayloadInit       string = ActionEnterState + types.JOB_STATE_UPLOAD_PAYLOAD_INIT
	ActionEnterStateUploadPayloadDoing      string = ActionEnterState + types.JOB_STATE_UPLOAD_PAYLOAD_DOING
	ActionEnterStateUploadPayloadDone       string = ActionEnterState + types.JOB_STATE_UPLOAD_PAYLOAD_DONE
	ActionEnterStateSealObjectInit          string = ActionEnterState + types.JOB_STATE_SEAL_OBJECT_INIT
	ActionEnterStateSealObjectSignatureDone string = ActionEnterState + types.JOB_STATE_SEAL_OBJECT_SIGNATURE_DONE
	ActionLeaveStateSealObjectSignatureDone string = ActionLeaveState + types.JOB_STATE_SEAL_OBJECT_SIGNATURE_DONE
	ActionEnterStateSealObjectDone          string = ActionEnterState + types.JOB_STATE_SEAL_OBJECT_DONE
	ActionInterruptEvent                    string = types.JOB_STATE_ERROR
	ActionBeforeEventAll                    string = ActionBeforeEvent + "event"
	ActionAfterEventAll                     string = ActionAfterEvent + "event"
	ActionAfterEventUploadPieceDoneEvent    string = ActionAfterEvent + UploadPieceDoneEvent
	ActionAfterEventUploadPieceNotifyEvent  string = ActionAfterEvent + UploadPieceNotifyEvent
	ActionEnterStateAllocSecondaryInit      string = ActionEnterState + types.JOB_STATE_ALLOC_SECONDARY_INIT
	ActionEnterStateAllocSecondaryDoing     string = ActionEnterState + types.JOB_STATE_ALLOC_SECONDARY_DOING
	ActionEnterStateAllocSecondaryDone      string = ActionEnterState + types.JOB_STATE_ALLOC_SECONDARY_DONE
)

// define FSM callback of the action

var UploadPayLoadFsmCallBack = fsm.Callbacks{
	ActionEnterStateUploadPayloadInit:       InitUploadPayloadJobContextFromDB,
	ActionEnterStateUploadPayloadDoing:      StartUploadPrimaryAndSecondaryJob,
	ActionEnterStateUploadPayloadDone:       UpdateUploadPayloadJobStateDoneToDB,
	ActionEnterStateSealObjectInit:          CreateIntegrityHashJob,
	ActionEnterStateSealObjectSignatureDone: UpdateIntegrityHashToDB,
	ActionLeaveStateSealObjectSignatureDone: CreateSealObjectJob,
	ActionEnterStateSealObjectDone:          UpdateUploadPayloadJobStateSealDoneToDB,
	ActionInterruptEvent:                    InterruptUploadPayloadJob,
	ActionBeforeEventAll:                    InspectUploadPayloadJobBeforeEvent,
	ActionAfterEventAll:                     InspectUploadPayloadJobAfterEvent,
}
var UploadPrimaryFsmCallBack = fsm.Callbacks{
	ActionEnterStateUploadPayloadInit:    InitUploadPrimaryJob,
	ActionEnterStateUploadPayloadDoing:   PopUploadPrimaryJob,
	ActionAfterEventUploadPieceDoneEvent: DonePrimaryPieceJob,
	ActionEnterStateUploadPayloadDone:    UpdateUploadPrimaryJobStateToDB,
	ActionBeforeEventAll:                 InspectUploadPrimaryJobBeforeEvent,
	ActionAfterEventAll:                  InspectUploadPrimaryJobAfterEvent,
}

var UploadSecondaryFsmCallBack = fsm.Callbacks{
	ActionEnterStateAllocSecondaryInit:   CreateAllocSecondaryJob,
	ActionEnterStateAllocSecondaryDoing:  ConsumeEvent,
	ActionEnterStateAllocSecondaryDone:   InitUploadSecondaryJob,
	ActionEnterStateUploadPayloadInit:    ConsumeEvent,
	ActionEnterStateUploadPayloadDoing:   PopUploadSecondaryJob,
	ActionAfterEventUploadPieceDoneEvent: DoneSecondaryPieceJob,
	ActionAfterEventUploadPieceDoneEvent: PopUploadSecondaryJobBySegmentId,
	ActionEnterStateUploadPayloadDone:    UpdateUploadSecondaryStateToDB,
	ActionBeforeEventAll:                 InspectUploadSecondaryJobBeforeEvent,
	ActionAfterEventAll:                  InspectUploadSecondaryJobAfterEvent,
}
