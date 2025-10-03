package utils

import "testing"

func TestMin(t *testing.T) {
	tests := []struct {
		name string
		a    int
		b    int
		want int
	}{
		{"first smaller", 1, 2, 1},
		{"second smaller", 5, 3, 3},
		{"equal", 4, 4, 4},
		{"negative numbers", -1, -5, -5},
		{"negative and positive", -1, 5, -1},
		{"zero and positive", 0, 5, 0},
		{"zero and negative", 0, -5, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Min(tt.a, tt.b); got != tt.want {
				t.Errorf("Min(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		name string
		a    int
		b    int
		want int
	}{
		{"first larger", 2, 1, 2},
		{"second larger", 3, 5, 5},
		{"equal", 4, 4, 4},
		{"negative numbers", -1, -5, -1},
		{"negative and positive", -1, 5, 5},
		{"zero and positive", 0, 5, 5},
		{"zero and negative", 0, -5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Max(tt.a, tt.b); got != tt.want {
				t.Errorf("Max(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestCalculateTotalPages(t *testing.T) {
	tests := []struct {
		name         string
		totalRows    int
		itemsPerPage int
		want         int
	}{
		{"exact division", 100, 10, 10},
		{"with remainder", 101, 10, 11},
		{"single page", 5, 10, 1},
		{"no items", 0, 10, 0},
		{"zero items per page", 100, 0, 0},
		{"negative items per page", 100, -5, 0},
		{"large numbers", 1000000, 25, 40000},
		{"one item", 1, 1, 1},
		{"one item large page", 1, 100, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateTotalPages(tt.totalRows, tt.itemsPerPage); got != tt.want {
				t.Errorf("CalculateTotalPages(%d, %d) = %d, want %d", tt.totalRows, tt.itemsPerPage, got, tt.want)
			}
		})
	}
}
