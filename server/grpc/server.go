package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"yunwei/global"
	agentModel "yunwei/model/agent"
	"yunwei/model/server"
	agentService "yunwei/service/agent"

	"google.golang.org/grpc"
)

// AgentGRPCServer Agent gRPC服务
type AgentGRPCServer struct {
	UnimplementedAgentServiceServer
	port           string
	agentManager   *agentService.AgentManager
	heartbeatMon   *agentService.HeartbeatMonitor
	versionManager *agentService.VersionManager
}

// NewAgentGRPCServer 创建gRPC服务
func NewAgentGRPCServer(port string) *AgentGRPCServer {
	am := agentService.NewAgentManager()
	return &AgentGRPCServer{
		port:           port,
		agentManager:   am,
		heartbeatMon:   am.GetHeartbeatMonitor(),
		versionManager: am.GetVersionManager(),
	}
}

// Start 启动服务
func (s *AgentGRPCServer) Start() error {
	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("gRPC监听失败: %w", err)
	}

	grpcServer := grpc.NewServer()
	RegisterAgentServiceServer(grpcServer, s)

	// 启动心跳监控
	s.agentManager.Start()

	go func() {
		grpcServer.Serve(lis)
	}()

	return nil
}

// Stop 停止服务
func (s *AgentGRPCServer) Stop() {
	s.agentManager.Stop()
}

// ==================== 基础接口 ====================

// Heartbeat 心跳
func (s *AgentGRPCServer) Heartbeat(ctx context.Context, req *HeartbeatRequest) (*HeartbeatResponse, error) {
	// 使用心跳监控处理
	heartbeatReq := &agentService.HeartbeatRequest{
		AgentID:        req.AgentId,
		IP:             req.Ip,
		Port:           int(req.Port),
		Version:        req.Version,
		Platform:       req.Platform,
		Arch:           req.Arch,
		UptimeSeconds:  req.UptimeSeconds,
		CPUUsage:       req.CpuUsage,
		MemoryUsage:    req.MemoryUsage,
		GoroutineCount: int(req.GoroutineCount),
		PendingTasks:   int(req.PendingTasks),
		RunningTasks:   int(req.RunningTasks),
		CompletedTasks: int(req.CompletedTasks),
	}

	resp, err := s.heartbeatMon.ProcessHeartbeat(heartbeatReq)
	if err != nil {
		return &HeartbeatResponse{Success: false, Message: err.Error()}, nil
	}

	return &HeartbeatResponse{
		Success:       resp.Success,
		Message:       resp.Message,
		NeedUpgrade:   resp.NeedUpgrade,
		UpgradeTaskId: uint32(resp.UpgradeTaskID),
		TargetVersion: resp.TargetVersion,
	}, nil
}

// ReportMetrics 上报指标
func (s *AgentGRPCServer) ReportMetrics(ctx context.Context, req *MetricsRequest) (*MetricsResponse, error) {
	var srv server.Server
	if err := global.DB.Where("agent_id = ?", req.AgentId).First(&srv).Error; err != nil {
		return &MetricsResponse{Success: false, Message: "未注册"}, nil
	}

	var metric server.ServerMetric
	if err := json.Unmarshal([]byte(req.MetricsJson), &metric); err != nil {
		return &MetricsResponse{Success: false, Message: "解析失败"}, nil
	}

	metric.ServerID = srv.ID
	global.DB.Create(&metric)

	// 更新服务器状态
	srv.CPUUsage = metric.CPUUsage
	srv.MemoryUsage = metric.MemoryUsage
	srv.DiskUsage = metric.DiskUsage
	srv.Load1 = metric.Load1
	srv.Load5 = metric.Load5
	srv.Load15 = metric.Load15
	now := time.Now()
	srv.LastHeartbeat = &now
	srv.AgentOnline = true
	global.DB.Save(&srv)

	// 更新 Agent 状态
	var ag agentModel.Agent
	if err := global.DB.Where("agent_id = ?", req.AgentId).First(&ag).Error; err == nil {
		ag.LastHeartbeat = &now
		ag.Status = agentModel.AgentStatusOnline
		global.DB.Save(&ag)

		// 记录 Agent 指标
		agentMetric := &agentModel.AgentMetric{
			AgentID:        ag.ID,
			ServerID:       ag.ServerID,
			CPUUsage:       metric.CPUUsage,
			MemoryUsage:    metric.MemoryUsage,
			GoroutineCount: 0, // 从请求中获取
		}
		global.DB.Create(agentMetric)
	}

	return &MetricsResponse{Success: true, Message: "OK"}, nil
}

