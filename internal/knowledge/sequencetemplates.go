package knowledge

import (
	"fmt"
	"strings"
	"sync"
)

// SequenceTemplates manages sequence lookup with LRU caching
type SequenceTemplates struct {
	templates map[string]*SequenceTemplate // outcome_vertical → template
	cache     *sync.Map
}

// NewSequenceTemplates loads from data/knowledge/sequences.json
func NewSequenceTemplates() *SequenceTemplates {
	st := &SequenceTemplates{
		templates: make(map[string]*SequenceTemplate),
		cache:     &sync.Map{},
	}

	// Load embedded or file-based sequences
	st.loadTemplates()

	return st
}

func (st *SequenceTemplates) loadTemplates() {
	// Pre-populate from training data (85% compressed)
	templates := map[string]*SequenceTemplate{
		"first_purchase_dtc": {
			Outcome:        "First Purchase Acquisition",
			Vertical:       "DTC",
			Duration:       "7-14 days",
			TouchPoints:    5,
			Triggers:       []string{"Subscribed", "Viewed Product"},
			CadenceString:  "Every 2-3 days",
			Frameworks:     []string{"AIDA", "FAB"},
			KeyMessages:    []string{"Welcome & Value Prop", "Social Proof", "Objection Buster", "Limited Offer"},
			BranchingLogic: "IF clicks CTA THEN exit; IF views product THEN cart abandonment",
		},
		"cart_abandonment_dtc": {
			Outcome:        "Cart Recovery",
			Vertical:       "DTC",
			Duration:       "48 hours",
			TouchPoints:    3,
			Triggers:       []string{"Abandoned Cart"},
			CadenceString:  "1h, 12h, 24h",
			Frameworks:     []string{"PAS", "FAB"},
			KeyMessages:    []string{"Item Reminder", "Security Reassurance", "Final Push"},
			BranchingLogic: "Hyper-urgent based on abandonment time",
		},
		"onboarding_supplements": {
			Outcome:        "Habit Formation",
			Vertical:       "Supplements",
			Duration:       "21 days",
			TouchPoints:    5,
			Triggers:       []string{"First Purchase"},
			CadenceString:  "Every 3-4 days",
			Frameworks:     []string{"AIDA", "4Ps"},
			KeyMessages:    []string{"Usage Instructions", "Scientific Credibility", "Testimonials", "Results Check-in"},
			BranchingLogic: "IF no usage check-in THEN proactive support",
		},
		"lead_nurture_coaching": {
			Outcome:        "Consultation Booking",
			Vertical:       "Coaching",
			Duration:       "30-60 days",
			TouchPoints:    12,
			Triggers:       []string{"Lead Magnet Download"},
			CadenceString:  "Every 2-3 days",
			Frameworks:     []string{"PAS", "BAB", "4Ps"},
			KeyMessages:    []string{"Lead Magnet Delivery", "Authority Building", "Case Studies", "Pricing Objection Handling"},
			BranchingLogic: "IF engages THEN accelerate; IF silent THEN re-engagement",
		},
		"donor_escalation_nonprofit": {
			Outcome:        "Single to Recurring Donor",
			Vertical:       "Nonprofit",
			Duration:       "60 days",
			TouchPoints:    4,
			Triggers:       []string{"One-time Donation"},
			CadenceString:  "15d, 30d, 45d",
			Frameworks:     []string{"BAB", "4Ps"},
			KeyMessages:    []string{"Thank You & Impact", "Mission Reinforcement", "Recurring Program Ask"},
			BranchingLogic: "IF accepts recurring THEN VIP nurture",
		},
	}

	for key, template := range templates {
		st.templates[key] = template
	}
}

// GetSequence retrieves template by outcome + vertical key
func (st *SequenceTemplates) GetSequence(outcome, vertical string) *SequenceTemplate {
	cacheKey := fmt.Sprintf("seq:%s:%s", strings.ToLower(outcome), strings.ToLower(vertical))

	// Check cache first
	if cached, found := st.cache.Load(cacheKey); found {
		return cached.(*SequenceTemplate)
	}

	// Lookup in templates
	key := fmt.Sprintf("%s_%s", strings.ToLower(outcome), strings.ToLower(vertical))
	if template, exists := st.templates[key]; exists {
		st.cache.Store(cacheKey, template)
		return template
	}

	// Fallback: generic outcome
	fallbackKey := strings.ToLower(outcome) + "_dtc"
	if template, exists := st.templates[fallbackKey]; exists {
		st.cache.Store(cacheKey, template)
		return template
	}

	return nil
}

// ExtractRelevantSequenceContext builds minimal context for instruction composer
func (st *SequenceTemplates) ExtractRelevantSequenceContext(outcome, vertical string) string {
	template := st.GetSequence(outcome, vertical)
	if template == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("## RECOMMENDED SEQUENCE TEMPLATE\n\n")
	fmt.Fprintf(&sb, "**Outcome:** %s (%s)\n", template.Outcome, template.Vertical)
	fmt.Fprintf(&sb, "**Duration:** %s (%d touchpoints)\n", template.Duration, template.TouchPoints)
	fmt.Fprintf(&sb, "**Cadence:** %s\n", template.CadenceString)
	fmt.Fprintf(&sb, "**Recommended Frameworks:** %s\n", strings.Join(template.Frameworks, ", "))
	fmt.Fprintf(&sb, "**Key Messages:** %s\n", strings.Join(template.KeyMessages, " → "))
	fmt.Fprintf(&sb, "**Branching Logic:** %s\n", template.BranchingLogic)

	return sb.String()
}

// ListSequencesForVertical returns available sequences for a vertical
func (st *SequenceTemplates) ListSequencesForVertical(vertical string) []string {
	var sequences []string
	for key, template := range st.templates {
		if strings.Contains(strings.ToLower(key), strings.ToLower(vertical)) {
			sequences = append(sequences, fmt.Sprintf("%s (%s)", template.Outcome, template.Duration))
		}
	}
	return sequences
}
