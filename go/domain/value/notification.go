package value

const (
	NotificationCategory_Taxicall   = "Taxicall"
	NotificationCategory_Payment    = "Payment"
	NotificationCategory_Driver     = "Driver"
	NotificationCategory_Settlement = "Settlement"
)

type Notification struct {
	Principal  string              `json:"principal"`
	MessageKey string              `json:"messageKey"`
	Category   string              `json:"category"`
	Message    NotificationMessage `json:"message"`
	Data       map[string]string   `json:"data"`
}

type NotificationMessage struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func NewNotification(principal, category, title, body, messageKey string, data map[string]string) Notification {
	return Notification{
		Principal:  principal,
		Category:   category,
		MessageKey: messageKey,
		Message: NotificationMessage{
			Title: title,
			Body:  body,
		},
		Data: data,
	}
}
