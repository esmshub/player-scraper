//go:build go1.20

package core

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
