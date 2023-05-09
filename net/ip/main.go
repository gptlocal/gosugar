package main

import (
	"encoding/hex"
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func main() {
	parseIPv4Header("450000540000400040010000c0a80001c0a800c7")
	parseIPv4Header("4500003c1c4640004006b1e6c0a80001c0a800c7")
}

func parseIPv4Header(ipHeaderHex string) error {
	ipHeaderBytes, _ := hex.DecodeString(ipHeaderHex)

	header, err := icmp.ParseIPv4Header(ipHeaderBytes)
	if err != nil {
		return err
	}
	fmt.Println(header)

	header2, err := ipv4.ParseHeader(ipHeaderBytes)
	if err != nil {
		return err
	}
	fmt.Println(header2)
	return nil
}
