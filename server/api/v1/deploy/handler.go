package deploy

import (
	"encoding/json"
	"net/http"
	"time"

	"yunwei/global"
	"yunwei/model/common/response"
	"yunwei/service/deploy/analyzer"
	"yunwei/service/deploy/planner"
	"yunwei/service/deploy/executor"
	"yunwei/service/deploy/config"

	"github.com/gin-gonic/gin"
)

// UploadProject 上传项目进行分析
func UploadProject(c *gin.Context) {
	file, err := c.FormFile("project")
	if err != nil {
		response.FailWithMessage("请上传项目文件", c)
		return
	}

	// 保存文件
	uploadPath := "/tmp/uploads/" + time.Now().Format("20060102")
	filename := uploadPath + "/" + file.Filename
	
	if err := c.SaveUploadedFile(file, filename); err != nil {
		response.FailWithMessage("保存文件失败: " + err.Error(), c)
		return
	}

	// 分析项目
	projectAnalyzer := analyzer.NewProjectAnalyzer()
	analysis, err := projectAnalyzer.Analyze(filename)
	if err != nil {
		response.FailWithMessage("项目分析失败: " + err.Error(), c)
		return
	}

	response.OkWithData(gin.H{
		"analysis": analysis,
		"message":  "项目上传并分析成功",
	}, c)
}

