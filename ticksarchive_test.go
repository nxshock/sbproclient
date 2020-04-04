package sbproclient

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenFileName(t *testing.T) {
	client := &Client{ServerLocation: time.Local}

	tests := []struct {
		Symbol           string
		Contract         string
		Date             time.Time
		ExpectedFileName string
	}{
		{"GC", "04-20", time.Date(2020, 01, 01, 0, 0, 0, 0, time.Local), "cache/GC/04-20/20200101.zip"},
		{"GC", "06-20", time.Date(2020, 02, 01, 0, 0, 0, 0, time.Local), "cache/GC/06-20/20200201.zip"},
	}

	for _, test := range tests {
		assert.Equal(t, test.ExpectedFileName, client.genFileName(test.Date, test.Symbol, test.Contract))
	}
}
