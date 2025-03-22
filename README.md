# RDPROXY

This is a redis proxy to make ACL more convenient. It prefixes keys with the ACL username before sending it to upstream and undoing it when it about to send downstream. This makes the connection appear truly support redis for multi tenancy.

At the moment this proxy does.

Your app can connect to this instance listening by default at port `6479`. 

## Envar Options

| Env | Default |
|:--|:--|
|`LISTEN`|`:6479`|
|`UPSTREAM_REDIS`|`:6379`|

## TODO 

+ RESP3 protocol (aka `HELLO 3`)
+ Pub/Sub implementations
+ Use cluster for read operations
+ Unit tests

## Acknowledgements

Some parts of the code are inspired from these related projects

- [redis-namespace](https://github.com/resque/redis-namespace/)
- [redigo](https://github.com/gomodule/redigo)
- [redcon](github.com/tidwall/redcon)
