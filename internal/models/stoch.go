package models

import "encoding/json"

type Stoch struct {
	RSI float64 `json:"rsi"`
	K   float64 `json:"k"`
	D   float64 `json:"d"`
}

type Oscillator struct {
	Symbol string            `json:"symbol"`
	Stoch  map[string]*Stoch `json:"stoch"`
}

func (s *Oscillator) String() string {
	b, _ := json.Marshal(s)
	return string(b)
}

func (o *Oscillator) GetRSI() float64 {
	if o.Stoch == nil || o.Stoch["1h"] == nil {
		return 0
	}
	return o.Stoch["1h"].RSI
}
