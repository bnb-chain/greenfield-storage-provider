package database

import (
	"context"
	"github.com/forbole/juno/v4/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (db *DB) SaveStreamRecord(ctx context.Context, streamRecord *models.StreamRecord) error {
	return nil
}

func (db *DB) SavePaymentAccount(ctx context.Context, paymentAccount *models.PaymentAccount) error {
	return nil
}

func (db *DB) SaveStreamRecordToSQL(ctx context.Context, streamRecord *models.StreamRecord) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table((&models.StreamRecord{}).TableName()).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "account"}},
		UpdateAll: true,
	}).Create(streamRecord).Statement
	return stat.SQL.String(), stat.Vars
}

func (db *DB) SavePaymentAccountToSQL(ctx context.Context, paymentAccount *models.PaymentAccount) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{
		DryRun: true,
	}).Table((&models.PaymentAccount{}).TableName()).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "addr"}},
		UpdateAll: true,
	}).Create(paymentAccount).Statement
	return stat.SQL.String(), stat.Vars
}