// AnalyzeProject 分析项目路径
func AnalyzeProject(c *gin.Context) {
	var req struct {
		Path string `json:"path" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	projectAnalyzer := analyzer.NewProjectAnalyzer()
	analysis, err := projectAnalyzer.Analyze(req.Path)
	if err != nil {
		response.FailWithMessage("项目分析失败: " + err.Error(), c)
		return
	}

	response.OkWithData(analysis, c)
}

// GetProjectAnalyses 获取项目分析列表
func GetProjectAnalyses(c *gin.Context) {
	analyses, err := analyzer.GetProjectAnalyses()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(analyses, c)
}

// GetProjectAnalysis 获取项目分析详情
func GetProjectAnalysis(c *gin.Context) {
	id := c.Param("id")
	analysis, err := analyzer.GetProjectAnalysis(parseInt(id))
	if err != nil {
		response.FailWithMessage("项目分析不存在", c)
		return
	}
	response.OkWithData(analysis, c)
}

// AnalyzeServers 分析服务器资源
func AnalyzeServers(c *gin.Context) {
	serverAnalyzer := analyzer.NewServerResourceAnalyzer()
	
	// 分析所有服务器
	pool, err := serverAnalyzer.AnalyzeAllServers()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(pool, c)
}

// GetServerCapabilities 获取服务器能力列表
func GetServerCapabilities(c *gin.Context) {
	capabilities, err := analyzer.GetServerCapabilities()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(capabilities, c)
}

// FindBestServers 查找最适合的服务器
func FindBestServers(c *gin.Context) {
	var req analyzer.ResourceRequirements
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	serverAnalyzer := analyzer.NewServerResourceAnalyzer()
	matches, err := serverAnalyzer.FindBestServers(&req, 10)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithData(matches, c)
}

// GenerateDeployPlan 生成部署方案
func GenerateDeployPlan(c *gin.Context) {
	var req struct {
		ProjectAnalysisID uint `json:"projectAnalysisId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	// 获取项目分析
	projectAnalysis, err := analyzer.GetProjectAnalysis(req.ProjectAnalysisID)
	if err != nil {
		response.FailWithMessage("项目分析不存在", c)
		return
	}

	// 分析服务器资源
	serverAnalyzer := analyzer.NewServerResourceAnalyzer()
	requirements := &analyzer.ResourceRequirements{
		MinCPU:    projectAnalysis.MinCPU,
		MinMemory: uint64(projectAnalysis.MinMemory),
		MinDisk:   uint64(projectAnalysis.MinDisk),
	}
	
	matches, err := serverAnalyzer.FindBestServers(requirements, projectAnalysis.ClusterSize)
	if err != nil {
		response.FailWithMessage("查找服务器失败: " + err.Error(), c)
		return
	}

	// 生成部署方案
	deployPlanner := planner.NewDeployPlanner()
	plan, err := deployPlanner.GeneratePlan(projectAnalysis, matches)
	if err != nil {
		response.FailWithMessage("生成部署方案失败: " + err.Error(), c)
		return
	}

	response.OkWithData(plan, c)
}

// GetDeployPlans 获取部署方案列表
func GetDeployPlans(c *gin.Context) {
	plans, err := planner.GetDeployPlans()
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(plans, c)
}

// GetDeployPlan 获取部署方案详情
func GetDeployPlan(c *gin.Context) {
	id := c.Param("id")
	plan, err := planner.GetDeployPlan(parseInt(id))
	if err != nil {
		response.FailWithMessage("部署方案不存在", c)
		return
	}
	response.OkWithData(plan, c)
}

// DeleteDeployPlan 删除部署方案
func DeleteDeployPlan(c *gin.Context) {
	id := c.Param("id")
	if err := planner.DeleteDeployPlan(parseInt(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(c)
}

// ExecuteDeploy 执行部署
func ExecuteDeploy(c *gin.Context) {
	id := c.Param("id")
	plan, err := planner.GetDeployPlan(parseInt(id))
	if err != nil {
		response.FailWithMessage("部署方案不存在", c)
		return
	}

	deployExecutor := executor.NewDeployExecutor()
	task, err := deployExecutor.Execute(plan)
	if err != nil {
		response.FailWithMessage("启动部署失败: " + err.Error(), c)
		return
	}

	response.OkWithData(task, c)
}

// GetDeployTask 获取部署任务详情
func GetDeployTask(c *gin.Context) {
	id := c.Param("id")
	task, err := executor.GetTask(parseInt(id))
	if err != nil {
		response.FailWithMessage("任务不存在", c)
		return
	}
	response.OkWithData(task, c)
}

// GetDeployTasks 获取部署任务列表
func GetDeployTasks(c *gin.Context) {
	planID := c.Query("planId")
	tasks, err := executor.GetTasks(parseInt(planID))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(tasks, c)
}

// GetTaskSteps 获取任务步骤
func GetTaskSteps(c *gin.Context) {
	id := c.Param("id")
	steps, err := executor.GetTaskSteps(parseInt(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(steps, c)
}

// PauseDeploy 暂停部署
func PauseDeploy(c *gin.Context) {
	id := c.Param("id")
	deployExecutor := executor.NewDeployExecutor()
	if err := deployExecutor.Pause(parseInt(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(c)
}

// ResumeDeploy 恢复部署
func ResumeDeploy(c *gin.Context) {
	id := c.Param("id")
	deployExecutor := executor.NewDeployExecutor()
	if err := deployExecutor.Resume(parseInt(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(c)
}

// RollbackDeploy 回滚部署
func RollbackDeploy(c *gin.Context) {
	id := c.Param("id")
	deployExecutor := executor.NewDeployExecutor()
	if err := deployExecutor.Rollback(parseInt(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(c)
}

// PreviewConfigs 预览配置
func PreviewConfigs(c *gin.Context) {
	id := c.Param("id")
	plan, err := planner.GetDeployPlan(parseInt(id))
	if err != nil {
		response.FailWithMessage("部署方案不存在", c)
		return
	}

	configGenerator := config.NewConfigGenerator()
	configs, err := configGenerator.GenerateAllConfigs(plan)
	if err != nil {
		response.FailWithMessage("生成配置失败: " + err.Error(), c)
		return
	}

	response.OkWithData(configs, c)
}

// GetServiceTopology 获取服务拓扑
func GetServiceTopology(c *gin.Context) {
	id := c.Param("id")
	plan, err := planner.GetDeployPlan(parseInt(id))
	if err != nil {
		response.FailWithMessage("部署方案不存在", c)
		return
	}

	var topology planner.ServiceTopology
	if plan.ServiceTopology != "" {
		json.Unmarshal([]byte(plan.ServiceTopology), &topology)
	}

	response.OkWithData(topology, c)
}

func parseInt(s string) uint {
	var result uint
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + uint(c-'0')
		}
	}
	return result
}
