package websocket

import (
        "encoding/json"
        "fmt"
        "net/http"
        "sync"
        "time"

        "yunwei/global"
        "yunwei/model/server"

        "github.com/gin-gonic/gin"
        "github.com/gorilla/websocket"
)

// MessageType 消息类型
type MessageType string

const (
        MessageTypeMetric    MessageType = "metric"
        MessageTypeAlert     MessageType = "alert"
        MessageTypeCommand   MessageType = "command"
        MessageTypeStatus    MessageType = "status"
        MessageTypeLog       MessageType = "log"
        MessageTypeDecision  MessageType = "decision"
        MessageTypeHeartbeat MessageType = "heartbeat"
)

// Message WebSocket消息
type Message struct {
        Type      MessageType `json:"type"`
        Timestamp time.Time   `json:"timestamp"`
        ServerID  uint        `json:"serverId,omitempty"`
        Data      interface{} `json:"data"`
}

// MetricMessage 指标消息
type MetricMessage struct {
        ServerID    uint    `json:"serverId"`
        ServerName  string  `json:"serverName"`
        CPUUsage    float64 `json:"cpuUsage"`
        MemoryUsage float64 `json:"memoryUsage"`
        DiskUsage   float64 `json:"diskUsage"`
        Load1       float64 `json:"load1"`
        Load5       float64 `json:"load5"`
        Load15      float64 `json:"load15"`
        NetIn       uint64  `json:"netIn"`
        NetOut      uint64  `json:"netOut"`
        Timestamp   time.Time `json:"timestamp"`
}

// AlertMessage 告警消息
type AlertMessage struct {
        AlertID   uint      `json:"alertId"`
        ServerID  uint      `json:"serverId"`
        Type      string    `json:"type"`
        Level     string    `json:"level"`
        Title     string    `json:"title"`
        Message   string    `json:"message"`
        Timestamp time.Time `json:"timestamp"`
}

// Client WebSocket客户端
type Client struct {
        ID         string
        Conn       *websocket.Conn
        Send       chan []byte
        ServerIDs  map[uint]bool // 订阅的服务器ID
        AllServers bool          // 订阅所有服务器
        UserID     uint
        mu         sync.Mutex
}

// Hub 连接管理中心
type Hub struct {
        Clients    map[*Client]bool
        Broadcast  chan []byte
        Register   chan *Client
        Unregister chan *Client
        mu         sync.RWMutex
}

// NewHub 创建Hub
func NewHub() *Hub {
        return &Hub{
                Clients:    make(map[*Client]bool),
                Broadcast:  make(chan []byte, 256),
                Register:   make(chan *Client),
                Unregister: make(chan *Client),
        }
}

// Run 运行Hub
func (h *Hub) Run() {
        for {
                select {
                case client := <-h.Register:
                        h.mu.Lock()
                        h.Clients[client] = true
                        h.mu.Unlock()
                        global.Logger.Info(fmt.Sprintf("WebSocket客户端连接: %s", client.ID))

                case client := <-h.Unregister:
                        h.mu.Lock()
                        if _, ok := h.Clients[client]; ok {
                                delete(h.Clients, client)
                                close(client.Send)
                        }
                        h.mu.Unlock()
                        global.Logger.Info(fmt.Sprintf("WebSocket客户端断开: %s", client.ID))

                case message := <-h.Broadcast:
                        h.mu.RLock()
                        for client := range h.Clients {
                                select {
                                case client.Send <- message:
                                default:
                                        close(client.Send)
                                        delete(h.Clients, client)
                                }
                        }
                        h.mu.RUnlock()
                }
        }
}

// BroadcastToServer 向订阅指定服务器的客户端广播
func (h *Hub) BroadcastToServer(serverID uint, message []byte) {
        h.mu.RLock()
        defer h.mu.RUnlock()

        for client := range h.Clients {
                if client.AllServers || client.ServerIDs[serverID] {
                        select {
                        case client.Send <- message:
                        default:
                                // 发送失败，关闭连接
                                close(client.Send)
                                delete(h.Clients, client)
                        }
                }
        }
}

// BroadcastToUser 向指定用户广播
func (h *Hub) BroadcastToUser(userID uint, message []byte) {
        h.mu.RLock()
        defer h.mu.RUnlock()

        for client := range h.Clients {
                if client.UserID == userID {
                        select {
                        case client.Send <- message:
                        default:
                                close(client.Send)
                                delete(h.Clients, client)
                        }
                }
        }
}

// GetClientCount 获取客户端数量
func (h *Hub) GetClientCount() int {
        h.mu.RLock()
        defer h.mu.RUnlock()
        return len(h.Clients)
}

// ReadPump 读取消息
func (c *Client) ReadPump(h *Hub, onMessage func([]byte)) {
        defer func() {
                h.Unregister <- c
                c.Conn.Close()
        }()

        c.Conn.SetReadLimit(512)
        c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
        c.Conn.SetPongHandler(func(string) error {
                c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
                return nil
        })

        for {
                _, message, err := c.Conn.ReadMessage()
                if err != nil {
                        break
                }

                // 处理客户端消息
                if onMessage != nil {
                        onMessage(message)
                }

                // 解析订阅消息
                var msg struct {
                        Type      string `json:"type"`
                        ServerID  uint   `json:"serverId"`
                        AllServers bool  `json:"allServers"`
                }
                if err := json.Unmarshal(message, &msg); err == nil {
                        c.mu.Lock()
                        switch msg.Type {
                        case "subscribe":
                                if msg.AllServers {
                                        c.AllServers = true
                                } else if msg.ServerID > 0 {
                                        c.ServerIDs[msg.ServerID] = true
                                }
                        case "unsubscribe":
                                if msg.ServerID > 0 {
                                        delete(c.ServerIDs, msg.ServerID)
                                }
                        }
                        c.mu.Unlock()
                }
        }
}

