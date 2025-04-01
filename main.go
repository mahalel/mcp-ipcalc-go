package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Create a new MCP server
	s := server.NewMCPServer(
		"IP Calculator",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)

	// Add an IP calculator tool
	ipCalcTool := mcp.NewTool("ipcalc",
		mcp.WithDescription("Calculate IP address information like ipcalc in Linux"),
		mcp.WithString("operation",
			mcp.Required(),
			mcp.Description("The operation to perform (info, split, netmask)"),
			mcp.Enum(
				"info",
				// "split",
				"netmask"),
		),
		mcp.WithString("ip",
			mcp.Required(),
			mcp.Description("IP address in CIDR notation (e.g., 192.168.1.0/24)"),
		),
		mcp.WithNumber("prefix",
			mcp.Description("New prefix length for subnet splitting (required for 'split' operation)"),
		),
	)

	// Add the IP calculator handler
	s.AddTool(ipCalcTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		op := request.Params.Arguments["operation"].(string)
		ipCIDR := request.Params.Arguments["ip"].(string)

		// Parse IP address and network
		ip, ipNet, err := net.ParseCIDR(ipCIDR)
		if err != nil {
			return nil, errors.New("Invalid IP address format. Please use CIDR notation (e.g., 192.168.1.0/24)")
		}

		switch op {
		case "info":
			return getIPInfo(ip, ipNet)
		// case "split":
		// 	// Get the new prefix length
		// 	prefixVal, ok := request.Params.Arguments["prefix"]
		// 	if !ok {
		// 		return nil, errors.New("Prefix is required for split operation")
		// 	}

		// 	newPrefix := int(prefixVal.(float64))
		// 	return splitSubnet(ip, ipNet, newPrefix)
		case "netmask":
			return getNetmask(ipNet)
		default:
			return nil, errors.New("Unknown operation")
		}
	})

	// Start the server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func getIPInfo(ip net.IP, ipNet *net.IPNet) (*mcp.CallToolResult, error) {
	mask := ipNet.Mask
	ones, bits := mask.Size()

	// Get network address
	network := ipNet.IP

	// Format output based on prefix length
	result := fmt.Sprintf("IP Address: %s\n", ip)
	result += fmt.Sprintf("Network: %s/%d\n", network, ones)

	// Convert mask to string format
	maskStr := fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3])
	result += fmt.Sprintf("Netmask: %s\n", maskStr)

	if ones == bits {
		// Special case for /32 (or /128 in IPv6)
		result += fmt.Sprintf("Prefix: /%d\n", ones)
		result += "Hosts/Net: 1\n"
	} else {
		// Calculate broadcast address for non-/32 networks
		broadcast := make(net.IP, len(network))
		copy(broadcast, network)
		for i := range broadcast {
			broadcast[i] |= ^mask[i]
		}
		result += fmt.Sprintf("Broadcast: %s\n", broadcast)
		result += fmt.Sprintf("Prefix: /%d\n", ones)

		// Calculate total hosts
		var totalHosts float64
		if bits == 32 { // IPv4
			totalHosts = max(float64(uint64(1)<<uint64(bits-ones))-2, 0)
		}
		result += fmt.Sprintf("Usable Hosts: %.0f\n", totalHosts)
	}

	return mcp.NewToolResultText(result), nil
}

// func splitSubnet(_ net.IP, ipNet *net.IPNet, newPrefix int) (*mcp.CallToolResult, error) {
// 	currentPrefix, totalBits := ipNet.Mask.Size()

// 	if newPrefix <= currentPrefix {
// 		return nil, errors.New("New prefix must be larger than current prefix")
// 	}

// 	if newPrefix > totalBits {
// 		return nil, errors.New("New prefix cannot be larger than address size")
// 	}

// 	// Calculate number of subnets
// 	numSubnets := 1 << (newPrefix - currentPrefix)

// 	result := fmt.Sprintf("Original network: %s/%d\n", ipNet.IP, currentPrefix)
// 	result += fmt.Sprintf("Splitting into %d subnets with prefix /%d:\n\n", numSubnets, newPrefix)

