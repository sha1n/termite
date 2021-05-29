[![Go](https://github.com/sha1n/termite/actions/workflows/go.yml/badge.svg)](https://github.com/sha1n/termite/actions/workflows/go.yml)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/sha1n/termite)
[![Go Report Card](https://goreportcard.com/badge/github.com/sha1n/termite)](https://goreportcard.com/report/github.com/sha1n/termite)
[![Coverage Status](https://coveralls.io/repos/github/sha1n/termite/badge.svg?branch=master&service=github)](https://coveralls.io/github/sha1n/termite?branch=master)
[![Release](https://img.shields.io/github/release/sha1n/termite.svg?style=flat-square)](https://github.com/sha1n/termite/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go report card](https://github.com/sha1n/termite/actions/workflows/go-report-card.yml/badge.svg)](https://github.com/sha1n/termite/actions/workflows/go-report-card.yml)
[![Release Drafter](https://github.com/sha1n/termite/actions/workflows/release-drafter.yml/badge.svg)](https://github.com/sha1n/termite/actions/workflows/release-drafter.yml)


<img src="docs/images/termite.png" width="96">

- [TERMite](#termite)
  - [Install](#install)
  - [Code Examples](#code-examples)
    - [Spinner Progress Example](#spinner-progress-example)
    - [Progress Bar Example](#progress-bar-example)
  - [Showcase](#showcase)

# TERMite
Termite is my playground for terminal app utilities and visual elements such as progress bars and indicators, cursor control and screen updates.

## Install
```bash
go get github.com/sha1n/termite
```

## Code Examples
### Spinner Progress Example
```go
terminal := termite.NewTerminal(true)
refreshInterval := time.Millisecond * 100
spinner := termite.NewSpinner(terminal, "Processing...", refreshInterval)

if _, e := spinner.Start(); e == nil {
  doWork()
  
  _ = spinner.Stop("Done!")
}

```

### Progress Bar Example
```go
terminal := termite.NewTerminal(true)
progressBar := termite.NewDefaultProgressBar(terminal, workItems)

if tick, cancel, err := progressBar.Start(); err == nil {
  defer cancel()
  
  doWork(tick)
}


```

## Showcase
The code for this demo can be found in [internal/main.go](internal/main.go) (`go run internal/main.go`). 

<img src="docs/images/termite_demo_800.gif" width="100%">
