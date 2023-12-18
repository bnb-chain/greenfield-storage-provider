package prefixtree

import (
	"context"
	"errors"
	"fmt"
	"strings"

	abci "github.com/cometbft/cometbft/abci/types"
	tmctypes "github.com/cometbft/cometbft/rpc/core/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/forbole/juno/v4/common"
	"github.com/forbole/juno/v4/log"

	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var (
	EventCreateObject       = proto.MessageName(&storagetypes.EventCreateObject{})
	EventDeleteObject       = proto.MessageName(&storagetypes.EventDeleteObject{})
	EventCancelCreateObject = proto.MessageName(&storagetypes.EventCancelCreateObject{})
	EventRejectSealObject   = proto.MessageName(&storagetypes.EventRejectSealObject{})
)

// BuildPrefixTreeEvents maps event types that trigger the creation or deletion of prefix tree nodes.
// If an event type is present and set to true in this map,
// it means that event will result in changes to the prefix tree structure.
var BuildPrefixTreeEvents = map[string]bool{
	EventCreateObject:       true,
	EventDeleteObject:       true,
	EventCancelCreateObject: true,
	EventRejectSealObject:   true,
}

func (m *Module) ExtractEventStatements(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, event sdk.Event) (map[string][]interface{}, error) {
	if !BuildPrefixTreeEvents[event.Type] {
		return nil, nil
	}

	typedEvent, err := sdk.ParseTypedEvent(abci.Event(event))
	if err != nil {
		log.Errorw("parse typed events error", "module", m.Name(), "event", event, "err", err)
		return nil, err
	}

	switch event.Type {
	case EventCreateObject:
		createObject, ok := typedEvent.(*storagetypes.EventCreateObject)
		if !ok {
			log.Errorw("type assert error", "type", "EventCreateObject", "event", typedEvent)
			return nil, errors.New("create object event assert error")
		}
		return m.handleCreateObject(ctx, createObject)
	case EventDeleteObject:
		deleteObject, ok := typedEvent.(*storagetypes.EventDeleteObject)
		if !ok {
			log.Errorw("type assert error", "type", "EventDeleteObject", "event", typedEvent)
			return nil, errors.New("delete object event assert error")
		}
		return m.handleDeleteObject(ctx, deleteObject)
	case EventCancelCreateObject:
		cancelObject, ok := typedEvent.(*storagetypes.EventCancelCreateObject)
		if !ok {
			log.Errorw("type assert error", "type", "EventCancelCreateObject", "event", typedEvent)
			return nil, errors.New("cancel create object event assert error")
		}
		return m.handleCancelCreateObject(ctx, cancelObject)
	case EventRejectSealObject:
		rejectSealObject, ok := typedEvent.(*storagetypes.EventRejectSealObject)
		if !ok {
			log.Errorw("type assert error", "type", "EventRejectSealObject", "event", typedEvent)
			return nil, errors.New("reject seal object event assert error")
		}
		return m.handleRejectSealObject(ctx, rejectSealObject)
	default:
		return nil, nil
	}
}

// HandleEvent handles the events relevant to the building of the PrefixTree.
// It checks the type of the event and calls the appropriate handler for it.
func (m *Module) HandleEvent(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, event sdk.Event) error {
	return nil
}

// handleCreateObject handles EventCreateObject.
// It builds the directory tree structure for the object if necessary.
func (m *Module) handleCreateObject(ctx context.Context, createObject *storagetypes.EventCreateObject) (map[string][]interface{}, error) {
	var nodes []*bsdb.SlashPrefixTreeNode
	objectPath := createObject.ObjectName
	bucketName := createObject.BucketName
	objectID := createObject.ObjectId

	// Split full path to get the directories
	pathParts := strings.Split(objectPath, "/")

	// Traverse from the deepest directory up to the root
	for i := len(pathParts) - 1; i > 0; i-- {
		path := strings.Join(pathParts[:i], "/") + "/"
		// Check if the current directory exists
		// If the context doesn't contain this folder
		if m.GetCtx(GenerateFolderExistKey(path)) == nil {
			// folder is not existed in the context value
			tree, err := m.db.GetPrefixTree(ctx, path, bucketName)
			if err != nil {
				log.Errorw("failed to get prefix tree", "error", err)
				return nil, err
			}
			if tree == nil {
				// Create the tree node and set it to exist in the context
				m.SetCtx(GenerateFolderExistKey(path), 1)
				// If the directory does not exist, create it
				newNode := &bsdb.SlashPrefixTreeNode{
					PathName:   strings.Join(pathParts[:i-1], "/") + "/",
					FullName:   path,
					Name:       pathParts[i-1] + "/",
					IsObject:   false,
					IsFolder:   true,
					BucketName: bucketName,
					ObjectName: "",
				}
				nodes = append(nodes, newNode)
			} else {
				// If the directory exists, we can break the loop
				break
			}
		} else {
			// If the context contains this folder
			if intVal, ok := m.GetCtx(GenerateFolderExistKey(path)).(int64); ok {
				// If intVal = 0 which means this folder already been deleted, we are supposed to create a new one
				if intVal == 0 {
					m.SetCtx(GenerateFolderExistKey(path), 1)
					// If the directory does not exist, create it
					newNode := &bsdb.SlashPrefixTreeNode{
						PathName:   strings.Join(pathParts[:i-1], "/") + "/",
						FullName:   path,
						Name:       pathParts[i-1] + "/",
						IsObject:   false,
						IsFolder:   true,
						BucketName: bucketName,
						ObjectName: "",
					}
					nodes = append(nodes, newNode)
				}
			} else {
				// If the directory exists, we can break the loop
				break
			}
		}
		// check if this folder count key exist in the context
		// if yes, we will get count from context and update it to count + 1
		// if not, we will generate folder count key and set value count + 1 in the context
		if m.GetCtx(GenerateFolderCountKey(path)) == nil {
			count, err := m.db.GetPrefixTreeCount(ctx, path, bucketName)
			if err != nil {
				log.Errorw("failed to get prefix tree count", "error", err)
				return nil, err
			}
			m.SetCtx(GenerateFolderCountKey(path), count+1)
		} else {
			if intVal, ok := m.GetCtx(GenerateFolderCountKey(path)).(int64); ok {
				m.SetCtx(GenerateFolderCountKey(path), intVal+1)
			}
		}
	}

	// if object is not exist in the context
	if m.GetCtx(GenerateObjectExistKey(common.BigToHash(objectID.BigInt()).String())) == nil {
		object, err := m.db.GetPrefixTreeObject(ctx, common.BigToHash(objectID.BigInt()), bucketName)
		if err != nil {
			log.Errorw("failed to get prefix tree object", "error", err)
			return nil, err
		}
		if object == nil {
			m.SetCtx(GenerateObjectExistKey(common.BigToHash(objectID.BigInt()).String()), 1)
			objectNode := &bsdb.SlashPrefixTreeNode{
				PathName:   strings.Join(pathParts[:len(pathParts)-1], "/") + "/",
				FullName:   objectPath,
				Name:       pathParts[len(pathParts)-1],
				IsObject:   true,
				IsFolder:   false,
				BucketName: bucketName,
				ObjectID:   common.BigToHash(objectID.BigInt()),
				ObjectName: objectPath,
			}
			nodes = append(nodes, objectNode)
		}
	} else {
		// If the context contains this object
		if intVal, ok := m.GetCtx(GenerateObjectExistKey(common.BigToHash(objectID.BigInt()).String())).(int); ok {
			// if intVal = 0 which means this folder already been deleted, we are supposed to create a new one
			if intVal == 0 {
				m.SetCtx(GenerateObjectExistKey(common.BigToHash(objectID.BigInt()).String()), 1)
				objectNode := &bsdb.SlashPrefixTreeNode{
					PathName:   strings.Join(pathParts[:len(pathParts)-1], "/") + "/",
					FullName:   objectPath,
					Name:       pathParts[len(pathParts)-1],
					IsObject:   true,
					IsFolder:   false,
					BucketName: bucketName,
					ObjectID:   common.BigToHash(objectID.BigInt()),
					ObjectName: objectPath,
				}
				nodes = append(nodes, objectNode)
			}
		}
	}
	if len(nodes) == 0 {
		return nil, nil
	}
	k, v := m.db.CreatePrefixTree(ctx, nodes)
	return map[string][]interface{}{
		k: v,
	}, nil
}

// handleDeleteObject handles EventDeleteObject.
// It removes the directory tree structure associated with the object.
func (m *Module) handleDeleteObject(ctx context.Context, deleteObject *storagetypes.EventDeleteObject) (map[string][]interface{}, error) {
	return m.deleteObject(ctx, deleteObject.ObjectName, deleteObject.BucketName, common.BigToHash(deleteObject.ObjectId.BigInt()))
}

// handleCancelCreateObject handles EventCancelCreateObject.
// It removes the directory tree structure associated with the object.
func (m *Module) handleCancelCreateObject(ctx context.Context, cancelCreateObject *storagetypes.EventCancelCreateObject) (map[string][]interface{}, error) {
	return m.deleteObject(ctx, cancelCreateObject.ObjectName, cancelCreateObject.BucketName, common.BigToHash(cancelCreateObject.ObjectId.BigInt()))
}

// handleRejectSealObject handles EventRejectSealObject.
// It removes the directory tree structure associated with the object.
func (m *Module) handleRejectSealObject(ctx context.Context, rejectSealObject *storagetypes.EventRejectSealObject) (map[string][]interface{}, error) {
	return m.deleteObject(ctx, rejectSealObject.ObjectName, rejectSealObject.BucketName, common.BigToHash(rejectSealObject.ObjectId.BigInt()))
}

// deleteObject according to the given object path and bucket name.
func (m *Module) deleteObject(ctx context.Context, objectPath, bucketName string, objectID common.Hash) (map[string][]interface{}, error) {
	var nodes []*bsdb.SlashPrefixTreeNode
	var err error
	// Split full path to get the directories
	pathParts := strings.Split(objectPath, "/")
	nodes = append(nodes, &bsdb.SlashPrefixTreeNode{
		FullName:   objectPath,
		IsObject:   true,
		BucketName: bucketName,
	})

	// Check and delete any empty parent directories
	for i := len(pathParts) - 1; i > 0; i-- {
		var count int64
		path := strings.Join(pathParts[:i], "/") + "/"
		// If the context don't contain this count, we need to update all the folder count under this path
		if m.GetCtx(GenerateFolderCountKey(path)) == nil {
			count, err = m.db.GetPrefixTreeCount(ctx, path, bucketName)
			if err != nil {
				log.Errorw("failed to get prefix tree count", "error", err)
				return nil, err
			}
			m.SetCtx(GenerateFolderCountKey(path), count-1)
		} else {
			// If the context contains this folder
			if intVal, ok := m.GetCtx(GenerateFolderCountKey(path)).(int64); ok {
				count = intVal
				m.SetCtx(GenerateFolderCountKey(path), intVal-1)
			}
		}
		if count <= 1 {
			nodes = append(nodes, &bsdb.SlashPrefixTreeNode{
				FullName:   path,
				IsObject:   false,
				BucketName: bucketName,
			})
			m.SetCtx(GenerateFolderExistKey(path), 0)
		} else {
			// Found a non-empty directory, stop here
			break
		}
	}
	if len(nodes) == 0 {
		return nil, nil
	}
	k, v := m.db.DeletePrefixTree(ctx, nodes)
	m.SetCtx(GenerateObjectExistKey(objectID.String()), 0)
	return map[string][]interface{}{
		k: v,
	}, nil
}

func GenerateFolderExistKey(folder string) string {
	return fmt.Sprintf("slash_prefix_tree:folder:exist:%s", folder)
}

func GenerateObjectExistKey(object string) string {
	return fmt.Sprintf("slash_prefix_tree:object:exist:%s", object)
}

func GenerateFolderCountKey(folder string) string {
	return fmt.Sprintf("slash_prefix_tree:folder:count:%s", folder)
}
