# Da Vinci - AI Email Sequence Creator

**Production-ready Golang chatbot** powered by Vertex AI (Gemini 2.5 Flash) for DTC email marketers. Creates high-conversion automated sequences using Buyers' Circles of Trust framework.

[![Go](https://img.shields.io/badge/Go-1.21-blue.svg)](https://golang.org)
[![Vertex AI](https://img.shields.io/badge/Vertex%20AI-Gemini%202.5-green)](https://cloud.google.com/vertex-ai)
[![License](https://img.shields.io/badge/License-MIT-yellow)](LICENSE)

## 🚀 Quick Start (5 minutes)

```bash
# 1. Clone & Setup
mkdir davinci-chatbot && cd davinci-chatbot
# Copy all 35 files from docs/DIRECTORYTREE.txt structure

# 2. Dependencies
go mod tidy

# 3. Configure (GCP required)
cp .env.example .env
# Edit .env: GCP_PROJECT_ID=your-project

# 4. Run Backend
make run
# Backend: http://localhost:8080

# 5. Test
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"currentMessage": "I sell supplements, nurture new customers"}'

Expected Output:

text
{
  "message": "Hi, I'm Da Vinci... What's your USP?",
  "workflowStep": 1,
  "identifiedVertical": "supplements"
}
Architecture
text
Frontend SPA (React) → Echo API → Orchestrator → Vertex AI (Gemini)
                                      ↓
                       Modular KB (85% token compression)
Key Features:

Stateless Design: Full context in each request

8-Step Workflow: Consultative selling → Sequence generation

Modular Prompts: Dynamic instruction injection (6 layers)

Vertical Intelligence: Supplements, Coaching, Ecommerce, etc.

Security: Prompt injection validation, compliance guards

Performance: 85% token reduction, LRU caching

📊 Performance Metrics
Metric	Value
Token Reduction	85% via KB extraction
Cost per Request	$0.02-0.04
Response Time	30-40% faster
Spam Rate Target	<0.3%
🛠️ Prerequisites
Go 1.21+

GCP Account + Vertex AI enabled

Service Account with roles/aiplatform.user

🔧 Development
bash
# Install
make setup

# Test
make test

# Dev with hot reload
make dev

# Docker
make docker-build docker-run
📁 Directory Structure
See docs/DIRECTORYTREE.txt (35 files, 12K LOC).

🎯 Supported Verticals
Nutritional Supplements

Online Coaching

Ecommerce/Fashion

Nonprofits

Beauty/Skincare

Subscription Boxes

Home Goods

HealthTech

🔒 Security & Compliance
CAN-SPAM, GDPR, CASL compliant

Prompt injection detection

Spam-trigger auditing

Tiered security responses

📈 Copywriting Frameworks
AIDA • PAS • FAB • BAB • 4Ps • Hero
