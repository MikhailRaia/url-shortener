package memory

import (
	"testing"
)

func TestStorage_Save(t *testing.T) {
	storage := NewStorage()
	originalURL := "https://example.com"

	id, err := storage.Save(originalURL)
	if err != nil {
		t.Errorf("Storage.Save() error = %v", err)
		return
	}

	if id == "" {
		t.Errorf("Storage.Save() returned empty ID")
	}

	savedURL, found := storage.Get(id)
	if !found {
		t.Errorf("Storage.Get() couldn't find URL for ID = %v", id)
	}

	if savedURL != originalURL {
		t.Errorf("Storage.Get() = %v, want %v", savedURL, originalURL)
	}
}

func TestStorage_Get(t *testing.T) {
	storage := NewStorage()
	originalURL := "https://example.com"

	id, _ := storage.Save(originalURL)

	tests := []struct {
		name      string
		id        string
		wantURL   string
		wantFound bool
	}{
		{
			name:      "Get existing URL",
			id:        id,
			wantURL:   originalURL,
			wantFound: true,
		},
		{
			name:      "Get non-existing URL",
			id:        "nonexistent",
			wantURL:   "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, gotFound := storage.Get(tt.id)

			if gotFound != tt.wantFound {
				t.Errorf("Storage.Get() found = %v, want %v", gotFound, tt.wantFound)
			}

			if gotURL != tt.wantURL {
				t.Errorf("Storage.Get() = %v, want %v", gotURL, tt.wantURL)
			}
		})
	}
}
