package collector

type Perf struct{}

func NewPerf() (*Perf, error) {
	return &Perf{}, nil
}
