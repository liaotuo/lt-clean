package sysutil

import "testing"

func TestDiskFreeBytes(t *testing.T) {
	result := DiskFreeBytes()
	if result == 0 {
		t.Fatal("DiskFreeBytes returned 0 — expected non-zero on a working macOS system")
	}
	t.Logf("DiskFreeBytes = %d bytes", result)
}

func TestParseSizeSuffix(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{"bytes no suffix", "100", 100, false},
		{"bytes lower b", "50b", 50, false},
		{"kilobytes K", "100K", 102400, false},
		{"kilobytes KB", "100KB", 102400, false},
		{"megabytes M", "1M", 1048576, false},
		{"megabytes MB", "1MB", 1048576, false},
		{"gigabytes G", "1G", 1073741824, false},
		{"gigabytes GB", "1GB", 1073741824, false},
		{"terabytes T", "1T", 1099511627776, false},
		{"terabytes TB", "1TB", 1099511627776, false},
		{"case insensitive m", "1m", 1048576, false},
		{"case insensitive mb", "1mb", 1048576, false},
		{"case insensitive g", "1g", 1073741824, false},
		{"case insensitive gb", "1gb", 1073741824, false},
		{"empty string", "", 0, true},
		{"letters only", "abc", 0, true},
		{"invalid suffix", "100X", 0, true},
		{"float suffix", "1.5M", 0, true},
		{"negative value", "-10M", 0, true},
		{"missing numeric", "M", 0, true},
		{"space before suffix", "100 M", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSizeSuffix(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSizeSuffix(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseSizeSuffix(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}