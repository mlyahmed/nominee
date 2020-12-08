package nominee

import "encoding/json"

type Nominee struct {
	Name    string
	Cluster string
	Address string
	Port    int
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
