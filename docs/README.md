[![Go](https://github.com/sha1n/termite/actions/workflows/go.yml/badge.svg)](https://github.com/sha1n/termite/actions/workflows/go.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/sha1n/termite.svg)](https://pkg.go.dev/github.com/sha1n/termite)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/sha1n/termite)
[![Go Report Card](https://goreportcard.com/badge/github.com/sha1n/termite)](https://goreportcard.com/report/github.com/sha1n/termite)
[![Coverage Status](https://coveralls.io/repos/github/sha1n/termite/badge.svg?branch=master&service=github)](https://coveralls.io/github/sha1n/termite?branch=master)
[![Release](https://img.shields.io/github/release/sha1n/termite.svg?style=flat-square)](https://github.com/sha1n/termite/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go report card](https://github.com/sha1n/termite/actions/workflows/go-report-card.yml/badge.svg)](https://github.com/sha1n/termite/actions/workflows/go-report-card.yml)
[![Release Drafter](https://github.com/sha1n/termite/actions/workflows/release-drafter.yml/badge.svg)](https://github.com/sha1n/termite/actions/workflows/release-drafter.yml)


<img src="images/termite.png" width="96">

- [TERMite](#termite)
  - [Install](#install)
  - [Examples](#examples)
    - [Spinner](#spinner)
    - [Progress Bar](#progress-bar)
    - [Matrix](#matrix)
  - [Showcase](#showcase)

# TERMite
Termite is my playground for terminal app utilities and visual elements such as progress bars and indicators, cursor control and screen updates.

## Install
```bash
go get github.com/sha1n/termite
```

## Examples
### Spinner
```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

refreshInterval := time.Millisecond * 100
spinner := termite.NewSpinner(termite.StdoutWriter, "Processing...", refreshInterval, termite.DefaultSpinnerFormatter())

if err := spinner.Start(ctx); err == nil {
  doWork()
  
  _ = spinner.Stop(context.Background(), "Done!")
}

// Or using the fluent builder
builder := termite.NewSpinnerBuilder().
	WithTitle("Processing...").
	WithInterval(time.Millisecond * 100)

spinner := builder.Build()
_ = spinner.Start(ctx)
```

### Progress Bar
```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

termWidthFn := func() int { w, _, _ := termite.GetTerminalDimensions(); return w }
progressBar := termite.NewProgressBar(termite.StdoutWriter, tickCount, termWidthFn, width, termite.DefaultProgressBarFormatter())

if tick, err := progressBar.Start(ctx); err == nil {
  doWork(tick)
}
```

### Matrix
```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

refreshInterval := time.Millisecond * 100
matrix := termite.NewMatrix(termite.StdoutWriter, refreshInterval)
done := matrix.Start(ctx)

// Allocate rows for concurrent tasks
rows := matrix.NewRange(3)
for i, row := range rows {
  go func(idx int, r termite.MatrixRow) {
    r.Update(fmt.Sprintf("Task %d: Running...", idx+1))
    doWork()
    r.Update(fmt.Sprintf("Task %d: Done!", idx+1))
  }(i, row)
}

// Wait for completion
cancel()
<-done
```

## Showcase
The code for this demo can be found in [cmd/demo/main.go](https://github.com/sha1n/termite/blob/master/cmd/demo/main.go) (`go run -mod=readonly ./cmd/demo`). 

<img src="images/termite_demo_800.gif" width="100%">