// RegisterAgent 注册Agent
func (s *AgentGRPCServer) RegisterAgent(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	// 使用 Agent 管理器注册
	registerReq := &agentService.RegisterRequest{
		AgentID:  req.AgentId,
		IP:       req.Ip,
		Platform: req.Os,
		Arch:     req.Arch,
		Version:  req.Version,
		Hostname: req.Hostname,
	}

	ag, err := s.agentManager.RegisterAgent(registerReq)
	if err != nil {
		return &RegisterResponse{Success: false, Message: err.Error()}, nil
	}

	// 同时更新/创建 Server 记录
	var srv server.Server
	result := global.DB.Where("host = ?", req.Ip).First(&srv)

	if result.Error != nil {
		srv = server.Server{
			Name:     req.Hostname,
			Hostname: req.Hostname,
			Host:     req.Ip,
			OS:       req.Os,
			Arch:     req.Arch,
			Kernel:   req.Kernel,
			CPUCores: int(req.CpuCores),
			AgentID:  req.AgentId,
			Status:   "online",
		}
		global.DB.Create(&srv)
	} else {
		srv.AgentID = req.AgentId
		srv.Hostname = req.Hostname
		now := time.Now()
		srv.LastHeartbeat = &now
		srv.AgentOnline = true
		global.DB.Save(&srv)
	}

	return &RegisterResponse{
		Success:  true,
		Message:  "注册成功",
		ServerId: uint32(srv.ID),
		AgentId:  uint32(ag.ID),
		Secret:   ag.AgentSecret,
	}, nil
}

// ExecuteCommand 执行命令
func (s *AgentGRPCServer) ExecuteCommand(ctx context.Context, req *CommandRequest) (*CommandResponse, error) {
	// TODO: 实现命令下发和执行
	return &CommandResponse{Success: true, Message: "已发送"}, nil
}

// ==================== 版本管理接口 ====================

// CheckUpgrade 检查升级
func (s *AgentGRPCServer) CheckUpgrade(ctx context.Context, req *CheckUpgradeRequest) (*CheckUpgradeResponse, error) {
	var ag agentModel.Agent
	if err := global.DB.Where("agent_id = ?", req.AgentId).First(&ag).Error; err != nil {
		return &CheckUpgradeResponse{Success: false, Message: "Agent未注册"}, nil
	}

	info, err := s.versionManager.CheckUpgrade(&ag)
	if err != nil {
		return &CheckUpgradeResponse{Success: false, Message: err.Error()}, nil
	}

	return &CheckUpgradeResponse{
		Success:      true,
		Message:      "OK",
		NeedUpgrade:  info.NeedUpgrade,
		CurrentVersion: info.CurrentVersion,
		TargetVersion:  info.TargetVersion,
		ForceUpdate:    info.ForceUpdate,
		Changelog:      info.Changelog,
		DownloadUrl:    info.DownloadURL,
		FileMd5:        info.FileMD5,
		FileSize:       info.FileSize,
	}, nil
}

// ReportUpgradeProgress 上报升级进度
func (s *AgentGRPCServer) ReportUpgradeProgress(ctx context.Context, req *UpgradeProgressRequest) (*UpgradeProgressResponse, error) {
	ue := s.agentManager.GetUpgradeEngine()

	switch req.Status {
	case "downloading":
		ue.HandleUpgradeProgress(uint(req.TaskId), int(req.Progress), "downloading", req.Message)
	case "installing":
		ue.HandleUpgradeProgress(uint(req.TaskId), int(req.Progress), "installing", req.Message)
	case "success":
		ue.HandleUpgradeSuccess(uint(req.TaskId), req.Output)
	case "failed":
		// 构造失败处理
		var task agentModel.AgentUpgradeTask
		if err := global.DB.First(&task, req.TaskId).Error; err == nil {
			task.Status = "failed"
			task.Error = req.Message
			task.Output = req.Output
			now := time.Now()
			task.CompletedAt = &now
			global.DB.Save(&task)
		}
	}

	return &UpgradeProgressResponse{Success: true, Message: "OK"}, nil
}

