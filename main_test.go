package main

import (
	"net"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// Helper function to extract text from mcp.CallToolResult
func getResultText(result *mcp.CallToolResult) string {
	if len(result.Content) == 0 {
		return ""
	}

	// Assume the first content element contains our text
	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		return ""
	}

	return textContent.Text
}

// parseResultText extracts values from the result text using regex
func parseResultText(text string, pattern string) string {
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(text)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

// extractHosts parses the "Usable Hosts" value from result text
func extractHosts(text string) int {
	hostsStr := parseResultText(text, `Usable Hosts: (\d+)`)
	hosts, _ := strconv.Atoi(hostsStr)
	return hosts
}

// TestGetIPInfo tests the getIPInfo function
func TestGetIPInfo(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		expectedNet    string
		expectedMask   string
		expectedBcast  string
		expectedHosts  int
		expectedPrefix int
		isHostRoute    bool // Flag to indicate a /32 host route
	}{
		{
			name:           "Class C Network",
			input:          "192.168.1.0/24",
			expectedNet:    "192.168.1.0",
			expectedMask:   "255.255.255.0",
			expectedBcast:  "192.168.1.255",
			expectedHosts:  254,
			expectedPrefix: 24,
			isHostRoute:    false,
		},
		{
			name:           "Class B Network",
			input:          "172.16.0.0/16",
			expectedNet:    "172.16.0.0",
			expectedMask:   "255.255.0.0",
			expectedBcast:  "172.16.255.255",
			expectedHosts:  65534,
			expectedPrefix: 16,
			isHostRoute:    false,
		},
		{
			name:           "Small Subnet",
			input:          "10.0.0.0/30",
			expectedNet:    "10.0.0.0",
			expectedMask:   "255.255.255.252",
			expectedBcast:  "10.0.0.3",
			expectedHosts:  2,
			expectedPrefix: 30,
			isHostRoute:    false,
		},
		{
			name:           "Single Host",
			input:          "192.168.1.1/32",
			expectedNet:    "192.168.1.1",
			expectedMask:   "255.255.255.255",
			expectedBcast:  "", // No broadcast for /32
			expectedHosts:  1,  // Exactly 1 host for /32
			expectedPrefix: 32,
			isHostRoute:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse input
			ip, ipNet, err := net.ParseCIDR(tc.input)
			if err != nil {
				t.Fatalf("Failed to parse CIDR: %v", err)
			}

			// Call the function
			result, err := getIPInfo(ip, ipNet)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Extract text result
			text := getResultText(result)

			// Validate network
			network := parseResultText(text, `Network: ([0-9.]+)/\d+`)
			if network != tc.expectedNet {
				t.Errorf("Expected network %s, got %s", tc.expectedNet, network)
			}

			// Validate netmask
			netmask := parseResultText(text, `Netmask: ([0-9.]+)`)
			if netmask != tc.expectedMask {
				t.Errorf("Expected netmask %s, got %s", tc.expectedMask, netmask)
			}

			// Validate broadcast (only for non-host routes)
			if !tc.isHostRoute {
				broadcast := parseResultText(text, `Broadcast: ([0-9.]+)`)
				if broadcast != tc.expectedBcast {
					t.Errorf("Expected broadcast %s, got %s", tc.expectedBcast, broadcast)
				}
			} else {
				// For /32, we shouldn't have a broadcast line at all
				if strings.Contains(text, "Broadcast:") {
					t.Errorf("Expected no broadcast for /32, but found broadcast line")
				}
			}

			// Validate prefix
			prefix := parseResultText(text, `Prefix: /(\d+)`)
			prefixInt, _ := strconv.Atoi(prefix)
			if prefixInt != tc.expectedPrefix {
				t.Errorf("Expected prefix /%d, got /%s", tc.expectedPrefix, prefix)
			}

			// Validate hosts
			if !tc.isHostRoute {
				hosts := extractHosts(text)
				if hosts != tc.expectedHosts {
					t.Errorf("Expected %d hosts, got %d", tc.expectedHosts, hosts)
				}
			} else {
				// For /32, we should see "Hosts/Net: 1" instead of "Usable Hosts"
				hostsNet := parseResultText(text, `Hosts/Net: (\d+)`)
				hostsNetInt, _ := strconv.Atoi(hostsNet)
				if hostsNetInt != tc.expectedHosts {
					t.Errorf("Expected Hosts/Net: %d, got %d", tc.expectedHosts, hostsNetInt)
				}
			}
		})
	}
}

// TestSplitSubnet tests the splitSubnet function
// func TestSplitSubnet(t *testing.T) {
// 	testCases := []struct {
// 		name            string
// 		input           string
// 		newPrefix       int
// 		expectedError   bool
// 		expectedSubnets int
// 		firstSubnet     string
// 	}{
// 		{
// 			name:            "Split /24 to /26",
// 			input:           "192.168.1.0/24",
// 			newPrefix:       26,
// 			expectedError:   false,
// 			expectedSubnets: 4,
// 			firstSubnet:     "192.168.1.0/26",
// 		},
// 		{
// 			name:            "Split /16 to /18",
// 			input:           "10.0.0.0/16",
// 			newPrefix:       18,
// 			expectedError:   false,
// 			expectedSubnets: 4,
// 			firstSubnet:     "10.0.0.0/18",
// 		},
// 		{
// 			name:            "Invalid prefix (smaller)",
// 			input:           "192.168.1.0/24",
// 			newPrefix:       16,
// 			expectedError:   true,
// 			expectedSubnets: 0,
// 			firstSubnet:     "",
// 		},
// 		{
// 			name:            "Invalid prefix (too large)",
// 			input:           "192.168.1.0/24",
// 			newPrefix:       33,
// 			expectedError:   true,
// 			expectedSubnets: 0,
// 			firstSubnet:     "",
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			// Parse input
// 			ip, ipNet, err := net.ParseCIDR(tc.input)
// 			if err != nil {
// 				t.Fatalf("Failed to parse CIDR: %v", err)
// 			}

