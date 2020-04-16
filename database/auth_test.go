package database

import (
	"testing"

	"github.com/camphor-/relaym-server/domain/entity"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestAuthRepository_Store(t *testing.T) {
	tests := []struct {
		name    string
		state   *entity.StateTemp
		wantErr bool
	}{
		{
			name: "正しく保存できる",
			state: &entity.StateTemp{
				State:       uuid.New().String(),
				RedirectURL: "https://example.com",
			},
			wantErr: false,
		},
	}
	dbMap, err := NewDB()
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := NewAuthRepository(dbMap)
			if err := r.Store(tt.state); (err != nil) != tt.wantErr {
				t.Errorf("Store() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got, err := r.FindStateByState(tt.state.State)
				if err != nil {
					t.Fatal(err)
				}
				if !cmp.Equal(got, tt.state) {
					t.Errorf("Store() got = %v, want %v", got, tt.state)
				}
			}
		})
	}
}

func TestAuthRepository_FindStateByState(t *testing.T) {
	tests := []struct {
		name    string
		state   string
		want    *entity.StateTemp
		wantErr bool
	}{
		{
			name:  "存在するStateを正しく取得できる",
			state: "state",
			want: &entity.StateTemp{
				State:       "state",
				RedirectURL: "https://example.com",
			},
			wantErr: false,
		},
		{
			name:    "存在しないStateを取得しようとするとエラーになる",
			state:   "not_found",
			want:    nil,
			wantErr: true,
		},
	}
	dbMap, err := NewDB()
	if err != nil {
		t.Fatal(err)
	}
	r := NewAuthRepository(dbMap)

	// Prepare
	truncateTable(t, dbMap)
	if err := r.Store(&entity.StateTemp{
		State:       "state",
		RedirectURL: "https://example.com",
	}); err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := r.FindStateByState(tt.state)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindStateByState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("FindStateByState() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthRepository_Delete(t *testing.T) {

	tests := []struct {
		name    string
		state   string
		wantErr bool
	}{
		{
			name:    "存在するStateを正しく削除できる",
			state:   "state",
			wantErr: false,
		},
		{
			name:    "存在しないStateを削除してもエラーにならない",
			state:   "not_found",
			wantErr: false,
		},
	}

	dbMap, err := NewDB()
	if err != nil {
		t.Fatal(err)
	}
	r := NewAuthRepository(dbMap)

	// Prepare
	truncateTable(t, dbMap)
	if err := r.Store(&entity.StateTemp{
		State:       "state",
		RedirectURL: "https://example.com",
	}); err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := r.Delete(tt.state); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
