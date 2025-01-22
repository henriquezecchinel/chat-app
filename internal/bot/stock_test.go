package bot

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchStockData(t *testing.T) {
	tests := []struct {
		name           string
		stockCode      string
		mockResponse   string
		expectedResult string
	}{
		{
			name:           "Valid stock code",
			stockCode:      "AAPL.US",
			mockResponse:   "Symbol,Date,Time,Open,High,Low,Close,Volume\nAAPL.US,2025-01-22,16:15:22,219.79,223.3528,219.79,222.4683,8385754",
			expectedResult: "AAPL.US quote is $219.79 per share",
		},
		{
			name:           "Invalid stock code format",
			stockCode:      "INVALID CODE",
			mockResponse:   "",
			expectedResult: "Invalid stock code: INVALID CODE",
		},
		{
			name:           "No data available",
			stockCode:      "BAD",
			mockResponse:   "Symbol,Date,Time,Open,High,Low,Close,Volume\nBAD.US,N/D,N/D,N/D,N/D,N/D,N/D,N/D",
			expectedResult: "No data available for stock code BAD",
		},
		{
			name:           "Unexpected data format",
			stockCode:      "AAPL",
			mockResponse:   "Unexpected,Data,Format",
			expectedResult: "No data available for stock code AAPL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, err := w.Write([]byte(tt.mockResponse))
				if err != nil {
					return
				}
			}))
			defer server.Close()

			result := FetchStockData(tt.stockCode)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
