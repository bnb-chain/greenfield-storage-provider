package bsdb

import (
	"path/filepath"
	"strings"

	"github.com/forbole/juno/v4/common"
	"gorm.io/gorm"
)

// ListObjects List objects by bucket name
func (b *BsDBImpl) ListObjects(bucketName, continuationToken, prefix string, maxKeys int) ([]*ListObjectsResult, error) {
	var (
		nodes       []*SlashPrefixTreeNode
		filters     []func(*gorm.DB) *gorm.DB
		res         []*ListObjectsResult
		pathName    string
		prefixQuery string
		limit       int
		err         error
	)
	// return NextContinuationToken by adding 1 additionally
	limit = maxKeys + 1
	strings.Split(prefix, "/")
	pathName, prefixQuery = processPath(prefix)
	if pathName != "" {
		filters = append(filters, PathNameFilter(pathName))
	}
	if prefixQuery != "" {
		filters = append(filters, NameFilter(prefixQuery))
	}
	if continuationToken != "" {
		filters = append(filters, FullNameFilter(continuationToken))
	}
	err = b.db.Table((&SlashPrefixTreeNode{}).TableName()).
		Where("bucket_name = ?", bucketName).
		Scopes(filters...).
		Order("full_name").
		Limit(limit).
		Find(&nodes).Error
	if err != nil {
		return nil, err
	}
	res, err = b.filterObjects(nodes)
	return res, err
}

// processPath takes in a string that is a path name, and returns two strings:
// the directory part of the path, and the file part of the path. If the path does not contain
// a "/", then the directory is "/" and the file is the path.
func processPath(pathName string) (string, string) {
	var (
		dir  string
		file string
	)

	if !strings.Contains(pathName, "/") {
		dir = "/"
		file = pathName
	} else {
		dir, file = filepath.Split(pathName)
	}

	return dir, file
}

// filterObjects filters a slice of SlashPrefixTreeNode for nodes which IsObject attribute is true,
// maps these objects by their ID and transforms them into a ListObjectsResult format.
// Returns a slice of ListObjectsResult containing filtered object data or an error if something goes wrong.
func (b *BsDBImpl) filterObjects(nodes []*SlashPrefixTreeNode) ([]*ListObjectsResult, error) {
	var (
		objectIDs  []common.Hash
		objects    []*Object
		res        []*ListObjectsResult
		objectsMap map[common.Hash]*Object
		err        error
	)

	//filter objects and query the info
	for _, node := range nodes {
		if node.IsObject {
			objectIDs = append(objectIDs, node.ObjectID)
		}
	}

	err = b.db.Table((&Object{}).TableName()).
		Where("object_id in (?)", objectIDs).
		Find(&objects).Error
	if err != nil {
		return nil, err
	}

	objectsMap = make(map[common.Hash]*Object)
	for _, object := range objects {
		objectsMap[object.ObjectID] = object
	}

	for _, node := range nodes {
		if node.IsObject {
			object := objectsMap[node.ObjectID]
			if object == nil {
				continue
			}
			res = append(res, &ListObjectsResult{
				PathName:   node.FullName,
				ResultType: ObjectName,
				Object:     object,
			})
		} else {
			res = append(res, &ListObjectsResult{
				PathName:   node.FullName,
				ResultType: CommonPrefix,
				Object:     &Object{},
			})
		}
	}
	return res, nil
}
