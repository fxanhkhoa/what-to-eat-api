package types

type Void struct{}

type Kv struct {
	Key   string
	Value string
}

var NULL Void

type HttpCompression struct {
	Threshold int `json:"threshold,omitempty" mapstructure:"threshold,omitempty" msgpack:"threshold,omitempty"`
}

type PerMessageDeflate struct {
	Threshold int `json:"threshold,omitempty" mapstructure:"threshold,omitempty" msgpack:"threshold,omitempty"`
}
