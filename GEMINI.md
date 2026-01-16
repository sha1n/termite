# Termite - AI Agent Context

## Project Overview
Terminal app utilities library for Go - provides progress bars, spinners, cursor control, and multi-line matrix layouts.

## Key Components
| Component | File | Purpose |
|-----------|------|---------|
| **Matrix** | `matrix.go` | Multi-row terminal layout for concurrent tasks |
| **Spinner** | `spinner.go` | Animated progress indicator |
| **ProgressBar** | `progress_bar.go` | Horizontal progress bar |
| **Cursor** | `cursor.go` | Terminal cursor control |

## Build & Verify
```bash
make format lint test   # Run all checks
go run ./internal       # Run demo
```

## Testing Conventions
- All tests use testify/assert
- Use `bytes.Buffer` to capture terminal output
- Tests verify behavior, not internal state

## Code Style
- Interfaces defined before implementations
- Formatters allow customization of visual elements
- Context-based cancellation for all async operations
