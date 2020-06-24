package web

import (
	"testing"
)

func Test_deployPreviewCorsMiddleware_IsDeployPreviewOrigin(t *testing.T) {

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
			name:   "別ののURLならtrue",
			origin: "https://deploy-preview-191--relaym2.netlify.app",
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newDeployPreviewCorsMiddleware()
			if got := m.IsDeployPreviewOrigin(tt.origin); got != tt.want {
				t.Errorf("IsDeployPreviewOrigin() = %v, want %v", got, tt.want)
			}
		})
	}
}
