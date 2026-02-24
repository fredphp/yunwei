package analyzer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"yunwei/global"
)

// ProjectType 项目类型
type ProjectType string

const (
	ProjectTypeFrontend    ProjectType = "frontend"    // 前端项目
	ProjectTypeBackend     ProjectType = "backend"     // 后端项目
	ProjectTypeMicroservice ProjectType = "microservice" // 微服务
	ProjectTypeDatabase    ProjectType = "database"    // 数据库
	ProjectTypeCache       ProjectType = "cache"       // 缓存服务
	ProjectTypeMQ          ProjectType = "mq"          // 消息队列
	ProjectTypeStatic      ProjectType = "static"      // 静态站点
	ProjectTypeDocker      ProjectType = "docker"      // Docker 项目
	ProjectTypeKubernetes  ProjectType = "kubernetes"  // K8s 项目
	ProjectTypeUnknown     ProjectType = "unknown"     // 未知类型
)

// TechStack 技术栈
type TechStack string

const (
	TechReact       TechStack = "react"
	TechVue         TechStack = "vue"
	TechAngular     TechStack = "angular"
	TechNextJS      TechStack = "nextjs"
	TechNuxtJS      TechStack = "nuxtjs"
	TechGo          TechStack = "go"
	TechJava        TechStack = "java"
	TechPython      TechStack = "python"
	TechNodeJS      TechStack = "nodejs"
	TechPHP         TechStack = "php"
	TechRuby        TechStack = "ruby"
	TechDotNet      TechStack = "dotnet"
	TechMySQL       TechStack = "mysql"
	TechPostgreSQL  TechStack = "postgresql"
	TechMongoDB     TechStack = "mongodb"
	TechRedis       TechStack = "redis"
	TechKafka       TechStack = "kafka"
	TechRabbitMQ    TechStack = "rabbitmq"
	TechElasticsearch TechStack = "elasticsearch"
)

// ProjectAnalysis 项目分析结果
type ProjectAnalysis struct {
	ID           uint                   `json:"id" gorm:"primarykey"`
	CreatedAt    string                 `json:"createdAt" gorm:"type:timestamp"`
	
	// 基本信息
	ProjectName  string                 `json:"projectName" gorm:"type:varchar(128)"`
	ProjectPath  string                 `json:"projectPath" gorm:"type:varchar(512)"`
	ProjectType  ProjectType            `json:"projectType" gorm:"type:varchar(32)"`
	TechStacks   string                 `json:"techStacks" gorm:"type:text"` // JSON数组
	
	// 资源需求
	MinCPU       int                    `json:"minCpu"`       // 最小CPU核心数
	MinMemory    int                    `json:"minMemory"`    // 最小内存(MB)
	MinDisk      int                    `json:"minDisk"`      // 最小磁盘(GB)
	RecCPU       int                    `json:"recCpu"`       // 推荐CPU核心数
	RecMemory    int                    `json:"recMemory"`    // 推荐内存(MB)
	RecDisk      int                    `json:"recDisk"`      // 推荐磁盘(GB)
	
	// 服务配置
	Services     string                 `json:"services" gorm:"type:text"`     // JSON 服务列表
	Dependencies string                 `json:"dependencies" gorm:"type:text"` // JSON 依赖列表
	Ports        string                 `json:"ports" gorm:"type:text"`        // JSON 端口列表
	Environments string                 `json:"environments" gorm:"type:text"` // JSON 环境变量
	
	// 部署信息
	BuildCommand string                 `json:"buildCommand" gorm:"type:text"`
	StartCommand string                 `json:"startCommand" gorm:"type:text"`
	Dockerfile   string                 `json:"dockerfile" gorm:"type:text"`
	DockerCompose string                `json:"dockerCompose" gorm:"type:text"`
	
	// 分布式配置
	NeedCluster   bool                  `json:"needCluster"`
	ClusterSize   int                   `json:"clusterSize"`   // 推荐集群规模
	NeedLB        bool                  `json:"needLb"`        // 需要负载均衡
	NeedDBCluster bool                  `json:"needDbCluster"` // 需要数据库集群
	NeedCache     bool                  `json:"needCache"`     // 需要缓存
	NeedMQ        bool                  `json:"needMq"`        // 需要消息队列
	
	// AI 建议
	AISuggestion string                 `json:"aiSuggestion" gorm:"type:text"`
	Confidence   float64                `json:"confidence"`
}

