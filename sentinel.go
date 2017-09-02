package cachex

// wencan
// 2017-08-31 15:33

type Sentinel struct {
	flag chan interface{}

	result interface{}
	err error
}

func NewSentinel() *Sentinel {
	return &Sentinel{
		flag: make(chan interface{}),
	}
}

func (s *Sentinel) Done(result interface{}, err error) {
	s.result = result
	s.err = err

	close(s.flag)
}

func (s *Sentinel) Wait() (result interface{}, err error) {
	<- s.flag

	return s.result, s.err
}

func (s *Sentinel) Destroy() {
	close(s.flag)
}