#!/bin/bash
# Quick start script for xDS example

set -e

echo "🚀 Starting xDS Example with Istio Pilot"
echo "========================================"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Error: Docker is not running"
    exit 1
fi

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Error: docker-compose is not installed"
    exit 1
fi

echo "✅ Docker is running"

# Start services
echo ""
echo "📦 Starting Istio Pilot and example services..."
docker-compose up -d istio-pilot

echo "⏳ Waiting for Istio Pilot to be ready..."
sleep 10

# Check Istio Pilot health
if curl -f http://localhost:15010/ready > /dev/null 2>&1; then
    echo "✅ Istio Pilot is ready"
else
    echo "⚠️  Warning: Istio Pilot may not be fully ready"
fi

# Start example services
echo ""
echo "📦 Starting example services..."
docker-compose up -d example-service-1 example-service-2

echo "⏳ Waiting for services to register..."
sleep 5

# Run client
echo ""
echo "🔍 Running example client..."
docker-compose up example-client

echo ""
echo "✅ Example complete!"
echo ""
echo "To view logs:"
echo "  docker-compose logs -f istio-pilot"
echo "  docker-compose logs -f example-service-1"
echo "  docker-compose logs -f example-client"
echo ""
echo "To stop all services:"
echo "  docker-compose down"
