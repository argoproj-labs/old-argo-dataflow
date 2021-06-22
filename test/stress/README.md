# Stress

## Pprof

https://golang.org/pkg/net/http/pprof/

```
go tool pprof -web http://localhost:3569/debug/pprof/allocs
go tool pprof -web http://localhost:3569/debug/pprof/profile
go tool pprof -web http://localhost:3569/debug/pprof/heap
go tool trace <(curl -s http://localhost:3569/debug/pprof/trace\?seconds\=5)
```