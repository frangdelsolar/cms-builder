package auth

func VerifyGodUser(envToken string, requestToken string) bool {
	return requestToken == envToken && requestToken != "" && envToken != ""
}
