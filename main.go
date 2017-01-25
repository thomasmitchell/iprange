package main

import (
	"fmt"
	"os"

	"github.com/starkandwayne/goutils/ansi"

	"net"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New("iprange", "Check if an IP is in range")

	rangeCom    = app.Command("range", "Check if IP is between two other ips")
	rangeTarget = rangeCom.Arg("target", "IP to check if is within range").Required().IP()
	minRange    = rangeCom.Arg("minrange", "Lowest possible IP to be considered in range").Required().IP()
	maxRange    = rangeCom.Arg("maxrange", "Highest possible IP to be considered in range").Required().IP()

	cidrCom    = app.Command("cidr", "Check if an ip is in a CIDR range")
	cidrTarget = cidrCom.Arg("target", "IP to check if is within range").Required().IP()
	cidrRange  = cidrCom.Arg("range", "CIDR notation for range to check").Required().String()
)

func main() {
	app.HelpFlag.Short('h')
	var message string
	var err error
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case "range":
		message, err = checkRange()
	case "cidr":
		message, err = checkCIDR()
	}

	if err != nil {
		ansi.Fprintf(os.Stderr, "@R{%s}\n", message)
		os.Exit(1)
	}
	ansi.Fprintf(os.Stderr, "@G{%s}\n", message)
}

func checkRange() (string, error) {
	targetValue := toNumber(*rangeTarget)
	if targetValue < toNumber(*minRange) || targetValue > toNumber(*maxRange) {
		return "IP not in range", fmt.Errorf("")
	}
	return "IP in range", nil
}

func toNumber(ip net.IP) int {
	numericIP := ip.To4()
	return (int(numericIP[0]) << 24) + (int(numericIP[1]) << 16) + (int(numericIP[2]) << 8) + int(numericIP[3])
}

func checkCIDR() (string, error) {
	_, network, err := net.ParseCIDR(*cidrRange)
	if err != nil {
		return "Could not parse CIDR range", err
	}
	if !network.Contains(*cidrTarget) {
		return "IP not in range", fmt.Errorf("")
	}
	return "IP in range", nil
}
