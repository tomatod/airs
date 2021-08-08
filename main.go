package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var (
	ip           *string
	region       *string
	service      *string
	clean        *bool
	compress     *bool
	listCidrs    *bool
	listRegions  *bool
	listServices *bool
)

type IPRanges struct {
	SyncToken    string        `json:"syncToken"`
	CreateDate   string        `json:"createDate"`
	Prefixes     []*Prefix     `json:"prefixes"`
	IPv6Prefixes []*IPv6Prefix `json:"ipv6_prefixes"`
}

type Prefix struct {
	IPPrefix           string `json:"ip_prefix"`
	Region             string `json:"region"`
	Service            string `json:"service"`
	NetworkBorderGroup string `json:"network_border_group"`
}

type IPv6Prefix struct {
	IPv6Prefix         string `json:"ipv6_prefix"`
	Region             string `json:"region"`
	Service            string `json:"service"`
	NetworkBorderGroup string `json:"network_border_group"`
}

type CIDRs struct {
	SyncToken  string   `json:"syncToken"`
	CreateDate string   `json:"createDate"`
	IPv4CIDRs  []string `json:"ipv4_cidrs"`
	IPv6CIDRs  []string `json:"ipv6_cidrs"`
}

type Regions struct {
	SyncToken  string   `json:"syncToken"`
	CreateDate string   `json:"createDate"`
	Regions    []string `json:"regions"`
}

type Services struct {
	SyncToken  string   `json:"syncToken"`
	CreateDate string   `json:"createDate"`
	Services   []string `json:"services"`
}

type Error struct {
	Phase   string `json:"phase"`
	Message string `json:"error_essage"`
}

const IP_RANGES_JSON_URL string = "https://ip-ranges.amazonaws.com/ip-ranges.json"

func init() {
	ip = flag.String("ip", "", "target IP for search")
	region = flag.String("region", "", "target region for search.")
	service = flag.String("service", "", "target service")
	clean = flag.Bool("clean", false, "wheather download new ip-ranges.json.")
	compress = flag.Bool("compress", false, "wheather compress output JSON")
	listCidrs = flag.Bool("ls-cidrs", false, "list all of CIDR.")
	listRegions = flag.Bool("ls-regions", false, "list all of region.")
	listServices = flag.Bool("ls-services", false, "list all of service ailias.")
	flag.Parse()
}

func main() {
	if err := downLoadJson(IP_RANGES_JSON_URL, getJsonName()); err != nil {
		fmt.Fprintln(os.Stderr, getErrorJson("download ip-ranges.json", err))
		os.Exit(1)
	}
	ipRanges, err := parseJson(getJsonName())
	if err != nil {
		fmt.Fprintln(os.Stderr, getErrorJson("unmarshal ip-ranges.json to Golang struct 'IPRanges'", err))
		os.Exit(1)
	}

	var result interface{}
	if *listCidrs {
		result, err = listCIDRDo(ipRanges)
	}
	if *listRegions {
		result, err = listRegionDo(ipRanges)
	}
	if *listServices {
		result, err = listServiceDo(ipRanges)
	}
	if !*listCidrs && !*listRegions && !*listServices {
		result, err = searchInfo(ipRanges)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, getErrorJson("comand run", err))
		os.Exit(1)
	}
	printResult(result)
}

func getJsonName() string {
	return filepath.Join(os.TempDir(), "ip-ranges-"+time.Now().Format("20060102")+".json")
}

