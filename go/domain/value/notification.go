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

type dataOnlyNotification struct {
	principal string
	data      map[string]string
}

func (d dataOnlyNotification) DataOnly() bool {
	return true
}

func (d dataOnlyNotification) Data() map[string]string {
	return d.data
}

func (d dataOnlyNotification) NotificationMessage() NotificationMessage {
	return NotificationMessage{}
}

func (d dataOnlyNotification) Principal() string {
	return d.principal
}

func NewDataOnlyNotification(principal string, data map[string]string) dataOnlyNotification {
	return dataOnlyNotification{
		principal: principal,
		data:      data,
	}
}

type notification struct {
	principal           string
	notificationMessage NotificationMessage
	data                map[string]string
}

func (d notification) DataOnly() bool {
	return false
}

func (d notification) Data() map[string]string {
	return d.data
}

func (d notification) NotificationMessage() NotificationMessage {
	return d.notificationMessage
}

func (d notification) Principal() string {
	return d.principal
}

func NewNotification(principal string, message NotificationMessage, data map[string]string) notification {
	return notification{
		principal:           principal,
		notificationMessage: message,
		data:                data,
	}
}
