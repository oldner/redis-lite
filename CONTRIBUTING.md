# Contributing to Redis-Lite

First off, thanks for taking the time to contribute! 🎉

We welcome pull requests for bug fixes, new commands (e.g., Lists, Sets), or performance improvements.

## How to Contribute

1. Fork the repository on GitHub.
2. Clone your fork locally.
3. Create a branch for your feature:
```bash
git checkout -b feature/amazing-command
```
4. Make your changes.
5. Run Tests: Ensure you haven't broken anything.
```bash
go test -v -race ./...
```
6. Commit & Push:
```
git commit -m "Add implementation for LPUSH command"
git push origin feature/amazing-command
```
7. Open a Pull Request against the main branch.

## Code Style

- We follow standard Go conventions.
- Please run `go fmt ./...` before committing.
- Ensure all public functions have comments explaining their purpose.

Reporting Bugs

If you find a bug, please open an issue using the Bug Report template. Include the steps to reproduce the issue.