func downLoadJson(url string, fileName string) error {
	if !*clean && existFile(fileName) {
		return nil
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func existFile(fileName string) bool {
	_, err := os.Stat(fileName)
	if err != nil {
		return false
	}
	return true
}

func parseJson(fileName string) (*IPRanges, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	ipRanges := IPRanges{}
	err = json.Unmarshal(data, &ipRanges)
	return &ipRanges, err
}

func getErrorJson(phase string, err error) string {
	if err == nil {
		return ""
	}
	e := Error{
		Phase:   phase,
		Message: err.Error(),
	}
	chars, _ := json.Marshal(e)
	return string(chars)
}

func printResult(result interface{}) {
	chars, err := json.Marshal(result)
	if err != nil {
		fmt.Fprintln(os.Stderr, getErrorJson("marshal Golang struct to JSON", err))
		os.Exit(1)
	}
	str := string(chars)
	if str == "null" {
		os.Exit(0)
	}

	if *compress {
		fmt.Println(str)
		os.Exit(0)
	}

	var buf bytes.Buffer
	err = json.Indent(&buf, chars, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, getErrorJson("indent JSON", err))
		os.Exit(1)
	}
	fmt.Println(buf.String())
	os.Exit(0)
}

func listCIDRDo(ipRanges *IPRanges) (*CIDRs, error) {
	result := CIDRs{
		SyncToken:  ipRanges.SyncToken,
		CreateDate: ipRanges.CreateDate,
		IPv4CIDRs:  []string{},
		IPv6CIDRs:  []string{},
	}
	for _, prefix := range ipRanges.Prefixes {
		result.IPv4CIDRs = append(result.IPv4CIDRs, prefix.IPPrefix)
	}
	for _, prefix := range ipRanges.IPv6Prefixes {
		result.IPv6CIDRs = append(result.IPv6CIDRs, prefix.IPv6Prefix)
	}
	return &result, nil
}

func listRegionDo(ipRanges *IPRanges) (*Regions, error) {
	result := Regions{
		SyncToken:  ipRanges.SyncToken,
		CreateDate: ipRanges.CreateDate,
		Regions:    []string{},
	}

	m := map[string]bool{}
	for _, prefix := range ipRanges.Prefixes {
		if _, ok := m[prefix.Region]; !ok {
			m[prefix.Region] = true
			result.Regions = append(result.Regions, prefix.Region)
		}
	}
	for _, prefix := range ipRanges.IPv6Prefixes {
		if _, ok := m[prefix.Region]; !ok {
			m[prefix.Region] = true
			result.Regions = append(result.Regions, prefix.Region)
		}
	}

	return &result, nil
}

func listServiceDo(ipRanges *IPRanges) (*Services, error) {
	result := Services{
		SyncToken:  ipRanges.SyncToken,
		CreateDate: ipRanges.CreateDate,
		Services:   []string{},
	}

	m := map[string]bool{}
	for _, prefix := range ipRanges.Prefixes {
		if _, ok := m[prefix.Service]; !ok {
			m[prefix.Service] = true
			result.Services = append(result.Services, prefix.Service)
		}
	}
	for _, prefix := range ipRanges.IPv6Prefixes {
		if _, ok := m[prefix.Service]; !ok {
			m[prefix.Service] = true
			result.Services = append(result.Services, prefix.Service)
		}
	}

	return &result, nil
}

func searchInfo(ipRanges *IPRanges) (*IPRanges, error) {
	ipStruct := net.ParseIP(*ip)
	if *ip != "" && ipStruct == nil {
		return nil, errors.New("specified IP address is invalied.")
	}

	result := IPRanges{}
	result.SyncToken = ipRanges.SyncToken
	result.CreateDate = ipRanges.CreateDate
	result.Prefixes = []*Prefix{}
	result.IPv6Prefixes = []*IPv6Prefix{}

	for _, prefix := range ipRanges.Prefixes {
		_, ipNet, err := net.ParseCIDR(prefix.IPPrefix)
		if err != nil {
			return nil, err
		}
		if *region != "" && prefix.Region != *region {
			continue
		}
		if *service != "" && prefix.Service != *service {
			continue
		}
		if *ip != "" {
			if ipNet.Contains(ipStruct) {
				result.Prefixes = append(result.Prefixes, prefix)
				return &result, nil
			}
			continue
		}
		result.Prefixes = append(result.Prefixes, prefix)
	}

	for _, prefix := range ipRanges.IPv6Prefixes {
		_, ipNet, err := net.ParseCIDR(prefix.IPv6Prefix)
		if err != nil {
			return nil, err
		}
		if *region != "" && prefix.Region != *region {
			continue
		}
		if *service != "" && prefix.Service != *service {
			continue
		}
		if *ip != "" {
			if ipNet.Contains(ipStruct) {
				result.IPv6Prefixes = append(result.IPv6Prefixes, prefix)
				return &result, nil
			}
			continue
		}
		result.IPv6Prefixes = append(result.IPv6Prefixes, prefix)
	}

	return &result, nil
}
