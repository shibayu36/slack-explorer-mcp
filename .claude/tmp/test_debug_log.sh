#!/bin/bash

# Test debug logging

echo "=== Testing with LOG_LEVEL=DEBUG ==="
LOG_LEVEL=DEBUG ./slack-explorer-mcp 2>&1 | head -20

echo ""
echo "=== Testing with LOG_LEVEL=INFO (default) ==="
LOG_LEVEL=INFO ./slack-explorer-mcp 2>&1 | head -20

echo ""
echo "=== Testing without LOG_LEVEL (default should be INFO) ==="
./slack-explorer-mcp 2>&1 | head -20