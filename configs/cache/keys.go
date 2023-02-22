package cache

func ProfileKey(user_id string) string {
	return "user:" + user_id + ":profile"
}
