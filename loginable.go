package miaosic

type QrLoginSession struct {
	Url string `json:"url"`
	Key string `json:"key"`
}

type QrLoginResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type Loginable interface {
	Login(username string, password string) error
	Logout() error
	IsLogin() bool
	RefreshLogin() error
	QrLogin() (*QrLoginSession, error)
	QrLoginVerify(qrlogin *QrLoginSession) (*QrLoginResult, error)
	RestoreSession(session string) error
	SaveSession() string
}
