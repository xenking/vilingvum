package domain

type User struct {
	Name     string
	Settings UserSetting
	ID       int64
	IsAdmin  bool
}

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
)

type UserSetting struct{}