// 			// Call the function
// 			result, err := splitSubnet(ip, ipNet, tc.newPrefix)

// 			// Check for expected errors
// 			if tc.expectedError {
// 				if err == nil {
// 					t.Fatalf("Expected error but got none")
// 				}
// 				return
// 			}

// 			// Should not have error for valid cases
// 			if err != nil {
// 				t.Fatalf("Unexpected error: %v", err)
// 			}

// 			text := getResultText(result)

// 			// Check if we got the expected number of subnets
// 			subnetsMsg := parseResultText(text, `Splitting into (\d+) subnets`)
// 			numSubnets, _ := strconv.Atoi(subnetsMsg)
// 			if numSubnets != tc.expectedSubnets {
// 				t.Errorf("Expected %d subnets, got %d", tc.expectedSubnets, numSubnets)
// 			}

// 			// Check the first subnet
// 			if tc.firstSubnet != "" {
// 				firstSubnet := parseResultText(text, `Subnet 1: ([0-9.]+/\d+)`)
// 				if firstSubnet != tc.firstSubnet {
// 					t.Errorf("Expected first subnet %s, got %s", tc.firstSubnet, firstSubnet)
// 				}
// 			}
// 		})
// 	}
// }

// TestGetNetmask tests the getNetmask function
func TestGetNetmask(t *testing.T) {
	testCases := []struct {
		name             string
		input            string
		expectedMask     string
		expectedWildcard string
		expectedHex      string
	}{
		{
			name:             "Class C Netmask",
			input:            "192.168.1.0/24",
			expectedMask:     "255.255.255.0",
			expectedWildcard: "0.0.0.255",
			expectedHex:      "0xFFFFFF00",
		},
		{
			name:             "Class B Netmask",
			input:            "172.16.0.0/16",
			expectedMask:     "255.255.0.0",
			expectedWildcard: "0.0.255.255",
			expectedHex:      "0xFFFF0000",
		},
		{
			name:             "Custom Netmask",
			input:            "10.0.0.0/27",
			expectedMask:     "255.255.255.224",
			expectedWildcard: "0.0.0.31",
			expectedHex:      "0xFFFFFFE0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse input
			_, ipNet, err := net.ParseCIDR(tc.input)
			if err != nil {
				t.Fatalf("Failed to parse CIDR: %v", err)
			}

			// Call the function
			result, err := getNetmask(ipNet)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			text := getResultText(result)

			// Validate netmask
			netmask := parseResultText(text, `Netmask: ([0-9.]+) =`)
			if netmask != tc.expectedMask {
				t.Errorf("Expected netmask %s, got %s", tc.expectedMask, netmask)
			}

			// Validate wildcard mask
			wildcard := parseResultText(text, `Wildcard: ([0-9.]+)`)
			if wildcard != tc.expectedWildcard {
				t.Errorf("Expected wildcard %s, got %s", tc.expectedWildcard, wildcard)
			}

			// Validate hex notation
			hexMask := parseResultText(text, `Hex netmask: (0x[0-9A-F]+)`)
			if hexMask != tc.expectedHex {
				t.Errorf("Expected hex netmask %s, got %s", tc.expectedHex, hexMask)
			}
		})
	}
}

// TestWildcardMask tests the wildcardMask helper function
func TestWildcardMask(t *testing.T) {
	testCases := []struct {
		name     string
		mask     net.IPMask
		expected []string
	}{
		{
			name:     "Class C mask",
			mask:     net.CIDRMask(24, 32),
			expected: []string{"0", "0", "0", "255"},
		},
		{
			name:     "Class B mask",
			mask:     net.CIDRMask(16, 32),
			expected: []string{"0", "0", "255", "255"},
		},
		{
			name:     "Custom mask",
			mask:     net.CIDRMask(20, 32),
			expected: []string{"0", "0", "15", "255"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := wildcardMask(tc.mask)

			if len(result) != len(tc.expected) {
				t.Fatalf("Expected %d elements, got %d", len(tc.expected), len(result))
			}

			for i, v := range result {
				if v != tc.expected[i] {
					t.Errorf("At index %d expected %s, got %s", i, tc.expected[i], v)
				}
			}
		})
	}
}

// Helper test for parsing functions
func TestHelperFunctions(t *testing.T) {
	// Sample output text
	sampleText := `IP Address: 192.168.1.1
Network: 192.168.1.0/24
Netmask: 255.255.255.0
Broadcast: 192.168.1.255
Prefix: /24
Usable Hosts: 254`

	t.Run("parseResultText", func(t *testing.T) {
		network := parseResultText(sampleText, `Network: ([0-9.]+)/\d+`)
		if network != "192.168.1.0" {
			t.Errorf("Expected '192.168.1.0', got '%s'", network)
		}
	})

	t.Run("extractHosts", func(t *testing.T) {
		hosts := extractHosts(sampleText)
		if hosts != 254 {
			t.Errorf("Expected 254 hosts, got %d", hosts)
		}
	})
}

// For mocking purposes, we may need to define this if mcp.NewToolResultText isn't available for testing
type mockContent struct {
	mcp.TextContent
}

func TestMockNewToolResultText(t *testing.T) {
	// This is just a placeholder to verify our mocking approach
	// You might need to adapt this based on the actual mcp package behavior
	result := mcp.NewToolResultText("test text")
	text := getResultText(result)
	if text != "test text" {
		t.Errorf("Mock text extraction failed, expected 'test text', got '%s'", text)
	}
}
