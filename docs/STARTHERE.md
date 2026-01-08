# 🚀 Start Here (1 Minute)

## Terminal Commands
```bash
# 1. Create project
mkdir davinci-chatbot && cd davinci-chatbot

# 2. Copy files from chat history (use FILES_CREATED.md checklist)

# 3. Install
go mod tidy

# 4. Configure
cp .env.example .env
# Edit GCP_PROJECT_ID, VERTEXAI_LOCATION=us-central1

# 5. Test
make test
make run

# 6. API Test
curl -X POST http://localhost:8080/chat \
  -d '{"currentMessage": "supplement brand nurture"}'


{
  "message": "Hi, I'm Da Vinci... What's your USP?",
  "workflowStep": 1,
  "identifiedVertical": "supplements"
}
