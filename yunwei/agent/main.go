package main

import (
	"agent/collector"
	"agent/executor"
	"agent/reporter"
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	serverAddr   = flag.String("server", "localhost:50051", "gRPC server address")
	agentName    = flag.String("name", "", "Agent name")
	interval     = flag.Int("interval", 10, "Metrics collection interval in seconds")
	dockerEnable = flag.Bool("docker", true, "Enable Docker monitoring")
	portsEnable  = flag.Bool("ports", true, "Enable port monitoring")
)

func main() {
	flag.Parse()

	// 生成 Agent ID
	agentID := *agentName
	if agentID == "" {
		hostname, _ := os.Hostname()
		agentID = hostname
	}

	log.Printf("===========================================")
	log.Printf("  AI-Ops Agent 启动")
	log.Printf("===========================================")
	log.Printf("  Agent ID:  %s", agentID)
	log.Printf("  Server:    %s", *serverAddr)
	log.Printf("  Interval:  %d seconds", *interval)
	log.Printf("  Docker:    %v", *dockerEnable)
	log.Printf("  Ports:     %v", *portsEnable)
	log.Printf("===========================================")

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 创建采集器
	coll := collector.NewCollector(&collector.Config{
		EnableDocker: *dockerEnable,
		EnablePorts:  *portsEnable,
	})

	// 创建执行器
	exec := executor.NewExecutor()

	// 创建上报器
	rep := reporter.NewReporter(*serverAddr, agentID, coll, exec)

	// 连接服务器
	for i := 0; i < 5; i++ {
		if err := rep.Connect(); err != nil {
			log.Printf("连接服务器失败: %v, %d秒后重试...", err, 5)
			time.Sleep(5 * time.Second)
		} else {
			log.Printf("已连接到服务器: %s", *serverAddr)
			break
		}
	}

	// 启动心跳
	go startHeartbeat(ctx, rep, agentID)

	// 启动指标采集
	go startMetricsCollection(ctx, rep, coll, *interval)

	// 启动任务执行器
	go startTaskExecutor(ctx, rep)

	// 等待退出信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Agent 正在关闭...")
	cancel()
	time.Sleep(1 * time.Second)
	log.Println("Agent 已停止")
}

// startHeartbeat 启动心跳
func startHeartbeat(ctx context.Context, rep *reporter.Reporter, agentID string) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// 立即发送一次心跳
	if err := rep.SendHeartbeat(ctx); err != nil {
		log.Printf("心跳发送失败: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := rep.SendHeartbeat(ctx); err != nil {
				log.Printf("心跳发送失败: %v", err)
			}
		}
	}
}

// startMetricsCollection 启动指标采集
func startMetricsCollection(ctx context.Context, rep *reporter.Reporter, coll *collector.Collector, interval int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metrics := coll.Collect()
			if err := rep.ReportMetrics(ctx, metrics); err != nil {
				log.Printf("指标上报失败: %v", err)
			}
		}
	}
}

// startTaskExecutor 启动任务执行器
func startTaskExecutor(ctx context.Context, rep *reporter.Reporter) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tasks, err := rep.ReceiveTasks(ctx)
			if err != nil {
				continue
			}
			for _, task := range tasks {
				go rep.ExecuteTask(ctx, task)
			}
		}
	}
}
