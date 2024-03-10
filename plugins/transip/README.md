# TransIp

This plugin will update A and AAAA records by using there [api](https://api.transip.nl/) for the configured hosts.

## Config

```toml
[plugin.transip]

## Required for creating api keys, see 
## https://www.transip.nl/cp/account/api/
private_key="""
-----BEGIN PRIVATE KEY-----
......
-----END PRIVATE KEY-----"""
## paramters for authenitcation
# https://api.transip.nl/rest/docs.html#header-authentication
signature.login="username"
# signature.read_only=false
# signature.expiration=0
# signature.label=""
# signature.global_key=false

## list of hosts to monitor and update
hosts = ["example.com"]

## enable support ipv6 and AAAA records
# ipv6  = false
```
