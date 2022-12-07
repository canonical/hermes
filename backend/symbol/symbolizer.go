package symbol

import "github.com/sirupsen/logrus"

type Symbolizer struct {
	ksymCache *KsymCache
}

func NewSymbolizer() (*Symbolizer, error) {
	ksymCache, err := NewKsymCache()
	if err != nil {
		return nil, err
	}
	return &Symbolizer{
		ksymCache: ksymCache,
	}, nil
}

func (symbolizer *Symbolizer) Symbolize(addr uint64) (string, error) {
	symbol, err := symbolizer.ksymCache.Resolve(addr)
	if err != nil {
		logrus.Errorf("Failed to resolve symbol, err [%s]", err)
		return "", err
	}
	return symbol, nil
}
