package stone

import (
	"github.com/looplab/fsm"

	ptypesv1pb "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
)

/*
   "JOB_STATE_CREATE_OBJECT_DONE" -> "JOB_STATE_UPLOAD_PAYLOAD_INIT" [ label = "UploadPayloadInitEvent" ];
   "JOB_STATE_UPLOAD_PAYLOAD_INIT" -> "JOB_STATE_UPLOAD_PRIMARY_DOING" [ label = "UploadPrimaryDoingEvent" ];
   "JOB_STATE_UPLOAD_PRIMARY_DOING" -> "JOB_STATE_UPLOAD_PRIMARY_DOING" [ label = "UploadPrimaryPieceDoneEvent" ];
   "JOB_STATE_UPLOAD_PRIMARY_DOING" -> "JOB_STATE_UPLOAD_PRIMARY_DONE" [ label = "UploadPrimaryDoneEvent" ];
   "JOB_STATE_UPLOAD_PRIMARY_DONE" -> "JOB_STATE_UPLOAD_SECONDARY_INIT" [ label = "UploadSecondaryInitEvent" ];
   "JOB_STATE_UPLOAD_SECONDARY_INIT" -> "JOB_STATE_UPLOAD_SECONDARY_DOING" [ label = "UploadSecondaryDoingEvent" ];
   "JOB_STATE_UPLOAD_SECONDARY_DOING" -> "JOB_STATE_UPLOAD_SECONDARY_DOING" [ label = "UploadSecondaryPieceDoneEvent" ];
   "JOB_STATE_UPLOAD_SECONDARY_DOING" -> "JOB_STATE_UPLOAD_SECONDARY_DONE" [ label = "UploadSecondaryDoneEvent" ];
   "JOB_STATE_UPLOAD_SECONDARY_DONE" -> "JOB_STATE_SEAL_OBJECT_INIT" [ label = "SealObjectInitEvent" ];
   "JOB_STATE_SEAL_OBJECT_INIT" -> "JOB_STATE_SEAL_OBJECT_TX_DOING" [ label = "SealObjectDoingEvent" ];
   "JOB_STATE_SEAL_OBJECT_TX_DOING" -> "JOB_STATE_SEAL_OBJECT_DONE" [ label = "SealObjectDoneEvent" ];
*/

// define FSM Event Name
var (
	UploadPayloadInitEvent        string = "UploadPayloadInitEvent"
	UploadPrimaryDoingEvent       string = "UploadPrimaryDoingEvent"
	UploadPrimaryPieceDoneEvent   string = "UploadPrimaryPieceDoneEvent"
	UploadPrimaryDoneEvent        string = "UploadPrimaryDoneEvent"
	UploadSecondaryInitEvent      string = "UploadSecondaryInitEvent"
	UploadSecondaryDoingEvent     string = "UploadSecondaryDoingEvent"
	UploadSecondaryPieceDoneEvent string = "UploadSecondaryPieceDoneEvent"
	UploadSecondaryDoneEvent      string = "UploadSecondaryDoneEvent"
	SealObjectInitEvent           string = "SealObjectInitEvent"
	SealObjectDoingEvent          string = "SealObjectDoingEvent"
	SealObjectDoneEvent           string = "SealObjectDoneEvent"
	InterruptEvent                string = "InterruptEvent"
)

// UploadPayloadFsmEvent define FSM event transitions
var UploadPayloadFsmEvent = fsm.Events{
	{Name: UploadPayloadInitEvent, Src: []string{ptypesv1pb.JOB_STATE_CREATE_OBJECT_DONE}, Dst: ptypesv1pb.JOB_STATE_UPLOAD_PRIMARY_INIT},
	{Name: UploadPrimaryDoingEvent, Src: []string{ptypesv1pb.JOB_STATE_UPLOAD_PRIMARY_INIT}, Dst: ptypesv1pb.JOB_STATE_UPLOAD_PRIMARY_DOING},
	{Name: UploadPrimaryPieceDoneEvent, Src: []string{ptypesv1pb.JOB_STATE_UPLOAD_PRIMARY_DOING}, Dst: ptypesv1pb.JOB_STATE_UPLOAD_PRIMARY_DOING},
	{Name: UploadPrimaryDoneEvent, Src: []string{ptypesv1pb.JOB_STATE_UPLOAD_PRIMARY_DOING}, Dst: ptypesv1pb.JOB_STATE_UPLOAD_PRIMARY_DONE},
	{Name: UploadSecondaryInitEvent, Src: []string{ptypesv1pb.JOB_STATE_UPLOAD_PRIMARY_DONE}, Dst: ptypesv1pb.JOB_STATE_UPLOAD_SECONDARY_INIT},
	{Name: UploadSecondaryDoingEvent, Src: []string{ptypesv1pb.JOB_STATE_UPLOAD_SECONDARY_INIT}, Dst: ptypesv1pb.JOB_STATE_UPLOAD_SECONDARY_DOING},
	{Name: UploadSecondaryPieceDoneEvent, Src: []string{ptypesv1pb.JOB_STATE_UPLOAD_SECONDARY_DOING}, Dst: ptypesv1pb.JOB_STATE_UPLOAD_SECONDARY_DOING},
	{Name: UploadSecondaryDoneEvent, Src: []string{ptypesv1pb.JOB_STATE_UPLOAD_SECONDARY_DOING}, Dst: ptypesv1pb.JOB_STATE_UPLOAD_SECONDARY_DONE},
	{Name: SealObjectInitEvent, Src: []string{ptypesv1pb.JOB_STATE_UPLOAD_SECONDARY_DONE}, Dst: ptypesv1pb.JOB_STATE_SEAL_OBJECT_INIT},
	{Name: SealObjectDoingEvent, Src: []string{ptypesv1pb.JOB_STATE_SEAL_OBJECT_INIT}, Dst: ptypesv1pb.JOB_STATE_SEAL_OBJECT_TX_DOING},
	{Name: SealObjectDoneEvent, Src: []string{ptypesv1pb.JOB_STATE_SEAL_OBJECT_TX_DOING}, Dst: ptypesv1pb.JOB_STATE_SEAL_OBJECT_DONE},
	{Name: InterruptEvent, Src: []string{
		ptypesv1pb.JOB_STATE_CREATE_OBJECT_DONE,
		ptypesv1pb.JOB_STATE_UPLOAD_PRIMARY_INIT,
		ptypesv1pb.JOB_STATE_UPLOAD_PRIMARY_DOING,
		ptypesv1pb.JOB_STATE_UPLOAD_PRIMARY_DONE,
		ptypesv1pb.JOB_STATE_UPLOAD_SECONDARY_INIT,
		ptypesv1pb.JOB_STATE_UPLOAD_SECONDARY_DOING,
		ptypesv1pb.JOB_STATE_UPLOAD_SECONDARY_DONE,
		ptypesv1pb.JOB_STATE_SEAL_OBJECT_INIT,
		ptypesv1pb.JOB_STATE_SEAL_OBJECT_TX_DOING,
		ptypesv1pb.JOB_STATE_SEAL_OBJECT_DONE},
		Dst: ptypesv1pb.JOB_STATE_ERROR},
}