func (ProjectAnalysis) TableName() string {
	return "project_analyses"
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Port        int               `json:"port"`
	HealthCheck string            `json:"healthCheck"`
	Env         map[string]string `json:"env"`
	Replicas    int               `json:"replicas"`
}

// Dependency 依赖信息
type Dependency struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Type    string `json:"type"` // runtime, dev, database, cache
}

// ProjectAnalyzer 项目分析器
type ProjectAnalyzer struct {
	analyzers map[ProjectType]TypeAnalyzer
}

// TypeAnalyzer 类型分析器接口
type TypeAnalyzer interface {
	Analyze(path string) (*ProjectAnalysis, error)
	Detect(path string) bool
}

// NewProjectAnalyzer 创建项目分析器
func NewProjectAnalyzer() *ProjectAnalyzer {
	return &ProjectAnalyzer{
		analyzers: map[ProjectType]TypeAnalyzer{
			ProjectTypeFrontend:    &FrontendAnalyzer{},
			ProjectTypeBackend:     &BackendAnalyzer{},
			ProjectTypeMicroservice: &MicroserviceAnalyzer{},
			ProjectTypeDocker:      &DockerAnalyzer{},
			ProjectTypeKubernetes:  &KubernetesAnalyzer{},
		},
	}
}

// Analyze 分析项目
func (a *ProjectAnalyzer) Analyze(projectPath string) (*ProjectAnalysis, error) {
	// 检测项目类型
	projectType := a.detectProjectType(projectPath)
	
	analysis := &ProjectAnalysis{
		ProjectPath: projectPath,
		ProjectType: projectType,
	}
	
	// 根据类型使用对应的分析器
	if analyzer, ok := a.analyzers[projectType]; ok {
		result, err := analyzer.Analyze(projectPath)
		if err != nil {
			return nil, err
		}
		analysis = result
	} else {
		// 通用分析
		a.genericAnalyze(projectPath, analysis)
	}
	
	// 分析资源需求
	a.analyzeResourceNeeds(analysis)
	
	// 保存分析结果
	global.DB.Create(analysis)
	
	return analysis, nil
}

// detectProjectType 检测项目类型
func (a *ProjectAnalyzer) detectProjectType(path string) ProjectType {
	// 检查是否有 K8s 配置
	if _, err := os.Stat(filepath.Join(path, "kubernetes")); err == nil {
		return ProjectTypeKubernetes
	}
	if files, _ := filepath.Glob(filepath.Join(path, "*.yaml")); len(files) > 0 {
		for _, f := range files {
			if strings.Contains(f, "deployment") || strings.Contains(f, "service") {
				return ProjectTypeKubernetes
			}
		}
	}
	
	// 检查是否有 Docker
	if _, err := os.Stat(filepath.Join(path, "Dockerfile")); err == nil {
		return ProjectTypeDocker
	}
	if _, err := os.Stat(filepath.Join(path, "docker-compose.yml")); err == nil {
		return ProjectTypeDocker
	}
	if _, err := os.Stat(filepath.Join(path, "docker-compose.yaml")); err == nil {
		return ProjectTypeDocker
	}
	
	// 检查前端项目
	if _, err := os.Stat(filepath.Join(path, "package.json")); err == nil {
		return ProjectTypeFrontend
	}
	
	// 检查后端项目
	if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
		return ProjectTypeBackend
	}
	if _, err := os.Stat(filepath.Join(path, "pom.xml")); err == nil {
		return ProjectTypeBackend
	}
	if _, err := os.Stat(filepath.Join(path, "requirements.txt")); err == nil {
		return ProjectTypeBackend
	}
	if _, err := os.Stat(filepath.Join(path, "Gemfile")); err == nil {
		return ProjectTypeBackend
	}
	if _, err := os.Stat(filepath.Join(path, "composer.json")); err == nil {
		return ProjectTypeBackend
	}
	
	// 检查微服务
	if _, err := os.Stat(filepath.Join(path, "services")); err == nil {
		return ProjectTypeMicroservice
	}
	if _, err := os.Stat(filepath.Join(path, "apps")); err == nil {
		return ProjectTypeMicroservice
	}
	
	return ProjectTypeUnknown
}

