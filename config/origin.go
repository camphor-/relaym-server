package config

import "regexp"

// IsReliableOrigin は受け取ったURLが信頼できるURLか確認します。
func IsReliableOrigin(url string) bool {
	if IsDev() {
		return isReliableOriginDev(url)
	}
	// ローカルは何でも良い
	if IsLocal() {
		return true
	}

	return isReliableOrigin(url)
}

func isReliableOrigin(url string) bool {
	return FrontendURL() == url
}

func isReliableOriginDev(url string) bool {
	re := regexp.MustCompile(`^https://[a-zA-Z0-9-_]+\.relaym\.pages\.dev$`)
	return re.MatchString(url)
}
