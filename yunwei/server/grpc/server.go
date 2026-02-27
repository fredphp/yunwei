package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"yunwei/global"
	"yunwei/model/server"

	"google.golang.org/grpc"
)

// AgentGRPCServer Agent gRPC服务
type AgentGRPCServer struct {
	UnimplementedAgentServiceServer
	port string
}

// NewAgentGRPCServer 创建gRPC服务
func NewAgentGRPCServer(port string) *AgentGRPCServer {
	return &AgentGRPCServer{
		port: port,
	}
}

// Start 启动gRPC服务
func (s *AgentGRPCServer) Start() error {
	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("gRPC监听失败: %w", err)
	}

	grpcServer := grpc.NewServer()
	RegisterAgentServiceServer(grpcServer, s)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			global.Logger.Error(fmt.Sprintf("gRPC服务错误: %v", err))
		}
	}()

	return nil
}

// Heartbeat 心跳
func (s *AgentGRPCServer) Heartbeat(ctx context.Context, req *HeartbeatRequest) (*HeartbeatResponse, error) {
	// 更新服务器状态
	var srv server.Server
	if err := global.DB.Where("agent_id = ?", req.AgentId).First(&srv).Error; err != nil {
		return &HeartbeatResponse{
			Success: false,
			Message: "服务器未注册",
		}, nil
	}

	now := time.Now()
	srv.LastHeartbeat = &now
	srv.AgentOnline = true
	global.DB.Save(&srv)

	return &HeartbeatResponse{
		Success: true,
		Message: "OK",
	}, nil
}

// ReportMetrics 上报指标
func (s *AgentGRPCServer) ReportMetrics(ctx context.Context, req *MetricsRequest) (*MetricsResponse, error) {
	// 查找服务器
	var srv server.Server
	if err := global.DB.Where("agent_id = ?", req.AgentId).First(&srv).Error; err != nil {
		return &MetricsResponse{
			Success: false,
			Message: "服务器未注册",
		}, nil
	}

	// 解析指标
	var metric server.ServerMetric
	if err := json.Unmarshal([]byte(req.MetricsJson), &metric); err != nil {
		return &MetricsResponse{
			Success: false,
			Message: "指标解析失败",
		}, nil
	}

	// 设置服务器ID
	metric.ServerID = srv.ID

	// 保存指标
	if err := global.DB.Create(&metric).Error; err != nil {
		return &MetricsResponse{
			Success: false,
			Message: "保存指标失败",
		}, nil
	}

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

	return &MetricsResponse{
		Success: true,
		Message: "OK",
	}, nil
}

// ReportDocker 上报Docker容器
func (s *AgentGRPCServer) ReportDocker(ctx context.Context, req *DockerRequest) (*DockerResponse, error) {
	var srv server.Server
	if err := global.DB.Where("agent_id = ?", req.AgentId).First(&srv).Error; err != nil {
		return &DockerResponse{
			Success: false,
			Message: "服务器未注册",
		}, nil
	}

	// 删除旧容器数据
	global.DB.Where("server_id = ?", srv.ID).Delete(&server.DockerContainer{})

	// 解析并保存新数据
	var containers []server.DockerContainer
	if err := json.Unmarshal([]byte(req.ContainersJson), &containers); err != nil {
		return &DockerResponse{
			Success: false,
			Message: "数据解析失败",
		}, nil
	}

	for i := range containers {
		containers[i].ServerID = srv.ID
	}

	if len(containers) > 0 {
		global.DB.Create(&containers)
	}

	return &DockerResponse{
		Success: true,
		Message: "OK",
	}, nil
}

// ReportPorts 上报端口信息
func (s *AgentGRPCServer) ReportPorts(ctx context.Context, req *PortsRequest) (*PortsResponse, error) {
	var srv server.Server
	if err := global.DB.Where("agent_id = ?", req.AgentId).First(&srv).Error; err != nil {
		return &PortsResponse{
			Success: false,
			Message: "服务器未注册",
		}, nil
	}

	// 删除旧端口数据
	global.DB.Where("server_id = ?", srv.ID).Delete(&server.PortInfo{})

	// 解析并保存新数据
	var ports []server.PortInfo
	if err := json.Unmarshal([]byte(req.PortsJson), &ports); err != nil {
		return &PortsResponse{
			Success: false,
			Message: "数据解析失败",
		}, nil
	}

	for i := range ports {
		ports[i].ServerID = srv.ID
	}

	if len(ports) > 0 {
		global.DB.Create(&ports)
	}

	return &PortsResponse{
		Success: true,
		Message: "OK",
	}, nil
}

// ExecuteCommand 执行命令
func (s *AgentGRPCServer) ExecuteCommand(ctx context.Context, req *CommandRequest) (*CommandResponse, error) {
	var srv server.Server
	if err := global.DB.Where("agent_id = ?", req.AgentId).First(&srv).Error; err != nil {
		return &CommandResponse{
			Success: false,
			Message: "服务器未注册",
		}, nil
	}

	// TODO: 通过Agent执行命令并返回结果
	// 实际实现需要与Agent建立双向通信

	// 记录日志
	log := server.ServerLog{
		ServerID: srv.ID,
		Type:     "command",
		Content:  req.Command,
	}

	startTime := time.Now()
	// 执行命令...
	log.Duration = time.Since(startTime).Milliseconds()

	global.DB.Create(&log)

	return &CommandResponse{
		Success: true,
		Message: "命令已发送",
	}, nil
}

