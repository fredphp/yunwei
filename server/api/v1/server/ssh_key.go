package server

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"strings"

	"yunwei/global"
	"yunwei/model/common/response"
	"yunwei/model/server"

	"github.com/gin-gonic/gin"
)

// GenerateFingerprint 生成密钥指纹
func GenerateFingerprint(keyContent string) string {
	// 提取公钥部分
	lines := strings.Split(keyContent, "\n")
	var keyData strings.Builder
	inKey := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "-----BEGIN") {
			inKey = true
			continue
		}
		if strings.Contains(line, "-----END") {
			inKey = false
			continue
		}
		if inKey {
			keyData.WriteString(line)
		}
	}

	if keyData.Len() == 0 {
		return "unknown"
	}

	// 解码 Base64 并计算 MD5
	// 简化处理：直接对 Base64 数据计算 MD5
	hash := md5.Sum([]byte(keyData.String()))
	return hex.EncodeToString(hash[:])
}

// GetSshKeyList 获取 SSH 密钥列表
func GetSshKeyList(c *gin.Context) {
	var keys []server.SshKey

	query := global.DB.Model(&server.SshKey{})

	// 搜索条件
	if name := c.Query("name"); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	// 分页
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	var total int64

	query.Count(&total)
	query.Order("created_at DESC")
	query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&keys)

	// 查询每个密钥关联的服务器数量
	result := make([]map[string]interface{}, 0, len(keys))
	for _, key := range keys {
		var count int64
		global.DB.Model(&server.Server{}).Where("ssh_key_id = ?", key.ID).Count(&count)

		result = append(result, map[string]interface{}{
			"id":          key.ID,
			"name":        key.Name,
			"filename":    key.Filename,
			"fingerprint": key.Fingerprint,
			"description": key.Description,
			"createdAt":   key.CreatedAt,
			"updatedAt":   key.UpdatedAt,
			"serverCount": count,
		})
	}

	response.OkWithData(gin.H{
		"list":     result,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}, c)
}

// GetSshKey 获取 SSH 密钥详情
func GetSshKey(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.FailWithMessage("无效的ID", c)
		return
	}

	var key server.SshKey
	if err := global.DB.First(&key, id).Error; err != nil {
		response.FailWithMessage("密钥不存在", c)
		return
	}

	// 查询关联的服务器
	var servers []server.Server
	global.DB.Where("ssh_key_id = ?", id).Select("id, name, host, status").Find(&servers)

	response.OkWithData(gin.H{
		"id":          key.ID,
		"name":        key.Name,
		"filename":    key.Filename,
		"fingerprint": key.Fingerprint,
		"description": key.Description,
		"createdAt":   key.CreatedAt,
		"updatedAt":   key.UpdatedAt,
		"servers":     servers,
	}, c)
}

