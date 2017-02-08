package main

import (
	"fmt"
	"os"
	"reflect"
	"strings"

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

	convertCom   = app.Command("convert", "Takes one or more CIDRs and gives the min and max IP address in the range(s)")
	convertRange = convertCom.Arg("ranges", "CIDR notation for ranges to convert").Required().Strings()

	subtractCom   = app.Command("subtract", "Subtract one CIDR from another and get the resulting range(s)")
	subMinuend    = subtractCom.Arg("minuend", "CIDR range from which to subtract").Required().String()
	subSubtrahend = subtractCom.Arg("subtrahend", "CIDR to subtract from the minuend").Required().String()

	optNewline = app.Flag("newline", "Print ranges as newline separated IPs").Short('n').Bool()
)

type netRange struct {
	Min net.IP
	Max net.IP
}

func netRangeFromInt(min, max int) netRange {
	return netRange{
		Min: toIP(min),
		Max: toIP(max),
	}
}

func (n netRange) String() string {
	separator := " - "
	if *optNewline {
		separator = "\n"
	}
	return fmt.Sprintf("%s%s%s\n", n.Min, separator, n.Max)
}

//Returns true if min is less than or equal to max
func (n netRange) Valid() bool {
	return toNumber(n.Min) <= toNumber(n.Max)
}

func main() {
	app.HelpFlag.Short('h')
	var message string
	var err error
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case "range":
		message, err = checkRange()
	case "cidr":
		message, err = checkCIDR()
	case "convert":
		message, err = convertCIDR()
	case "subtract":
		message, err = subtractCIDR()
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
	return (int(numericIP[0]) << 24) +
		(int(numericIP[1]) << 16) +
		(int(numericIP[2]) << 8) +
		int(numericIP[3])
}

func toIP(numIP int) net.IP {
	return net.IP{
		byte(numIP & 0xFF000000 >> 24),
		byte(numIP & 0x00FF0000 >> 16),
		byte(numIP & 0x0000FF00 >> 8),
		byte(numIP & 0x000000FF),
	}
}

func reverseMask(mask net.IPMask) net.IPMask {
	return net.IPMask{
		mask[0] ^ 0xFF,
		mask[1] ^ 0xFF,
		mask[2] ^ 0xFF,
		mask[3] ^ 0xFF,
	}
}

func getMaxCIDRIP(network *net.IPNet) net.IP {
	return toIP(toNumber(network.IP) + toNumber(net.IP(reverseMask(network.Mask))))
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

func convertCIDR() (string, error) {
	var retString string
	for _, r := range *convertRange {
		_, network, err := net.ParseCIDR(r)
		if err != nil {
			return "Could not parse CIDR range", err
		}
		retString = fmt.Sprintf("%s%s", retString, netRange{
			Min: network.IP,
			Max: getMaxCIDRIP(network),
		})
	}
	return strings.TrimRight(fmt.Sprintf("%s", retString), "\n"), nil
}

//Currently overengineered, but may want to support non-CIDR ranges later...
func subtractCIDR() (string, error) {
	_, minuend, err := net.ParseCIDR(*subMinuend)
	if err != nil {
		return "Could not parse minuend CIDR", err
	}
	_, subtrahend, err := net.ParseCIDR(*subSubtrahend)
	if err != nil {
		return "Could not parse subtrahend CIDR", err
	}

	var overrideOrigRange bool
	var difference []netRange
	//Just terrible variable names
	lowerMax := toNumber(subtrahend.IP)
	upperMin := toNumber(getMaxCIDRIP(subtrahend))
	minuendMin := toNumber(minuend.IP)
	minuendMax := toNumber(getMaxCIDRIP(minuend))

	//Get left of subtrahend
	if lowerMax < minuendMax {
		if lowerMax == 0 { //0.0.0.0 ... avoid wraparound
			overrideOrigRange = true
			goto skipLeft
		}
		leftRange := netRangeFromInt(minuendMin, lowerMax-1)
		if leftRange.Valid() {
			difference = append(difference, leftRange)
			overrideOrigRange = true
		}
	}
skipLeft:

	//Get right of subtrahend
	if upperMin > minuendMin {
		if upperMin == 4294967295 { //2^32 - 1, aka 255.255.255.255... avoid wraparound
			overrideOrigRange = true
			goto skipRight
		}
		rightRange := netRangeFromInt(upperMin+1, minuendMax)
		if rightRange.Valid() {
			difference = append(difference, rightRange)
			overrideOrigRange = true
		}
	}
skipRight:

	if !overrideOrigRange && !reflect.DeepEqual(minuend, subtrahend) {
		difference = append(difference, netRangeFromInt(minuendMin, minuendMax))
	}

	var returnString string
	//Create string from ranges
	for _, r := range difference {
		returnString = fmt.Sprintf("%s%s", returnString, r)
	}

	//Get right of subtrahend
	return strings.TrimRight(returnString, "\n"), nil
}
