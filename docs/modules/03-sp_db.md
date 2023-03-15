# SP DB(Storage Provider Database)

SP store needs to implement [SPDB](../../store/sqldb/database.go) interface. SQL database(MySQL) is used by default.
The following mainly introduces the data schemas corresponding to several core interfaces.

## JobContext

JobContext records the context of uploading an payload data, it contains two tables: JobTable and ObjectTable.

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
  // default job type
  JOB_TYPE_UNSPECIFIED = 0;
  // upload object
  JOB_TYPE_UPLOAD_OBJECT = 1;
  // delete object
  JOB_TYPE_DELETE_OBJECT = 2;
}

enum JobState {
  // default job state
  JOB_STATE_INIT_UNSPECIFIED = 0;

  // uploading payload data to primary SP
  JOB_STATE_UPLOAD_OBJECT_DOING = 1;
  // upload payload data to primary SP has done
  JOB_STATE_UPLOAD_OBJECT_DONE = 2;
  // failed to upload primary SP
  JOB_STATE_UPLOAD_OBJECT_ERROR = 3;

  // alloc secondary SPs is doing
  JOB_STATE_ALLOC_SECONDARY_DOING = 4;
  // alloc secondary SPs has done
  JOB_STATE_ALLOC_SECONDARY_DONE = 5;
  // failed to alloc secondary SPs
  JOB_STATE_ALLOC_SECONDARY_ERROR = 6;

  // replicating payload data to secondary SPs
  JOB_STATE_REPLICATE_OBJECT_DOING = 7;
  // replicate payload data to secondary SPs has done
  JOB_STATE_REPLICATE_OBJECT_DONE = 8;
  // failed to replicate payload data to secondary SPs
  JOB_STATE_REPLICATE_OBJECT_ERROR = 9;

  // signing seal object transaction
  JOB_STATE_SIGN_OBJECT_DOING = 10;
  // sign seal object transaction has done
  JOB_STATE_SIGN_OBJECT_DONE = 11;
  // failed to sign seal object transaction
  JOB_STATE_SIGN_OBJECT_ERROR = 12;

  // seal object transaction is doing on chain
  JOB_STATE_SEAL_OBJECT_TX_DOING = 13;
  // seal object transaction has done
  JOB_STATE_SEAL_OBJECT_DONE = 14;
  // failed to run seal object transaction
  JOB_STATE_SEAL_OBJECT_ERROR = 15;
}
```

### Object Table

ObjectTable stores basic information about an upload object metadata.

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
  // default redundancy type is replica
  REDUNDANCY_REPLICA_TYPE = 0;
  // redundancy type is ec
  REDUNDANCY_EC_TYPE = 1;
  // redundancy type is inline type
  REDUNDANCY_INLINE_TYPE = 2;
}

enum ObjectStatus {
  // default object status is initialized
  OBJECT_STATUS_INIT = 0;
  // object status is in service
  OBJECT_STATUS_IN_SERVICE = 1;
}

enum SourceType {
  // default source type that object is origin
  SOURCE_TYPE_ORIGIN = 0;
  // object is from bsc cross chain
  SOURCE_TYPE_BSC_CROSS_CHAIN = 1;
}
```

## Integrity

For each object there are some pieces root hashes stored on greenfield chain to keep data integrity. And for the pieces of an object stored on a specific SP, the SP keeps these pieces' hashes, which are used for storage proof.

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
