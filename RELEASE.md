# Release Process

## Prerequisites

- Ensure all changes are committed
- Run tests locally: `go test ./...`

## Release Steps

```bash
./release.sh <version>
```

Example:
```bash
./release.sh 1.0.1
```

## After Release

1. Create a GitHub release from the tag
2. Build binaries for different platforms if needed
3. Update changelog if maintained
