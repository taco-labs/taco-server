package value

const (
	NotificationCategory_Taxicall = "Taxicall"
	NotificationCategory_Payment  = "Payment"
	NotificationCategory_Driver   = "Driver"
)

type Notification struct {
	Principal string              `json:"principal"`
	Category  string              `json:"category"`
	Message   NotificationMessage `json:"message"`
	Data      map[string]string   `json:"data"`
}

type NotificationMessage struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func NewNotification(principal, category, title, body string, data map[string]string) Notification {
	return Notification{
		Principal: principal,
		Category:  category,
		Message: NotificationMessage{
			Title: title,
			Body:  body,
		},
		Data: data,
	}
}
