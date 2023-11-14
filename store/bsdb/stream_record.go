package bsdb

import "time"

// GetPrimarySPStreamRecordBySpID return primary SP's stream records
func (b *BsDBImpl) GetPrimarySPStreamRecordBySpID(spID uint32) ([]*PrimarySpIncomeMeta, error) {
	var (
		primarySpIncomeMetaList []*PrimarySpIncomeMeta
		query                   string
		err                     error
	)
	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()
	query = " SELECT s.* , g.global_virtual_group_family_id FROM `stream_records` s join `global_virtual_group_families` g on g.virtual_payment_address = s.account where g.primary_sp_id=? and g.removed = false"
	err = b.db.Raw(query, spID).Find(&primarySpIncomeMetaList).Error

	return primarySpIncomeMetaList, err
}

// GetSecondarySPStreamRecordBySpID return secondary SP's stream records
func (b *BsDBImpl) GetSecondarySPStreamRecordBySpID(spID uint32) ([]*SecondarySpIncomeMeta, error) {
	var (
		secondarySpIncomeMetaList []*SecondarySpIncomeMeta
		query                     string
		err                       error
	)
	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()
	query = "SELECT s.* , g.global_virtual_group_id FROM `stream_records` s join `global_virtual_groups` g on g.virtual_payment_address = s.account where FIND_IN_SET(?, g.secondary_sp_ids) > 0 and removed = false; "
	err = b.db.Raw(query, spID).Find(&secondarySpIncomeMetaList).Error

	return secondarySpIncomeMetaList, err
}
