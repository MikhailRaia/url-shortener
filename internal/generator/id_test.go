package generator

import (
	"testing"
)

func TestGenerateID(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		wantErr bool
	}{
		{
			name:    "Generate ID with length 8",
			length:  8,
			wantErr: false,
		},
		{
			name:    "Generate ID with length 16",
			length:  16,
			wantErr: false,
		},
		{
			name:    "Generate ID with length 0",
			length:  0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateID(tt.length)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(got) != tt.length {
					t.Errorf("GenerateID() returned ID with length = %v, want %v", len(got), tt.length)
				}

				got2, _ := GenerateID(tt.length)
				if got == got2 && tt.length > 0 {
					t.Errorf("GenerateID() generated the same ID twice: %v", got)
				}
			}
		})
	}
}
