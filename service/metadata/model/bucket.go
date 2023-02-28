package model

type Bucket struct {
	// owner is the account address of bucket creator, it is also the bucket owner.
	Owner string `json:"owner"`
	// bucket_name is a globally unique name of bucket
	BucketName string `json:"bucketName"`
	// is_public define the highest permissions for bucket. When the bucket is public, everyone can get the object in it.
	IsPublic bool `json:"isPublic"`
	// id is the unique identification for bucket.
	Id         string `json:"id"`
	SourceType int    `json:"sourceType"`
	// create_at define the block number when the bucket created.
	CreateAt int64 `json:"createAt"`
	// payment_address is the address of the payment account
	PaymentAddress string `json:"paymentAddress"`
	// primary_sp_address is the address of the primary sp. Objects belongs to this bucket will never
	// leave this SP, unless you explicitly shift them to another SP.
	PrimarySpAddress string `json:"primarySpAddress"`
	// read_quota defines the traffic quota for read
	ReadQuota        int   `json:"readQuota"`
	PaymentPriceTime int64 `json:"paymentPriceTime"`
	//PaymentOutFlows  []OutFlowInUSD `json:"paymentOutFlows"`
}

//type OutFlowInUSD struct {
//	// SP(service provider) stream account address
//	SpAddress string `json:"sp_address,omitempty"`
//	// flow rate in USD
//	Rate github_com_cosmos_cosmos_sdk_types.Int `json:"rate"`
//}

func (a *Bucket) TableName() string {
	return "bucket"
}
