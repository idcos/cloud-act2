# GoSlaves

GoSlaves is a simple golang's library which can handle wide list of tasks asynchronously.

[![GoDoc](https://godoc.org/github.com/themester/GoSlaves?status.svg)](https://godoc.org/github.com/themester/GoSlaves)
[![Go Report Card](https://goreportcard.com/badge/github.com/themester/goslaves)](https://goreportcard.com/report/github.com/themester/goslaves)

![alt text](https://raw.githubusercontent.com/themester/GoSlaves/master/logo.png)

Installation
------------

```
$ go get -u -v -x github.com/themester/GoSlaves
```

Benchmark
---------

After a lot of benchmarks and the following enhancings of the package I got this results:

```
$ go test -bench=. -benchmem -benchtime=10s
goos: linux
goarch: amd64
BenchmarkGrPool-2      	10000000	      1183 ns/op	      40 B/op	       1 allocs/op
BenchmarkTunny-2       	10000000	      2538 ns/op	      32 B/op	       2 allocs/op
BenchmarkSlavePool-2   	20000000	      1107 ns/op	      16 B/op	       1 allocs/op
```

GrPool is a little faster than SlavePool. This is caused because GoSlaves resize the stack adapting it to work demand.

Example
-------
```go
func main() {
  ch := make(chan int, 20)
  cs := make(chan struct{})
  sp := &slaves.SlavePool{
    Work: func(obj interface{}) {
      ch <- obj.(int)
    },
  }
  sp.Open()
  defer sp.Close()

  go func() {
    p := 0
    for range ch {
      p++
    }
    if p == 20 {
      cs <- struct{}{}
    } else {
      panic(
        fmt.Sprintf("Bad test: %s", p),
      )
    }
  }()

  for i := 0; i < 20; i++ {
    sp.Serve(i)
  }
  time.Sleep(time.Second)
  close(ch)

  select {
  case <-cs:
  case <-time.After(time.Second * 2):
    t.Fatal("timeout")
  }

```
