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
dev

dev init
```

Todo
----

```
[x] - add flags
[x] - add include exclude files in watcher
[x] - add dev init with toml file like air cause it is hard to rembember the command
[ ] - executor should return the exit error so that we can log if the process finishes
[ ] - when refresh is about to happen, server is down we do polling. we could check the executor if it is active and only poll if so?
[ ] - refactor out server out of runtimepolling
[ ] - change to websocket so we can control restarts better
[ ] - add backoff and timeout when 
[ ] - add custom exec that user can provide in config
```
