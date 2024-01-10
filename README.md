Dev
===

Dev helps ya do go development stuff

Install
-------
```
$ go install github.com/rdkal/dev
```

Usage 
-----
```
$ dev

```


Todo
----

[ ] - add include exclude files in watcher
[ ] - add flags
[ ] - executor should return the exit error so that we can log if the process finishes
[ ] - when refresh is about to happen, server is down, we retry alot. we could check the executor if it is active
[ ] - refactor out server out of runtime
[ ] - change to websocket so we can control restarts better
