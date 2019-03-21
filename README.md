# brightness

Control brightness.

## Usage

Display help

```sh
brightness -help
```

Get value of current brightness

```sh
brightness -get
```

Set to max

```sh
brightness -set max
```

Increment 10%

```sh
brightness -inc
```

## Requirements

- Linux
  - Permission of read `/sys/class/backlight/*/max_brightness`
  - Permission of read write `/sys/class/backlight/*/brightness`

## Supported environments

- Arch Linux

## Installation

```sh
go get github.com/yaeshimo/brightness/cmd/brightness
```

## License

MIT
