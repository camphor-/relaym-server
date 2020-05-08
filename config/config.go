package config

import "os"

// IsLocal はローカル環境がどうか返します。
func IsLocal() bool {
	return os.Getenv("ENV") == "local"
}

// Port はサーバのポート番号を取得します。
func Port() string {
	return os.Getenv("PORT")
}
