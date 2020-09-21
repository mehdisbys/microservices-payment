package handlers

import (
	"testing"
)

func TestRequestHandler_checkAuth(t *testing.T) {
	type fields struct {
		userTokens map[string]string
	}
	type args struct {
		r UserCheckAuthRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "happy path auth",
			fields: fields{userTokens: map[string]string{"1": "abc"}},
			args: args{UserCheckAuthRequest{
				UserID: "1",
				Token:  "abc",
			}},
		},
		{
			name:   "invalid auth",
			fields: fields{userTokens: map[string]string{"1": "abc"}},
			args: args{UserCheckAuthRequest{
				UserID: "1",
				Token:  "abcd",
			}},
			wantErr: true,
		},
		{
			name:   "no tokens loaded",
			fields: fields{userTokens: map[string]string{"1": "abc"}},
			args: args{UserCheckAuthRequest{
				UserID: "1",
				Token:  "abcd",
			}},
			wantErr: true,
		},
		{
			name:    "no auth provided",
			fields:  fields{userTokens: map[string]string{"1": "abc"}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &RequestHandler{
				userTokens: tt.fields.userTokens,
			}
			if err := s.checkAuth(tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("checkAuth() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
