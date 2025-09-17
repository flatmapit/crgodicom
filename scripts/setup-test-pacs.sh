#!/bin/bash

# Setup script for CRGoDICOM test PACS servers
# This script sets up two Orthanc PACS servers for testing

set -e

echo "Setting up CRGoDICOM test PACS servers..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "Error: Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    echo "Error: docker-compose is not installed. Please install docker-compose and try again."
    exit 1
fi

# Stop any existing containers
echo "Stopping any existing test PACS containers..."
docker-compose -f docker-compose.test.yml down

# Start the test PACS servers
echo "Starting test PACS servers..."
docker-compose -f docker-compose.test.yml up -d

# Wait for services to be ready
echo "Waiting for PACS servers to be ready..."
sleep 10

# Check if services are running
echo "Checking PACS server status..."
docker-compose -f docker-compose.test.yml ps

# Test connectivity
echo "Testing connectivity to PACS servers..."
echo "Orthanc1 (port 4900):"
if curl -s http://localhost:4900/system > /dev/null; then
    echo "  ✓ Orthanc1 is responding"
else
    echo "  ✗ Orthanc1 is not responding"
fi

echo "Orthanc2 (port 4901):"
if curl -s http://localhost:4901/system > /dev/null; then
    echo "  ✓ Orthanc2 is responding"
else
    echo "  ✗ Orthanc2 is not responding"
fi

echo ""
echo "Test PACS servers setup complete!"
echo ""
echo "PACS Server Details:"
echo "  Orthanc1: http://localhost:4900 (AET: ORTHANC1)"
echo "  Orthanc2: http://localhost:4901 (AET: ORTHANC2)"
echo ""
echo "To stop the servers: docker-compose -f docker-compose.test.yml down"
echo "To view logs: docker-compose -f docker-compose.test.yml logs"
