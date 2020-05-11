package database

import (
	"reflect"
	"testing"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/go-gorp/gorp/v3"
)

func TestNewSessionRepository(t *testing.T) {
	// Prepare
	dbMap, err := NewDB()
	if err != nil {
		t.Fatal(err)
	}
	dbMap.AddTableWithName(sessionDTO{}, "sessions")
	truncateTable(t, dbMap)
	if err := dbMap.Insert(&sessionDTO{
		ID:         "existing_session_id",
		name:       "existing_session_name",
		creator_id: "existing_creator_id",
		queue_head: 0,
		state_type: "",
	}); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		args args
		want *SessionRepository
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSessionRepository(tt.args.dbMap); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSessionRepository() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSessionRepository_FindByID(t *testing.T) {
	type fields struct {
		dbMap *gorp.DbMap
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *entity.Session
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &SessionRepository{
				dbMap: tt.fields.dbMap,
			}
			got, err := r.FindByID(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("SessionRepository.FindByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SessionRepository.FindByID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSessionRepository_Store(t *testing.T) {
	type fields struct {
		dbMap *gorp.DbMap
	}
	type args struct {
		session *entity.Session
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &SessionRepository{
				dbMap: tt.fields.dbMap,
			}
			if err := r.Store(tt.args.session); (err != nil) != tt.wantErr {
				t.Errorf("SessionRepository.Store() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