// genericAnalyze 通用分析
func (a *ProjectAnalyzer) genericAnalyze(path string, analysis *ProjectAnalysis) {
	analysis.ProjectName = filepath.Base(path)
	
	// 扫描目录结构
	filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		// 跳过隐藏目录和 node_modules 等
		if strings.Contains(filePath, "node_modules") || 
		   strings.Contains(filePath, ".git") ||
		   strings.Contains(filePath, "vendor") {
			return filepath.SkipDir
		}
		
		return nil
	})
}

// analyzeResourceNeeds 分析资源需求
func (a *ProjectAnalyzer) analyzeResourceNeeds(analysis *ProjectAnalysis) {
	// 根据项目类型设置默认资源需求
	switch analysis.ProjectType {
	case ProjectTypeFrontend:
		analysis.MinCPU = 1
		analysis.MinMemory = 512
		analysis.MinDisk = 5
		analysis.RecCPU = 2
		analysis.RecMemory = 1024
		analysis.RecDisk = 10
		
	case ProjectTypeBackend:
		analysis.MinCPU = 1
		analysis.MinMemory = 512
		analysis.MinDisk = 10
		analysis.RecCPU = 4
		analysis.RecMemory = 2048
		analysis.RecDisk = 20
		
	case ProjectTypeMicroservice:
		analysis.MinCPU = 2
		analysis.MinMemory = 1024
		analysis.MinDisk = 20
		analysis.RecCPU = 8
		analysis.RecMemory = 4096
		analysis.RecDisk = 50
		analysis.NeedCluster = true
		analysis.ClusterSize = 3
		analysis.NeedLB = true
		
	case ProjectTypeDatabase:
		analysis.MinCPU = 2
		analysis.MinMemory = 2048
		analysis.MinDisk = 50
		analysis.RecCPU = 8
		analysis.RecMemory = 8192
		analysis.RecDisk = 200
		analysis.NeedDBCluster = true
	}
}

// FrontendAnalyzer 前端项目分析器
type FrontendAnalyzer struct{}

func (a *FrontendAnalyzer) Detect(path string) bool {
	_, err := os.Stat(filepath.Join(path, "package.json"))
	return err == nil
}

