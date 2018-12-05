package lrucache

// wencan
// 2017-09-02 09:01

import (
	"container/list"
)

type listEntry struct {
	key   interface{}
	value interface{}
}

type ListMap struct {
	sequence *list.List
	mapping  map[interface{}]*list.Element
}

func NewListMap() *ListMap {
	return &ListMap{
		sequence: list.New(),
		mapping:  make(map[interface{}]*list.Element),
	}
}

func (m *ListMap) Get(key interface{}) (value interface{}, ok bool) {
	elem, ok := m.mapping[key]
	if ok {
		entry := elem.Value.(*listEntry)
		return entry.value, ok
	}
	return nil, false
}

func (m *ListMap) Pop(key interface{}) (value interface{}, ok bool) {
	elem, ok := m.mapping[key]
	if ok {
		delete(m.mapping, key)
		m.sequence.Remove(elem)

		entry := elem.Value.(*listEntry)
		return entry.value, true
	}
	return nil, false
}

func (m *ListMap) Len() int {
	return m.sequence.Len()
}

func (m *ListMap) Front() (value interface{}, ok bool) {
	elem := m.sequence.Front()
	if elem != nil {
		entry := elem.Value.(*listEntry)
		return entry.value, true
	}
	return nil, false
}

func (m *ListMap) Back() (value interface{}, ok bool) {
	elem := m.sequence.Back()
	if elem != nil {
		entry := elem.Value.(*listEntry)
		return entry.value, true
	}
	return nil, false
}

func (m *ListMap) PopFront() (value interface{}, ok bool) {
	elem := m.sequence.Front()
	if elem != nil {
		entry := elem.Value.(*listEntry)

		m.sequence.Remove(elem)
		delete(m.mapping, entry.key)

		return entry.value, true
	}
	return nil, false
}

func (m *ListMap) PopBack() (value interface{}, ok bool) {
	elem := m.sequence.Back()
	if elem != nil {
		entry := elem.Value.(*listEntry)

		m.sequence.Remove(elem)
		delete(m.mapping, entry.key)

		return entry.value, true
	}
	return nil, false
}

func (m *ListMap) PushFront(key interface{}, value interface{}) {
	entry := &listEntry{
		key:   key,
		value: value,
	}

	elem := m.sequence.PushFront(entry)
	m.mapping[key] = elem
}

func (m *ListMap) PushBack(key interface{}, value interface{}) {
	entry := &listEntry{
		key:   key,
		value: value,
	}

	elem := m.sequence.PushBack(entry)
	m.mapping[key] = elem
}

func (m *ListMap) MoveToFront(key interface{}) {
	elem, ok := m.mapping[key]
	if ok {
		m.sequence.MoveToFront(elem)
	}
}

func (m *ListMap) MoveToBack(key interface{}) {
	elem, ok := m.mapping[key]
	if ok {
		m.sequence.MoveToBack(elem)
	}
}

func (m *ListMap) Clear() {
	m.mapping = make(map[interface{}]*list.Element)
	m.sequence = list.New()
}
