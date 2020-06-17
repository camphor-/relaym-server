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

// CORSAllowOrigin はCORSのAllow Originを取得します。
func CORSAllowOrigin() string {
	return os.Getenv("CORS_ALLOW_ORIGIN")
}

// FrontendURL はフロントエンドサーバのURLを取得します。
func FrontendURL() string {
	return os.Getenv("FRONTEND_URL")
}
