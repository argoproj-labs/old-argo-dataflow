# Stress

## Pprof

https://golang.org/pkg/net/http/pprof/

```
go tool pprof -web http://127.0.0.1:3569/debug/pprof/allocs
go tool pprof -web http://127.0.0.1:3569/debug/pprof/heap
go tool pprof -web http://127.0.0.1:3569/debug/pprof/profile\?seconds\=10
go tool trace <(curl -s http://127.0.0.1:3569/debug/pprof/trace\?seconds\=10)
```