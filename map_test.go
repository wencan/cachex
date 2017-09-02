package cachex

// wencan
// 2017-09-02 09:50

import (
	"testing"
	"time"
)

func TestListMap_PushBack(t *testing.T) {
	m := NewListMap()

	t1 := time.Now().Nanosecond()
	t2 := time.Now().Nanosecond()
	m.PushBack(1, t1)
	m.PushFront(2, t2)

	v, _ := m.Get(1)
	if v != t1 {
		t.Fatal("ListMap::Get error")
	}

	v1, _ := m.Back()
	v2, _ := m.Front()
	if t1 != v1 {
		t.FailNow()
	}
	if t2 != v2 {
		t.FailNow()
	}
	if m.Len() != 2 {
		t.Fatal("ListMap:Len error")
	}
}

func TestListMap_PopFront(t *testing.T) {
	m := NewListMap()

	t1 := time.Now().Nanosecond()
	t2 := time.Now().Nanosecond()
	m.PushBack(1, t1)
	m.PushFront(2, t2)

	v, _ := m.PopFront()
	if v != t2 {
		t.Fatal("ListMap::PopFront error")
	}

	v, _ = m.Front()
	if v != t1 {
		t.FailNow()
	}
	if m.Len() != 1 {
		t.FailNow()
	}

	m.PopFront()
	m.PopFront()
	if m.Len() != 0 {
		t.FailNow()
	}
}

func TestListMap_MoveToFront(t *testing.T) {
	m := NewListMap()

	t1 := time.Now().Nanosecond()
	t2 := time.Now().Nanosecond()
	m.PushBack(1, t1)
	m.PushFront(2, t2)

	m.MoveToFront(1)

	v1, _ := m.Back()
	v2, _ := m.Front()
	if v1 != t2 {
		t.FailNow()
	}
	if v2 != t1 {
		t.FailNow()
	}
}
