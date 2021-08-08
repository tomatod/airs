# Airs (AWS IP ranges search)
Airs is CLI tool for search AWS IP ranges. When you check error log, Haven't you ever thought "Which service does this IP address belong to?" ? Using this tool, You can find out where the IP address is in [ip-ranges.json](https://ip-ranges.amazonaws.com/ip-ranges.json), and it has other functions.

## Related page
- AWS IP address ranges: https://docs.aws.amazon.com/general/latest/gr/aws-ip-ranges.html
- ip-ranges.json: https://ip-ranges.amazonaws.com/ip-ranges.json

## Get started
```
go get github.com/tomatod/airs
```

## Usage
```
# find out the affiliation of an IP address
$ airs -ip xxx.xxx.xxx.xxx
{
  "syncToken": "1234567890",
  "createDate": "2021-01-02-03-04-05",
  "prefixes": [
    {
      "ip_prefix": "xxx.xxx.xxx.xxx/xx",
      "region": "ap-northeast-1",
      "service": "EC2",
      "network_border_group": "ap-northeast-1"
    }
  ],
  "ipv6_prefixes": []
}

# extract a specific region
$ airs -region ap-northeast-1

# extract a specific service
$ airs -service EC2

# list all regions 
$ airs -ls-regions

# list all services
$ airs -ls-services

# list all CIDRs
$ airs -ls-cidrs

# help command
$ airs -h
Usage of airs:
  -clean
        wheather download new ip-ranges.json.
  -compress
        wheather compress output JSON
  -ip string
        target IP for search
  -ls-cidrs
        list all of CIDR.
  -ls-regions
        list all of region.
  -ls-services
        list all of service ailias.
  -region string
        target region for search.
  -service string
        target service
```
