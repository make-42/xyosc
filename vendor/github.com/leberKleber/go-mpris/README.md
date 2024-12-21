# go-mpris

[![Go](https://github.com/leberKleber/go-mpris/workflows/go/badge.svg)](https://github.com/leberKleber/go-mpris/actions?query=workflow%3Ago)
[![GoDoc](https://godoc.org/github.com/leberKleber/go-mpris?status.png)](https://godoc.org/github.com/leberKleber/go-mpris)
[![Go Report Card](https://goreportcard.com/badge/github.com/leberKleber/go-mpris)](https://goreportcard.com/report/github.com/leberKleber/go-mpris)
[![codecov](https://codecov.io/gh/leberKleber/go-mpris/branch/main/graph/badge.svg)](https://codecov.io/gh/leberKleber/go-mpris)

go-mpris is an implementation of the mpris dbus interface written in go (golang).
Implemented and tested against version 2.2. See: https://specifications.freedesktop.org/mpris-spec/2.2.

## Example

Example cli has been implemented.

```shell
git clone git@github.com:leberKleber/go-mpris.git

go build examples/cli.go

./cli-client
```

## Features:

### Player

https://specifications.freedesktop.org/mpris-spec/2.2/Player_Interface.html

#### Methods

| method      | library path                                                            | implemented        |
|-------------|-------------------------------------------------------------------------|--------------------|
| Next        | `mpris.Player.Next()`                                                   | :heavy_check_mark: |
| Previous    | `mpris.Player.Previous()`                                               | :heavy_check_mark: |
| Pause       | `mpris.Player.Pause()`                                                  | :heavy_check_mark: |
| PlayPause   | `mpris.Player.PlayPause()`                                              | :heavy_check_mark: |
| Stop        | `mpris.Player.Stop()`                                                   | :heavy_check_mark: |
| Seek        | `mpris.Player.SeekTo(<offset> int64)`¹                                  | :heavy_check_mark: |
| SetPosition | `mpris.Player.SetPosition(<trackID> dbus.ObjectPath, <position> int64)` | :heavy_check_mark: |
| OpenUri     | `mpris.Player.OpenUri(<uri> string)`                                    | :heavy_check_mark: |

¹ Could not be named Seek, it's a reserved function name.

#### Properties

| property       | library path                                                            | implemented        |
|----------------|-------------------------------------------------------------------------|--------------------|
| PlaybackStatus | `mpris.Player.PlaybackStatus() (mpris.PlaybackStatus, error)`           | :heavy_check_mark: |
| LoopStatus     | `mpris.Player.LoopStatus() (mpris.LoopStatus, error)`                   | :heavy_check_mark: |
| LoopStatus     | `mpris.Player.SetLoopStatus(<loopStatus> mpris.LoopStatus) error`       | :heavy_check_mark: |
| Rate           | `mpris.Player.Rate() (float64, error)`                                  | :heavy_check_mark: |
| Rate           | `mpris.Player.SetRate(<rate> float64) error`                            | :heavy_check_mark: |
| Shuffle        | `mpris.Player.Shuffle() (bool, error)`                                  | :heavy_check_mark: |
| Shuffle        | `mpris.Player.SetShuffle(<shuffle> bool) error`                         | :heavy_check_mark: |
| Metadata       | `mpris.Player.Metadata() (mpris.Metadata, error)`                       | :heavy_check_mark: |
| Volume         | `mpris.Player.Volume() (float64, error)`                                | :heavy_check_mark: |
| Volume         | `mpris.Player.SetVolume(<volume> float64) (error)`                      | :heavy_check_mark: |
| Position       | `mpris.Player.Position() (int64, error)`                                | :heavy_check_mark: |
| Position       | `mpris.Player.SetPosition(<trackID> dbus.ObjectPath, <position> int64)` | :heavy_check_mark: |
| MinimumRate    | `mpris.Player.MinimumRate() (float64, error)`                           | :heavy_check_mark: |
| MaximumRate    | `mpris.Player.MaximumRate() (float64, error)`                           | :heavy_check_mark: |
| CanGoNext      | `mpris.Player.CanGoNext() (bool, error)`                                | :heavy_check_mark: |
| CanGoPrevious  | `mpris.Player.CanGoPrevious() (bool, error)`                            | :heavy_check_mark: |
| CanPlay        | `mpris.Player.CanPlay() (bool, error)`                                  | :heavy_check_mark: |
| CanPause       | `mpris.Player.CanPause() (bool, error)`                                 | :heavy_check_mark: |
| CanSeek        | `mpris.Player.CanSeek() (bool, error)`                                  | :heavy_check_mark: |
| CanControl     | `mpris.Player.CanControl(bool, error)`                                  | :heavy_check_mark: |

#### Signals

| signal | library path                                                      | implemented        |
|--------|-------------------------------------------------------------------|--------------------|
| Seeked | `mpris.Player.Seeked(<ctx> context.Context) (<-chan int, error) ` | :heavy_check_mark: |

## Development

### Versioning

This library follows the semantic versioning concept.

### Commits

Commits should follow the conventional commit rules.  
See: https://conventionalcommits.org.

### Mocks

Mocks will be generated with `github.com/matryer/moq`. It can be installed with
`go install github.com/matryer/moq@latest`. Generation can be triggered with `go generate ./...`.

### Go Docs

Read the docs at https://pkg.go.dev/github.com/leberKleber/go-mpris