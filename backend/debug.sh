curl -H "X-Debug-Token: secret-token" http://localhost:6060/debug/pprof/heap > heap.out
go tool pprof heap.out

curl -H "X-Debug-Token: secret-token" http://localhost:6060/debug/pprof/profile > profile.out
go tool pprof profile.out

go tool pprof http://localhost:6060/debug/pprof/profile?token=secret-token