func (a *FrontendAnalyzer) Analyze(path string) (*ProjectAnalysis, error) {
	analysis := &ProjectAnalysis{
		ProjectPath: path,
		ProjectType: ProjectTypeFrontend,
		ProjectName: filepath.Base(path),
	}
	
	// 读取 package.json
	packageJSON, err := os.ReadFile(filepath.Join(path, "package.json"))
	if err != nil {
		return nil, err
	}
	
	var pkg struct {
		Name        string            `json:"name"`
		Scripts     map[string]string `json:"scripts"`
		Dependencies map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	
	if err := json.Unmarshal(packageJSON, &pkg); err != nil {
		return nil, err
	}
	
	analysis.ProjectName = pkg.Name
	
	// 检测技术栈
	var techStacks []TechStack
	for dep := range pkg.Dependencies {
		switch {
		case strings.Contains(dep, "react"):
			techStacks = append(techStacks, TechReact)
		case strings.Contains(dep, "vue"):
			techStacks = append(techStacks, TechVue)
		case strings.Contains(dep, "angular"):
			techStacks = append(techStacks, TechAngular)
		case strings.Contains(dep, "next"):
			techStacks = append(techStacks, TechNextJS)
		case strings.Contains(dep, "nuxt"):
			techStacks = append(techStacks, TechNuxtJS)
		}
	}
	
	techJSON, _ := json.Marshal(techStacks)
	analysis.TechStacks = string(techJSON)
	
	// 设置构建和启动命令
	if pkg.Scripts["build"] != "" {
		analysis.BuildCommand = "npm run build"
	}
	if pkg.Scripts["start"] != "" {
		analysis.StartCommand = "npm run start"
	}
	
	// 检查是否是 SSR（需要更多资源）
	isSSR := false
	for _, ts := range techStacks {
		if ts == TechNextJS || ts == TechNuxtJS {
			isSSR = true
			break
		}
	}
	
	if isSSR {
		analysis.MinCPU = 2
		analysis.MinMemory = 1024
		analysis.RecCPU = 4
		analysis.RecMemory = 2048
	} else {
		analysis.MinCPU = 1
		analysis.MinMemory = 256
		analysis.RecCPU = 2
		analysis.RecMemory = 512
	}
	
	// 检查是否需要集群（高流量场景）
	if isSSR {
		analysis.NeedCluster = true
		analysis.ClusterSize = 2
		analysis.NeedLB = true
	}
	
	return analysis, nil
}

// BackendAnalyzer 后端项目分析器
type BackendAnalyzer struct{}

func (a *BackendAnalyzer) Detect(path string) bool {
	files := []string{"go.mod", "pom.xml", "requirements.txt", "Gemfile", "composer.json"}
	for _, f := range files {
		if _, err := os.Stat(filepath.Join(path, f)); err == nil {
			return true
		}
	}
	return false
}

func (a *BackendAnalyzer) Analyze(path string) (*ProjectAnalysis, error) {
	analysis := &ProjectAnalysis{
		ProjectPath: path,
		ProjectType: ProjectTypeBackend,
		ProjectName: filepath.Base(path),
	}
	
	var techStacks []TechStack
	var dependencies []Dependency
	
	// 检测 Go 项目
	if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
		techStacks = append(techStacks, TechGo)
		analysis.BuildCommand = "go build -o main ."
		analysis.StartCommand = "./main"
		
		// 解析 go.mod
		goMod, _ := os.ReadFile(filepath.Join(path, "go.mod"))
		lines := strings.Split(string(goMod), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "\t") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					dependencies = append(dependencies, Dependency{
						Name:    parts[0],
						Version: parts[1],
						Type:    "runtime",
					})
				}
			}
		}
	}
	
	// 检测 Java 项目
	if _, err := os.Stat(filepath.Join(path, "pom.xml")); err == nil {
		techStacks = append(techStacks, TechJava)
		analysis.BuildCommand = "mvn clean package"
		analysis.StartCommand = "java -jar target/*.jar"
	}
	
	// 检测 Python 项目
	if _, err := os.Stat(filepath.Join(path, "requirements.txt")); err == nil {
		techStacks = append(techStacks, TechPython)
		analysis.StartCommand = "python main.py"
		
		// 检查是否有 Django/Flask
		reqFile, _ := os.ReadFile(filepath.Join(path, "requirements.txt"))
		if strings.Contains(string(reqFile), "django") {
			analysis.StartCommand = "python manage.py runserver 0.0.0.0:8000"
		}
		if strings.Contains(string(reqFile), "gunicorn") {
			analysis.StartCommand = "gunicorn main:app"
		}
	}
	
	// 检测 Node.js 后端
	if _, err := os.Stat(filepath.Join(path, "package.json")); err == nil {
		// 检查是否是纯后端
		packageJSON, _ := os.ReadFile(filepath.Join(path, "package.json"))
		var pkg struct {
			Dependencies map[string]string `json:"dependencies"`
		}
		json.Unmarshal(packageJSON, &pkg)
		
		for dep := range pkg.Dependencies {
			if strings.Contains(dep, "express") || 
			   strings.Contains(dep, "koa") ||
			   strings.Contains(dep, "fastify") ||
			   strings.Contains(dep, "nestjs") {
				techStacks = append(techStacks, TechNodeJS)
				break
			}
		}
	}
	
	// 检查数据库依赖
	dbDeps := []string{"mysql", "postgres", "mongodb", "redis", "elasticsearch"}
	needDB := false
	for _, dep := range dependencies {
		for _, dbDep := range dbDeps {
			if strings.Contains(strings.ToLower(dep.Name), dbDep) {
				needDB = true
				analysis.NeedCache = strings.Contains(dbDep, "redis")
				break
			}
		}
	}
	
	analysis.NeedDBCluster = needDB
	
	// 分析端口
	if data, err := os.ReadFile(filepath.Join(path, "config.yaml")); err == nil {
		if strings.Contains(string(data), "port:") {
			analysis.Ports = `["8080"]`
		}
	}
	
	techJSON, _ := json.Marshal(techStacks)
	analysis.TechStacks = string(techJSON)
	
	// 后端项目推荐配置
	analysis.MinCPU = 1
	analysis.MinMemory = 512
	analysis.MinDisk = 10
	analysis.RecCPU = 4
	analysis.RecMemory = 2048
	analysis.RecDisk = 20
	
	// 集群推荐
	analysis.NeedCluster = true
	analysis.ClusterSize = 2
	analysis.NeedLB = true
	
	return analysis, nil
}

