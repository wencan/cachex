package cachex

// 哨兵机制实现。
// 用于解决缓存失效风暴问题
// wencan
// 2017-08-31 15:33

// Sentinel 哨兵。一个生产者，多个消费者等待生产者完成并提交结果
type Sentinel struct {
	flag chan interface{}

	result interface{}
	err    error
}

// NewSentinel 新建哨兵
func NewSentinel() *Sentinel {
	return &Sentinel{
		flag: make(chan interface{}),
	}
}

// Done 生产者提交结果，消费者将等待到提交的结果
func (s *Sentinel) Done(result interface{}, err error) {
	s.result = result
	s.err = err

	close(s.flag)
}

// Wait 消费者等待生产者提交结果
func (s *Sentinel) Wait() (result interface{}, err error) {
	<-s.flag

	return s.result, s.err
}

// Destroy 销毁
func (s *Sentinel) Destroy() {
	close(s.flag)
}
