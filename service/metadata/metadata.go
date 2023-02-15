package metadata

import (
	"context"
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
func (metadata *Metadata) Start(ctx context.Context) error {
	v1 := metadata.engine.Group("/api/v1")
	metadata.router.InitHandlers(v1)
	return nil
}

// Stop implement the lifecycle interface
func (metadata *Metadata) Stop(ctx context.Context) error {
	return nil
}

func NewMetadataService(cfg *MetadataConfig, ctx context.Context) (metadata *Metadata, err error) {
	store, _ := store.NewStore(cfg.DBConfig)
	metadata = &Metadata{
		ctx:    ctx,
		store:  store,
		engine: gin.Default(),
		router: router.NewRouter(store),
	}
	//err = metadata.initDB()
	return
}

//func (metadata *Metadata) Init() {
//	v1 := metadata.engine.Group("/api/v1")
//	metadata.router.InitHandlers(v1)
//}

//func (s *Server) Run() error {
//	//s.startLoop(s.ctx)
//	return s.engine.Run("127.0.0.1:8081")
//}

//func (s *Server) startLoop(ctx context.Context) {
//	go wait.UntilWithContext(ctx, s.emitStatistics, time.Hour)
//}
