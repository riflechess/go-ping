# go-ping

A lightweight, real-time ICMP ping dashboard written in Go. It allows you to continuously monitor the reachability and round-trip time (RTT) of one or more hosts, with a colorful ANSI-based terminal UI and synthetic host pattern expansion.

## Features

* **Real-time ANSI dashboard**: Continuously updates ping results in-place, with color-coded statuses and sparklines for RTT.
* **Synthetic host patterns**: Expand patterns like `host-[001:003]` or `app-[us,eu]-[01:02]` into multiple hosts automatically.
* **Configurable interval & timeout**: Control ping frequency and per-ping timeout.
* **Flat host list output**: With `-printhosts`, print just the expanded hosts and exit (no pinging/UI).
* **Graceful shutdown**: Handles `SIGINT`/`SIGTERM` to exit cleanly.

## Installation

Make sure you have Go 1.23 or later installed.

```bash
# Clone the repository
git clone https://github.com/riflechess/go-ping.git
cd go-ping

# Build the binary
go build -o go-ping ./cmd

# (Optional) Install to $GOPATH/bin
go install ./cmd
```

## Usage

```bash
go-ping [flags] [host ...]
# or
go run ./cmd -s "mysite-[dca,dcb,dcc]-[01:02]" example.com
```

If you haven’t built the binary yet, you can use `go run`:

```bash
go run ./cmd -timeout 2000 -i 2 -s "app-[us,eu]-[01:02]"
```

### Flags

| Flag          | Type   | Default | Description                                 |
| ------------- | ------ | ------- | ------------------------------------------- |
| `-s`          | string | —       | Synthetic host pattern. Repeatable.         |
| `-i`          | int    | 1       | Interval between ping rounds (in seconds).  |
| `-timeout`    | int    | 5000    | Per-ping timeout in milliseconds.           |
| `-printhosts` | bool   | false   | Print the expanded host list only and exit. |
| `-help`       | —      | —       | Show usage information.                     |

#### Synthetic Patterns

Patterns allow you to define ranges or comma-delimited lists inside brackets:

* Numeric ranges: `[01:03]` → `01`, `02`, `03`
* Lists: `[us,eu,ap]` → `us`, `eu`, `ap`
* Combined: `site-[01:02]-[us,eu]` → `site-01-us`, `site-01-eu`, `site-02-us`, `site-02-eu`

### Examples

* Ping two static hosts every second (default):

  ```bash
  go-ping example.com 8.8.8.8
  ```

* Ping a synthetic pattern every 2 seconds with 1s timeout:

  ```bash
  go-ping -s "mysite-[dca,dcb]-[01:03]" -i 2 -timeout 1000
  ```

* Flat print of hosts without pinging:

  ```bash
  go-ping -s "mysite-[dca,dcb]-[01:02]" -printhosts
  # Output:
  # mysite-dca-01
  # mysite-dca-02
  # mysite-dcb-01
  # mysite-dcb-02
  ```

## Testing

Unit tests cover synthetic pattern expansion. Run them with:

```bash
go test ./cmd
```

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.

* Please run `go fmt` and `go vet` before committing.
* Add tests for new features.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
