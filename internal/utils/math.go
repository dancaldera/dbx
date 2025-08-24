package utils

// Min returns the minimum of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximum of two integers
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// CalculateTotalPages computes total pages for pagination
func CalculateTotalPages(totalRows, itemsPerPage int) int {
	if itemsPerPage <= 0 {
		return 0
	}
	return (totalRows + itemsPerPage - 1) / itemsPerPage
}
