davinci-chatbot/
├── cmd/api/main.go                  # Entry point
├── internal/                         # Core application logic
│   ├── api/handlers/chat.go         # HTTP handlers
│   ├── knowledge/knowledge_base.go   # Knowledge base manager
│   ├── instruction/composer.go       # 6-layer prompt builder
│   ├── orchestrator/                 # Request processing
│   ├── vertex/client.go              # Gemini integration
│   ├── validation/                   # Input/output validation
│   └── models/chat.go                # Data models
├── data/knowledge/                   # Pre-built knowledge base
│   ├── frameworks.json               # 6 copywriting frameworks
│   ├── sequences.json                # Sequence templates
│   └── verticals.json                # Industry guidance
├── public/                           # Frontend (ready for your SPA)
├── go.mod                            # Dependencies
├── go.sum                            # Dependency checksums
├── Makefile                          # Build commands
├── .env.example                      # Environment template
├── test_api.sh                       # API test suite
├── setup.sh                          # Automated setup script
├── README.md                         # Full documentation
├── QUICKSTART.md                     # 5-minute setup
└── INSTALLATION_SUMMARY.md           # Detailed setu
