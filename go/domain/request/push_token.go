package request

type CreatePushTokenRequest struct {
	PrincipalId string
	FcmToken    string
}

type UpdatePushTokenRequest struct {
	PrincipalId string
	FcmToken    string
}