// 	// Calculate subnet size (in address count)
// 	subnetSize := 1 << (totalBits - newPrefix)

// 	// Start with the base network address
// 	currentIP := make(net.IP, len(ipNet.IP))
// 	copy(currentIP, ipNet.IP)

// 	// Iterate and print subnets
// 	for i := 0; i < numSubnets && i < 50; i++ {
// 		// Create a new subnet with the new prefix
// 		subnet := &net.IPNet{
// 			IP:   currentIP,
// 			Mask: net.CIDRMask(newPrefix, totalBits),
// 		}

// 		// Calculate first usable address
// 		firstUsable := make(net.IP, len(currentIP))
// 		copy(firstUsable, currentIP)
// 		if totalBits == 32 { // For IPv4
// 			firstUsable[3]++
// 		}

// 		// Calculate last usable address
// 		lastUsable := make(net.IP, len(currentIP))
// 		copy(lastUsable, currentIP)
// 		for j := range lastUsable {
// 			lastUsable[j] |= ^subnet.Mask[j]
// 		}
// 		if totalBits == 32 { // For IPv4
// 			lastUsable[3]--
// 		}

// 		result += fmt.Sprintf("Subnet %d: %s/%d\n", i+1, subnet.IP, newPrefix)
// 		result += fmt.Sprintf("  First usable: %s\n", firstUsable)
// 		result += fmt.Sprintf("  Last usable: %s\n", lastUsable)
// 		result += fmt.Sprintf("  # of hosts: %d\n\n", subnetSize-2)

// 		// Calculate the next subnet address by adding the subnet size to the current subnet
// 		incremented := false
// 		for j := len(currentIP) - 1; j >= 0 && !incremented; j-- {
// 			var add uint
// 			if j == len(currentIP)-1 {
// 				add = uint(subnetSize)
// 			}
// 			sum := uint(currentIP[j]) + add
// 			currentIP[j] = byte(sum & 0xFF)
// 			if sum>>8 == 0 {
// 				incremented = true
// 			}
// 		}
// 	}

// 	if numSubnets > 50 {
// 		result += fmt.Sprintf("... and %d more subnets (output limited to 50)\n", numSubnets-50)
// 	}

// 	return mcp.NewToolResultText(result), nil
// }

func getNetmask(ipNet *net.IPNet) (*mcp.CallToolResult, error) {
	mask := ipNet.Mask
	ones, bits := mask.Size()

	// Convert to binary string
	binStr := ""
	for range ones {
		binStr += "1"
	}
	for i := ones; i < bits; i++ {
		binStr += "0"
	}

	// Break into chunks for presentation
	var binChunks []string
	for i := 0; i < bits; i += 8 {
		end := min(i+8, len(binStr))
		binChunks = append(binChunks, binStr[i:end])
	}

	// Convert mask to decimal dotted format
	decimals := make([]string, len(mask))
	for i, b := range mask {
		decimals[i] = strconv.Itoa(int(b))
	}

	// Convert mask to hex format
	hexes := make([]string, len(mask))
	for i, b := range mask {
		hexes[i] = fmt.Sprintf("%02X", b)
	}

	result := fmt.Sprintf("Address: %s\n", ipNet.IP)
	result += fmt.Sprintf("Netmask: %s = %s\n", strings.Join(decimals, "."), strings.Join(binChunks, "."))
	result += fmt.Sprintf("Wildcard: %s\n", strings.Join(wildcardMask(mask), "."))
	result += fmt.Sprintf("Hex netmask: 0x%s\n", strings.Join(hexes, ""))
	result += fmt.Sprintf("Prefix: /%d\n", ones)

	return mcp.NewToolResultText(result), nil
}

func wildcardMask(mask net.IPMask) []string {
	wildcards := make([]string, len(mask))
	for i, b := range mask {
		wildcards[i] = strconv.Itoa(int(^b & 0xff))
	}
	return wildcards
}
