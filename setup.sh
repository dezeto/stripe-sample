#!/bin/bash

echo "🚀 Setting up Stripe Integration..."

# Create backend .env file if it doesn't exist
if [ ! -f "apps/backend/.env" ]; then
    echo "📝 Creating backend .env file..."
    cp apps/backend/.env.example apps/backend/.env
    echo "⚠️  Please update apps/backend/.env with your Stripe secret key!"
fi

# Create frontend .env.local file if it doesn't exist
if [ ! -f "apps/web/.env.local" ]; then
    echo "📝 Frontend .env.local already exists"
else
    echo "✅ Frontend environment configured"
fi

echo ""
echo "🔧 Setup Instructions:"
echo "1. Update apps/backend/.env with your Stripe secret key"
echo "2. Start backend: cd apps/backend && go run server.go"
echo "3. Start frontend: cd apps/web && npm run dev"
echo "4. Visit http://localhost:3000"
echo ""
echo "📚 Test Cards:"
echo "   Success: 4242424242424242"
echo "   Decline: 4000000000000002"
echo ""
