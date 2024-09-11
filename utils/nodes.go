package utils

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	rpc_client "github.com/unifralabs/unifra-benchmark-tool/rpc_client"
	tooltypes "github.com/unifralabs/unifra-benchmark-tool/types"
)

// ParseNodes parses the given nodes according to input specification
func ParseNodes(nodes []string, verbose bool, requestMetadata bool) (map[string]tooltypes.Node, error) {
	if verbose {
		PrintHeader("Gathering node data...")
	}

	newNodes := make(map[string]tooltypes.Node)

	nodesChan := make([]tooltypes.Node, 0)

	for _, node := range nodes {
		parsedNode, err := ParseNode(node, requestMetadata)
		if err == nil {
			nodesChan = append(nodesChan, parsedNode)
		}
	}

	for _, node := range nodesChan {
		newNodes[node.Name] = node
	}

	if verbose {
		PrintNodesTable(newNodes)
	}

	return newNodes, nil
}

// PrintNodesTable prints a table of nodes
func PrintNodesTable(nodes map[string]tooltypes.Node) {
	// Implementation of printing table goes here
	// You may want to use a table formatting library for Go
}

// ParseNode parses a single node according to input specification
func ParseNode(node string, requestMetadata bool) (tooltypes.Node, error) {
	prefixes := []string{"http", "https", "ws", "wss"}

	// Parse name and URL
	var name, url string
	if strings.Contains(node, "=") {
		parts := strings.SplitN(node, "=", 2)
		name, url = parts[0], parts[1]
	} else {
		name, url = node, node
	}

	// Parse remote and URL
	var remote string
	if strings.Contains(url, ":") {
		parts := strings.SplitN(url, ":", 2)
		if !contains(prefixes, parts[0]) {
			if _, err := fmt.Sscanf(parts[1], "%d", new(int)); err != nil {
				remote, url = parts[0], parts[1]
			}
		}
	}

	// Add missing prefix
	if !hasPrefix(url, prefixes) {
		if strings.HasPrefix(url, "localhost") || isIP(url) {
			url = "http://" + url
		} else {
			url = "https://" + url
		}
	}

	var clientVersion, network string
	if requestMetadata {
		clientVersion = getNodeClientVersion(url, remote)
		network = "ethereum"
	}

	return tooltypes.Node{
		Name:          name,
		URL:           url,
		Remote:        remote,
		ClientVersion: clientVersion,
		Network:       network,
	}, nil
}

func getNodeClientVersion(url, remote string) string {
	if remote == "" {
		// Implement RPC call to get client version
		client, err := rpc_client.NewRpcClient(url)
		if err != nil {
			return ""
		}

		version, err := client.GetClientVersion()
		if err != nil {
			return ""
		}
		return version
	}

	cmd := exec.Command("ssh", remote, fmt.Sprintf("curl -X POST -H 'Content-Type: application/json' -d '{\"jsonrpc\": \"2.0\", \"method\": \"web3_clientVersion\", \"params\": [], \"id\": 1}' %s", url))
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	var response struct {
		Result string `json:"result"`
	}
	if err := json.Unmarshal(output, &response); err != nil {
		return ""
	}

	return response.Result
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func hasPrefix(s string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

func isIP(s string) bool {
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if _, err := fmt.Sscanf(part, "%d", new(int)); err != nil {
			return false
		}
	}
	return true
}