// GetAgentConfig 获取Agent配置
func (s *AgentGRPCServer) GetAgentConfig(ctx context.Context, req *AgentConfigRequest) (*AgentConfigResponse, error) {
	var ag agentModel.Agent
	if err := global.DB.Where("agent_id = ?", req.AgentId).First(&ag).Error; err != nil {
		return &AgentConfigResponse{Success: false, Message: "Agent未注册"}, nil
	}

	config, err := s.agentManager.GetAgentConfig(ag.ID)
	if err != nil {
		return &AgentConfigResponse{Success: false, Message: err.Error()}, nil
	}

	configJSON, _ := json.Marshal(config.Config)

	return &AgentConfigResponse{
		Success:        true,
		Message:        "OK",
		AutoUpgrade:    config.AutoUpgrade,
		UpgradeChannel: config.UpgradeChannel,
		AutoRecover:    config.AutoRecover,
		GrayGroup:      config.GrayGroup,
		GrayWeight:     int32(config.GrayWeight),
		ConfigJson:     string(configJSON),
		ConfigHash:     config.ConfigHash,
	}, nil
}

// ==================== 流式接口 ====================

// StreamHeartbeat 流式心跳
func (s *AgentGRPCServer) StreamHeartbeat(stream AgentService_StreamHeartbeatServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}

		// 处理心跳
		heartbeatReq := &agentService.HeartbeatRequest{
			AgentID:       req.AgentId,
			IP:            req.Ip,
			Version:       req.Version,
			Platform:      req.Platform,
			Arch:          req.Arch,
			UptimeSeconds: req.UptimeSeconds,
			CPUUsage:      req.CpuUsage,
			MemoryUsage:   req.MemoryUsage,
		}

		resp, err := s.heartbeatMon.ProcessHeartbeat(heartbeatReq)
		if err != nil {
			stream.Send(&HeartbeatStreamResponse{
				Success: false,
				Message: err.Error(),
			})
			continue
		}

		stream.Send(&HeartbeatStreamResponse{
			Success:       resp.Success,
			Message:       resp.Message,
			NeedUpgrade:   resp.NeedUpgrade,
			UpgradeTaskId: uint32(resp.UpgradeTaskID),
			TargetVersion: resp.TargetVersion,
		})
	}
}

// CommandStream 命令流
func (s *AgentGRPCServer) CommandStream(stream AgentService_CommandStreamServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}

		// 处理命令执行结果
		// TODO: 实现命令结果处理

		stream.Send(&CommandStreamResponse{
			Success: true,
			Message: "received",
		})
	}
}

// ==================== gRPC 接口定义 ====================

type AgentServiceServer interface {
	// 基础接口
	Heartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error)
	ReportMetrics(context.Context, *MetricsRequest) (*MetricsResponse, error)
	RegisterAgent(context.Context, *RegisterRequest) (*RegisterResponse, error)
	ExecuteCommand(context.Context, *CommandRequest) (*CommandResponse, error)

	// 版本管理
	CheckUpgrade(context.Context, *CheckUpgradeRequest) (*CheckUpgradeResponse, error)
	ReportUpgradeProgress(context.Context, *UpgradeProgressRequest) (*UpgradeProgressResponse, error)
	GetAgentConfig(context.Context, *AgentConfigRequest) (*AgentConfigResponse, error)

	// 流式接口
	StreamHeartbeat(AgentService_StreamHeartbeatServer) error
	CommandStream(AgentService_CommandStreamServer) error
}

type UnimplementedAgentServiceServer struct{}

func (UnimplementedAgentServiceServer) Heartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error) {
	return nil, fmt.Errorf("not implemented")
}
func (UnimplementedAgentServiceServer) ReportMetrics(context.Context, *MetricsRequest) (*MetricsResponse, error) {
	return nil, fmt.Errorf("not implemented")
}
func (UnimplementedAgentServiceServer) RegisterAgent(context.Context, *RegisterRequest) (*RegisterResponse, error) {
	return nil, fmt.Errorf("not implemented")
}
func (UnimplementedAgentServiceServer) ExecuteCommand(context.Context, *CommandRequest) (*CommandResponse, error) {
	return nil, fmt.Errorf("not implemented")
}
func (UnimplementedAgentServiceServer) CheckUpgrade(context.Context, *CheckUpgradeRequest) (*CheckUpgradeResponse, error) {
	return nil, fmt.Errorf("not implemented")
}
func (UnimplementedAgentServiceServer) ReportUpgradeProgress(context.Context, *UpgradeProgressRequest) (*UpgradeProgressResponse, error) {
	return nil, fmt.Errorf("not implemented")
}
func (UnimplementedAgentServiceServer) GetAgentConfig(context.Context, *AgentConfigRequest) (*AgentConfigResponse, error) {
	return nil, fmt.Errorf("not implemented")
}
func (UnimplementedAgentServiceServer) StreamHeartbeat(AgentService_StreamHeartbeatServer) error {
	return fmt.Errorf("not implemented")
}
func (UnimplementedAgentServiceServer) CommandStream(AgentService_CommandStreamServer) error {
	return fmt.Errorf("not implemented")
}

