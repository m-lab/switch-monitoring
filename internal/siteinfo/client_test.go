package siteinfo

import (
	"reflect"
	"testing"

	"github.com/m-lab/switch-monitoring/internal"
)

type mockHTTPProvider struct {
}

func TestNew(t *testing.T) {
	type args struct {
		projectID  string
		httpClient internal.HTTPProvider
	}
	tests := []struct {
		name string
		args args
		want *Client
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.projectID, tt.args.httpClient); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Switches(t *testing.T) {
	type fields struct {
		ProjectID  string
		httpClient internal.HTTPProvider
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Client{
				ProjectID:  tt.fields.ProjectID,
				httpClient: tt.fields.httpClient,
			}
			got, err := s.Switches()
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Switches() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.Switches() = %v, want %v", got, tt.want)
			}
		})
	}
}
