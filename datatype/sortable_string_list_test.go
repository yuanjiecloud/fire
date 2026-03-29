package datatype

import (
	"sort"
	"testing"
)

func TestSortableStringList_Len(t *testing.T) {
	tests := []struct {
		name string
		list SortableStringList
		want int
	}{
		{"empty", SortableStringList{}, 0},
		{"one", SortableStringList{"a"}, 1},
		{"three", SortableStringList{"c", "a", "b"}, 3},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.list.Len(); got != tc.want {
				t.Errorf("Len() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestSortableStringList_Less(t *testing.T) {
	list := SortableStringList{"banana", "apple", "cherry"}
	if !list.Less(1, 0) {
		t.Error("Less(1,0): 'apple' < 'banana' should be true")
	}
	if list.Less(0, 1) {
		t.Error("Less(0,1): 'banana' < 'apple' should be false")
	}
	if list.Less(0, 0) {
		t.Error("Less(0,0): equal elements should not be Less")
	}
}

func TestSortableStringList_Swap(t *testing.T) {
	list := SortableStringList{"x", "y", "z"}
	list.Swap(0, 2)
	if list[0] != "z" || list[2] != "x" {
		t.Errorf("Swap(0,2) failed, got %v", list)
	}
}

func TestSortableStringList_Sort(t *testing.T) {
	list := SortableStringList{"cherry", "apple", "banana"}
	sort.Sort(list)
	expected := SortableStringList{"apple", "banana", "cherry"}
	for i, v := range expected {
		if list[i] != v {
			t.Errorf("after Sort index %d: got %q, want %q", i, list[i], v)
		}
	}
}

func TestSortableStringList_SortAlreadySorted(t *testing.T) {
	list := SortableStringList{"a", "b", "c"}
	sort.Sort(list)
	if list[0] != "a" || list[1] != "b" || list[2] != "c" {
		t.Errorf("already-sorted list changed: %v", list)
	}
}

func TestSortableStringList_SortSingleElement(t *testing.T) {
	list := SortableStringList{"only"}
	sort.Sort(list)
	if len(list) != 1 || list[0] != "only" {
		t.Errorf("single-element sort failed: %v", list)
	}
}

func TestSortableStringList_SortEmpty(t *testing.T) {
	var list SortableStringList
	sort.Sort(list)
	if len(list) != 0 {
		t.Errorf("empty sort produced non-empty list: %v", list)
	}
}