func RegisterAgentServiceServer(s *grpc.Server, srv AgentServiceServer) {
	s.RegisterService(&_AgentService_serviceDesc, srv)
}

var _AgentService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "AgentService",
	HandlerType: (*AgentServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "Heartbeat", Handler: _AgentService_Heartbeat_Handler},
		{MethodName: "ReportMetrics", Handler: _AgentService_ReportMetrics_Handler},
		{MethodName: "RegisterAgent", Handler: _AgentService_RegisterAgent_Handler},
		{MethodName: "ExecuteCommand", Handler: _AgentService_ExecuteCommand_Handler},
		{MethodName: "CheckUpgrade", Handler: _AgentService_CheckUpgrade_Handler},
		{MethodName: "ReportUpgradeProgress", Handler: _AgentService_ReportUpgradeProgress_Handler},
		{MethodName: "GetAgentConfig", Handler: _AgentService_GetAgentConfig_Handler},
	},
	Streams: []grpc.StreamDesc{
		{StreamName: "StreamHeartbeat", Handler: _AgentService_StreamHeartbeat_Handler},
		{StreamName: "CommandStream", Handler: _AgentService_CommandStream_Handler},
	},
	Metadata: "agent.proto",
}

// ==================== Handler 函数 ====================

func _AgentService_Heartbeat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HeartbeatRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	return srv.(AgentServiceServer).Heartbeat(ctx, in)
}

func _AgentService_ReportMetrics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MetricsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	return srv.(AgentServiceServer).ReportMetrics(ctx, in)
}

func _AgentService_RegisterAgent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	return srv.(AgentServiceServer).RegisterAgent(ctx, in)
}

func _AgentService_ExecuteCommand_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CommandRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	return srv.(AgentServiceServer).ExecuteCommand(ctx, in)
}

func _AgentService_CheckUpgrade_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckUpgradeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	return srv.(AgentServiceServer).CheckUpgrade(ctx, in)
}

func _AgentService_ReportUpgradeProgress_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpgradeProgressRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	return srv.(AgentServiceServer).ReportUpgradeProgress(ctx, in)
}

func _AgentService_GetAgentConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AgentConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	return srv.(AgentServiceServer).GetAgentConfig(ctx, in)
}

func _AgentService_StreamHeartbeat_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(AgentServiceServer).StreamHeartbeat(&agentServiceStreamHeartbeatServer{stream})
}

func _AgentService_CommandStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(AgentServiceServer).CommandStream(&agentServiceCommandStreamServer{stream})
}

// ==================== 流式接口辅助类型 ====================

type AgentService_StreamHeartbeatServer interface {
	Send(*HeartbeatStreamResponse) error
	Recv() (*HeartbeatStreamRequest, error)
	grpc.ServerStream
}

type agentServiceStreamHeartbeatServer struct {
	grpc.ServerStream
}

