package types

type Node struct {
	Name          string      `json:"name"`
	URL           string      `json:"url"`
	Remote        string      `json:"remote"`
	ClientVersion string      `json:"client_version"`
	Network       interface{} `json:"network"` // Can be string, int, or nil
}

type Nodes map[string]Node
