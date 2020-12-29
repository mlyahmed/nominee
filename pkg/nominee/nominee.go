package nominee

import "encoding/json"

// StopChan ...
type StopChan <-chan struct{}

// Nominee ...
type Nominee struct {
	ElectionKey string
	Name        string
	Address     string
	Port        int64
}

// Marshal ...
func (n Nominee) Marshal() string {
	data, _ := json.Marshal(n)
	return string(data)
}

// Unmarshal ...
func Unmarshal(data []byte) (Nominee, error) {
	value := Nominee{}
	err := json.Unmarshal(data, &value)
	return value, err
}
