package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	len   int
	front *ListItem
	back  *ListItem
}

func NewList() List {
	return new(list)
}

func (l *list) Len() int {
	return l.len
}

func (l *list) Front() *ListItem {
	return l.front
}

func (l *list) Back() *ListItem {
	return l.back
}

func (l *list) PushFront(v interface{}) *ListItem {
	l.len++
	newItem := &ListItem{Value: v}
	if l.front == nil {
		l.front = newItem
		l.back = newItem
		return newItem
	}
	newItem.Next = l.front
	l.front.Prev = newItem
	l.front = newItem
	return newItem
}

func (l *list) PushBack(v interface{}) *ListItem {
	l.len++
	newItem := &ListItem{Value: v}
	if l.back == nil {
		l.front = newItem
		l.back = newItem
		return newItem
	}
	newItem.Prev = l.back
	l.back.Next = newItem
	l.back = newItem
	return newItem
}

func (l *list) Remove(i *ListItem) {
	if i == nil {
		return
	}
	l.len--
	if i.Prev != nil {
		i.Prev.Next = i.Next
	} else {
		l.front = i.Next
	}
	if i.Next != nil {
		i.Next.Prev = i.Prev
	} else {
		l.back = i.Prev
	}
}

func (l *list) MoveToFront(i *ListItem) {
	if i == nil || i == l.front {
		return
	}
	l.Remove(i)
	i.Next = l.front
	l.front.Prev = i
	l.front = i
	l.len++
}