// RegisterAgent 注册Agent
func (s *AgentGRPCServer) RegisterAgent(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	// 查找服务器
	var srv server.Server
	result := global.DB.Where("host = ?", req.Ip).First(&srv)

	if result.Error != nil {
		// 创建新服务器
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
		// 更新服务器信息
		srv.AgentID = req.AgentId
		srv.Hostname = req.Hostname
		srv.OS = req.Os
		srv.Arch = req.Arch
		srv.Kernel = req.Kernel
		srv.CPUCores = int(req.CpuCores)
		now := time.Now()
		srv.LastHeartbeat = &now
		srv.AgentOnline = true
		global.DB.Save(&srv)
	}

	return &RegisterResponse{
		Success:  true,
		Message:  "注册成功",
		ServerId: uint32(srv.ID),
	}, nil
}

// Agent服务定义 - Protocol Buffers 生成的代码
// 由于没有proto文件，这里手动定义接口

type AgentServiceServer interface {
	Heartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error)
	ReportMetrics(context.Context, *MetricsRequest) (*MetricsResponse, error)
	ReportDocker(context.Context, *DockerRequest) (*DockerResponse, error)
	ReportPorts(context.Context, *PortsRequest) (*PortsResponse, error)
	ExecuteCommand(context.Context, *CommandRequest) (*CommandResponse, error)
	RegisterAgent(context.Context, *RegisterRequest) (*RegisterResponse, error)
}

type UnimplementedAgentServiceServer struct{}

func (UnimplementedAgentServiceServer) Heartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error) {
	return nil, fmt.Errorf("not implemented")
}
func (UnimplementedAgentServiceServer) ReportMetrics(context.Context, *MetricsRequest) (*MetricsResponse, error) {
	return nil, fmt.Errorf("not implemented")
}
func (UnimplementedAgentServiceServer) ReportDocker(context.Context, *DockerRequest) (*DockerResponse, error) {
	return nil, fmt.Errorf("not implemented")
}
func (UnimplementedAgentServiceServer) ReportPorts(context.Context, *PortsRequest) (*PortsResponse, error) {
	return nil, fmt.Errorf("not implemented")
}
func (UnimplementedAgentServiceServer) ExecuteCommand(context.Context, *CommandRequest) (*CommandResponse, error) {
	return nil, fmt.Errorf("not implemented")
}
func (UnimplementedAgentServiceServer) RegisterAgent(context.Context, *RegisterRequest) (*RegisterResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func RegisterAgentServiceServer(s *grpc.Server, srv AgentServiceServer) {
	s.RegisterService(&_AgentService_serviceDesc, srv)
}

var _AgentService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "AgentService",
	HandlerType: (*AgentServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Heartbeat",
			Handler:    _AgentService_Heartbeat_Handler,
		},
		{
			MethodName: "ReportMetrics",
			Handler:    _AgentService_ReportMetrics_Handler,
		},
		{
			MethodName: "ReportDocker",
			Handler:    _AgentService_ReportDocker_Handler,
		},
		{
			MethodName: "ReportPorts",
			Handler:    _AgentService_ReportPorts_Handler,
		},
		{
			MethodName: "ExecuteCommand",
			Handler:    _AgentService_ExecuteCommand_Handler,
		},
		{
			MethodName: "RegisterAgent",
			Handler:    _AgentService_RegisterAgent_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "agent.proto",
}

// Handler函数
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

func _AgentService_ReportDocker_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DockerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	return srv.(AgentServiceServer).ReportDocker(ctx, in)
}

func _AgentService_ReportPorts_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PortsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	return srv.(AgentServiceServer).ReportPorts(ctx, in)
}

func _AgentService_ExecuteCommand_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CommandRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	return srv.(AgentServiceServer).ExecuteCommand(ctx, in)
}

func _AgentService_RegisterAgent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	return srv.(AgentServiceServer).RegisterAgent(ctx, in)
}

// 请求/响应消息定义
type HeartbeatRequest struct {
	AgentId string `json:"agentId"`
}

type HeartbeatResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type MetricsRequest struct {
	AgentId    string `json:"agentId"`
	MetricsJson string `json:"metricsJson"`
}

type MetricsResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type DockerRequest struct {
	AgentId      string `json:"agentId"`
	ContainersJson string `json:"containersJson"`
}

type DockerResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type PortsRequest struct {
	AgentId   string `json:"agentId"`
	PortsJson string `json:"portsJson"`
}

type PortsResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type CommandRequest struct {
	AgentId string `json:"agentId"`
	Command string `json:"command"`
}

type CommandResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Output  string `json:"output"`
}

type RegisterRequest struct {
	AgentId  string `json:"agentId"`
	Ip       string `json:"ip"`
	Hostname string `json:"hostname"`
	Os       string `json:"os"`
	Arch     string `json:"arch"`
	Kernel   string `json:"kernel"`
	CpuCores uint32 `json:"cpuCores"`
}

type RegisterResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	ServerId uint32 `json:"serverId"`
}
