package cpuanalyzer

import "errors"

// 环形队列
type CircleQueue struct {
	length int   // 队列长度
	head   int   // 指向队列首 0
	tail   int   // 指向队列尾 0
	data   []interface{} // 数组 => 模拟队列
}

func NewCircleQueue(length int) *CircleQueue {
	return &CircleQueue{
		length: length,
		data:   make([]interface{}, length),
	}
}

func (s *CircleQueue)GetByIndex(index int)(val interface{}, err error) {
	val = s.data[(s.head + index)% s.length]
	return
}

func (s *CircleQueue)UpdateByIndex(index int, val interface{}){
	s.data[(s.head + index)% s.length] = val
	return
}

// 队列是否为空
func (s *CircleQueue) IsEmpty() bool {
	return s.head == s.tail
}

// 队列是否已经满了, 采用空一格的方式
func (s *CircleQueue) IsFull() (res bool) {
	return s.head == (s.tail+1)%s.length
}

// 环形队列长度
func (s *CircleQueue) Len() (res int) {
	return (s.tail - s.head + s.length) % s.length
}

// 队列尾新增元素
func (s *CircleQueue) Push(val interface{}) (b error) {
	if s.IsFull() {
		return errors.New("队列已满！")
	}
	s.data[s.tail] = val
	s.tail = (s.tail + 1) % s.length
	return
}

// 队列头弹出元素
func (s *CircleQueue) Pop() (val interface{}, err error) {
	if s.IsEmpty() {
		return nil, errors.New("队列为空！")
	}
	val = s.data[s.head]
	s.head = (s.head + 1) % s.length
	return
}

func (s *CircleQueue) Each(fn func(node interface{})) {
	for i := s.head; i < s.tail+s.length; i++ {
		fn(s.data[i%s.length])
	}
}

// 清空
func (s *CircleQueue) Clear() {
	s.head = 0
	s.tail = 0
	s.data = make([]interface{}, s.length)
}
