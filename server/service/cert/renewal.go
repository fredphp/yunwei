package cert

import (
        "crypto/x509"
        "encoding/json"
        "encoding/pem"
        "fmt"
        "io/ioutil"
        "os"
        "os/exec"
        "strings"
        "time"

        "yunwei/global"
        "yunwei/service/ai/llm"
        "yunwei/model/notify"
)

// CertStatus 证书状态
type CertStatus string

const (
        CertStatusValid    CertStatus = "valid"
        CertStatusExpiring CertStatus = "expiring" // 即将过期
        CertStatusExpired  CertStatus = "expired"
        CertStatusRenewing CertStatus = "renewing"
        CertStatusFailed   CertStatus = "failed"
)

// RenewalStatus 续期状态
type RenewalStatus string

const (
        RenewalStatusPending   RenewalStatus = "pending"
        RenewalStatusRunning   RenewalStatus = "running"
        RenewalStatusSuccess   RenewalStatus = "success"
        RenewalStatusFailed    RenewalStatus = "failed"
)

// CertProvider 证书提供商
type CertProvider string

const (
        ProviderLetsEncrypt CertProvider = "letsencrypt"
        ProviderZeroSSL     CertProvider = "zerossl"
        ProviderCustom      CertProvider = "custom"
        ProviderACME        CertProvider = "acme"
)

// Certificate 证书
type Certificate struct {
        ID          uint      `json:"id" gorm:"primarykey"`
        CreatedAt   time.Time `json:"createdAt"`
        UpdatedAt   time.Time `json:"updatedAt"`

        Name        string       `json:"name" gorm:"type:varchar(64)"`
        Domain      string       `json:"domain" gorm:"type:varchar(256);index"`
        SANs        string       `json:"sans" gorm:"type:text"` // Subject Alternative Names (JSON数组)
        Provider    CertProvider `json:"provider" gorm:"type:varchar(16)"`

        // 证书文件路径
        CertPath    string `json:"certPath" gorm:"type:varchar(256)"`
        KeyPath     string `json:"keyPath" gorm:"type:varchar(256)"`
        ChainPath   string `json:"chainPath" gorm:"type:varchar(256)"`
        FullChainPath string `json:"fullChainPath" gorm:"type:varchar(256)"`

        // 证书信息
        SerialNumber string    `json:"serialNumber" gorm:"type:varchar(64)"`
        Issuer       string    `json:"issuer" gorm:"type:varchar(128)"`
        NotBefore    time.Time `json:"notBefore"`
        NotAfter     time.Time `json:"notAfter"`
        DaysLeft     int       `json:"daysLeft"`

        // 状态
        Status       CertStatus `json:"status" gorm:"type:varchar(16)"`

        // 自动续期配置
        AutoRenew      bool  `json:"autoRenew"`
        RenewBefore    int   `json:"renewBefore"` // 提前多少天续期
        RenewalCount   int   `json:"renewalCount"`
        LastRenewAt    *time.Time `json:"lastRenewAt"`
        NextRenewAt    *time.Time `json:"nextRenewAt"`

        // ACME 配置
        ACMEEmail     string `json:"acmeEmail" gorm:"type:varchar(64)"`
        ACMEServer    string `json:"acmeServer" gorm:"type:varchar(256)"`
        DNSProvider   string `json:"dnsProvider" gorm:"type:varchar(32)"` // cloudflare, aliyun, tencent
        DNSCredentials string `json:"dnsCredentials" gorm:"type:text"` // 加密存储

        // 部署配置
        DeployTarget   string `json:"deployTarget" gorm:"type:varchar(32)"` // nginx, apache, k8s, cdn
        DeployConfig   string `json:"deployConfig" gorm:"type:text"` // JSON
        LastDeployAt   *time.Time `json:"lastDeployAt"`
}

func (Certificate) TableName() string {
        return "certificates"
}