// MicroserviceAnalyzer 微服务项目分析器
type MicroserviceAnalyzer struct{}

func (a *MicroserviceAnalyzer) Detect(path string) bool {
	// 检查是否有 services 或 apps 目录
	if _, err := os.Stat(filepath.Join(path, "services")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(path, "apps")); err == nil {
		return true
	}
	return false
}

func (a *MicroserviceAnalyzer) Analyze(path string) (*ProjectAnalysis, error) {
	analysis := &ProjectAnalysis{
		ProjectPath:  path,
		ProjectType:  ProjectTypeMicroservice,
		ProjectName:  filepath.Base(path),
		NeedCluster:  true,
		NeedLB:       true,
		NeedDBCluster: true,
		NeedCache:    true,
	}
	
	// 扫描服务目录
	serviceDirs := []string{"services", "apps", "packages"}
	var services []ServiceInfo
	
	for _, dir := range serviceDirs {
		servicePath := filepath.Join(path, dir)
		if entries, err := os.ReadDir(servicePath); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					services = append(services, ServiceInfo{
						Name: entry.Name(),
						Type: "microservice",
						Port: 8000 + len(services),
						Replicas: 2,
					})
				}
			}
		}
	}
	
	servicesJSON, _ := json.Marshal(services)
	analysis.Services = string(servicesJSON)
	
	// 微服务资源需求
	analysis.ClusterSize = len(services) * 2
	analysis.MinCPU = 2 * len(services)
	analysis.MinMemory = 1024 * len(services)
	analysis.MinDisk = 20
	analysis.RecCPU = 4 * len(services)
	analysis.RecMemory = 2048 * len(services)
	analysis.RecDisk = 50
	
	// 检查是否有 API Gateway
	if _, err := os.Stat(filepath.Join(path, "gateway")); err == nil {
		analysis.Services = analysis.Services + `,{"name":"gateway","type":"gateway","port":80}`
	}
	
	// 检查是否有消息队列需求
	if _, err := os.Stat(filepath.Join(path, "queue")); err == nil {
		analysis.NeedMQ = true
	}
	
	return analysis, nil
}

// DockerAnalyzer Docker 项目分析器
type DockerAnalyzer struct{}

func (a *DockerAnalyzer) Detect(path string) bool {
	files := []string{"Dockerfile", "docker-compose.yml", "docker-compose.yaml"}
	for _, f := range files {
		if _, err := os.Stat(filepath.Join(path, f)); err == nil {
			return true
		}
	}
	return false
}

