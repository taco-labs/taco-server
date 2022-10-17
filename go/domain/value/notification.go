package value

type Notification struct {
	Principal string
	Message   NotificationMessage
	Data      map[string]string
}

type NotificationMessage struct {
	Title string
	Body  string
}
