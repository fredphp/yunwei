package notify

// Notifier 通知器接口 - 供所有服务使用
type Notifier interface {
	SendPatrolReport(record interface{}) error
	SendAlert(alert interface{}) error
	SendMessage(title, content string) error
}

// NotifierAdapter 适配器，用于适配具体实现
type NotifierAdapter struct {
	SendPatrolReportFunc func(record interface{}) error
	SendAlertFunc        func(alert interface{}) error
	SendMessageFunc      func(title, content string) error
}

func (a *NotifierAdapter) SendPatrolReport(record interface{}) error {
	if a.SendPatrolReportFunc != nil {
		return a.SendPatrolReportFunc(record)
	}
	return nil
}

func (a *NotifierAdapter) SendAlert(alert interface{}) error {
	if a.SendAlertFunc != nil {
		return a.SendAlertFunc(alert)
	}
	return nil
}

func (a *NotifierAdapter) SendMessage(title, content string) error {
	if a.SendMessageFunc != nil {
		return a.SendMessageFunc(title, content)
	}
	return nil
}
