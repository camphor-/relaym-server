package web

import (
	"testing"
)

func Test_deployPreviewCorsMiddleware_IsDeployPreviewOrigin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		origin string
		want   bool
	}{
		{
			name:   "正しいデプロイプレビューのURLならtrue",
			origin: "https://deploy-preview-191--relaym.netlify.app",
			want:   true,
		},
		{
			name:   "別ののURLならfalse",
			origin: "https://deploy-preview-191--relaym2.netlify.app",
			want:   false,
		},
		{
			name:   "前後に変な文字が入ってもfalse",
			origin: "https://evil.example.com/https://deploy-preview-191--relaym.netlify.app/hoge",
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newDeployPreviewCorsMiddleware(nil, true)
			if got := m.IsDeployPreviewOrigin(tt.origin); got != tt.want {
				t.Errorf("IsDeployPreviewOrigin() = %v, want %v", got, tt.want)
			}
		})
	}
}
