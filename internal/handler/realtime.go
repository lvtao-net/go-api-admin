package handler

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lvtao/go-gin-api-admin/internal/service"
	"github.com/lvtao/go-gin-api-admin/pkg/response"
)

type RealtimeHandler struct {
	clients    map[string]map[*chan string]bool
	clientsMux sync.RWMutex
	subs       map[string]map[*chan string]bool // 订阅特定集合
	subsMux    sync.RWMutex
}

type SubscriptionRequest struct {
	Collection string `json:"collection"`
	RecordID   uint64 `json:"recordId"`
}

type RealtimeMessage struct {
	Action     string                 `json:"action"`
	Collection string                 `json:"collection"`
	RecordID   uint64                 `json:"recordId,omitempty"`
	Record     map[string]interface{} `json:"record,omitempty"`
	Timestamp  int64                  `json:"timestamp"`
}

var (
	globalHandler *RealtimeHandler
	handlerOnce   sync.Once
)

func GetRealtimeHandler() *RealtimeHandler {
	handlerOnce.Do(func() {
		globalHandler = NewRealtimeHandler()
		// 启动事件转发
		go globalHandler.forwardEvents()
	})
	return globalHandler
}

func NewRealtimeHandler() *RealtimeHandler {
	return &RealtimeHandler{
		clients: make(map[string]map[*chan string]bool),
		subs:    make(map[string]map[*chan string]bool),
	}
}

// forwardEvents 转发事件到订阅者
func (h *RealtimeHandler) forwardEvents() {
	ch := service.Subscribe()
	for event := range ch {
	msg := RealtimeMessage{
		Action:     event.Action,
		Collection: event.Collection,
		RecordID:   event.RecordID,
		Record:     event.Record,
		Timestamp:  time.Now().Unix(),
	}

		data, _ := json.Marshal(msg)
		message := fmt.Sprintf("data: %s\n\n", string(data))

		// 广播给所有订阅该集合的客户端
		h.subsMux.RLock()
		for collection, channels := range h.subs {
			if collection == event.Collection {
				for ch := range channels {
					select {
					case *ch <- message:
					default:
					}
				}
			}
		}
		h.subsMux.RUnlock()

		// 也广播给所有客户端
		h.clientsMux.RLock()
		for _, channels := range h.clients {
			for ch := range channels {
				select {
				case *ch <- message:
				default:
				}
			}
		}
		h.clientsMux.RUnlock()
	}
}

// Connect 建立 SSE 连接
func (h *RealtimeHandler) Connect(c *gin.Context) {
	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 创建消息通道
	messageChan := make(chan string, 100)
	clientChan := &messageChan

	// 生成客户端ID
	clientID := c.Query("clientId")
	if clientID == "" {
		clientID = fmt.Sprintf("client_%d", time.Now().UnixNano())
	}

	// 注册客户端
	h.clientsMux.Lock()
	if h.clients[clientID] == nil {
		h.clients[clientID] = make(map[*chan string]bool)
	}
	h.clients[clientID][clientChan] = true
	h.clientsMux.Unlock()

	// 发送连接成功消息
	connectMsg := fmt.Sprintf("event: connected\ndata: {\"clientId\":\"%s\",\"status\":\"connected\"}\n\n", clientID)
	c.Writer.WriteString(connectMsg)
	c.Writer.Flush()

	// 启动心跳
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.Writer.WriteString(": heartbeat\n\n")
				c.Writer.Flush()
			}
		}
	}()

	// 监听消息
	closeChan := c.Request.Context().Done()
	for {
		select {
		case <-closeChan:
			// 移除客户端
			h.clientsMux.Lock()
			delete(h.clients[clientID], clientChan)
			if len(h.clients[clientID]) == 0 {
				delete(h.clients, clientID)
			}
			h.clientsMux.Unlock()
			return
		case msg := <-messageChan:
			c.Writer.WriteString(msg)
			c.Writer.Flush()
		}
	}
}

// Subscribe 订阅
func (h *RealtimeHandler) Subscribe(c *gin.Context) {
	var req SubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	clientID := c.Query("clientId")
	if clientID == "" {
		clientID = fmt.Sprintf("client_%d", time.Now().UnixNano())
	}

	// 创建消息通道
	messageChan := make(chan string, 100)
	clientChan := &messageChan

	// 注册订阅
	h.subsMux.Lock()
	if h.subs[req.Collection] == nil {
		h.subs[req.Collection] = make(map[*chan string]bool)
	}
	h.subs[req.Collection][clientChan] = true
	h.subsMux.Unlock()

	// 发送订阅成功消息
	response.Success(c, gin.H{
		"message":    "Subscribed successfully",
		"collection": req.Collection,
		"recordId":   req.RecordID,
	})

	// 监听消息（后台）
	go func() {
		for {
			select {
			case msg := <-messageChan:
				_ = msg
			}
		}
	}()
}

// Unsubscribe 取消订阅
func (h *RealtimeHandler) Unsubscribe(c *gin.Context) {
	var req SubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	h.subsMux.Lock()
	if h.subs[req.Collection] != nil {
		delete(h.subs, req.Collection)
	}
	h.subsMux.Unlock()

	response.Success(c, gin.H{
		"message":    "Unsubscribed successfully",
		"collection": req.Collection,
	})
}

// Broadcast 广播消息（内部使用）
func (h *RealtimeHandler) Broadcast(collection string, recordID uint64, record map[string]interface{}) {
	msg := RealtimeMessage{
		Action:     "create",
		Collection: collection,
		RecordID:   recordID,
		Record:     record,
		Timestamp:  time.Now().Unix(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	message := fmt.Sprintf("data: %s\n\n", string(data))

	// 广播给所有订阅该集合的客户端
	h.clientsMux.RLock()
	defer h.clientsMux.RUnlock()

	for _, channels := range h.clients {
		for ch := range channels {
			select {
			case *ch <- message:
			default:
			}
		}
	}
}
