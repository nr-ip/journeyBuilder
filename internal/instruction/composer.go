package instruction

import (
	"fmt"
	"strings"
)

// ComposeInstructions builds the final prompt through 6 modular layers.
func (c *ComposerConfig) ComposeInstructions() string {
	var sb strings.Builder

	// Layer 1: Base System Instructions
	sb.WriteString(c.BaseSystemPrompt)
	sb.WriteString("\n\n")

	// Layer 2: Security & Compliance
	sb.WriteString(c.buildComplianceLayer())
	sb.WriteString("\n\n")

	// Layer 3: Workflow Step Context
	sb.WriteString(c.buildWorkflowStepContext())
	sb.WriteString("\n\n")

	// Layer 4: Knowledge Context (from KB)
	sb.WriteString(c.KnowledgeContext)
	sb.WriteString("\n\n")

	// Layer 5: Output Format Specifications
	sb.WriteString(c.buildOutputFormatContext())
	sb.WriteString("\n\n")

	// Layer 6: User Context (stateless)
	sb.WriteString(c.buildUserContextLayer())
	sb.WriteString("\n\n")

	return sb.String()
}

// buildComplianceLayer adds security and compliance instructions.
func (c *ComposerConfig) buildComplianceLayer() string {
	return `SECURITY & COMPLIANCE MANDATE:
- CAN-SPAM: Include unsubscribe link and physical address in every email
- GDPR/CASL: No personal data collection without explicit consent
- Spam Rate Target: <0.3% - Avoid trigger words, use balanced design
- Subject Lines: 40 chars max, personalized where possible
`
}

// buildWorkflowStepContext injects step-specific instructions.
func (c *ComposerConfig) buildWorkflowStepContext() string {
	stepDescriptions := map[WorkflowStep]string{
		StepIntroduction:         "STEP 1: Introduce yourself and ask for USP/ICP",
		StepDiscovery:            "STEP 2: Next tell the user, Tell me about your product's Unique Selling Proposition (USP) and its Ideal Customer Profile (ICP)",
		StepValidation:           "STEP 3: Respond back with your understanding summary of the USP and the ICP.  Ask the user to confirm and continue to the next step when answered in the affirmative.",
		StepFrameworkApplication: "STEP 4: Next ask the user, Who is your intended audience according to The Buyers' Circles of Trust(tm)?  If you're not sure, tell me who you want to target and I'll identify which Circle of Trust it is.",
		StepCircleConfirmation:   "STEP 5: Confirm to the user your understanding of which Circle of Trust is the intended audience of the automated email sequence",
		StepGoalSetting:          "STEP 6: Next ask the user, What is the desired outcome of your automated email sequence?",
		StepAnalysis:             "STEP 7: Compare the desired outcome against the Circle of Trust of the intended audience. Tell the user whether the desired outcome is appropriate to the Circle of Trust. Provide your analysis ONCE and stop. Do NOT repeat this analysis. Do NOT say 'Let's move to STEP 8' or announce transitions. After providing the analysis, the system will automatically advance to Step 8.",
		StepExecution:            "STEP 8: IMMEDIATELY GENERATE THE COMPLETE EMAIL SEQUENCE NOW. DO NOT announce transitions, say 'Let's move to STEP 8', or make any introductory statements. DO NOT ASK ANY QUESTIONS. DO NOT ask about tone, subject lines, individual emails, CTAs, delays, or anything else. You have all the information you need. Start with the table that includes Email #, Subject Line, AND Day Delay (required for every row). Then provide full content for each email. Begin generating immediately - no announcements, no questions, no confirmations, no asking for preferences. Just generate the sequence.",
	}

	return fmt.Sprintf("CURRENT WORKFLOW STEP: %s\nFOCUS YOUR RESPONSE ON THIS STEP ONLY.", stepDescriptions[c.WorkflowStep])
}

// buildOutputFormatContext enforces response structure.
func (c *ComposerConfig) buildOutputFormatContext() string {
	format := fmt.Sprintf(`
OUTPUT FORMAT REQUIREMENTS:
- Type: %s
- Max Email Length: %d chars
- Readability: %s level
`, c.OutputFormat.Type, c.OutputFormat.MaxEmailLength, c.OutputFormat.ReadabilityLevel)

	if c.OutputFormat.IncludeTable {
		format += fmt.Sprintf(`
- REQUIRED TABLE FORMAT (for Step 8 - Execution):
| %s |
| %s |

CRITICAL TABLE REQUIREMENTS:
- EVERY row MUST include ALL three columns: Email #, Subject Line, AND Day Delay
- Day Delay MUST be a number (0, 1, 2, 3, etc.) indicating days to wait before sending this email
- Email 1 MUST have Day Delay = 0 (sent immediately)
- Subsequent emails MUST have increasing Day Delay values based on the cadence:
  * If cadence is "Every 2-3 days": Use delays like 0, 2, 5, 7, 10 (incrementing by 2-3 days)
  * If cadence is "1 hour, 12 hours, 24 hours": Convert to days (0, 0, 1) - round hours to nearest day
  * If cadence is "Every day": Use delays like 0, 1, 2, 3, 4
  * Always start with 0 for the first email
- Day Delay values MUST follow the cadence from the sequence template if available
- DO NOT leave Day Delay blank, empty, or use text - it MUST be a numeric value for every email
- Example table row: | 1 | Welcome to our product | 0 |

REQUIRED ACTION: Generate the complete sequence immediately. Start directly with the table, then provide all email content. 
Each row must include: Email number, Subject line (max 40 chars), and Day Delay (number of days to wait before sending this email).
- AFTER THE TABLE, provide the full email content for each email in the sequence.
- Each email must be clearly labeled (e.g., "Email 1:", "Email 2:", etc.).
- Subject lines must be personalized, compelling, and under 40 characters.
- Email content must be concise (max %d chars), compliant (CAN-SPAM, GDPR), and written at %s readability level.
- Delays should be realistic and follow the cadence from the sequence template if available.`,
			strings.Join(c.OutputFormat.TableColumns, " | "),
			strings.Repeat("--- | ", len(c.OutputFormat.TableColumns)),
			c.OutputFormat.MaxEmailLength,
			c.OutputFormat.ReadabilityLevel)
	}

	return format
}

// buildUserContextLayer injects extracted context.
func (c *ComposerConfig) buildUserContextLayer() string {
	ctx := c.UserContext
	var sb strings.Builder

	if ctx.ExtractedUSP != "" {
		sb.WriteString(fmt.Sprintf("EXTRACTED USP: %s\n", ctx.ExtractedUSP))
	}
	if ctx.ExtractedICP != "" {
		sb.WriteString(fmt.Sprintf("EXTRACTED ICP: %s\n", ctx.ExtractedICP))
	}
	if ctx.IdentifiedVertical != "" {
		sb.WriteString(fmt.Sprintf("DETECTED VERTICAL: %s\n", ctx.IdentifiedVertical))
	}
	if ctx.CurrentCircleOfTrust != "" {
		sb.WriteString(fmt.Sprintf("CURRENT CIRCLE: %s\n", ctx.CurrentCircleOfTrust))
	}
	if ctx.ProposedOutcome != "" {
		sb.WriteString(fmt.Sprintf("PROPOSED OUTCOME: %s\n", ctx.ProposedOutcome))
	}

	return sb.String()
}
