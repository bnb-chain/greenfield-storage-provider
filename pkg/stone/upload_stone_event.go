package stone

import (
	types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	"github.com/looplab/fsm"
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
	{Name: UploadPayloadInitEvent, Src: []string{types.JOB_STATE_CREATE_OBJECT_DONE}, Dst: types.JOB_STATE_UPLOAD_PRIMARY_INIT},
	{Name: UploadPrimaryDoingEvent, Src: []string{types.JOB_STATE_UPLOAD_PRIMARY_INIT}, Dst: types.JOB_STATE_UPLOAD_PRIMARY_DOING},
	{Name: UploadPrimaryPieceDoneEvent, Src: []string{types.JOB_STATE_UPLOAD_PRIMARY_DOING}, Dst: types.JOB_STATE_UPLOAD_PRIMARY_DOING},
	{Name: UploadPrimaryDoneEvent, Src: []string{types.JOB_STATE_UPLOAD_PRIMARY_DOING}, Dst: types.JOB_STATE_UPLOAD_PRIMARY_DONE},
	{Name: UploadSecondaryInitEvent, Src: []string{types.JOB_STATE_UPLOAD_PRIMARY_DONE}, Dst: types.JOB_STATE_UPLOAD_SECONDARY_INIT},
	{Name: UploadSecondaryDoingEvent, Src: []string{types.JOB_STATE_UPLOAD_SECONDARY_INIT}, Dst: types.JOB_STATE_UPLOAD_SECONDARY_DOING},
	{Name: UploadSecondaryPieceDoneEvent, Src: []string{types.JOB_STATE_UPLOAD_SECONDARY_DOING}, Dst: types.JOB_STATE_UPLOAD_SECONDARY_DOING},
	{Name: UploadSecondaryDoneEvent, Src: []string{types.JOB_STATE_UPLOAD_SECONDARY_DOING}, Dst: types.JOB_STATE_UPLOAD_SECONDARY_DONE},
	{Name: SealObjectInitEvent, Src: []string{types.JOB_STATE_UPLOAD_SECONDARY_DONE}, Dst: types.JOB_STATE_SEAL_OBJECT_INIT},
	{Name: SealObjectDoingEvent, Src: []string{types.JOB_STATE_SEAL_OBJECT_INIT}, Dst: types.JOB_STATE_SEAL_OBJECT_TX_DOING},
	{Name: SealObjectDoneEvent, Src: []string{types.JOB_STATE_SEAL_OBJECT_TX_DOING}, Dst: types.JOB_STATE_SEAL_OBJECT_DONE},
	{Name: InterruptEvent, Src: []string{}, Dst: types.JOB_STATE_ERROR},
}

// define FSM action, the Action associated with callback
var (
	ActionBeforeEvent                        string = "before_"
	ActionLeaveState                         string = "leave_"
	ActionEnterState                         string = "enter_"
	ActionAfterEvent                         string = "after_"
	ActionEnterStateUploadPrimaryInit        string = ActionEnterState + types.JOB_STATE_UPLOAD_PRIMARY_INIT
	ActionEnterStateUploadPrimaryDoing       string = ActionEnterState + types.JOB_STATE_UPLOAD_PRIMARY_DOING
	ActionAfterEventUploadPrimaryPieceDone   string = ActionAfterEvent + UploadPrimaryPieceDoneEvent
	ActionEnterUploadPrimaryDone             string = ActionEnterState + types.JOB_STATE_UPLOAD_PRIMARY_DONE
	ActionEnterUploadSecondaryInit           string = ActionEnterState + types.JOB_STATE_UPLOAD_SECONDARY_INIT
	ActionEnterUploadSecondaryDoing          string = ActionEnterState + types.JOB_STATE_UPLOAD_SECONDARY_INIT
	ActionAfterEventUploadSecondaryPieceDone string = ActionAfterEvent + UploadSecondaryPieceDoneEvent
	ActionEnterUploadSecondaryDone           string = ActionEnterState + UploadSecondaryDoneEvent
	ActionEnterSealObjectInit                string = ActionEnterState + types.JOB_STATE_SEAL_OBJECT_INIT
	ActionEnterSealObjectDoing               string = ActionEnterState + types.JOB_STATE_SEAL_OBJECT_TX_DOING
	ActionEnterSealObjectDone                string = ActionEnterState + types.JOB_STATE_SEAL_OBJECT_DONE
	ActionAfterEventInterrupt                string = ActionAfterEvent + InterruptEvent
	ActionBeforeEventAll                     string = ActionBeforeEvent + "event"
	ActionAfterEventAll                      string = ActionAfterEvent + "event"
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
