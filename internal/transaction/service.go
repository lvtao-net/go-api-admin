package transaction

import (
	"fmt"

	"github.com/lvtao/go-gin-api-admin/internal/model"
	"github.com/lvtao/go-gin-api-admin/internal/repository"
	"github.com/lvtao/go-gin-api-admin/pkg/database"
)

// Service 事务服务
type Service struct {
	engine     *Engine
	collection *repository.CollectionRepository
}

// NewService 创建事务服务
func NewService() *Service {
	return &Service{
		engine:     NewEngine(),
		collection: repository.NewCollectionRepository(database.GetDB()),
	}
}

// Execute 执行事务
func (s *Service) Execute(collectionName string, params map[string]interface{}, userID uint) (*TransactionResult, error) {
	// 获取集合配置
	collection, err := s.collection.GetByName(collectionName)
	if err != nil {
		return nil, fmt.Errorf("事务集合不存在: %s", collectionName)
	}

	return s.engine.Execute(collection, params, userID)
}

// ExecuteWithCollection 使用集合对象执行事务
func (s *Service) ExecuteWithCollection(collection *model.Collection, params map[string]interface{}, userID uint) (*TransactionResult, error) {
	return s.engine.Execute(collection, params, userID)
}