// define FSM action, the Action associated with callback
var (
	ActionBeforeEvent                        = "before_"
	ActionLeaveState                         = "leave_"
	ActionEnterState                         = "enter_"
	ActionAfterEvent                         = "after_"
	ActionEnterStateUploadPrimaryInit        = ActionEnterState + ptypesv1pb.JOB_STATE_UPLOAD_PRIMARY_INIT
	ActionEnterStateUploadPrimaryDoing       = ActionEnterState + ptypesv1pb.JOB_STATE_UPLOAD_PRIMARY_DOING
	ActionAfterEventUploadPrimaryPieceDone   = ActionAfterEvent + UploadPrimaryPieceDoneEvent
	ActionEnterUploadPrimaryDone             = ActionEnterState + ptypesv1pb.JOB_STATE_UPLOAD_PRIMARY_DONE
	ActionEnterUploadSecondaryInit           = ActionEnterState + ptypesv1pb.JOB_STATE_UPLOAD_SECONDARY_INIT
	ActionEnterUploadSecondaryDoing          = ActionEnterState + ptypesv1pb.JOB_STATE_UPLOAD_SECONDARY_DOING
	ActionAfterEventUploadSecondaryPieceDone = ActionAfterEvent + UploadSecondaryPieceDoneEvent
	ActionEnterUploadSecondaryDone           = ActionEnterState + UploadSecondaryDoneEvent
	ActionEnterSealObjectInit                = ActionEnterState + ptypesv1pb.JOB_STATE_SEAL_OBJECT_INIT
	ActionEnterSealObjectDoing               = ActionEnterState + ptypesv1pb.JOB_STATE_SEAL_OBJECT_TX_DOING
	ActionEnterSealObjectDone                = ActionEnterState + ptypesv1pb.JOB_STATE_SEAL_OBJECT_DONE
	ActionAfterEventInterrupt                = ActionAfterEvent + InterruptEvent
	ActionBeforeEventAll                     = ActionBeforeEvent + "event"
	ActionAfterEventAll                      = ActionAfterEvent + "event"
)

// UploadPayLoadFsmCallBack map the action that event occurs or state changes trigger and the callback
var UploadPayLoadFsmCallBack = fsm.Callbacks{
	ActionEnterStateUploadPrimaryInit:        EnterStateUploadPrimaryInit,
	ActionEnterStateUploadPrimaryDoing:       EnterStateUploadPrimaryDoing,
	ActionAfterEventUploadPrimaryPieceDone:   AfterUploadPrimaryPieceDone,
	ActionEnterUploadPrimaryDone:             EnterUploadPrimaryDone,
	ActionEnterUploadSecondaryInit:           EnterUploadSecondaryInit,
	ActionEnterUploadSecondaryDoing:          EnterUploadSecondaryDoing,
	ActionAfterEventUploadSecondaryPieceDone: AfterUploadSecondaryPieceDone,
	ActionEnterUploadSecondaryDone:           EnterUploadSecondaryDone,
	ActionEnterSealObjectInit:                EnterSealObjectInit,
	ActionEnterSealObjectDoing:               EnterSealObjectDoing,
	ActionEnterSealObjectDone:                EnterSealObjectDone,
	ActionAfterEventInterrupt:                AfterInterrupt,
	ActionBeforeEventAll:                     ShowStoneInfo,
	ActionAfterEventAll:                      ShowStoneInfo,
}
