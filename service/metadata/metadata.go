package metadata

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata/router"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata/store"
	"github.com/gin-gonic/gin"
)

type Metadata struct {
	ctx    context.Context
	name   string
	store  store.IStore
	engine *gin.Engine
	router *router.Router
}

// Name implement the lifecycle interface
func (metadata *Metadata) Name() string {
	return metadata.name
}

// Start implement the lifecycle interface
// to delete api/v1
func (metadata *Metadata) Start(ctx context.Context) error {
	v1 := metadata.engine.Group("")
	metadata.router.InitHandlers(v1)

	go metadata.Run()
	log.Debug("metadata service succeed to start")
	return nil
}

func (metadata *Metadata) Run() {
	if err := metadata.engine.Run("127.0.0.1:9733"); err != nil {
		log.Errorw("failed to listen", "err", err)
		return
	}
}

// Stop implement the lifecycle interface
func (metadata *Metadata) Stop(ctx context.Context) error {
	return nil
}

func NewMetadataService(cfg *MetadataConfig, ctx context.Context) (metadata *Metadata, err error) {
	store, _ := store.NewStore(cfg.SpDBConfig)
	metadata = &Metadata{
		name:   model.MetadataService,
		ctx:    ctx,
		store:  store,
		engine: gin.Default(),
		router: router.NewRouter(store),
	}
	// err = metadata.initDB()
	return
}
