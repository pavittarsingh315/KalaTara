package cache

import "time"

const (
	ProfileExp            = time.Hour * 3
	NewUserConfirmCodeExp = time.Minute * 5
	PasswordResetCodeEXP  = time.Minute * 5
)
