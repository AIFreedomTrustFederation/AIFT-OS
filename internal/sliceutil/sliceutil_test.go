package sliceutil

import (
	"testing"
)

func TestContainsFound(t *testing.T) {
	if !Contains([]string{"a", "b", "c"}, "b") {
		t.Error("should find b")
	}
}

func TestContainsNotFound(t *testing.T) {
	if Contains([]string{"a", "b", "c"}, "d") {
		t.Error("should not find d")
	}
}

func TestContainsEmpty(t *testing.T) {
	if Contains([]string{}, "a") {
		t.Error("should not find in empty slice")
	}
}

func TestContainsNil(t *testing.T) {
	if Contains(nil, "a") {
		t.Error("should not find in nil slice")
	}
}

func TestUniqueBasic(t *testing.T) {
	got := Unique([]string{"b", "a", "b", "c", "a"})
	want := []string{"a", "b", "c"}
	assertSliceEqual(t, "basic", got, want)
}

func TestUniqueEmptyStrings(t *testing.T) {
	got := Unique([]string{"a", "", "b", ""})
	want := []string{"a", "b"}
	assertSliceEqual(t, "empty strings", got, want)
}

func TestUniqueAlreadyUnique(t *testing.T) {
	got := Unique([]string{"x", "y", "z"})
	want := []string{"x", "y", "z"}
	assertSliceEqual(t, "already unique", got, want)
}

func TestUniqueNil(t *testing.T) {
	got := Unique(nil)
	if len(got) != 0 {
		t.Errorf("Unique(nil) = %v, want empty", got)
	}
}

func TestUniqueEmpty(t *testing.T) {
	got := Unique([]string{})
	if len(got) != 0 {
		t.Errorf("Unique([]) = %v, want empty", got)
	}
}

func TestUniqueAllEmpty(t *testing.T) {
	got := Unique([]string{"", "", ""})
	if len(got) != 0 {
		t.Errorf("Unique(all empty) = %v, want empty", got)
	}
}

func TestUniqueSorted(t *testing.T) {
	got := Unique([]string{"z", "a", "m"})
	want := []string{"a", "m", "z"}
	assertSliceEqual(t, "sorted", got, want)
}

func TestSortedBoolMapKeys(t *testing.T) {
	m := map[string]bool{"cherry": true, "apple": true, "banana": true}
	got := SortedBoolMapKeys(m)
	want := []string{"apple", "banana", "cherry"}
	assertSliceEqual(t, "bool map keys", got, want)
}

func TestSortedBoolMapKeysEmpty(t *testing.T) {
	got := SortedBoolMapKeys(map[string]bool{})
	if len(got) != 0 {
		t.Errorf("SortedBoolMapKeys(empty) = %v, want empty", got)
	}
}

func TestSortedIntMapKeys(t *testing.T) {
	m := map[string]int{"cherry": 3, "apple": 1, "banana": 2}
	got := SortedIntMapKeys(m)
	want := []string{"apple", "banana", "cherry"}
	assertSliceEqual(t, "int map keys", got, want)
}

func TestSortedIntMapKeysEmpty(t *testing.T) {
	got := SortedIntMapKeys(map[string]int{})
	if len(got) != 0 {
		t.Errorf("SortedIntMapKeys(empty) = %v, want empty", got)
	}
}

func assertSliceEqual(t *testing.T, name string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s: got %d items %v, want %d items %v", name, len(got), got, len(want), want)
		return
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%s: [%d] = %q, want %q", name, i, got[i], want[i])
		}
	}
}
