package gfspapp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
)

func TestGfSpBaseApp_TaskTimeout1(t *testing.T) {
	t.Log("some basic cases")
	ctrl := gomock.NewController(t)
	m := coretask.NewMockTask(ctrl)
	m.EXPECT().Type().Return(coretask.TypeTaskUnknown).AnyTimes()
	cases := []struct {
		name         string
		task         coretask.Task
		size         int64
		wantedResult int64
	}{
		{
			name:         "create bucket approval task",
			task:         &gfsptask.GfSpCreateBucketApprovalTask{},
			wantedResult: NotUseTimeout,
		},
		{
			name:         "create object approval task",
			task:         &gfsptask.GfSpCreateObjectApprovalTask{},
			wantedResult: NotUseTimeout,
		},
		{
			name:         "replicate piece approval task",
			task:         &gfsptask.GfSpReplicatePieceApprovalTask{},
			size:         1,
			wantedResult: NotUseTimeout,
		},
		{
			name:         "unknown task",
			task:         m,
			wantedResult: NotUseTimeout,
		},
	}
	g := setup(t)
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := g.TaskTimeout(tt.task, 0)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestGfSpBaseApp_TaskTimeout2(t *testing.T) {
	t.Log("upload object task")
	cases := []struct {
		name         string
		size         uint64
		wantedResult int64
	}{
		{
			name:         "1",
			size:         0,
			wantedResult: MinUploadTime,
		},
		{
			name:         "2",
			size:         301 * MinSpeed,
			wantedResult: MaxUploadTime,
		},
		{
			name:         "3",
			size:         150 * MinSpeed,
			wantedResult: 150,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			g := GfSpBaseApp{}
			task := &gfsptask.GfSpUploadObjectTask{}
			result := g.TaskTimeout(task, tt.size)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestGfSpBaseApp_TaskTimeout3(t *testing.T) {
	t.Log("replicate piece task")
	cases := []struct {
		name         string
		size         uint64
		wantedResult int64
	}{
		{
			name:         "1",
			size:         80 * MinSpeed,
			wantedResult: MinReplicateTime,
		},
		{
			name:         "2",
			size:         501 * MinSpeed,
			wantedResult: MaxReplicateTime,
		},
		{
			name:         "3",
			size:         100 * MinSpeed,
			wantedResult: 100,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			g := GfSpBaseApp{}
			task := &gfsptask.GfSpReplicatePieceTask{}
			result := g.TaskTimeout(task, tt.size)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestGfSpBaseApp_TaskTimeout4(t *testing.T) {
	t.Log("receive piece task")
	cases := []struct {
		name         string
		size         uint64
		wantedResult int64
	}{
		{
			name:         "1",
			size:         80 * MinSpeed,
			wantedResult: MinReceiveTime,
		},
		{
			name:         "2",
			size:         301 * MinSpeed,
			wantedResult: MaxReceiveTime,
		},
		{
			name:         "3",
			size:         100 * MinSpeed,
			wantedResult: 100,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			g := GfSpBaseApp{}
			task := &gfsptask.GfSpReceivePieceTask{}
			result := g.TaskTimeout(task, tt.size)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestGfSpBaseApp_TaskTimeout5(t *testing.T) {
	t.Log("seal object task")
	cases := []struct {
		name              string
		sealObjectTimeout int64
		wantedResult      int64
	}{
		{
			name:              "1",
			sealObjectTimeout: 80,
			wantedResult:      MinSealObjectTime,
		},
		{
			name:              "2",
			sealObjectTimeout: 301,
			wantedResult:      MaxSealObjectTime,
		},
		{
			name:              "3",
			sealObjectTimeout: 100,
			wantedResult:      100,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			g := GfSpBaseApp{sealObjectTimeout: tt.sealObjectTimeout}
			task := &gfsptask.GfSpSealObjectTask{}
			result := g.TaskTimeout(task, 0)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestGfSpBaseApp_TaskTimeout6(t *testing.T) {
	t.Log("download object task")
	cases := []struct {
		name         string
		size         uint64
		wantedResult int64
	}{
		{
			name:         "1",
			size:         0,
			wantedResult: MinDownloadTime,
		},
		{
			name:         "2",
			size:         301 * MinSpeed,
			wantedResult: MaxDownloadTime,
		},
		{
			name:         "3",
			size:         100 * MinSpeed,
			wantedResult: 100,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			g := GfSpBaseApp{}
			task := &gfsptask.GfSpDownloadObjectTask{}
			result := g.TaskTimeout(task, tt.size)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestGfSpBaseApp_TaskTimeout7(t *testing.T) {
	t.Log("challenge piece task")
	cases := []struct {
		name         string
		size         uint64
		wantedResult int64
	}{
		{
			name:         "1",
			size:         0,
			wantedResult: MinDownloadTime,
		},
		{
			name:         "2",
			size:         301 * MinSpeed,
			wantedResult: MaxDownloadTime,
		},
		{
			name:         "3",
			size:         100 * MinSpeed,
			wantedResult: 100,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			g := GfSpBaseApp{}
			task := &gfsptask.GfSpChallengePieceTask{}
			result := g.TaskTimeout(task, tt.size)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestGfSpBaseApp_TaskTimeout8(t *testing.T) {
	t.Log("gc object task")
	cases := []struct {
		name            string
		gcObjectTimeout int64
		wantedResult    int64
	}{
		{
			name:            "1",
			gcObjectTimeout: 80,
			wantedResult:    MinGCObjectTime,
		},
		{
			name:            "2",
			gcObjectTimeout: 601,
			wantedResult:    MaxGCObjectTime,
		},
		{
			name:            "3",
			gcObjectTimeout: 400,
			wantedResult:    400,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			g := GfSpBaseApp{gcObjectTimeout: tt.gcObjectTimeout}
			task := &gfsptask.GfSpGCObjectTask{}
			result := g.TaskTimeout(task, 0)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestGfSpBaseApp_TaskTimeout9(t *testing.T) {
	t.Log("gc zombie task")
	cases := []struct {
		name            string
		gcZombieTimeout int64
		wantedResult    int64
	}{
		{
			name:            "1",
			gcZombieTimeout: 80,
			wantedResult:    MinGCZombieTime,
		},
		{
			name:            "2",
			gcZombieTimeout: 601,
			wantedResult:    MaxGCZombieTime,
		},
		{
			name:            "3",
			gcZombieTimeout: 400,
			wantedResult:    400,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			g := GfSpBaseApp{gcZombieTimeout: tt.gcZombieTimeout}
			task := &gfsptask.GfSpGCZombiePieceTask{}
			result := g.TaskTimeout(task, 0)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestGfSpBaseApp_TaskTimeout10(t *testing.T) {
	t.Log("gc meta task")
	cases := []struct {
		name          string
		gcMetaTimeout int64
		wantedResult  int64
	}{
		{
			name:          "1",
			gcMetaTimeout: 80,
			wantedResult:  MinGCMetaTime,
		},
		{
			name:          "2",
			gcMetaTimeout: 601,
			wantedResult:  MaxGCMetaTime,
		},
		{
			name:          "3",
			gcMetaTimeout: 400,
			wantedResult:  400,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			g := GfSpBaseApp{gcMetaTimeout: tt.gcMetaTimeout}
			task := &gfsptask.GfSpGCMetaTask{}
			result := g.TaskTimeout(task, 0)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestGfSpBaseApp_TaskTimeout11(t *testing.T) {
	t.Log("recover piece task")
	cases := []struct {
		name         string
		size         uint64
		wantedResult int64
	}{
		{
			name:         "1",
			size:         0,
			wantedResult: MinRecoveryTime,
		},
		{
			name:         "2",
			size:         51 * MinSpeed,
			wantedResult: MaxRecoveryTime,
		},
		{
			name:         "3",
			size:         20 * MinSpeed,
			wantedResult: 21,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			g := GfSpBaseApp{}
			task := &gfsptask.GfSpRecoverPieceTask{}
			result := g.TaskTimeout(task, tt.size)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestGfSpBaseApp_TaskTimeout12(t *testing.T) {
	t.Log("migrate gvg task")
	cases := []struct {
		name              string
		migrateGVGTimeout int64
		wantedResult      int64
	}{
		{
			name:              "1",
			migrateGVGTimeout: 800,
			wantedResult:      MinMigrateGVGTime,
		},
		{
			name:              "2",
			migrateGVGTimeout: 3601,
			wantedResult:      MaxMigrateGVGTime,
		},
		{
			name:              "3",
			migrateGVGTimeout: 2000,
			wantedResult:      2000,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			g := GfSpBaseApp{migrateGVGTimeout: tt.migrateGVGTimeout}
			task := &gfsptask.GfSpMigrateGVGTask{}
			result := g.TaskTimeout(task, 0)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestGfSpBaseApp_TaskMaxRetry1(t *testing.T) {
	t.Log("some basic cases")
	ctrl := gomock.NewController(t)
	m := coretask.NewMockTask(ctrl)
	m.EXPECT().Type().Return(coretask.TypeTaskUnknown).AnyTimes()
	cases := []struct {
		name         string
		task         coretask.Task
		wantedResult int64
	}{
		{
			name:         "create bucket approval task",
			task:         &gfsptask.GfSpCreateBucketApprovalTask{},
			wantedResult: NotUseRetry,
		},
		{
			name:         "create object approval task",
			task:         &gfsptask.GfSpCreateObjectApprovalTask{},
			wantedResult: NotUseRetry,
		},
		{
			name:         "replicate piece approval task",
			task:         &gfsptask.GfSpReplicatePieceApprovalTask{},
			wantedResult: NotUseRetry,
		},
		{
			name:         "upload object task",
			task:         &gfsptask.GfSpUploadObjectTask{},
			wantedResult: NotUseRetry,
		},
		{
			name:         "download object task",
			task:         &gfsptask.GfSpDownloadObjectTask{},
			wantedResult: NotUseRetry,
		},
		{
			name:         "challenge task",
			task:         &gfsptask.GfSpChallengePieceTask{},
			wantedResult: NotUseRetry,
		},
		{
			name:         "unknown task",
			task:         m,
			wantedResult: NotUseRetry,
		},
	}
	g := setup(t)
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := g.TaskMaxRetry(tt.task)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestGfSpBaseApp_TaskMaxRetry2(t *testing.T) {
	t.Log("replicate piece task")
	cases := []struct {
		name           string
		replicateRetry int64
		wantedResult   int64
	}{
		{
			name:           "1",
			replicateRetry: 2,
			wantedResult:   MinReplicateRetry,
		},
		{
			name:           "2",
			replicateRetry: 7,
			wantedResult:   MaxReplicateRetry,
		},
		{
			name:           "3",
			replicateRetry: 5,
			wantedResult:   5,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			g := GfSpBaseApp{replicateRetry: tt.replicateRetry}
			task := &gfsptask.GfSpReplicatePieceTask{}
			result := g.TaskMaxRetry(task)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestGfSpBaseApp_TaskMaxRetry3(t *testing.T) {
	t.Log("receive piece task")
	cases := []struct {
		name                string
		receiveConfirmRetry int64
		wantedResult        int64
	}{
		{
			name:                "1",
			receiveConfirmRetry: 0,
			wantedResult:        MinReceiveConfirmRetry,
		},
		{
			name:                "2",
			receiveConfirmRetry: 5,
			wantedResult:        MaxReceiveConfirmRetry,
		},
		{
			name:                "3",
			receiveConfirmRetry: 2,
			wantedResult:        2,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			g := GfSpBaseApp{receiveConfirmRetry: tt.receiveConfirmRetry}
			task := &gfsptask.GfSpReceivePieceTask{}
			result := g.TaskMaxRetry(task)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestGfSpBaseApp_TaskMaxRetry4(t *testing.T) {
	t.Log("seal object task")
	cases := []struct {
		name            string
		sealObjectRetry int64
		wantedResult    int64
	}{
		{
			name:            "1",
			sealObjectRetry: 2,
			wantedResult:    MinSealObjectRetry,
		},
		{
			name:            "2",
			sealObjectRetry: 11,
			wantedResult:    MaxSealObjectRetry,
		},
		{
			name:            "3",
			sealObjectRetry: 7,
			wantedResult:    7,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			g := GfSpBaseApp{sealObjectRetry: tt.sealObjectRetry}
			task := &gfsptask.GfSpSealObjectTask{}
			result := g.TaskMaxRetry(task)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestGfSpBaseApp_TaskMaxRetry5(t *testing.T) {
	t.Log("gc object task")
	cases := []struct {
		name          string
		gcObjectRetry int64
		wantedResult  int64
	}{
		{
			name:          "1",
			gcObjectRetry: 2,
			wantedResult:  MinGCObjectRetry,
		},
		{
			name:          "2",
			gcObjectRetry: 6,
			wantedResult:  MaxGCObjectRetry,
		},
		{
			name:          "3",
			gcObjectRetry: 4,
			wantedResult:  4,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			g := GfSpBaseApp{gcObjectRetry: tt.gcObjectRetry}
			task := &gfsptask.GfSpGCObjectTask{}
			result := g.TaskMaxRetry(task)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestGfSpBaseApp_TaskMaxRetry6(t *testing.T) {
	t.Log("gc zombie piece task")
	cases := []struct {
		name          string
		gcZombieRetry int64
		wantedResult  int64
	}{
		{
			name:          "1",
			gcZombieRetry: 2,
			wantedResult:  MinGCObjectRetry,
		},
		{
			name:          "2",
			gcZombieRetry: 6,
			wantedResult:  MaxGCObjectRetry,
		},
		{
			name:          "3",
			gcZombieRetry: 4,
			wantedResult:  4,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			g := GfSpBaseApp{gcZombieRetry: tt.gcZombieRetry}
			task := &gfsptask.GfSpGCZombiePieceTask{}
			result := g.TaskMaxRetry(task)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestGfSpBaseApp_TaskMaxRetry7(t *testing.T) {
	t.Log("gc meta task")
	cases := []struct {
		name         string
		gcMetaRetry  int64
		wantedResult int64
	}{
		{
			name:         "1",
			gcMetaRetry:  2,
			wantedResult: MinGCObjectRetry,
		},
		{
			name:         "2",
			gcMetaRetry:  6,
			wantedResult: MaxGCObjectRetry,
		},
		{
			name:         "3",
			gcMetaRetry:  4,
			wantedResult: 4,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			g := GfSpBaseApp{gcMetaRetry: tt.gcMetaRetry}
			task := &gfsptask.GfSpGCMetaTask{}
			result := g.TaskMaxRetry(task)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestGfSpBaseApp_TaskMaxRetry8(t *testing.T) {
	t.Log("recover piece task")
	cases := []struct {
		name          string
		recoveryRetry int64
		wantedResult  int64
	}{
		{
			name:          "1",
			recoveryRetry: 1,
			wantedResult:  MinRecoveryRetry,
		},
		{
			name:          "2",
			recoveryRetry: 5,
			wantedResult:  MaxRecoveryRetry,
		},
		{
			name:          "3",
			recoveryRetry: 2,
			wantedResult:  2,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			g := GfSpBaseApp{recoveryRetry: tt.recoveryRetry}
			task := &gfsptask.GfSpRecoverPieceTask{}
			result := g.TaskMaxRetry(task)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestGfSpBaseApp_TaskMaxRetry9(t *testing.T) {
	t.Log("migrate gvg task")
	cases := []struct {
		name            string
		migrateGVGRetry int64
		wantedResult    int64
	}{
		{
			name:            "1",
			migrateGVGRetry: 1,
			wantedResult:    MinMigrateGVGRetry,
		},
		{
			name:            "2",
			migrateGVGRetry: 5,
			wantedResult:    MaxMigrateGVGRetry,
		},
		{
			name:            "3",
			migrateGVGRetry: 2,
			wantedResult:    2,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			g := GfSpBaseApp{migrateGVGRetry: tt.migrateGVGRetry}
			task := &gfsptask.GfSpMigrateGVGTask{}
			result := g.TaskMaxRetry(task)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestGfSpBaseApp_TaskPriority(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := coretask.NewMockTask(ctrl)
	m.EXPECT().Type().Return(coretask.TypeTaskUnknown).AnyTimes()
	cases := []struct {
		name         string
		task         coretask.Task
		wantedResult coretask.TPriority
	}{
		{
			name:         "create bucket approval task",
			task:         &gfsptask.GfSpCreateBucketApprovalTask{},
			wantedResult: coretask.UnSchedulingPriority,
		},
		{
			name:         "migrate bucket approval task",
			task:         &gfsptask.GfSpMigrateBucketApprovalTask{},
			wantedResult: coretask.UnSchedulingPriority,
		},
		{
			name:         "create object approval task",
			task:         &gfsptask.GfSpCreateObjectApprovalTask{},
			wantedResult: coretask.UnSchedulingPriority,
		},
		{
			name:         "replicate piece approval task",
			task:         &gfsptask.GfSpReplicatePieceApprovalTask{},
			wantedResult: coretask.UnSchedulingPriority,
		},
		{
			name:         "upload object task",
			task:         &gfsptask.GfSpUploadObjectTask{},
			wantedResult: coretask.UnSchedulingPriority,
		},
		{
			name:         "replicate piece task",
			task:         &gfsptask.GfSpReplicatePieceTask{},
			wantedResult: coretask.MaxTaskPriority,
		},
		{
			name:         "receive piece task",
			task:         &gfsptask.GfSpReceivePieceTask{},
			wantedResult: coretask.DefaultSmallerPriority / 4,
		},
		{
			name:         "seal object task",
			task:         &gfsptask.GfSpSealObjectTask{},
			wantedResult: coretask.DefaultSmallerPriority,
		},
		{
			name:         "download object task",
			task:         &gfsptask.GfSpDownloadObjectTask{},
			wantedResult: coretask.UnSchedulingPriority,
		},
		{
			name:         "challenge task",
			task:         &gfsptask.GfSpChallengePieceTask{},
			wantedResult: coretask.UnSchedulingPriority,
		},
		{
			name:         "gc object task",
			task:         &gfsptask.GfSpGCObjectTask{},
			wantedResult: coretask.UnSchedulingPriority,
		},
		{
			name:         "gc zombie piece task",
			task:         &gfsptask.GfSpGCZombiePieceTask{},
			wantedResult: coretask.UnSchedulingPriority,
		},
		{
			name:         "gc meta task",
			task:         &gfsptask.GfSpGCMetaTask{},
			wantedResult: coretask.UnSchedulingPriority,
		},
		{
			name:         "recover piece task",
			task:         &gfsptask.GfSpRecoverPieceTask{},
			wantedResult: coretask.DefaultSmallerPriority / 4,
		},
		{
			name:         "migrate gvg task",
			task:         &gfsptask.GfSpMigrateGVGTask{},
			wantedResult: coretask.DefaultSmallerPriority,
		},
		{
			name:         "unknown task",
			task:         m,
			wantedResult: coretask.UnSchedulingPriority,
		},
	}
	g := setup(t)
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := g.TaskPriority(tt.task)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestGfSpBaseApp_TaskPriorityLevel(t *testing.T) {
	cases := []struct {
		name         string
		task         coretask.Task
		wantedResult coretask.TPriorityLevel
	}{
		{
			name:         "create bucket approval task",
			task:         &gfsptask.GfSpCreateBucketApprovalTask{Task: &gfsptask.GfSpTask{TaskPriority: 70}},
			wantedResult: coretask.TLowPriorityLevel,
		},
		{
			name:         "migrate bucket approval task",
			task:         &gfsptask.GfSpMigrateBucketApprovalTask{Task: &gfsptask.GfSpTask{TaskPriority: 200}},
			wantedResult: coretask.THighPriorityLevel,
		},
		{
			name:         "create object approval task",
			task:         &gfsptask.GfSpCreateObjectApprovalTask{Task: &gfsptask.GfSpTask{TaskPriority: 100}},
			wantedResult: coretask.TMediumPriorityLevel,
		},
	}
	g := setup(t)
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := g.TaskPriorityLevel(tt.task)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}
