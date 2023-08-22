package bsdb

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/forbole/juno/v4/common"
	"github.com/stretchr/testify/assert"
)

func TestContinuationTokenFilter(t *testing.T) {
	db, mock := setupDB(t)

	expectedSQL := "SELECT * FROM `objects` WHERE object_name >= ?"
	mock.ExpectQuery(expectedSQL).WithArgs("token123").WillReturnRows(sqlmock.NewRows([]string{}))

	db.db.Table(ObjectTableName).Scopes(ContinuationTokenFilter("token123")).Find(&[]struct{}{})
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPrefixFilter(t *testing.T) {
	db, mock := setupDB(t)

	expectedSQL := "SELECT * FROM `objects` WHERE object_name LIKE ?"
	mock.ExpectQuery(expectedSQL).WithArgs("prefix%").WillReturnRows(sqlmock.NewRows([]string{}))

	db.db.Table(ObjectTableName).Scopes(PrefixFilter("prefix")).Find(&[]struct{}{})
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPathNameFilter(t *testing.T) {
	db, mock := setupDB(t)

	expectedSQL := "SELECT * FROM `slash_prefix_tree_nodes` WHERE path_name = ?"
	mock.ExpectQuery(expectedSQL).WithArgs("path123").WillReturnRows(sqlmock.NewRows([]string{}))

	db.db.Table(PrefixTreeTableName).Scopes(PathNameFilter("path123")).Find(&[]struct{}{})
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNameFilter(t *testing.T) {
	db, mock := setupDB(t)

	expectedSQL := "SELECT * FROM `slash_prefix_tree_nodes` WHERE name like ?"
	mock.ExpectQuery(expectedSQL).WithArgs("name123%").WillReturnRows(sqlmock.NewRows([]string{}))

	db.db.Table(PrefixTreeTableName).Scopes(NameFilter("name123")).Find(&[]struct{}{})
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFullNameFilter(t *testing.T) {
	db, mock := setupDB(t)

	expectedSQL := "SELECT * FROM `slash_prefix_tree_nodes` WHERE full_name >= ?"
	mock.ExpectQuery(expectedSQL).WithArgs("fullName123").WillReturnRows(sqlmock.NewRows([]string{}))

	db.db.Table(PrefixTreeTableName).Scopes(FullNameFilter("fullName123")).Find(&[]struct{}{})
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSourceTypeFilter(t *testing.T) {
	db, mock := setupDB(t)

	expectedSQL := "SELECT * FROM `objects` WHERE source_type = ?"
	mock.ExpectQuery(expectedSQL).WithArgs("sourceType123").WillReturnRows(sqlmock.NewRows([]string{}))

	db.db.Table(ObjectTableName).Scopes(SourceTypeFilter("sourceType123")).Find(&[]struct{}{})
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRemovedFilter(t *testing.T) {
	db, mock := setupDB(t)

	expectedSQL := "SELECT * FROM `objects` WHERE removed = ?"
	mock.ExpectQuery(expectedSQL).WithArgs(true).WillReturnRows(sqlmock.NewRows([]string{}))

	db.db.Table(ObjectTableName).Scopes(RemovedFilter(true)).Find(&[]struct{}{})
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBucketIDStartAfterFilter(t *testing.T) {
	db, mock := setupDB(t)

	bucketID := common.HexToHash("1")

	expectedSQL := "SELECT * FROM `objects` WHERE bucket_id > ?"
	mock.ExpectQuery(expectedSQL).WithArgs(bucketID).WillReturnRows(sqlmock.NewRows([]string{}))

	db.db.Table(ObjectTableName).Scopes(BucketIDStartAfterFilter(bucketID)).Find(&[]struct{}{})
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestObjectIDStartAfterFilter(t *testing.T) {
	db, mock := setupDB(t)

	objectID := common.HexToHash("0")

	expectedSQL := "SELECT * FROM `objects` WHERE object_id > ?"
	mock.ExpectQuery(expectedSQL).WithArgs(objectID).WillReturnRows(sqlmock.NewRows([]string{}))

	db.db.Table(ObjectTableName).Scopes(ObjectIDStartAfterFilter(objectID)).Find(&[]struct{}{})
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGroupIDStartAfterFilter(t *testing.T) {
	db, mock := setupDB(t)

	groupID := common.HexToHash("0")

	expectedSQL := "SELECT * FROM `objects` WHERE group_id > ?"
	mock.ExpectQuery(expectedSQL).WithArgs(groupID).WillReturnRows(sqlmock.NewRows([]string{}))

	db.db.Table(ObjectTableName).Scopes(GroupIDStartAfterFilter(groupID)).Find(&[]struct{}{})
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGroupAccountIDStartAfterFilter(t *testing.T) {
	db, mock := setupDB(t)

	accountID := common.HexToAddress("1")
	specialID := common.HexToAddress("0")

	expectedSQL := "SELECT * FROM `objects` WHERE account_id > ? and account_id != ?"
	mock.ExpectQuery(expectedSQL).WithArgs(accountID, specialID).WillReturnRows(sqlmock.NewRows([]string{}))

	db.db.Table(ObjectTableName).Scopes(GroupAccountIDStartAfterFilter(accountID)).Find(&[]struct{}{})
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateAtFilter(t *testing.T) {
	db, mock := setupDB(t)

	createAt := int64(1629225600)

	expectedSQL := "SELECT * FROM `objects` WHERE create_at <= ?"
	mock.ExpectQuery(expectedSQL).WithArgs(createAt).WillReturnRows(sqlmock.NewRows([]string{}))

	db.db.Table(ObjectTableName).Scopes(CreateAtFilter(createAt)).Find(&[]struct{}{})
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWithLimit(t *testing.T) {
	db, mock := setupDB(t)

	limit := 10

	expectedSQL := "SELECT * FROM `objects` LIMIT 10"
	mock.ExpectQuery(expectedSQL).WillReturnRows(sqlmock.NewRows([]string{}))

	db.db.Table(ObjectTableName).Scopes(WithLimit(limit)).Find(&[]struct{}{})
	assert.NoError(t, mock.ExpectationsWereMet())
}
