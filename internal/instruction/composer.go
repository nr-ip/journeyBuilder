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
		StepDiscovery:            "STEP 2: Confirm USP and ICP, ask for outcome",
		StepValidation:           "STEP 3: Validate understanding, recommend Circle of Trust",
		StepFrameworkApplication: "STEP 4: Select framework (PAS/AIDA) based on circle",
		StepCircleConfirmation:   "STEP 5: Confirm Buyer Circle and cadence",
		StepGoalSetting:          "STEP 6: Define specific sequence goal",
		StepAnalysis:             "STEP 7: Analyze and recommend triggers/branching",
		StepExecution:            "STEP 8: Generate complete sequence with table",
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
- REQUIRED TABLE FORMAT:
| %s |
| %s |`, strings.Join(c.OutputFormat.TableColumns, " | "), strings.Repeat("--- | ", len(c.OutputFormat.TableColumns)))
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
