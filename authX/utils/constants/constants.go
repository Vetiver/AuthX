package constants

const ServiceOK = "OK"

const (
	UserRoleAdmin = "admin"
	UserRoleBase  = "user"
)

const (
	EventTypeRegister        = "user_registered"
	EventTypeSignIn          = "user_signIn"
	EventTypeAuthLoginFailed = "auth_login_failed"
)

var BaseRoles = []string{"admin", "user"}
