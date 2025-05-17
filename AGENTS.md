# gping Agent Documentation (agents.md)

This document is designed to help language models (LLMs) and automation agents understand the structure, purpose, and operation of the `gping` project.

## 🧭 Overview

`gping` is a terminal-based network monitoring tool written in Go that allows real-time pinging of multiple hosts with a rich terminal display. It shows per-host:

- IP resolution
- Ping status (responding or timeout)
- Round-trip time (RTT)
- A rolling success/failure history
- A visual RTT sparkline graph using Unicode block characters

It's a modern, extensible re-imagining of the classic `ping` utility for multiple hosts, built for observability and scriptability.

---

## 📂 Project Structure

```
.
├── README.md            # Project intro and usage
├── cmd/
│   ├── main.go          # Entry point: all app logic and CLI flags
│   └── main_test.go     # Unit tests for pattern expansion logic
├── go.mod               # Go module file (dependency tracking)
└── go.sum               # Module hash file
```

---

## ⚙️ Main Features

### ✨ Real-Time Dashboard

- Clears and redraws screen on every cycle.
- Displays all hosts in a stable row order.

### 🧠 Synthetic Host Expansion

Supports synthetic host patterns:

- Ranges: `web-[001:005]` → `web-001` to `web-005`
- Selections: `web-[us1,us2]` → `web-us1`, `web-us2`
- Multi-dimensions: `web-[us1,us2]-[001:003]`

### 📊 Terminal Output Columns

```
Host               IP                 Status       RTT        History    RTT Graph
------------------------------------------------------------------------------------------
google.com         142.250.191.174    Responding   27.605ms   ++         ▇█
microsoft.com      13.107.246.38      Timeout      --------   --         ----------
```

- **Host**: Original input or synthetic-expanded name
- **IP**: Resolved IPv4 address
- **Status**: Responding, Timeout, or Resolve Error
- **RTT**: Ping latency (or '--------' on timeout)
- **History**: Last 10 pings, color-coded (`+` for success, `-` for timeout)
- **RTT Graph**: Mini inline graph showing recent RTT trend

### 📌 Additional Features

- `-s` flag supports multiple synthetic patterns
- `-i` flag controls ping interval in seconds
- `-timeout` flag (in milliseconds) sets MaxRTT per ping
- Graceful exit via CTRL+C with summary
- Unit tested host expansion logic

---

## 🧪 Agent Testing Guidance

Agents can:

- Generate synthetic host patterns with brackets and commas
- Parse history/sparkline data to infer host stability
- Use `main.go` to modify ping strategy or output style
- Extend `main_test.go` with new synthetic expansion patterns for fuzzing

---

## 🔧 Ideas for Extension

Agents may wish to:

- Add `-json` or `-csv` output mode
- Persist historical ping data to disk
- Export metrics via Prometheus
- Add alert thresholds for high latency or failure
- Introduce a Web UI or TUI scrollable view

---

## 🤖 Agent Summary

`gping` is an interactive, headless-friendly network probe app that simulates `ping` across many hosts at once. It is suitable for:

- Real-time dashboards
- SRE observability pipelines
- Remote diagnostics
- Interactive demos or monitoring UIs

It is highly readable, extensible, and idiomatic Go — well-suited for LLM and automation agent integration or learning use.