func (x *agentServiceStreamHeartbeatServer) Send(m *HeartbeatStreamResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *agentServiceStreamHeartbeatServer) Recv() (*HeartbeatStreamRequest, error) {
	m := new(HeartbeatStreamRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

type AgentService_CommandStreamServer interface {
	Send(*CommandStreamResponse) error
	Recv() (*CommandStreamRequest, error)
	grpc.ServerStream
}

type agentServiceCommandStreamServer struct {
	grpc.ServerStream
}

func (x *agentServiceCommandStreamServer) Send(m *CommandStreamResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *agentServiceCommandStreamServer) Recv() (*CommandStreamRequest, error) {
	m := new(CommandStreamRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ==================== 消息定义 ====================

// HeartbeatRequest 心跳请求
type HeartbeatRequest struct {
	AgentId        string  `json:"agentId"`
	Ip             string  `json:"ip"`
	Port           int32   `json:"port"`
	Version        string  `json:"version"`
	Platform       string  `json:"platform"`
	Arch           string  `json:"arch"`
	UptimeSeconds  int64   `json:"uptimeSeconds"`
	CpuUsage       float64 `json:"cpuUsage"`
	MemoryUsage    float64 `json:"memoryUsage"`
	GoroutineCount int32   `json:"goroutineCount"`
	PendingTasks   int32   `json:"pendingTasks"`
	RunningTasks   int32   `json:"runningTasks"`
	CompletedTasks int32   `json:"completedTasks"`
}

// HeartbeatResponse 心跳响应
type HeartbeatResponse struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	NeedUpgrade   bool   `json:"needUpgrade"`
	UpgradeTaskId uint32 `json:"upgradeTaskId"`
	TargetVersion string `json:"targetVersion"`
}

// HeartbeatStreamRequest 流式心跳请求
type HeartbeatStreamRequest struct {
	AgentId        string  `json:"agentId"`
	Ip             string  `json:"ip"`
	Version        string  `json:"version"`
	Platform       string  `json:"platform"`
	Arch           string  `json:"arch"`
	UptimeSeconds  int64   `json:"uptimeSeconds"`
	CpuUsage       float64 `json:"cpuUsage"`
	MemoryUsage    float64 `json:"memoryUsage"`
}

// HeartbeatStreamResponse 流式心跳响应
type HeartbeatStreamResponse struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	NeedUpgrade   bool   `json:"needUpgrade"`
	UpgradeTaskId uint32 `json:"upgradeTaskId"`
	TargetVersion string `json:"targetVersion"`
}

// MetricsRequest 指标上报请求
type MetricsRequest struct {
	AgentId    string `json:"agentId"`
	MetricsJson string `json:"metricsJson"`
}

// MetricsResponse 指标上报响应
type MetricsResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	AgentId  string `json:"agentId"`
	Ip       string `json:"ip"`
	Hostname string `json:"hostname"`
	Os       string `json:"os"`
	Arch     string `json:"arch"`
	Kernel   string `json:"kernel"`
	CpuCores uint32 `json:"cpuCores"`
	Version  string `json:"version"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	ServerId uint32 `json:"serverId"`
	AgentId  uint32 `json:"agentId"`
	Secret   string `json:"secret"`
}

// CommandRequest 命令请求
type CommandRequest struct {
	AgentId string `json:"agentId"`
	Command string `json:"command"`
	Timeout int32  `json:"timeout"`
}

// CommandResponse 命令响应
type CommandResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Output    string `json:"output"`
	ExitCode  int32  `json:"exitCode"`
	Duration  int64  `json:"duration"`
}

// CommandStreamRequest 命令流请求
type CommandStreamRequest struct {
	AgentId   string `json:"agentId"`
	CommandId string `json:"commandId"`
	Status    string `json:"status"`
	Output    string `json:"output"`
	ExitCode  int32  `json:"exitCode"`
}

// CommandStreamResponse 命令流响应
type CommandStreamResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	CommandId string `json:"commandId"`
}

// CheckUpgradeRequest 检查升级请求
type CheckUpgradeRequest struct {
	AgentId string `json:"agentId"`
}

// CheckUpgradeResponse 检查升级响应
type CheckUpgradeResponse struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	NeedUpgrade    bool   `json:"needUpgrade"`
	CurrentVersion string `json:"currentVersion"`
	TargetVersion  string `json:"targetVersion"`
	ForceUpdate    bool   `json:"forceUpdate"`
	Changelog      string `json:"changelog"`
	DownloadUrl    string `json:"downloadUrl"`
	FileMd5        string `json:"fileMd5"`
	FileSize       int64  `json:"fileSize"`
}

// UpgradeProgressRequest 升级进度请求
type UpgradeProgressRequest struct {
	TaskId   uint32  `json:"taskId"`
	AgentId  string  `json:"agentId"`
	Status   string  `json:"status"`
	Progress float32 `json:"progress"`
	Message  string  `json:"message"`
	Output   string  `json:"output"`
}

// UpgradeProgressResponse 升级进度响应
type UpgradeProgressResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// AgentConfigRequest Agent配置请求
type AgentConfigRequest struct {
	AgentId string `json:"agentId"`
}

// AgentConfigResponse Agent配置响应
type AgentConfigResponse struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	AutoUpgrade    bool   `json:"autoUpgrade"`
	UpgradeChannel string `json:"upgradeChannel"`
	AutoRecover    bool   `json:"autoRecover"`
	GrayGroup      string `json:"grayGroup"`
	GrayWeight     int32  `json:"grayWeight"`
	ConfigJson     string `json:"configJson"`
	ConfigHash     string `json:"configHash"`
}