// RenewalRecord 续期记录
type RenewalRecord struct {
        ID        uint      `json:"id" gorm:"primarykey"`
        CreatedAt time.Time `json:"createdAt"`

        CertID     uint         `json:"certId" gorm:"index"`
        Cert       *Certificate `json:"cert" gorm:"foreignKey:CertID"`

        Status     RenewalStatus `json:"status" gorm:"type:varchar(16)"`

        // 续期前信息
        OldSerialNumber string    `json:"oldSerialNumber"`
        OldNotAfter     time.Time `json:"oldNotAfter"`

        // 续期后信息
        NewSerialNumber string    `json:"newSerialNumber"`
        NewNotAfter     time.Time `json:"newNotAfter"`

        // 执行信息
        Method       string `json:"method" gorm:"type:varchar(32)"` // acme, manual, import
        Commands     string `json:"commands" gorm:"type:text"`
        ExecutionLog string `json:"executionLog" gorm:"type:text"`
        ErrorMessage string `json:"errorMessage" gorm:"type:text"`

        // AI 决策
        AIDecision   string  `json:"aiDecision" gorm:"type:text"`
        AIConfidence float64 `json:"aiConfidence"`

        // 时间
        StartedAt   *time.Time `json:"startedAt"`
        CompletedAt *time.Time `json:"completedAt"`
        Duration    int64      `json:"duration"` // 毫秒
}

func (RenewalRecord) TableName() string {
        return "cert_renewal_records"
}

// CertRenewalManager 证书续期管理器
type CertRenewalManager struct {
        llmClient *llm.GLM5Client
        notifier  notify.Notifier
        executor  CertExecutor
}

// CertExecutor 证书执行器接口
type CertExecutor interface {
        GetCertInfo(certPath string) (*CertInfo, error)
        ExecuteACME(domain, email, server string, dnsProvider string, dnsCreds string) error
        DeployCert(cert *Certificate) error
        ReloadService(service string) error
}

// CertInfo 证书信息
type CertInfo struct {
        SerialNumber string
        Issuer       string
        NotBefore    time.Time
        NotAfter     time.Time
        DNSNames     []string
}

// NewCertRenewalManager 创建证书续期管理器
func NewCertRenewalManager() *CertRenewalManager {
        return &CertRenewalManager{}
}

// SetLLMClient 设置 LLM 客户端
func (m *CertRenewalManager) SetLLMClient(client *llm.GLM5Client) {
        m.llmClient = client
}

// SetNotifier 设置通知器
func (m *CertRenewalManager) SetNotifier(notifier notify.Notifier) {
        m.notifier = notifier
}

// SetExecutor 设置执行器
func (m *CertRenewalManager) SetExecutor(executor CertExecutor) {
        m.executor = executor
}

// CheckAllCertificates 检查所有证书
func (m *CertRenewalManager) CheckAllCertificates() error {
        var certs []Certificate
        global.DB.Find(&certs)

        for i := range certs {
                m.CheckCertificate(&certs[i])
        }

        return nil
}

// CheckCertificate 检查证书状态
func (m *CertRenewalManager) CheckCertificate(cert *Certificate) error {
        // 读取证书文件
        certInfo, err := m.parseCertFile(cert.CertPath)
        if err != nil {
                cert.Status = CertStatusFailed
                global.DB.Save(cert)
                return err
        }

        // 更新证书信息
        cert.SerialNumber = certInfo.SerialNumber
        cert.Issuer = certInfo.Issuer
        cert.NotBefore = certInfo.NotBefore
        cert.NotAfter = certInfo.NotAfter

        // 计算剩余天数
        daysLeft := int(time.Until(certInfo.NotAfter).Hours() / 24)
        cert.DaysLeft = daysLeft

        // 更新状态
        if daysLeft <= 0 {
                cert.Status = CertStatusExpired
        } else if daysLeft <= cert.RenewBefore {
                cert.Status = CertStatusExpiring
        } else {
                cert.Status = CertStatusValid
        }

        global.DB.Save(cert)

        // 检查是否需要自动续期
        if cert.AutoRenew && cert.Status == CertStatusExpiring {
                _, err := m.RenewCertificate(cert.ID)
                return err
        }

        return nil
}

