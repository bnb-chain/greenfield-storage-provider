# SP DB(Storage Provider Database)

SP uses traditional SQL database(MySQL) to store some important data. Here mainly introduce three databse table schemas: JobTable, ObjectTable and IntegrityTable.

## Job Context

Job context contains two tables: JobTable and ObjectTable.

### Job Table

JobTable describes some important data about job type and job state. Every operation in SP is a job which drives by state machine.

Below is the schema of `Jobtable`:

```go
// JobTable table schema
type JobTable struct {
    // JobID defines the unique id of a job
	JobID        uint64 `gorm:"primary_key;autoIncrement"`
    // JobType defines the type of a job
	JobType      int32
    // JobState defines the state of a job
	JobState     int32
    // JobErrorCode defines the error code when a job abnormal termination
	JobErrorCode uint32
    // CreatedTime defines the job create time, used to jobs garbage collection
	CreatedTime  time.Time
    // ModifiedTime defines the job last modified time, used to judge timeout
	ModifiedTime time.Time
}
```

Below is the enum of `Jobtype and JobState`:

```protobuf
enum JobType {
  JOB_TYPE_UNSPECIFIED = 0;
  JOB_TYPE_UPLOAD_OBJECT = 1;
  JOB_TYPE_DELETE_OBJECT = 2;
}

enum JobState {
  JOB_STATE_INIT_UNSPECIFIED = 0;

  JOB_STATE_UPLOAD_OBJECT_DOING = 1;
  JOB_STATE_UPLOAD_OBJECT_DONE = 2;
  JOB_STATE_UPLOAD_OBJECT_ERROR = 3;

  JOB_STATE_ALLOC_SECONDARY_DOING = 4;
  JOB_STATE_ALLOC_SECONDARY_DONE = 5;
  JOB_STATE_ALLOC_SECONDARY_ERROR = 6;

  JOB_STATE_REPLICATE_OBJECT_DOING = 7;
  JOB_STATE_REPLICATE_OBJECT_DONE = 8;
  JOB_STATE_REPLICATE_OBJECT_ERROR = 9;

  JOB_STATE_SIGN_OBJECT_DOING = 10;
  JOB_STATE_SIGN_OBJECT_DONE = 11;
  JOB_STATE_SIGN_OBJECT_ERROR = 12;

  JOB_STATE_SEAL_OBJECT_TX_DOING = 13;
  JOB_STATE_SEAL_OBJECT_DONE = 14;
  JOB_STATE_SEAL_OBJECT_ERROR = 15;
}
```

### Object Table

ObjectTable stores basic information about an object.

Below is the schema of `ObjectTable`:

```go
// ObjectTable table schema
type ObjectTable struct {
    // ObjectID defines the unique ID of an obejct
	ObjectID             uint64 `gorm:"primary_key"`
    // JobID defines the unique id of a job.
	JobID                uint64 `gorm:"index:job_to_object"`
    // Owner defines the owner of an object
	Owner                string
    // BucketName deinfes the bucket name to which an object belongs
	BucketName           string
    // ObjectName defines the object name
	ObjectName           string
    // PayloadSize defines the obejct size
	PayloadSize          uint64
    // IsPublic defines an object is public
	IsPublic             bool
    // ContentType defines an obejct content type
	ContentType          string
    // CreatedAtHeight defines an obejct created at which chain height 
	CreatedAtHeight      int64
    // ObjectStatus defines object status
	ObjectStatus         int32
    // RedundancyType defines the redundancy type of an object used
	RedundancyType       int32
    // SourceType defines the source type of an object
	SourceType           int32
    // SpIntegrityHash defines sp inetgirty hash
	SpIntegrityHash      string
    // SecondarySpAddresses defines secondary sp addresses
	SecondarySpAddresses string
}
```

Below is the enum of `RedundancyType, ObjectStatus and SourceType`:

```protobuf
enum RedundancyType {
  REDUNDANCY_REPLICA_TYPE = 0;
  REDUNDANCY_EC_TYPE = 1;
  REDUNDANCY_INLINE_TYPE = 2;
}

enum ObjectStatus {
  OBJECT_STATUS_INIT = 0;
  OBJECT_STATUS_IN_SERVICE = 1;
}

enum SourceType {
  SOURCE_TYPE_ORIGIN = 0;
  SOURCE_TYPE_BSC_CROSS_CHAIN = 1;
}
```

## Integrity

Integrity is used to record an object integrity hash for users' to challenge whether his/her data is right.

### Integrity Table

Below is the schema of `IntegrityMetaTable`:

```go
// IntegrityMetaTable table schema
type IntegrityMetaTable struct {
    // ObjectID defines the unique ID of an obejct
	ObjectID      uint64 `gorm:"primary_key"`
    // PieceHashList defines the piece hash list of an obejct by using sha256
	PieceHashList string
    // IntegrityHash defines the integrity hash of an object
	IntegrityHash string
    // Signature defines the signature of an obejct's IntegrityHash by using Secondary SP's private key
	Signature     string
}
```
