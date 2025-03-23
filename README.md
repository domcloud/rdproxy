# RDPROXY

This is a redis proxy to make ACL more convenient. It prefixes keys with the ACL username before sending it to upstream and undoing it when it about to send downstream. This makes the connecting clients appear like the whole service is dedicated for that client while actually the redis server is shared or set for multi tenancy.

Your app can connect to this instance listening by default at port `6479`. 

If your framework doesn't have a way to log in via `AUTH username password`, you can log in using legacy password like `AUTH username:::password` (note the double colon). This nonstandard way of log in only works with this proxy and not with actual redis instance.

## What it does do

Let's assume this software runs on port 6479 while the upstream Redis is on port 6379.

When you call `GET foo:bar` to port 6479, without any `AUTH` command, it will send `GET default:foo:bar` to upstream Redis, since the current user is `default`. Let's notate this as `|GET foo:bar| > |GET default:foo:bar|`.

This is how it works when it executed serially:

```
|GET foo| > |GET default:foo|
|AUTH foo:::bar| > |AUTH foo bar|
|GET baz| > |GET foo:baz|
|AUTH user pass| > |AUTH user pass|
|GET foo| > |GET user:foo|
|SET foo bar| > |GET user:foo bar|
|KEYS foo:*| > |KEYS user:foo:*|
|SCAN 0| > |SCAN 0 MATCH user:*|
|EVAL "return redis.call('KEYS', KEYS[1])" 1 *| > |EVAL "return redis.call('KEYS', KEYS[1])" 1 user:*|
```

The return values of some commands like KEYS and SCAN will be "revived" (e.g. from `user:foo:bar` to `foo:bar`) so redis clients wouldn't need to adapt. 

Note that the revival values doesn't work for KEYS ran via EVAL, that means you should only access key names provided via `KEYS` otherwise your lua script won't work properly.

## Envar Options

| Env | Default |
|:--|:--|
|`LISTEN`|`:6479`|
|`UPSTREAM_REDIS`|`:6379`|

## TODO 

+ RESP3 protocol (aka `HELLO 3`)
+ Use cluster for read operations
+ Unit tests

## Acknowledgements

Some parts of the code are inspired from these related projects

- [redis-namespace](https://github.com/resque/redis-namespace/)
- [redigo](https://github.com/gomodule/redigo)
- [redcon](github.com/tidwall/redcon)
