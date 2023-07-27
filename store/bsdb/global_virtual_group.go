package bsdb

import (
	"errors"
	"fmt"
	"time"

	"github.com/forbole/juno/v4/common"
	"gorm.io/gorm"
)

// GetGlobalVirtualGroupByGvgID get global virtual group by gvg id
func (b *BsDBImpl) GetGlobalVirtualGroupByGvgID(gvgID uint32) (*GlobalVirtualGroup, error) {
	var (
		gvg     *GlobalVirtualGroup
		filters []func(*gorm.DB) *gorm.DB
		err     error
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

	filters = append(filters, RemovedFilter(false))
	err = b.db.Table((&GlobalVirtualGroup{}).TableName()).
		Select("*").
		Where("global_virtual_group_id = ?", gvgID).
		Scopes(filters...).
		Take(&gvg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return gvg, err
}

// ListGvgByPrimarySpID list gvg by primary sp id
func (b *BsDBImpl) ListGvgByPrimarySpID(spID uint32) ([]*GlobalVirtualGroup, error) {
	var (
		gvg     []*GlobalVirtualGroup
		filters []func(*gorm.DB) *gorm.DB
		err     error
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

	filters = append(filters, RemovedFilter(false))
	err = b.db.Table((&GlobalVirtualGroup{}).TableName()).
		Select("*").
		Where("primary_sp_id = ?", spID).
		Scopes(filters...).
		Find(&gvg).Error

	return gvg, err
}

// ListGvgBySecondarySpID list gvg by secondary sp id
func (b *BsDBImpl) ListGvgBySecondarySpID(spID uint32) ([]*GlobalVirtualGroup, error) {
	var (
		gvg   []*GlobalVirtualGroup
		query string
		err   error
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

	query = fmt.Sprintf("select * from global_virtual_groups where FIND_IN_SET('%d', secondary_sp_ids) > 0 and removed = false; ", spID)
	err = b.db.Raw(query).Find(&gvg).Error

	return gvg, err
}

// GetGvgByBucketAndLvgID get global virtual group by lvg id and bucket id
func (b *BsDBImpl) GetGvgByBucketAndLvgID(bucketID common.Hash, lvgID uint32) (*GlobalVirtualGroup, error) {
	var (
		gvg *GlobalVirtualGroup
		lvg *LocalVirtualGroup
		err error
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

	lvg, err = b.GetLvgByBucketAndLvgID(bucketID, lvgID)
	if err != nil || lvg == nil {
		return nil, err
	}

	err = b.db.Table((&GlobalVirtualGroup{}).TableName()).
		Select("*").
		Where("global_virtual_group_id = ? and removed = false", lvg.GlobalVirtualGroupId).
		Take(&gvg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return gvg, err
}

// ListGvgByBucketID list global virtual group by bucket id
func (b *BsDBImpl) ListGvgByBucketID(bucketID common.Hash) ([]*GlobalVirtualGroup, error) {
	var (
		globalGroups []*GlobalVirtualGroup
		localGroups  []*GlobalVirtualGroup
		gvgID        []uint32
		filters      []func(*gorm.DB) *gorm.DB
		err          error
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

	filters = append(filters, RemovedFilter(false))
	err = b.db.Table((&LocalVirtualGroup{}).TableName()).
		Select("*").
		Where("bucket_id = ?", bucketID).
		Scopes(filters...).
		Find(&localGroups).Error
	if err != nil || len(localGroups) == 0 {
		return nil, err
	}

	gvgID = make([]uint32, len(localGroups))
	for i, group := range localGroups {
		gvgID[i] = group.GlobalVirtualGroupId
	}
	err = b.db.Table((&GlobalVirtualGroup{}).TableName()).
		Select("*").
		Where("global_virtual_group_id in (?)", gvgID).
		Take(&globalGroups).Error

	return globalGroups, err
}
