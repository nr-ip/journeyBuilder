# JourneyBuilder Application Flow & Architecture

## Table of Contents
1. [High-Level Architecture](#high-level-architecture)
2. [Request Flow Diagram](#request-flow-diagram)
3. [Information Flow Between Components](#information-flow-between-components)
4. [Prompt Engineering Flow (6-Layer Composition)](#prompt-engineering-flow-6-layer-composition)
5. [Data Flow to/from Gemini AI](#data-flow-tofrom-gemini-ai)
6. [Component Details](#component-details)

---

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLIENT (Browser)                        │
│                    Vue.js SPA (index.html)                      │
└────────────────────────────┬────────────────────────────────────┘
                             │ HTTP POST /api/chat
                             │ JSON: {currentMessage, conversationHistory}
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    HTTP SERVER (Gorilla Mux)                    │
│  Port: 8080 | CORS Enabled | Static File Serving                │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    API HANDLERS (handlers.go)                   │
│  HandleChat() → Validates Request → Calls Orchestrator          │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    ORCHESTRATOR (Core Logic)                     │
│  Coordinates: Validation → Context → Prompt → AI → Response    │
└────────────────────────────┬────────────────────────────────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
        ▼                    ▼                    ▼
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│  Validation  │    │  Knowledge   │    │  Instruction │
│  (Security)  │    │     Base     │    │   Composer   │
└──────────────┘    └──────────────┘    └──────────────┘
        │                    │                    │
        └────────────────────┼────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    GEMINI SERVICE (gemini.go)                   │
│  google.golang.org/genai → Gemini 2.5 Flash API                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## Request Flow Diagram

### Complete Request Lifecycle

```
1. CLIENT REQUEST
   └─> POST /api/chat
       Body: {
         "currentMessage": "I sell protein supplements",
         "conversationHistory": [
           {"role": "user", "content": "Hello"},
           {"role": "model", "content": "Hi! Tell me about your business..."}
         ]
       }

2. HTTP HANDLER (handlers.HandleChat)
   ├─> Parse JSON request
   ├─> Validate orchestrator is initialized
   └─> Call: orchestrator.ProcessChatRequest(ctx, req, false)

3. ORCHESTRATOR.ProcessChatRequest (9 Steps)
   │
   ├─> STEP 1: INPUT VALIDATION
   │   └─> inputValidator.ValidateInput(req.CurrentMessage)
   │       ├─> Check prompt injection patterns
   │       ├─> Check jailbreak patterns
   │       └─> Validate length & content
   │       └─> ❌ If invalid: Return error response
   │
   ├─> STEP 2: BUILD STATELESS CONTEXT
   │   └─> contextBuilder.BuildContext(req)
   │       ├─> Extract USP (Unique Selling Proposition)
   │       │   └─> Regex: "usp:", "unique selling proposition", etc.
   │       ├─> Extract ICP (Ideal Customer Profile)
   │       │   └─> Regex: "icp:", "target audience", "we sell to", etc.
   │       ├─> Identify Vertical
   │       │   └─> Keyword matching: supplements, coaching, ecommerce, etc.
   │       ├─> Extract Circle of Trust
   │       │   └─> Keywords: stranger, follower, customer, advocate
   │       └─> Extract Proposed Outcome
   │           └─> Regex: "goal:", "outcome:", "I want to", etc.
   │
   ├─> STEP 3: DETERMINE WORKFLOW STEP
   │   └─> contextBuilder.DetermineWorkflowStep(userCtx)
   │       └─> Based on filled fields count:
   │           ├─> 0 fields → StepIntroduction
   │           ├─> 1 field  → StepDiscovery
   │           ├─> 2 fields → StepValidation
   │           ├─> 3 fields → StepFrameworkApplication
   │           ├─> 4 fields → StepAnalysis
   │           └─> 5+ fields → StepExecution
   │
   ├─> STEP 4: EXTRACT KNOWLEDGE FROM KB
   │   └─> kb.ExtractRelevantContext(outcome, vertical, stepStr)
   │       ├─> Get frameworks for current step
   │       │   └─> Example: StepExecution → ["4Ps", "FAB"]
   │       ├─> Get sequence template (if outcome + vertical exist)
   │       │   └─> Lookup: "first_purchase_supplements"
   │       ├─> Get vertical guidance
   │       │   └─> Characteristics, principles, considerations
   │       └─> Return: Formatted knowledge context string
   │
   ├─> STEP 5: COMPOSE MODULAR PROMPT (6 Layers)
   │   └─> instruction.ComposerConfig.ComposeInstructions()
   │       ├─> Layer 1: Base System Prompt (Da Vinci persona)
   │       ├─> Layer 2: Security & Compliance
   │       ├─> Layer 3: Workflow Step Context
   │       ├─> Layer 4: Knowledge Context (from KB)
   │       ├─> Layer 5: Output Format Specifications
   │       └─> Layer 6: User Context (extracted data)
   │       └─> Return: Complete system prompt string
   │
   ├─> STEP 6: BUILD GEMINI REQUEST
   │   └─> services.RequestBuilder{
   │       ├─> SystemPrompt: composedPrompt (from Step 5)
   │       ├─> UserMessage: req.CurrentMessage
   │       ├─> ConversationHistory: converted messages
   │       ├─> Temperature: 0.7
   │       └─> MaxTokens: 1500
   │       }
   │
   ├─> STEP 7: CALL GEMINI AI
   │   └─> geminiService.SendRequest(ctx, geminiReq)
   │       ├─> Build genai.Content array (conversation history)
   │       ├─> Set SystemInstruction in config
   │       ├─> Retry logic (3 attempts for 503/429 errors)
   │       └─> Return: Response{Text: "AI generated response"}
   │
   ├─> STEP 8: VALIDATE OUTPUT
   │   └─> outputValidator.ValidateResponse(resp.Text, currentStep)
   │       ├─> Check spam keywords
   │       └─> Log issues (non-blocking)
   │
   └─> STEP 9: BUILD RESPONSE
       └─> models.ChatResponse{
           ├─> Message: resp.Text
           ├─> WorkflowStep: int(currentStep)
           ├─> ExtractedUSP: userCtx.ExtractedUSP
           ├─> ExtractedICP: userCtx.ExtractedICP
           ├─> IdentifiedVertical: userCtx.IdentifiedVertical
           ├─> CurrentCircle: userCtx.CurrentCircleOfTrust
           └─> ProposedOutcome: userCtx.ProposedOutcome
           }

4. HTTP RESPONSE
   └─> JSON Response to Client
       {
         "message": "Based on your protein supplements business...",
         "workflowStep": 2,
         "extractedUSP": "High-quality protein",
         "extractedICP": "Fitness enthusiasts",
         "identifiedVertical": "supplements",
         "currentCircle": "stranger",
         "proposedOutcome": "First purchase"
       }
```

---

## Information Flow Between Components

### Component Interaction Map

```
┌─────────────────────────────────────────────────────────────────────┐
│                         MAIN (cmd/api/main.go)                      │
│  Initializes all services and wires dependencies                    │
└────────────────────┬────────────────────────────────────────────────┘
                     │
        ┌────────────┼────────────┐
        │            │            │
        ▼            ▼            ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│   Gemini     │ │  Knowledge   │ │  Validation  │
│   Service    │ │     Base     │ │  (Input/     │
│              │ │              │ │   Output)    │
└──────────────┘ └──────────────┘ └──────────────┘
        │            │            │
        └────────────┼────────────┘
                     │
                     ▼
        ┌────────────────────────┐
        │     ORCHESTRATOR       │
        │  (Central Coordinator) │
        └────────────┬───────────┘
                     │
        ┌────────────┼────────────┐
        │            │            │
        ▼            ▼            ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│   Context    │ │  Instruction │ │   Handler    │
│   Builder    │ │   Composer   │ │   (HTTP)     │
└──────────────┘ └──────────────┘ └──────────────┘
```

### Data Flow Details

#### 1. **Initialization Flow** (main.go)
```
main()
  ├─> Load .env file
  ├─> Initialize GeminiService
  │   └─> Reads: GEMINI_API_KEY, GEMINI_MODEL
  ├─> Initialize KnowledgeBase
  │   └─> Loads: frameworks.json, sequence.json, verticals.json
  ├─> Initialize Validators
  │   ├─> InputValidator (prompt injection detection)
  │   └─> OutputValidator (spam detection)
  ├─> Initialize Orchestrator
  │   └─> Wires: GeminiService, KnowledgeBase, Validators
  └─> Start HTTP server (Gorilla Mux)
```

#### 2. **Request Processing Flow** (orchestrator.go)
```
ProcessChatRequest()
  ├─> Input: models.ChatRequest
  │   ├─> currentMessage: string
  │   └─> conversationHistory: []Message
  │
  ├─> Validation Layer
  │   └─> InputValidator.ValidateInput()
  │       └─> Returns: error (if malicious)
  │
  ├─> Context Extraction Layer
  │   └─> ContextBuilder.BuildContext()
  │       └─> Returns: UserContext
  │           ├─> ExtractedUSP
  │           ├─> ExtractedICP
  │           ├─> IdentifiedVertical
  │           ├─> CurrentCircleOfTrust
  │           └─> ProposedOutcome
  │
  ├─> Workflow Determination
  │   └─> ContextBuilder.DetermineWorkflowStep()
  │       └─> Returns: WorkflowStep (0-7)
  │
  ├─> Knowledge Extraction
  │   └─> KnowledgeBase.ExtractRelevantContext()
  │       ├─> Queries: frameworks.json
  │       ├─> Queries: sequence.json
  │       ├─> Queries: verticals.json
  │       └─> Returns: formatted knowledge string
  │
  ├─> Prompt Composition
  │   └─> InstructionComposer.ComposeInstructions()
  │       └─> Returns: complete system prompt
  │
  ├─> AI Service Call
  │   └─> GeminiService.SendRequest()
  │       ├─> Builds: genai.Content array
  │       ├─> Sets: SystemInstruction
  │       ├─> Calls: genai.Client.Models.GenerateContent()
  │       └─> Returns: Response{Text}
  │
  ├─> Output Validation
  │   └─> OutputValidator.ValidateResponse()
  │       └─> Checks: spam keywords
  │
  └─> Response Building
      └─> Returns: models.ChatResponse
```

---

## Prompt Engineering Flow (6-Layer Composition)

### Layer-by-Layer Breakdown

The prompt composition follows a **6-layer modular architecture** designed for:
- **Token Efficiency**: Only relevant knowledge is injected
- **Maintainability**: Each layer can be updated independently
- **Flexibility**: Different workflows can use different layer combinations

```
┌─────────────────────────────────────────────────────────────────┐
│                    FINAL COMPOSED PROMPT                        │
│  (Sent to Gemini as SystemInstruction)                          │
└─────────────────────────────────────────────────────────────────┘
                            ▲
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
        ▼                   ▼                   ▼
┌──────────────┐   ┌──────────────┐   ┌──────────────┐
│   LAYER 1    │   │   LAYER 2    │   │   LAYER 3    │
│ Base System  │   │  Security &  │   │  Workflow    │
│   Prompt     │   │  Compliance  │   │   Step       │
└──────────────┘   └──────────────┘   └──────────────┘
        │                   │                   │
        └───────────────────┼───────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
        ▼                   ▼                   ▼
┌──────────────┐   ┌──────────────┐   ┌──────────────┐
│   LAYER 4    │   │   LAYER 5    │   │   LAYER 6    │
│  Knowledge   │   │   Output     │   │    User      │
│   Context    │   │   Format     │   │   Context    │
└──────────────┘   └──────────────┘   └──────────────┘
```

### Layer 1: Base System Prompt
**Source**: `internal/instruction/base_prompt.go`

**Content**:
- Da Vinci persona definition
- Core mission statement
- Domain expertise (verticals, frameworks, outcomes)
- 8-step workflow overview
- Security & compliance guidelines
- Response format requirements

**Purpose**: Establishes the AI's identity and core capabilities

**Example**:
```
You are Da Vinci, the world's leading Email Lifecycle Strategist...
Your expertise spans consumer psychology, technical deliverability...
```

### Layer 2: Security & Compliance
**Source**: `internal/instruction/composer.go` → `buildComplianceLayer()`

**Content**:
- CAN-SPAM requirements
- GDPR/CASL compliance
- Spam rate targets (<0.3%)
- Subject line constraints (40 chars max)

**Purpose**: Ensures legal compliance and deliverability

**Example**:
```
SECURITY & COMPLIANCE MANDATE:
- CAN-SPAM: Include unsubscribe link and physical address
- GDPR/CASL: No personal data collection without consent
- Spam Rate Target: <0.3%
```

### Layer 3: Workflow Step Context
**Source**: `internal/instruction/composer.go` → `buildWorkflowStepContext()`

**Content**: Step-specific instructions based on current workflow step

**Purpose**: Focuses AI response on current conversation stage

**Workflow Steps**:
1. **StepIntroduction**: "Introduce yourself and ask for USP/ICP"
2. **StepDiscovery**: "Confirm USP and ICP, ask for outcome"
3. **StepValidation**: "Validate understanding, recommend Circle of Trust"
4. **StepFrameworkApplication**: "Select framework (PAS/AIDA) based on circle"
5. **StepCircleConfirmation**: "Confirm Buyer Circle and cadence"
6. **StepGoalSetting**: "Define specific sequence goal"
7. **StepAnalysis**: "Analyze and recommend triggers/branching"
8. **StepExecution**: "Generate complete sequence with table"

**Example** (StepExecution):
```
CURRENT WORKFLOW STEP: STEP 8: Generate complete sequence with table
FOCUS YOUR RESPONSE ON THIS STEP ONLY.
```

### Layer 4: Knowledge Context (from KB)
**Source**: `internal/knowledge/knowledge_base.go` → `ExtractRelevantContext()`

**Content**: Dynamically extracted from knowledge base based on:
- Current workflow step → Recommended frameworks
- Proposed outcome + Vertical → Sequence template
- Identified vertical → Vertical guidance

**Purpose**: Injects domain-specific knowledge (85% token compression)

**Extraction Process**:
```
1. Get frameworks for step
   └─> StepExecution → ["4Ps", "FAB"]

2. Get sequence template (if outcome + vertical exist)
   └─> "first_purchase_supplements" → Template with:
       - Duration: "7-14 days"
       - Touch Points: 5
       - Cadence: "Every 2-3 days"
       - Key Messages: [...]

3. Get vertical guidance
   └─> "supplements" → Characteristics, principles, considerations
```

**Example Output**:
```
## APPLICABLE COPYWRITING FRAMEWORKS

**4Ps (Product, Price, Place, Promotion):** BOFU conversion

## SEQUENCE TEMPLATE

Outcome: First Purchase
Duration: 7-14 days
Touch Points: 5
Cadence: Every 2-3 days
Key Messages: Value proposition, Social proof, Urgency

## VERTICAL GUIDANCE

Characteristics: Health-focused, FDA-regulated, Subscription models
Key Principles: Educational content, Trust-building, Testimonials
```

### Layer 5: Output Format Specifications
**Source**: `internal/instruction/composer.go` → `buildOutputFormatContext()`

**Content**: Response structure requirements

**Purpose**: Ensures consistent, structured AI responses

**Format Options**:
- Type: "text"
- Max Email Length: 220 chars
- Readability: Grade 6 level
- Include Table: true/false (based on step)
- Table Columns: ["Email #", "Subject Line", "Day Delay"]

**Example**:
```
OUTPUT FORMAT REQUIREMENTS:
- Type: text
- Max Email Length: 220 chars
- Readability: Grade6 level
- REQUIRED TABLE FORMAT:
| Email # | Subject Line | Day Delay |
| --- | --- | --- |
```

### Layer 6: User Context (Stateless)
**Source**: `internal/instruction/composer.go` → `buildUserContextLayer()`

**Content**: Extracted user data from conversation

**Purpose**: Personalizes AI response with user's business context

**Extracted Fields**:
- ExtractedUSP: "High-quality protein supplements"
- ExtractedICP: "Fitness enthusiasts aged 25-40"
- IdentifiedVertical: "supplements"
- CurrentCircleOfTrust: "stranger"
- ProposedOutcome: "First purchase"

**Example**:
```
EXTRACTED USP: High-quality protein supplements
EXTRACTED ICP: Fitness enthusiasts aged 25-40
DETECTED VERTICAL: supplements
CURRENT CIRCLE: stranger
PROPOSED OUTCOME: First purchase
```

### Complete Composed Prompt Example

```
[LAYER 1: Base System Prompt]
You are Da Vinci, the world's leading Email Lifecycle Strategist...

[LAYER 2: Security & Compliance]
SECURITY & COMPLIANCE MANDATE:
- CAN-SPAM: Include unsubscribe link and physical address...

[LAYER 3: Workflow Step]
CURRENT WORKFLOW STEP: STEP 8: Generate complete sequence with table
FOCUS YOUR RESPONSE ON THIS STEP ONLY.

[LAYER 4: Knowledge Context]
## APPLICABLE COPYWRITING FRAMEWORKS
**4Ps (Product, Price, Place, Promotion):** BOFU conversion

## SEQUENCE TEMPLATE
Outcome: First Purchase
Duration: 7-14 days
...

[LAYER 5: Output Format]
OUTPUT FORMAT REQUIREMENTS:
- Type: text
- Max Email Length: 220 chars
...

[LAYER 6: User Context]
EXTRACTED USP: High-quality protein supplements
EXTRACTED ICP: Fitness enthusiasts aged 25-40
DETECTED VERTICAL: supplements
CURRENT CIRCLE: stranger
PROPOSED OUTCOME: First purchase
```

---

## Data Flow to/from Gemini AI

### Request Structure

```
┌─────────────────────────────────────────────────────────────────┐
│              GeminiService.SendRequest()                        │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│              genai.GenerateContentConfig                         │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ SystemInstruction: &genai.Content{                      │  │
│  │   Parts: [{Text: composedPrompt}]                        │  │
│  │ }                                                         │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ Temperature: 0.7 (float32 pointer)                      │  │
│  │ MaxOutputTokens: 1500 (int32)                            │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│              genai.Content Array (Conversation History)          │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ [                                                         │  │
│  │   {Role: "user", Parts: [{Text: "Hello"}]},             │  │
│  │   {Role: "model", Parts: [{Text: "Hi! Tell me..."}]},    │  │
│  │   {Role: "user", Parts: [{Text: "I sell protein..."}]}   │  │
│  │ ]                                                         │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│         genai.Client.Models.GenerateContent()                   │
│  Parameters:                                                    │
│  - ctx: context.Context                                         │
│  - model: "gemini-2.5-flash"                                    │
│  - contents: []*genai.Content (conversation history)           │
│  - config: *genai.GenerateContentConfig                        │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│              Retry Logic (3 attempts)                           │
│  - Attempt 1: Immediate                                        │
│  - Attempt 2: After 1 second (if 503/429)                     │
│  - Attempt 3: After 2 seconds (if 503/429)                     │
│  - Retryable errors: 503, 429, UNAVAILABLE, RESOURCE_EXHAUSTED │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│              genai.GenerateContentResponse                      │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ Candidates: [                                            │  │
│  │   {                                                       │  │
│  │     Content: {                                           │  │
│  │       Parts: [{Text: "Based on your protein..."}]       │  │
│  │     }                                                    │  │
│  │   }                                                      │  │
│  │ ]                                                        │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│              Response Extraction                                │
│  resp.Candidates[0].Content.Parts[0].Text                     │
│  └─> "Based on your protein supplements business..."           │
└─────────────────────────────────────────────────────────────────┘
```

### Message Role Conversion

**Input** (from frontend):
```json
{
  "role": "user",
  "content": "Hello"
}
{
  "role": "ai",
  "content": "Hi! Tell me about your business..."
}
```

**Conversion** (in gemini.go):
```go
// Convert "ai" role to "model" as required by Gemini API
if role == "ai" || role == "assistant" {
    role = "model"
}
```

**Output** (to Gemini):
```go
&genai.Content{
    Role: "user",  // or "model"
    Parts: []*genai.Part{{Text: "Hello"}}
}
```

---

## Component Details

### 1. HTTP Handler Layer
**File**: `internal/api/handlers/handlers.go`

**Responsibilities**:
- Parse HTTP requests
- Validate request format
- Call orchestrator
- Format HTTP responses

**Key Function**:
```go
HandleChat(w http.ResponseWriter, r *http.Request)
  ├─> Parse JSON: models.ChatRequest
  ├─> Validate orchestrator initialized
  ├─> Call: orchestrator.ProcessChatRequest()
  └─> Return: JSON response
```

### 2. Orchestrator Layer
**File**: `internal/orchestrator/orchestrator.go`

**Responsibilities**:
- Coordinate all processing steps
- Manage workflow state
- Build AI requests
- Handle errors

**Key Function**:
```go
ProcessChatRequest(ctx, req, _)
  ├─> Input validation
  ├─> Context building
  ├─> Workflow determination
  ├─> Knowledge extraction
  ├─> Prompt composition
  ├─> AI service call
  ├─> Output validation
  └─> Response building
```

### 3. Context Builder
**File**: `internal/orchestrator/context_builder.go`

**Responsibilities**:
- Extract business context from conversation
- Identify vertical
- Determine workflow step
- Stateless processing (no server-side storage)

**Extraction Methods**:
- `extractUSP()`: Regex patterns for USP detection
- `extractICP()`: Regex patterns for ICP detection
- `identifyVertical()`: Keyword matching for verticals
- `extractCircleOfTrust()`: Keyword matching for circles
- `extractProposedOutcome()`: Regex patterns for outcomes

### 4. Knowledge Base
**File**: `internal/knowledge/knowledge_base.go`

**Responsibilities**:
- Load and cache knowledge data
- Extract relevant context
- Provide framework recommendations
- Provide sequence templates

**Data Sources**:
- `frameworks.json`: Copywriting frameworks (AIDA, PAS, FAB, etc.)
- `sequence.json`: Pre-built sequence templates
- `verticals.json`: Vertical-specific guidance

**Key Function**:
```go
ExtractRelevantContext(outcome, vertical, step)
  ├─> Get frameworks for step
  ├─> Get sequence template (if available)
  ├─> Get vertical guidance (if available)
  └─> Return: formatted knowledge string
```

### 5. Instruction Composer
**File**: `internal/instruction/composer.go`

**Responsibilities**:
- Compose 6-layer prompt
- Build compliance layer
- Build workflow step context
- Build output format context
- Build user context layer

**Key Function**:
```go
ComposeInstructions()
  ├─> Layer 1: Base system prompt
  ├─> Layer 2: Security & compliance
  ├─> Layer 3: Workflow step context
  ├─> Layer 4: Knowledge context
  ├─> Layer 5: Output format
  └─> Layer 6: User context
```

### 6. Gemini Service
**File**: `internal/services/gemini.go`

**Responsibilities**:
- Initialize Gemini client
- Build API requests
- Handle retries
- Extract responses

**Key Function**:
```go
SendRequest(ctx, req)
  ├─> Build genai.Content array (conversation history)
  ├─> Set SystemInstruction
  ├─> Set Temperature & MaxTokens
  ├─> Call: genai.Client.Models.GenerateContent()
  ├─> Retry on 503/429 errors (3 attempts)
  └─> Extract and return response text
```

### 7. Validation Layer
**Files**: 
- `internal/validation/input_validator.go`
- `internal/validation/output_validator.go`

**Input Validator**:
- Detects prompt injection attacks
- Detects jailbreak attempts
- Validates input length
- Checks for code injection

**Output Validator**:
- Checks for spam keywords
- Validates response quality
- Non-blocking (logs issues)

---

## Token Optimization Strategy

### Knowledge Base Compression (85% reduction)

**Traditional Approach**:
```
Full knowledge base: ~50,000 tokens
- All frameworks
- All sequences
- All verticals
- All examples
```

**Optimized Approach**:
```
Extracted context: ~7,500 tokens (85% reduction)
- Only relevant frameworks for current step
- Only matching sequence template
- Only relevant vertical guidance
- No examples unless needed
```

**Implementation**:
1. **Step-based framework selection**: Only frameworks relevant to current workflow step
2. **Outcome+Vertical matching**: Only sequence templates matching user's outcome and vertical
3. **Conditional loading**: Vertical guidance only if vertical is identified

**Example**:
```
Step: StepExecution
Outcome: "First Purchase"
Vertical: "supplements"

Extracted:
- Frameworks: ["4Ps", "FAB"] (2 frameworks, not all 6)
- Sequence: "first_purchase_supplements" (1 template, not all 20)
- Vertical: "supplements" guidance (1 vertical, not all 7)
```

---

## Error Handling & Retry Logic

### Retry Strategy (gemini.go)

**Retryable Errors**:
- 503 (Service Unavailable)
- 429 (Rate Limit)
- UNAVAILABLE
- RESOURCE_EXHAUSTED

**Retry Pattern**:
```
Attempt 1: Immediate
  └─> If error → Wait 1 second
Attempt 2: After 1 second
  └─> If error → Wait 2 seconds
Attempt 3: After 2 seconds
  └─> If error → Return error
```

**Non-Retryable Errors**:
- 400 (Bad Request)
- 401 (Unauthorized)
- 403 (Permission Denied)
- 404 (Not Found)

---

## Summary

### Key Design Principles

1. **Stateless Architecture**: All context extracted from request, no server-side storage
2. **Modular Prompt Engineering**: 6-layer composition for maintainability
3. **Token Optimization**: 85% reduction via selective knowledge extraction
4. **Security First**: Input validation, output validation, compliance built-in
5. **Workflow-Driven**: 8-step conversation workflow guides AI responses
6. **Vertical-Aware**: Industry-specific knowledge injection
7. **Retry Resilience**: Automatic retry for transient API errors

### Data Flow Summary

```
Client Request
  └─> HTTP Handler
      └─> Orchestrator
          ├─> Input Validation
          ├─> Context Extraction
          ├─> Workflow Determination
          ├─> Knowledge Extraction
          ├─> Prompt Composition (6 layers)
          ├─> Gemini AI Call
          ├─> Output Validation
          └─> Response Building
              └─> HTTP Response
```

### Prompt Engineering Summary

```
6-Layer Prompt Composition:
1. Base System Prompt (Da Vinci persona)
2. Security & Compliance (CAN-SPAM, GDPR)
3. Workflow Step Context (current step focus)
4. Knowledge Context (frameworks, sequences, verticals)
5. Output Format (structure requirements)
6. User Context (extracted business data)
```

---

*Document generated from codebase analysis*
*Last updated: Based on current codebase structure*
