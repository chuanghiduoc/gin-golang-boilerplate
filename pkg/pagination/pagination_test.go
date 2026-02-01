package pagination

import (
	"testing"
)

func TestRequest_Normalize(t *testing.T) {
	tests := []struct {
		name         string
		req          Request
		wantPage     int
		wantPageSize int
	}{
		{
			name:         "zero values",
			req:          Request{Page: 0, PageSize: 0},
			wantPage:     DefaultPage,
			wantPageSize: DefaultPageSize,
		},
		{
			name:         "negative values",
			req:          Request{Page: -1, PageSize: -5},
			wantPage:     DefaultPage,
			wantPageSize: DefaultPageSize,
		},
		{
			name:         "valid values",
			req:          Request{Page: 2, PageSize: 20},
			wantPage:     2,
			wantPageSize: 20,
		},
		{
			name:         "exceeds max page size",
			req:          Request{Page: 1, PageSize: 200},
			wantPage:     1,
			wantPageSize: MaxPageSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.req.Normalize()
			if tt.req.Page != tt.wantPage {
				t.Errorf("Page = %v, want %v", tt.req.Page, tt.wantPage)
			}
			if tt.req.PageSize != tt.wantPageSize {
				t.Errorf("PageSize = %v, want %v", tt.req.PageSize, tt.wantPageSize)
			}
		})
	}
}

func TestRequest_Offset(t *testing.T) {
	tests := []struct {
		name string
		req  Request
		want int32
	}{
		{
			name: "first page",
			req:  Request{Page: 1, PageSize: 10},
			want: 0,
		},
		{
			name: "second page",
			req:  Request{Page: 2, PageSize: 10},
			want: 10,
		},
		{
			name: "third page with 20 items per page",
			req:  Request{Page: 3, PageSize: 20},
			want: 40,
		},
		{
			name: "zero page (normalized)",
			req:  Request{Page: 0, PageSize: 10},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.req.Offset(); got != tt.want {
				t.Errorf("Offset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequest_Limit(t *testing.T) {
	tests := []struct {
		name string
		req  Request
		want int32
	}{
		{
			name: "default",
			req:  Request{Page: 1, PageSize: 10},
			want: 10,
		},
		{
			name: "custom size",
			req:  Request{Page: 1, PageSize: 25},
			want: 25,
		},
		{
			name: "zero (normalized)",
			req:  Request{Page: 1, PageSize: 0},
			want: int32(DefaultPageSize),
		},
		{
			name: "exceeds max (capped)",
			req:  Request{Page: 1, PageSize: 200},
			want: int32(MaxPageSize),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.req.Limit(); got != tt.want {
				t.Errorf("Limit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewResult(t *testing.T) {
	tests := []struct {
		name           string
		total          int64
		req            *Request
		wantTotal      int64
		wantPage       int
		wantPageSize   int
		wantTotalPages int
	}{
		{
			name:           "exact division",
			total:          100,
			req:            &Request{Page: 1, PageSize: 10},
			wantTotal:      100,
			wantPage:       1,
			wantPageSize:   10,
			wantTotalPages: 10,
		},
		{
			name:           "with remainder",
			total:          95,
			req:            &Request{Page: 1, PageSize: 10},
			wantTotal:      95,
			wantPage:       1,
			wantPageSize:   10,
			wantTotalPages: 10,
		},
		{
			name:           "single page",
			total:          5,
			req:            &Request{Page: 1, PageSize: 10},
			wantTotal:      5,
			wantPage:       1,
			wantPageSize:   10,
			wantTotalPages: 1,
		},
		{
			name:           "empty result",
			total:          0,
			req:            &Request{Page: 1, PageSize: 10},
			wantTotal:      0,
			wantPage:       1,
			wantPageSize:   10,
			wantTotalPages: 0,
		},
		{
			name:           "normalized request",
			total:          50,
			req:            &Request{Page: 0, PageSize: 0},
			wantTotal:      50,
			wantPage:       DefaultPage,
			wantPageSize:   DefaultPageSize,
			wantTotalPages: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewResult(tt.total, tt.req)
			if result.Total != tt.wantTotal {
				t.Errorf("Total = %v, want %v", result.Total, tt.wantTotal)
			}
			if result.Page != tt.wantPage {
				t.Errorf("Page = %v, want %v", result.Page, tt.wantPage)
			}
			if result.PageSize != tt.wantPageSize {
				t.Errorf("PageSize = %v, want %v", result.PageSize, tt.wantPageSize)
			}
			if result.TotalPages != tt.wantTotalPages {
				t.Errorf("TotalPages = %v, want %v", result.TotalPages, tt.wantTotalPages)
			}
		})
	}
}
