Dev
===

Dev helps ya do go development stuff

Install
-------
```
go install github.com/rdkal/dev
```

Usage 
-----
```
dev go run cmd/main.go
```

Todo
----

```
[x] - add flags
[x] - add include exclude files in watcher
[ ] - executor should return the exit error so that we can log if the process finishes
[ ] - when refresh is about to happen, server is down we do polling. we could check the executor if it is active and only poll if so?
[ ] - refactor out server out of runtimepolling
[ ] - change to websocket so we can control restarts better
[ ] - add backoff and timeout when 
```
