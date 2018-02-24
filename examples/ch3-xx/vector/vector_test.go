// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

package vector

import "testing"

func TestFind(t *testing.T) {
	vec := []int{1, 2, 3, 4, 5, 6, 7, 8}
	if result := Find(vec, 5); result != true {
		t.Errorf("Could not find number in vector, got: %v", result)
	}

	if result := Find(vec, 10); result != false {
		t.Errorf("Returned true when false was expected")
	}
}

func TestSum(t *testing.T) {
	vec1 := []int32{1, 2, 3, 5}
	vec2 := []int32{1, 2, 3, 5}

	result := SumVec(vec1, vec2)

	if result[0] != 2 {
		t.Errorf("Expected 2, got %v, result was: %v", result[0], result)
	}

	if result[1] != 4 {
		t.Errorf("Expected 4, got %v, result was: %v", result[0], result)
	}

	if result[2] != 6 {
		t.Errorf("Expected 6, got %v, result was: %v", result[0], result)
	}

	if result[3] != 10 {
		t.Errorf("Expected 10, got %v, result was: %v", result[0], result)
	}
}
