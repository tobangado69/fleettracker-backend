#!/bin/bash
# Quick fix script for Swagger blank page issue

echo "🔧 Fixing Swagger Blank Page Issue..."
echo ""
echo "This will:"
echo "1. Ensure Swagger docs are generated"
echo "2. Rebuild backend Docker container"
echo "3. Restart backend with Swagger included"
echo ""
read -p "Continue? (y/n) " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ Cancelled"
    exit 1
fi

echo ""
echo "📚 Step 1/4: Checking Swagger docs..."
if [ -f "docs/docs.go" ]; then
    echo "✅ Swagger docs found"
else
    echo "⚠️  Swagger docs not found, generating..."
    make swagger
fi

echo ""
echo "🐳 Step 2/4: Stopping old backend container..."
docker-compose stop backend

echo ""
echo "🔨 Step 3/4: Rebuilding backend container (this may take 1-2 minutes)..."
docker-compose build --no-cache backend

echo ""
echo "🚀 Step 4/4: Starting backend container..."
docker-compose up -d backend

echo ""
echo "⏳ Waiting for backend to be healthy (15 seconds)..."
sleep 15

echo ""
echo "✅ Testing Swagger..."
echo ""

# Test health endpoint
echo "Testing health endpoint..."
HEALTH=$(curl -s http://localhost:8080/health)
if [ $? -eq 0 ]; then
    echo "✅ Health endpoint: OK"
else
    echo "❌ Health endpoint: Failed"
fi

# Test Swagger JSON
echo "Testing Swagger JSON endpoint..."
SWAGGER=$(curl -s http://localhost:8080/swagger/doc.json | head -c 100)
if [ ! -z "$SWAGGER" ]; then
    echo "✅ Swagger JSON: OK"
else
    echo "❌ Swagger JSON: Failed"
fi

echo ""
echo "🎉 Fix complete!"
echo ""
echo "📚 Open Swagger UI in your browser:"
echo "   http://localhost:8080/swagger/index.html"
echo ""
echo "If still blank, try:"
echo "1. Hard refresh: Ctrl+Shift+R (Windows) or Cmd+Shift+R (Mac)"
echo "2. Check logs: make docker-logs-backend"
echo "3. Full reset: make docker-clean && make docker-setup"
echo ""