// WritePump 发送消息
func (c *Client) WritePump() {
        ticker := time.NewTicker(30 * time.Second)
        defer func() {
                ticker.Stop()
                c.Conn.Close()
        }()

        for {
                select {
                case message, ok := <-c.Send:
                        c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
                        if !ok {
                                c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
                                return
                        }

                        w, err := c.Conn.NextWriter(websocket.TextMessage)
                        if err != nil {
                                return
                        }
                        w.Write(message)

                        // 批量发送
                        n := len(c.Send)
                        for i := 0; i < n; i++ {
                                w.Write([]byte{'\n'})
                                w.Write(<-c.Send)
                        }

                        if err := w.Close(); err != nil {
                                return
                        }

                case <-ticker.C:
                        c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
                        if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                                return
                        }
                }
        }
}

// WebSocketService WebSocket服务
type WebSocketService struct {
        Hub *Hub
}

// NewWebSocketService 创建WebSocket服务
func NewWebSocketService() *WebSocketService {
        return &WebSocketService{
                Hub: NewHub(),
        }
}

// Start 启动服务
func (s *WebSocketService) Start() {
        go s.Hub.Run()
}

// PushMetric 推送指标
func (s *WebSocketService) PushMetric(serverID uint, metric *server.ServerMetric, serverName string) {
        msg := MetricMessage{
                ServerID:    serverID,
                ServerName:  serverName,
                CPUUsage:    metric.CPUUsage,
                MemoryUsage: metric.MemoryUsage,
                DiskUsage:   metric.DiskUsage,
                Load1:       metric.Load1,
                Load5:       metric.Load5,
                Load15:      metric.Load15,
                Timestamp:   time.Now(),
        }

        s.pushMessage(MessageTypeMetric, serverID, msg)
}

// PushAlert 推送告警
func (s *WebSocketService) PushAlert(alert AlertMessage) {
        s.pushMessage(MessageTypeAlert, alert.ServerID, alert)
}

// PushStatus 推送状态变化
func (s *WebSocketService) PushStatus(serverID uint, status string) {
        s.pushMessage(MessageTypeStatus, serverID, map[string]interface{}{
                "serverId": serverID,
                "status":   status,
        })
}

// PushLog 推送日志
func (s *WebSocketService) PushLog(serverID uint, logType, content string) {
        s.pushMessage(MessageTypeLog, serverID, map[string]interface{}{
                "serverId": serverID,
                "type":     logType,
                "content":  content,
        })
}

// PushDecision 推送AI决策
func (s *WebSocketService) PushDecision(serverID uint, decision interface{}) {
        s.pushMessage(MessageTypeDecision, serverID, decision)
}

// PushCommand 推送命令执行结果
func (s *WebSocketService) PushCommand(serverID uint, command, output, status string) {
        s.pushMessage(MessageTypeCommand, serverID, map[string]interface{}{
                "serverId": serverID,
                "command":  command,
                "output":   output,
                "status":   status,
        })
}

// pushMessage 推送消息
func (s *WebSocketService) pushMessage(msgType MessageType, serverID uint, data interface{}) {
        msg := Message{
                Type:      msgType,
                Timestamp: time.Now(),
                ServerID:  serverID,
                Data:      data,
        }

        jsonData, err := json.Marshal(msg)
        if err != nil {
                return
        }

        if serverID > 0 {
                s.Hub.BroadcastToServer(serverID, jsonData)
        } else {
                s.Hub.Broadcast <- jsonData
        }
}

// GetConnectedClients 获取连接的客户端数量
func (s *WebSocketService) GetConnectedClients() int {
        return s.Hub.GetClientCount()
}

// Broadcast 广播消息给所有客户端
func (s *WebSocketService) Broadcast(msgType MessageType, data interface{}) {
        s.pushMessage(msgType, 0, data)
}

// MonitorAndPush 监控并推送
func (s *WebSocketService) MonitorAndPush() {
        ticker := time.NewTicker(5 * time.Second)
        defer ticker.Stop()

        for range ticker.C {
                // 获取所有在线服务器
                var servers []server.Server
                global.DB.Where("agent_online = ?", true).Find(&servers)

                for _, srv := range servers {
                        // 获取最新指标
                        var metric server.ServerMetric
                        if err := global.DB.Where("server_id = ?", srv.ID).
                                Order("created_at DESC").
                                First(&metric).Error; err != nil {
                                continue
                        }

                        // 推送指标
                        s.PushMetric(srv.ID, &metric, srv.Name)
                }
        }
}

// upgrader WebSocket升级器
var upgrader = websocket.Upgrader{
        ReadBufferSize:  1024,
        WriteBufferSize: 1024,
        CheckOrigin: func(r *http.Request) bool {
                return true // 允许所有来源
        },
}

// HandleWebSocket 处理WebSocket连接
func (s *WebSocketService) HandleWebSocket(c *gin.Context) {
        conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
        if err != nil {
                return
        }

        client := &Client{
                ID:        generateClientID(),
                Conn:      conn,
                Send:      make(chan []byte, 256),
                ServerIDs: make(map[uint]bool),
                UserID:    0, // TODO: 从JWT获取
        }

        s.Hub.Register <- client

        // 启动读写协程
        go client.WritePump()
        go client.ReadPump(s.Hub, nil)
}

// generateClientID 生成客户端ID
func generateClientID() string {
        return fmt.Sprintf("%d", time.Now().UnixNano())
}
