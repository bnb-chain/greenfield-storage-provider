package database

import (
	"context"
	"github.com/forbole/juno/v4/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (db *DB) SaveGVG(ctx context.Context, gvg *models.GlobalVirtualGroup) error {
	return nil
}

func (db *DB) UpdateGVG(ctx context.Context, gvg *models.GlobalVirtualGroup) error {
	return nil
}

func (db *DB) SaveLVG(ctx context.Context, lvg *models.LocalVirtualGroup) error {
	return nil
}

func (db *DB) UpdateLVG(ctx context.Context, lvg *models.LocalVirtualGroup) error {
	return nil
}

func (db *DB) SaveVGF(ctx context.Context, vgf *models.GlobalVirtualGroupFamily) error {
	return nil
}

func (db *DB) UpdateVGF(ctx context.Context, vgf *models.GlobalVirtualGroupFamily) error {
	return nil
}

func (db *DB) SaveGVGToSQL(ctx context.Context, gvg *models.GlobalVirtualGroup) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table((&models.GlobalVirtualGroup{}).TableName()).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "global_virtual_group_id"}},
		UpdateAll: true,
	}).Create(gvg).Statement
	return stat.SQL.String(), stat.Vars
}

func (db *DB) UpdateGVGToSQL(ctx context.Context, gvg *models.GlobalVirtualGroup) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table((&models.GlobalVirtualGroup{}).TableName()).Where("global_virtual_group_id = ?", gvg.GlobalVirtualGroupId).Updates(gvg).Statement
	return stat.SQL.String(), stat.Vars
}

func (db *DB) SaveLVGToSQL(ctx context.Context, lvg *models.LocalVirtualGroup) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table((&models.LocalVirtualGroup{}).TableName()).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "local_virtual_group_id"}},
		UpdateAll: true,
	}).Create(lvg).Statement
	return stat.SQL.String(), stat.Vars
}

func (db *DB) UpdateLVGToSQL(ctx context.Context, lvg *models.LocalVirtualGroup) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table((&models.LocalVirtualGroup{}).TableName()).Where("local_virtual_group_id = ? and bucket_id = ?", lvg.LocalVirtualGroupId, lvg.BucketID).Updates(lvg).Statement
	return stat.SQL.String(), stat.Vars
}

func (db *DB) SaveVGFToSQL(ctx context.Context, vgf *models.GlobalVirtualGroupFamily) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table((&models.GlobalVirtualGroupFamily{}).TableName()).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "global_virtual_group_family_id"}},
		UpdateAll: true,
	}).Create(vgf).Statement
	return stat.SQL.String(), stat.Vars
}

func (db *DB) UpdateVGFToSQL(ctx context.Context, vgf *models.GlobalVirtualGroupFamily) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table((&models.GlobalVirtualGroupFamily{}).TableName()).Where("global_virtual_group_family_id = ?", vgf.GlobalVirtualGroupFamilyId).Updates(vgf).Statement
	return stat.SQL.String(), stat.Vars
}
