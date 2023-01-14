package v1

func (x *ObjectInfo) SafeCopy() *ObjectInfo {
	if x == nil {
		return nil
	}
	return &ObjectInfo{
		Owner:          x.GetOwner(),
		BucketName:     x.GetBucketName(),
		ObjectName:     x.GetObjectName(),
		Size:           x.GetSize(),
		Checksum:       x.GetChecksum(),
		IsPrivate:      x.GetIsPrivate(),
		ContentType:    x.GetContentType(),
		JobId:          x.GetJobId(),
		Height:         x.GetHeight(),
		TxHash:         x.GetTxHash(),
		ObjectId:       x.GetObjectId(),
		RedundancyType: x.GetRedundancyType(),
		PrimarySp:      x.GetPrimarySp().SafeCopy(),
	}
}

func (x *StorageProviderInfo) SafeCopy() *StorageProviderInfo {
	if x == nil {
		return nil
	}
	return &StorageProviderInfo{
		SpId:      x.GetSpId(),
		Idx:       x.GetIdx(),
		Checksum:  x.GetChecksum(),
		Signature: x.GetSignature(),
	}
}
