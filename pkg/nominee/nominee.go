package nominee

import "encoding/json"

type StopChan <-chan error

type Nominee struct {
	ElectionKey string
	Name        string
	Cluster     string
	Address     string
	Port        int64
}

func (n Nominee) Marshal() string {
	data, _ := json.Marshal(n)
	return string(data)
}

func Unmarshal(data []byte) (Nominee, error) {
	value := Nominee{}
	err := json.Unmarshal(data, &value)
	return value, err
}
