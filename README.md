# brightness

Provides brightness control.

akari is tool for control brightness from cli.

## Usage

Display help

```sh
akari -help
```

Get value of current brightness

```sh
akari -get
```

Set to max

```sh
akari -set max
```

Increment 10%

```sh
akari -inc
```

## Available

- Arch Linux

## Requirements

- Linux
  - Permission of read `/sys/class/backlight/*/max_brightness`
  - Permission of read write `/sys/class/backlight/*/brightness`

## Installation

```sh
go get github.com/yaeshimo/brightness/cmd/akari
```

## License

MIT
