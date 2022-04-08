package l

import "testing"

// There is not much to test here.
func TestSetOutput(t *testing.T) {
	if err := SetOutput("file:///not/existing/folder/pointing/nowhere"); err == nil {
		t.Errorf("SetOutput(...) = nil, want error")
	}
}