// AddSshKey 添加 SSH 密钥
func AddSshKey(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Filename    string `json:"filename"`
		KeyContent  string `json:"keyContent" binding:"required"`
		Passphrase  string `json:"passphrase"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 验证是否为有效的 PEM 格式
	if !strings.Contains(req.KeyContent, "-----BEGIN") || !strings.Contains(req.KeyContent, "-----END") {
		response.FailWithMessage("无效的 PEM 格式，请上传有效的 SSH 私钥文件", c)
		return
	}

	// 生成指纹
	fingerprint := GenerateFingerprint(req.KeyContent)

	// 检查是否已存在相同指纹的密钥
	var existingKey server.SshKey
	if global.DB.Where("fingerprint = ?", fingerprint).First(&existingKey).Error == nil {
		response.FailWithMessage("该密钥已存在，请勿重复添加", c)
		return
	}

	key := server.SshKey{
		Name:        req.Name,
		Filename:    req.Filename,
		KeyContent:  req.KeyContent,
		Passphrase:  req.Passphrase,
		Fingerprint: fingerprint,
		Description: req.Description,
	}

	if err := global.DB.Create(&key).Error; err != nil {
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}

	response.OkWithData(gin.H{
		"id":          key.ID,
		"name":        key.Name,
		"filename":    key.Filename,
		"fingerprint": key.Fingerprint,
		"description": key.Description,
		"createdAt":   key.CreatedAt,
	}, c)
}

// UpdateSshKey 更新 SSH 密钥
func UpdateSshKey(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.FailWithMessage("无效的ID", c)
		return
	}

	var key server.SshKey
	if err := global.DB.First(&key, id).Error; err != nil {
		response.FailWithMessage("密钥不存在", c)
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}

	// 如果更新了密钥内容，需要重新验证和生成指纹
	if keyContent, ok := req["keyContent"].(string); ok && keyContent != "" {
		if !strings.Contains(keyContent, "-----BEGIN") || !strings.Contains(keyContent, "-----END") {
			response.FailWithMessage("无效的 PEM 格式", c)
			return
		}
		req["fingerprint"] = GenerateFingerprint(keyContent)
	}

	if err := global.DB.Model(&key).Updates(req).Error; err != nil {
		response.FailWithMessage("更新失败", c)
		return
	}

	response.OkWithData(gin.H{
		"id":          key.ID,
		"name":        key.Name,
		"filename":    key.Filename,
		"fingerprint": key.Fingerprint,
		"description": key.Description,
	}, c)
}

// DeleteSshKey 删除 SSH 密钥
func DeleteSshKey(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.FailWithMessage("无效的ID", c)
		return
	}

	// 检查是否有服务器正在使用此密钥
	var count int64
	global.DB.Model(&server.Server{}).Where("ssh_key_id = ?", id).Count(&count)
	if count > 0 {
		response.FailWithMessage("无法删除：有 "+strconv.FormatInt(count, 10)+" 个服务器正在使用此密钥", c)
		return
	}

	if err := global.DB.Delete(&server.SshKey{}, id).Error; err != nil {
		response.FailWithMessage("删除失败", c)
		return
	}

	response.Ok(nil, c)
}

// UploadSshKey 上传 SSH 密钥文件
func UploadSshKey(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.FailWithMessage("文件上传失败: "+err.Error(), c)
		return
	}

	// 验证文件扩展名
	filename := file.Filename
	if !strings.HasSuffix(filename, ".pem") && !strings.HasSuffix(filename, ".key") {
		response.FailWithMessage("请上传 .pem 或 .key 格式的文件", c)
		return
	}

	// 读取文件内容
	content, err := file.Open()
	if err != nil {
		response.FailWithMessage("文件读取失败", c)
		return
	}
	defer content.Close()

	buf := make([]byte, file.Size)
	_, err = content.Read(buf)
	if err != nil {
		response.FailWithMessage("文件读取失败", c)
		return
	}

	keyContent := string(buf)

	// 验证是否为有效的 PEM 格式
	if !strings.Contains(keyContent, "-----BEGIN") || !strings.Contains(keyContent, "-----END") {
		response.FailWithMessage("无效的 PEM 格式，请上传有效的 SSH 私钥文件", c)
		return
	}

	// 获取其他参数
	name := c.PostForm("name")
	if name == "" {
		name = strings.TrimSuffix(filename, ".pem")
		name = strings.TrimSuffix(name, ".key")
	}
	description := c.PostForm("description")

	// 生成指纹
	fingerprint := GenerateFingerprint(keyContent)

	// 检查是否已存在相同指纹的密钥
	var existingKey server.SshKey
	if global.DB.Where("fingerprint = ?", fingerprint).First(&existingKey).Error == nil {
		response.FailWithMessage("该密钥已存在，请勿重复添加", c)
		return
	}

	key := server.SshKey{
		Name:        name,
		Filename:    filename,
		KeyContent:  keyContent,
		Fingerprint: fingerprint,
		Description: description,
	}

	if err := global.DB.Create(&key).Error; err != nil {
		response.FailWithMessage("保存失败: "+err.Error(), c)
		return
	}

	response.OkWithData(gin.H{
		"id":          key.ID,
		"name":        key.Name,
		"filename":    key.Filename,
		"fingerprint": key.Fingerprint,
		"description": key.Description,
		"createdAt":   key.CreatedAt,
	}, c)
}
