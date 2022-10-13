package value

type Notification interface {
	DataOnly() bool
	Data() map[string]string
	NotificationMessage() NotificationMessage
	Principal() string
}

type NotificationMessage struct {
	Title string
	Body  string
}

type DataOnlyNotification struct {
	principal string
	data      map[string]string
}

func (d DataOnlyNotification) DataOnly() bool {
	return true
}

func (d DataOnlyNotification) Data() map[string]string {
	return d.data
}

func (d DataOnlyNotification) NotificationMessage() NotificationMessage {
	return NotificationMessage{}
}

func (d DataOnlyNotification) Principal() string {
	return d.principal
}

func NewDataOnlyNotification(principal string, data map[string]string) DataOnlyNotification {
	return DataOnlyNotification{
		principal: principal,
		data:      data,
	}
}

type TitledNotification struct {
	principal           string
	notificationMessage NotificationMessage
	data                map[string]string
}

func (d TitledNotification) DataOnly() bool {
	return false
}

func (d TitledNotification) Data() map[string]string {
	return d.data
}

func (d TitledNotification) NotificationMessage() NotificationMessage {
	return d.notificationMessage
}

func (d TitledNotification) Principal() string {
	return d.principal
}

func NewTitledNotification(principal string, message NotificationMessage, data map[string]string) TitledNotification {
	return TitledNotification{
		principal:           principal,
		notificationMessage: message,
		data:                data,
	}
}
