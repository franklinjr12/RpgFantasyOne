package pure

import "testing"

func TestDepthLessOrdering(t *testing.T) {
	if !DepthLess(10, 5, 1, 20, 1, 0) {
		t.Fatal("expected lower depth y to sort first")
	}
	if !DepthLess(10, 3, 1, 10, 9, 0) {
		t.Fatal("expected lower depth x to sort first when y matches")
	}
	if !DepthLess(10, 3, 1, 10, 3, 2) {
		t.Fatal("expected lower stable id to sort first when depth matches")
	}
}
