package toolkit

import "testing"

func TestTools_RandomString(t *testing.T) {
	var testTools Tools

	s := testTools.RandomString(10)
	if len(s) != 10 {
		t.Errorf("RandomString failed, length of string is %d, expected 10", len(s))
	}
}
