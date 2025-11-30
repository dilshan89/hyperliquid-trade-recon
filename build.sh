#!/bin/bash

set -e  # Exit on error

echo "========================================="
echo "Building Hyperliquid Trade Reconciliation"
echo "========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Step 1: Build React frontend
echo -e "${BLUE}Step 1/3: Building React frontend...${NC}"
cd frontend
npm run build
cd ..
echo -e "${GREEN}✓ Frontend build complete${NC}"
echo ""

# Step 2: Copy build to backend for embedding
echo -e "${BLUE}Step 2/3: Preparing files for embedding...${NC}"
rm -rf backend/frontend/build
mkdir -p backend/frontend
cp -r frontend/build backend/frontend/
echo -e "${GREEN}✓ Files prepared for embedding${NC}"
echo ""

# Step 3: Build Go binary with embedded frontend
echo -e "${BLUE}Step 3/3: Building Go binary with embedded frontend...${NC}"
cd backend
CGO_ENABLED=0 go build -o ../hyperliquid-recon main.go
cd ..
echo -e "${GREEN}✓ Binary build complete${NC}"
echo ""

# Cleanup
echo -e "${BLUE}Cleaning up...${NC}"
rm -rf backend/frontend
echo -e "${GREEN}✓ Cleanup complete${NC}"
echo ""

echo "========================================="
echo -e "${GREEN}Build successful!${NC}"
echo "========================================="
echo ""
echo "Single executable created: ./hyperliquid-recon"
echo ""
echo "To run the application:"
echo "  ./hyperliquid-recon"
echo ""
echo "Then open your browser at: http://localhost:8080"
echo ""