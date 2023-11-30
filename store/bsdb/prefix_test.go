package bsdb

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/forbole/juno/v4/common"
	"github.com/stretchr/testify/assert"
)

const bucketName = "testBucket"

func TestBsDBImpl_ListObjectsSuccess(t *testing.T) {
	prefix := "testPrefix/"
	maxKeys := 2
	expectedObjects := []*ListObjectsResult{
		{PathName: "testPath1/obj1", ResultType: "object", Object: &Object{ObjectName: "obj1", ObjectID: common.HexToHash("1")}},
		{PathName: "testPath2/obj2", ResultType: "object", Object: &Object{ObjectName: "obj2", ObjectID: common.HexToHash("2")}},
	}

	s, mock := setupDB(t)

	// Expectation for GetPrefixesTableName
	mock.ExpectQuery(fmt.Sprintf("SELECT * FROM `%s` WHERE bucket_name = ? AND path_name = ? ORDER BY full_name LIMIT 3", GetPrefixesTableName(bucketName))).
		WithArgs(bucketName, prefix).
		WillReturnRows(
			sqlmock.NewRows([]string{"path_name", "full_name", "name", "is_object", "is_folder", "bucket_name", "object_id", "object_name"}).
				AddRow("testPath1/", "testPath1/obj1", "obj1", true, false, bucketName, common.HexToHash("1"), "obj1").
				AddRow("testPath2/", "testPath2/obj2", "obj2", true, false, bucketName, common.HexToHash("2"), "obj2"))

	// Expectations for GetObjectsTableName
	mock.ExpectQuery(fmt.Sprintf("SELECT * FROM `%s` WHERE object_id in (?,?)", GetObjectsTableName(bucketName))).
		WithArgs(common.HexToHash("1"), common.HexToHash("2")).
		WillReturnRows(
			sqlmock.NewRows([]string{"object_id", "object_name"}).
				AddRow(common.HexToHash("1"), "obj1").
				AddRow(common.HexToHash("2"), "obj2"))

	res, err := s.ListObjects(bucketName, "", prefix, maxKeys)

	assert.Nil(t, err)
	assert.Equal(t, expectedObjects, res)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBsDBImpl_ListObjectsWithContinuationToken(t *testing.T) {
	prefix := "testPrefix/"
	maxKeys := 2
	continuationToken := "testPath2/obj2"
	expectedObjects := []*ListObjectsResult{
		{PathName: "testPath2/obj2", ResultType: "object", Object: &Object{ObjectName: "obj2", ObjectID: common.HexToHash("2")}},
	}

	s, mock := setupDB(t)

	// Expectation for GetPrefixesTableName
	mock.ExpectQuery(fmt.Sprintf("SELECT * FROM `%s` WHERE bucket_name = ? AND path_name = ? AND full_name >= ? ORDER BY full_name LIMIT 3", GetPrefixesTableName(bucketName))).
		WithArgs(bucketName, prefix, continuationToken).
		WillReturnRows(
			sqlmock.NewRows([]string{"path_name", "full_name", "name", "is_object", "is_folder", "bucket_name", "object_id", "object_name"}).
				AddRow("testPath2/", "testPath2/obj2", "obj2", true, false, bucketName, common.HexToHash("2"), "obj2"))

	// Expectations for GetObjectsTableName
	mock.ExpectQuery(fmt.Sprintf("SELECT * FROM `%s` WHERE object_id in (?)", GetObjectsTableName(bucketName))).
		WithArgs(common.HexToHash("2")).
		WillReturnRows(
			sqlmock.NewRows([]string{"object_id", "object_name"}).
				AddRow(common.HexToHash("2"), "obj2"))

	res, err := s.ListObjects(bucketName, continuationToken, prefix, maxKeys)

	assert.Nil(t, err)
	assert.Equal(t, expectedObjects, res)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBsDBImpl_ListObjectsWithEmptyResult(t *testing.T) {
	prefix := "testPrefix/"
	maxKeys := 2

	s, mock := setupDB(t)

	// Expect the first query on slash_prefix_tree_nodes_31
	expectedSQL1 := "SELECT * FROM `slash_prefix_tree_nodes_31` WHERE bucket_name = ? AND path_name = ? ORDER BY full_name LIMIT 3"
	mock.ExpectQuery(expectedSQL1).
		WithArgs(bucketName, prefix).
		WillReturnRows(sqlmock.NewRows([]string{"path_name", "full_name", "name", "is_object", "is_folder", "bucket_name", "object_id", "object_name"}))

	mock.ExpectQuery(fmt.Sprintf("SELECT * FROM `%s` WHERE object_id in (?)", GetObjectsTableName(bucketName))).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(
			sqlmock.NewRows([]string{"object_id", "object_name"}))

	res, err := s.ListObjects(bucketName, "", prefix, maxKeys)
	assert.Nil(t, err)
	assert.Empty(t, res)
}

func TestBsDBImpl_ListObjectsWithError(t *testing.T) {
	prefix := "testPrefix/"
	maxKeys := 2

	s, mock := setupDB(t)

	expectedSQL1 := "SELECT * FROM `slash_prefix_tree_nodes_31` WHERE bucket_name = ? AND path_name = ? ORDER BY full_name LIMIT 3"
	mock.ExpectQuery(expectedSQL1).
		WithArgs(bucketName, prefix).
		WillReturnError(errors.New("test error"))

	_, err := s.ListObjects(bucketName, "", prefix, maxKeys)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "test error")
}

func TestBsDBImpl_ListObjectsWithPathName(t *testing.T) {
	prefix := "testPath/"
	maxKeys := 2
	expectedObjects := []*ListObjectsResult{
		{PathName: "testPath2/obj2", ResultType: "object", Object: &Object{ObjectName: "obj2", ObjectID: common.HexToHash("2")}},
	}

	s, mock := setupDB(t)

	// Expect the first query on slash_prefix_tree_nodes_31
	expectedSQL1 := "SELECT * FROM `slash_prefix_tree_nodes_31` WHERE bucket_name = ? AND path_name = ? ORDER BY full_name LIMIT 3"
	mock.ExpectQuery(expectedSQL1).
		WithArgs(bucketName, prefix).
		WillReturnRows(
			sqlmock.NewRows([]string{"path_name", "full_name", "name", "is_object", "is_folder", "bucket_name", "object_id", "object_name"}).
				AddRow("testPath2/", "testPath2/obj2", "obj2", true, false, bucketName, common.HexToHash("2"), "obj2"))

	mock.ExpectQuery(fmt.Sprintf("SELECT * FROM `%s` WHERE object_id in (?)", GetObjectsTableName(bucketName))).
		WithArgs(common.HexToHash("2")).
		WillReturnRows(
			sqlmock.NewRows([]string{"object_id", "object_name"}).
				AddRow(common.HexToHash("2"), "obj2"))

	res, err := s.ListObjects(bucketName, "", prefix, maxKeys)
	assert.Nil(t, err)
	assert.Equal(t, expectedObjects, res)
}

func TestBsDBImpl_ListObjectsWithPrefixQuery(t *testing.T) {
	prefix := "testPrefix/testQuery"
	maxKeys := 2
	expectedObjects := []*ListObjectsResult{
		{PathName: "testPath2/obj2", ResultType: "object", Object: &Object{ObjectName: "obj2", ObjectID: common.HexToHash("2")}},
	}

	s, mock := setupDB(t)

	// Split the prefix into the path and the query for the LIKE clause.
	splitIdx := strings.LastIndex(prefix, "/")
	pathName := prefix[:splitIdx+1]
	nameQuery := prefix[splitIdx+1:] + "%"

	// Modify the expected SQL to include the name LIKE clause
	expectedSQL1 := "SELECT * FROM `slash_prefix_tree_nodes_31` WHERE bucket_name = ? AND path_name = ? AND name like ? ORDER BY full_name LIMIT 3"
	mock.ExpectQuery(expectedSQL1).
		WithArgs(bucketName, pathName, nameQuery).
		WillReturnRows(
			sqlmock.NewRows([]string{"path_name", "full_name", "name", "is_object", "is_folder", "bucket_name", "object_id", "object_name"}).
				AddRow("testPath2", "testPath2/obj2", "obj2", true, false, bucketName, common.HexToHash("2"), "obj2"))

	// Set expectation for the object details query on the respective objects_xx table
	mock.ExpectQuery(fmt.Sprintf("SELECT * FROM `%s` WHERE object_id in (?)", GetObjectsTableName(bucketName))).
		WithArgs(common.HexToHash("2")).
		WillReturnRows(
			sqlmock.NewRows([]string{"object_id", "object_name"}).
				AddRow(common.HexToHash("2"), "obj2"))

	res, err := s.ListObjects(bucketName, "", prefix, maxKeys)
	assert.Nil(t, err)
	assert.Equal(t, expectedObjects, res)
}

func TestBsDBImpl_ListObjectsWithAllConditions(t *testing.T) {
	prefix := "testPath/testQuery"
	maxKeys := 2
	continuationToken := "testToken"
	expectedObjects := []*ListObjectsResult{
		{PathName: "testPath2/obj2", ResultType: "object", Object: &Object{ObjectName: "obj2", ObjectID: common.HexToHash("2")}},
	}

	s, mock := setupDB(t)

	// Split the prefix into the path and the query for the LIKE clause.
	splitIdx := strings.LastIndex(prefix, "/")
	pathName := prefix[:splitIdx+1]
	nameQuery := prefix[splitIdx+1:] + "%"

	expectedSQL := "SELECT * FROM `slash_prefix_tree_nodes_31` WHERE bucket_name = ? AND path_name = ? AND name like ? AND full_name >= ? ORDER BY full_name LIMIT 3"
	mock.ExpectQuery(expectedSQL).
		WithArgs(bucketName, pathName, nameQuery, continuationToken).
		WillReturnRows(
			sqlmock.NewRows([]string{"path_name", "full_name", "name", "is_object", "is_folder", "bucket_name", "object_id", "object_name"}).
				AddRow("testPath2", "testPath2/obj2", "obj2", true, false, bucketName, common.HexToHash("2"), "obj2"))

	mock.ExpectQuery(fmt.Sprintf("SELECT * FROM `%s` WHERE object_id in (?)", GetObjectsTableName(bucketName))).
		WithArgs(common.HexToHash("2")).
		WillReturnRows(
			sqlmock.NewRows([]string{"object_id", "object_name"}).
				AddRow(common.HexToHash("2"), "obj2"))

	res, err := s.ListObjects(bucketName, continuationToken, prefix, maxKeys)
	assert.Nil(t, err)
	assert.Equal(t, expectedObjects, res)
}
