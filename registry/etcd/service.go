package etcd

import (
	"encoding/json"

	"github.com/spcent/x/registry"
)

func marshal(si *registry.ServiceInstance) (string, error) {
	data, err := json.Marshal(si)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func unmarshal(data []byte) (si *registry.ServiceInstance, err error) {
	err = json.Unmarshal(data, &si)
	return
}