// parseCertFile 解析证书文件
func (m *CertRenewalManager) parseCertFile(certPath string) (*CertInfo, error) {
        data, err := ioutil.ReadFile(certPath)
        if err != nil {
                return nil, fmt.Errorf("读取证书文件失败: %w", err)
        }

        block, _ := pem.Decode(data)
        if block == nil {
                return nil, fmt.Errorf("解析 PEM 失败")
        }

        cert, err := x509.ParseCertificate(block.Bytes)
        if err != nil {
                return nil, fmt.Errorf("解析证书失败: %w", err)
        }

        return &CertInfo{
                SerialNumber: cert.SerialNumber.String(),
                Issuer:       cert.Issuer.CommonName,
                NotBefore:    cert.NotBefore,
                NotAfter:     cert.NotAfter,
                DNSNames:     cert.DNSNames,
        }, nil
}

// RenewCertificate 续期证书
func (m *CertRenewalManager) RenewCertificate(certID uint) (*RenewalRecord, error) {
        var cert Certificate
        if err := global.DB.First(&cert, certID).Error; err != nil {
                return nil, fmt.Errorf("证书不存在")
        }

        record := &RenewalRecord{
                CertID:          certID,
                Status:          RenewalStatusPending,
                OldSerialNumber: cert.SerialNumber,
                OldNotAfter:     cert.NotAfter,
                Method:          "acme",
        }

        global.DB.Create(record)

        // 更新证书状态
        cert.Status = CertStatusRenewing
        global.DB.Save(&cert)

        return m.executeRenewal(record, &cert)
}

// executeRenewal 执行续期
func (m *CertRenewalManager) executeRenewal(record *RenewalRecord, cert *Certificate) (*RenewalRecord, error) {
        record.Status = RenewalStatusRunning
        now := time.Now()
        record.StartedAt = &now
        global.DB.Save(record)

        var commands []string
        var allOutput []string
        success := false

        // 使用 certbot 或 acme.sh 续期
        switch cert.Provider {
        case ProviderLetsEncrypt, ProviderZeroSSL, ProviderACME:
                // 使用 ACME 协议续期
                cmd := m.buildACMECommand(cert)
                commands = append(commands, cmd)

                output, err := m.executeCommand(cmd)
                allOutput = append(allOutput, output)

                if err != nil {
                        allOutput = append(allOutput, fmt.Sprintf("ACME 续期失败: %s", err.Error()))
                        record.ErrorMessage = err.Error()
                } else {
                        success = true
                        allOutput = append(allOutput, "ACME 续期成功")

                        // 部署证书
                        if cert.DeployTarget != "" {
                                deployOutput, err := m.deployCertificate(cert)
                                if err != nil {
                                        allOutput = append(allOutput, fmt.Sprintf("部署失败: %s", err.Error()))
                                        success = false
                                } else {
                                        allOutput = append(allOutput, deployOutput)
                                }
                        }
                }

        default:
                // 手动续期
                record.ErrorMessage = "不支持的证书提供商，需要手动续期"
        }

        // 更新记录
        commandsJSON, _ := json.Marshal(commands)
        record.Commands = string(commandsJSON)
        record.ExecutionLog = strings.Join(allOutput, "\n")

        completedAt := time.Now()
        record.CompletedAt = &completedAt
        if record.StartedAt != nil {
                record.Duration = completedAt.Sub(*record.StartedAt).Milliseconds()
        }

        if success {
                record.Status = RenewalStatusSuccess

                // 更新证书信息
                certInfo, _ := m.parseCertFile(cert.CertPath)
                if certInfo != nil {
                        cert.SerialNumber = certInfo.SerialNumber
                        cert.NotBefore = certInfo.NotBefore
                        cert.NotAfter = certInfo.NotAfter
                        cert.DaysLeft = int(time.Until(certInfo.NotAfter).Hours() / 24)
                        cert.Status = CertStatusValid

                        record.NewSerialNumber = certInfo.SerialNumber
                        record.NewNotAfter = certInfo.NotAfter
                }

                cert.RenewalCount++
                cert.LastRenewAt = &completedAt

                // 计算下次续期时间
                nextRenew := cert.NotAfter.AddDate(0, 0, -cert.RenewBefore)
                cert.NextRenewAt = &nextRenew

                global.DB.Save(cert)

                // 发送通知
                if m.notifier != nil {
                        m.notifier.SendMessage("证书续期成功",
                                fmt.Sprintf("域名 %s 的证书已成功续期，新证书有效期至 %s",
                                        cert.Domain, cert.NotAfter.Format("2006-01-02")))
                }
        } else {
                record.Status = RenewalStatusFailed
                cert.Status = CertStatusFailed
                global.DB.Save(cert)

                // 发送通知
                if m.notifier != nil {
                        m.notifier.SendMessage("证书续期失败",
                                fmt.Sprintf("域名 %s 的证书续期失败，请手动处理", cert.Domain))
                }
        }

        global.DB.Save(record)

        return record, nil
}

