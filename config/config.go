package config

import "os"

// IsLocalはローカル環境がどうか返します。
func IsLocal() bool {
	return os.Getenv("ENV") == "local"
}
