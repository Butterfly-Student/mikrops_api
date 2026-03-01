#!/bin/bash

# Script untuk menjalankan MikroTik integration test
# Usage: ./test-mikrotik.sh [ip] [username] [password]

set -e

# Default values
MIKROTIK_IP="${1:-192.168.233.1}"
MIKROTIK_USER="${2:-admin}"
MIKROTIK_PASS="${3:-r00t}"

echo "=========================================="
echo "MikroTik Integration Test"
echo "=========================================="
echo ""
echo "Target MikroTik: $MIKROTIK_IP:8728"
echo "Username: $MIKROTIK_USER"
echo ""

# Check if running from correct directory
if [ ! -f "go.mod" ]; then
    echo "Error: Must run from project root directory"
    exit 1
fi

# Export environment variables
export MIKROTIK_TEST_IP=$MIKROTIK_IP
export MIKROTIK_TEST_USER=$MIKROTIK_USER
export MIKROTIK_TEST_PASS=$MIKROTIK_PASS

# Check prerequisites
echo "Checking prerequisites..."

# Check PostgreSQL
if ! docker ps | grep -q postgres; then
    echo "Starting PostgreSQL container..."
    docker run -d --name postgres-test \
        -e POSTGRES_PASSWORD=postgres \
        -e POSTGRES_DB=test_db \
        -p 5432:5432 \
        postgres:15-alpine 2>/dev/null || true
    sleep 5
fi

# Check RabbitMQ
if ! docker ps | grep -q rabbitmq; then
    echo "Starting RabbitMQ container..."
    docker run -d --name rabbitmq-test \
        -p 5672:5672 \
        -p 15672:15672 \
        rabbitmq:3-alpine 2>/dev/null || true
    sleep 5
fi

echo "✓ Prerequisites check complete"
echo ""

# Run tests
echo "Running integration tests..."
echo ""

cd tests/integration

go test -v -timeout 120s ./... -run TestMikrotikSyncIntegration 2>&1 | tee test-results.log

TEST_EXIT_CODE=${PIPESTATUS[0]}

echo ""
echo "=========================================="
if [ $TEST_EXIT_CODE -eq 0 ]; then
    echo "✓ All tests PASSED"
else
    echo "✗ Some tests FAILED"
    echo "Check test-results.log for details"
fi
echo "=========================================="

exit $TEST_EXIT_CODE