// buildACMECommand 构建 ACME 命令
func (m *CertRenewalManager) buildACMECommand(cert *Certificate) string {
        // 使用 acme.sh 或 certbot
        switch cert.DNSProvider {
        case "cloudflare":
                return fmt.Sprintf("acme.sh --renew -d %s --dns dns_cf", cert.Domain)
        case "aliyun":
                return fmt.Sprintf("acme.sh --renew -d %s --dns dns_ali", cert.Domain)
        case "tencent":
                return fmt.Sprintf("acme.sh --renew -d %s --dns dns_tencent", cert.Domain)
        case "http":
                return fmt.Sprintf("certbot renew --cert-name %s", cert.Domain)
        default:
                return fmt.Sprintf("acme.sh --renew -d %s", cert.Domain)
        }
}

// executeCommand 执行命令
func (m *CertRenewalManager) executeCommand(cmd string) (string, error) {
        execCmd := exec.Command("sh", "-c", cmd)
        output, err := execCmd.CombinedOutput()
        return string(output), err
}

// deployCertificate 部署证书
func (m *CertRenewalManager) deployCertificate(cert *Certificate) (string, error) {
        var outputs []string

        switch cert.DeployTarget {
        case "nginx":
                // 重载 Nginx
                output, err := m.executeCommand("nginx -t && systemctl reload nginx")
                if err != nil {
                        return "", fmt.Errorf("重载 Nginx 失败: %w", err)
                }
                outputs = append(outputs, "Nginx 已重载: "+output)

        case "apache":
                output, err := m.executeCommand("apachectl configtest && systemctl reload httpd")
                if err != nil {
                        return "", fmt.Errorf("重载 Apache 失败: %w", err)
                }
                outputs = append(outputs, "Apache 已重载: "+output)

        case "k8s":
                // 更新 Kubernetes Secret
                output, err := m.executeCommand(fmt.Sprintf(
                        "kubectl create secret tls %s-tls --cert=%s --key=%s -n default --dry-run=client -o yaml | kubectl apply -f -",
                        cert.Domain, cert.CertPath, cert.KeyPath))
                if err != nil {
                        return "", fmt.Errorf("更新 K8s Secret 失败: %w", err)
                }
                outputs = append(outputs, "K8s Secret 已更新: "+output)

        case "cdn":
                // 推送到 CDN (阿里云/腾讯云等)
                output, err := m.pushCertToCDN(cert)
                if err != nil {
                        return "", fmt.Errorf("推送到 CDN 失败: %w", err)
                }
                outputs = append(outputs, "CDN 证书已更新: "+output)
        }

        now := time.Now()
        cert.LastDeployAt = &now
        global.DB.Save(cert)

        return strings.Join(outputs, "\n"), nil
}

// pushCertToCDN 推送证书到 CDN
func (m *CertRenewalManager) pushCertToCDN(cert *Certificate) (string, error) {
        // 解析部署配置
        var config struct {
                Provider string `json:"provider"`
                Domain   string `json:"domain"`
                Region   string `json:"region"`
        }

        if cert.DeployConfig != "" {
                json.Unmarshal([]byte(cert.DeployConfig), &config)
        }

        switch config.Provider {
        case "aliyun":
                return m.executeCommand(fmt.Sprintf(
                        "aliyun cdn SetDomainServerCertificate --DomainName %s --CertName %s --Certificate '%s' --PrivateKey '%s'",
                        config.Domain, cert.Name, cert.CertPath, cert.KeyPath))
        case "tencent":
                return m.executeCommand(fmt.Sprintf(
                        "tccli cdn SetHttpsInfo --Domain %s --CertId %s",
                        config.Domain, cert.Name))
        default:
                return "", fmt.Errorf("不支持的 CDN 提供商")
        }
}

