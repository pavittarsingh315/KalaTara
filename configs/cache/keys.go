package cache

// Key format:
//  1. "U" meaning "user"
//  2. user_id of user
//  3. "P" meaning "profile"
func ProfileKey(user_id string) string {
	return "U:" + user_id + ":P"
}

// Key format:
//  1. "NU" meaning "new user"
//  2. contact of registering user
//  3. "CC" meaning "confirmation code"
func NewUserConfirmCodeKey(contact string) string {
	return "NU:" + contact + ":CC"
}