func (a *DockerAnalyzer) Analyze(path string) (*ProjectAnalysis, error) {
	analysis := &ProjectAnalysis{
		ProjectPath: path,
		ProjectType: ProjectTypeDocker,
		ProjectName: filepath.Base(path),
	}
	
	// 读取 Dockerfile
	if dockerfile, err := os.ReadFile(filepath.Join(path, "Dockerfile")); err == nil {
		analysis.Dockerfile = string(dockerfile)
		
		// 分析 Dockerfile 内容
		content := string(dockerfile)
		if strings.Contains(content, "EXPOSE") {
			lines := strings.Split(content, "\n")
			var ports []string
			for _, line := range lines {
				if strings.HasPrefix(line, "EXPOSE") {
					ports = append(ports, strings.TrimSpace(strings.TrimPrefix(line, "EXPOSE")))
				}
			}
			portsJSON, _ := json.Marshal(ports)
			analysis.Ports = string(portsJSON)
		}
	}
	
	// 读取 docker-compose
	composeFiles := []string{"docker-compose.yml", "docker-compose.yaml"}
	for _, f := range composeFiles {
		if data, err := os.ReadFile(filepath.Join(path, f)); err == nil {
			analysis.DockerCompose = string(data)
			break
		}
	}
	
	// Docker 项目默认配置
	analysis.MinCPU = 1
	analysis.MinMemory = 512
	analysis.RecCPU = 2
	analysis.RecMemory = 1024
	
	return analysis, nil
}

// KubernetesAnalyzer K8s 项目分析器
type KubernetesAnalyzer struct{}

func (a *KubernetesAnalyzer) Detect(path string) bool {
	// 检查 kubernetes 目录或 k8s 配置文件
	if _, err := os.Stat(filepath.Join(path, "kubernetes")); err == nil {
		return true
	}
	files, _ := filepath.Glob(filepath.Join(path, "*.yaml"))
	for _, f := range files {
		data, _ := os.ReadFile(f)
		content := string(data)
		if strings.Contains(content, "apiVersion:") && 
		   (strings.Contains(content, "Deployment") || 
		    strings.Contains(content, "Service")) {
			return true
		}
	}
	return false
}

func (a *KubernetesAnalyzer) Analyze(path string) (*ProjectAnalysis, error) {
	analysis := &ProjectAnalysis{
		ProjectPath: path,
		ProjectType: ProjectTypeKubernetes,
		ProjectName: filepath.Base(path),
	}
	
	// 扫描 K8s 配置文件
	var services []ServiceInfo
	
	// 扫描 kubernetes 目录
	k8sDir := filepath.Join(path, "kubernetes")
	if entries, err := os.ReadDir(k8sDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".yaml") || strings.HasSuffix(entry.Name(), ".yml")) {
				filePath := filepath.Join(k8sDir, entry.Name())
				data, _ := os.ReadFile(filePath)
				content := string(data)
				
				// 解析 Deployment
				if strings.Contains(content, "kind: Deployment") {
					// 提取服务名
					lines := strings.Split(content, "\n")
					for _, line := range lines {
						if strings.Contains(line, "name:") && !strings.Contains(line, "namespace") {
							name := strings.TrimSpace(strings.Split(line, ":")[1])
							services = append(services, ServiceInfo{
								Name: name,
								Type: "deployment",
							})
							break
						}
					}
				}
			}
		}
	}
	
	servicesJSON, _ := json.Marshal(services)
	analysis.Services = string(servicesJSON)
	
	// K8s 项目需要集群
	analysis.NeedCluster = true
	analysis.ClusterSize = len(services) * 2
	analysis.NeedLB = true
	
	return analysis, nil
}

// GetProjectAnalysis 获取项目分析结果
func GetProjectAnalysis(id uint) (*ProjectAnalysis, error) {
	var analysis ProjectAnalysis
	err := global.DB.First(&analysis, id).Error
	return &analysis, err
}

// GetProjectAnalyses 获取项目分析列表
func GetProjectAnalyses() ([]ProjectAnalysis, error) {
	var analyses []ProjectAnalysis
	err := global.DB.Order("created_at DESC").Find(&analyses).Error
	return analyses, err
}

// DeleteProjectAnalysis 删除项目分析
func DeleteProjectAnalysis(id uint) error {
	return global.DB.Delete(&ProjectAnalysis{}, id).Error
}