// AddCertificate 添加证书
func AddCertificate(cert *Certificate) error {
        // 检查证书文件是否存在
        if _, err := os.Stat(cert.CertPath); os.IsNotExist(err) {
                return fmt.Errorf("证书文件不存在: %s", cert.CertPath)
        }

        // 解析证书信息
        certInfo, err := NewCertRenewalManager().parseCertFile(cert.CertPath)
        if err != nil {
                return fmt.Errorf("解析证书失败: %w", err)
        }

        cert.SerialNumber = certInfo.SerialNumber
        cert.Issuer = certInfo.Issuer
        cert.NotBefore = certInfo.NotBefore
        cert.NotAfter = certInfo.NotAfter
        cert.DaysLeft = int(time.Until(certInfo.NotAfter).Hours() / 24)

        if cert.DaysLeft <= 0 {
                cert.Status = CertStatusExpired
        } else if cert.DaysLeft <= cert.RenewBefore {
                cert.Status = CertStatusExpiring
        } else {
                cert.Status = CertStatusValid
        }

        // 设置默认续期时间
        if cert.RenewBefore == 0 {
                cert.RenewBefore = 30 // 默认提前30天续期
        }

        // 计算下次续期时间
        nextRenew := cert.NotAfter.AddDate(0, 0, -cert.RenewBefore)
        cert.NextRenewAt = &nextRenew

        return global.DB.Create(cert).Error
}

// GetCertificates 获取证书列表
func GetCertificates() ([]Certificate, error) {
        var certs []Certificate
        err := global.DB.Find(&certs).Error
        return certs, err
}

// GetCertificate 获取证书
func GetCertificate(id uint) (*Certificate, error) {
        var cert Certificate
        err := global.DB.First(&cert, id).Error
        return &cert, err
}

// UpdateCertificate 更新证书
func UpdateCertificate(cert *Certificate) error {
        return global.DB.Save(cert).Error
}

// DeleteCertificate 删除证书
func DeleteCertificate(id uint) error {
        return global.DB.Delete(&Certificate{}, id).Error
}

// GetRenewalHistory 获取续期历史
func GetRenewalHistory(certID uint, limit int) ([]RenewalRecord, error) {
        var records []RenewalRecord
        query := global.DB.Model(&RenewalRecord{}).Order("created_at DESC")
        if certID > 0 {
                query = query.Where("cert_id = ?", certID)
        }
        if limit > 0 {
                query = query.Limit(limit)
        }
        err := query.Find(&records).Error
        return records, err
}

// MonitorCertificates 监控证书
func (m *CertRenewalManager) MonitorCertificates() {
        ticker := time.NewTicker(6 * time.Hour)
        defer ticker.Stop()

        for range ticker.C {
                m.CheckAllCertificates()
        }
}

// RequestNewCert 申请新证书
func (m *CertRenewalManager) RequestNewCert(domain, email string, dnsProvider string) (*Certificate, error) {
        // 执行 ACME 申请
        cmd := fmt.Sprintf("acme.sh --issue -d %s --dns %s", domain, dnsProvider)
        output, err := m.executeCommand(cmd)
        if err != nil {
                return nil, fmt.Errorf("申请证书失败: %w\n%s", err, output)
        }

        // 创建证书记录
        cert := &Certificate{
                Name:         domain,
                Domain:       domain,
                Provider:     ProviderLetsEncrypt,
                CertPath:     fmt.Sprintf("/etc/acme.sh/%s/%s.cer", domain, domain),
                KeyPath:      fmt.Sprintf("/etc/acme.sh/%s/%s.key", domain, domain),
                FullChainPath: fmt.Sprintf("/etc/acme.sh/%s/fullchain.cer", domain),
                ACMEEmail:    email,
                DNSProvider:  dnsProvider,
                AutoRenew:    true,
                RenewBefore:  30,
        }

        if err := AddCertificate(cert); err != nil {
                return nil, err
        }

        return cert, nil
}
