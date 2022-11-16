package value

type Notification struct {
	Principal string              `json:"principal"`
	Message   NotificationMessage `json:"message"`
	Data      map[string]string   `json:"data"`
}

type NotificationMessage struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}
