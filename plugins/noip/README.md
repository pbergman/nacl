# NoIP

update no-ip entries when ip changes by using there [ddns api](https://www.noip.com/integrate/request)

## Config

```toml
[plugin.noip]

## Support for multiple groups, each   
## group should be a separated table
## https://my.noip.com/dynamic-dns/groups
[[plugin.noip.group]]
hostname="..."
username="..."
password="..."
ddns_host="dynupdate.no-ip.com"
##  support/monitor ipv6 records
# ipv6=false
```
