#!/bin/bash
# Setup script to configure Codex to use FreeLLM by default

echo "Setting up Codex to use FreeLLM endpoint..."

# Create config directory if it doesn't exist
mkdir -p ~/.codex

# Create config file
cat > ~/.codex/config.json << 'EOF'
{
  "providers": {
    "openai": {
      "apiKey": "sk-freellm",
      "baseURL": "http://localhost:4000/v1"
    }
  },
  "defaultModel": "poolside/laguna-xs.2:free",
  "approvalPolicy": "never",
  "sandbox": {
    "workspace": true,
    "write": true
  }
}
EOF

echo "Created ~/.codex/config.json"

# Create shell alias
cat >> ~/.bashrc << 'EOF'

# Codex FreeLLM alias - always use FreeLLM endpoint
alias codex='OPENAI_BASE_URL=http://localhost:4000/v1 OPENAI_API_KEY=sk-freellm codex --model poolside/laguna-xs.2:free'
EOF

echo "Added codex alias to ~/.bashrc"
echo ""
echo "To activate immediately, run:"
echo "  source ~/.bashrc"
echo ""
echo "Or start codex with:"
echo "  ./codex-freellm.sh"
