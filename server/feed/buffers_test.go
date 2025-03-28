package feed

import (
	"reflect"
	"testing"
)

func TestBufferWrite(t *testing.T) {
	b := NewBuffer[int](3)
	b.Write(1)
	b.Write(2)
	b.Write(3)

	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(b.items, expected) {
		t.Errorf("item error, expected=%v, got=%v", expected, b.items)
	}
	b.Write(7)
	expected = []int{7, 2, 3}
	if !reflect.DeepEqual(b.items, expected) {
		t.Errorf("item error, expected=%v, got=%v", expected, b.items)
	}
	b.Write(8)
	expected = []int{7, 8, 3}
	if !reflect.DeepEqual(b.items, expected) {
		t.Errorf("item error, expected=%v, got=%v", expected, b.items)
	}
	b.Write(9)
	expected = []int{7, 8, 9}
	if !reflect.DeepEqual(b.items, expected) {
		t.Errorf("item error, expected=%v, got=%v", expected, b.items)
	}
}
