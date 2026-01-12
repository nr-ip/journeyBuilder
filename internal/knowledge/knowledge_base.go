package knowledge

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// Framework represents a copywriting framework
type Framework struct {
	Name          string   `json:"name"`           // AIDA, PAS, FAB, etc.
	Acronym       string   `json:"acronym"`        // "AIDA"
	Components    []string `json:"components"`     // ["Attention", "Interest", "Desire", "Action"]
	BestFor       []string `json:"best_for"`       // ["TOFU", "story-driven content"]
	EmotionalTone string   `json:"emotional_tone"` // "positive", "urgent", "transformational"
	FunnelStage   string   `json:"funnel_stage"`   // TOFU, MOFU, BOFU
	Example       string   `json:"example"`        // Real example
	Criticisms    string   `json:"criticisms"`     // Limitations
	Score         float64  `json:"score"`          // Relevance score (optional, used by FrameworkLookup)
}

// SequenceTemplate represents a pre-built outcome→sequence mapping
type SequenceTemplate struct {
	Outcome        string   `json:"outcome"`  // "First Purchase", "Churn Prevention", etc.
	Vertical       string   `json:"vertical"` // "DTC", "Supplements", "Coaching"
	Duration       string   `json:"duration"` // "7-14 days", "30 days"
	TouchPoints    int      `json:"touch_points"`
	Triggers       []string `json:"triggers"`
	CadenceString  string   `json:"cadence"`    // "Every 2-3 days"
	Frameworks     []string `json:"frameworks"` // Which frameworks to apply
	KeyMessages    []string `json:"key_messages"`
	BranchingLogic string   `json:"branching_logic"` // If/Then logic
}

// KnowledgeBase holds all compressed training data
type KnowledgeBase struct {
	Frameworks        map[string]*Framework        `json:"frameworks"`
	SequenceTemplates map[string]*SequenceTemplate `json:"sequences"`
	VerticalGuides    map[string]VerticalGuidance  `json:"verticals"`
	cache             *sync.Map                    // LRU cache for frequent lookups
}

type VerticalGuidance struct {
	VerticalName         string   `json:"vertical_name"`
	Characteristics      []string `json:"characteristics"`
	KeyPrinciples        []string `json:"key_principles"`
	CommonOutcomes       []string `json:"common_outcomes"`
	UniqueConsiderations string   `json:"unique_considerations"`
}

// NewKnowledgeBase loads and initializes the KB from JSON
func NewKnowledgeBase(frameworksPath, sequencesPath, verticalsPath string) (*KnowledgeBase, error) {
	kb := &KnowledgeBase{
		Frameworks:        make(map[string]*Framework),
		SequenceTemplates: make(map[string]*SequenceTemplate),
		VerticalGuides:    make(map[string]VerticalGuidance),
		cache:             &sync.Map{},
	}

	// Load frameworks
	if err := loadJSON(frameworksPath, &kb.Frameworks); err != nil {
		return nil, fmt.Errorf("failed to load frameworks: %w", err)
	}

	// Load sequences
	if err := loadJSON(sequencesPath, &kb.SequenceTemplates); err != nil {
		return nil, fmt.Errorf("failed to load sequences: %w", err)
	}

	// Load verticals
	if err := loadJSON(verticalsPath, &kb.VerticalGuides); err != nil {
		return nil, fmt.Errorf("failed to load verticals: %w", err)
	}

	return kb, nil
}

// GetFramework retrieves a copywriting framework with caching
func (kb *KnowledgeBase) GetFramework(name string) *Framework {
	cacheKey := fmt.Sprintf("fw:%s", name)

	if cached, found := kb.cache.Load(cacheKey); found {
		return cached.(*Framework)
	}

	framework := kb.Frameworks[strings.ToLower(name)]
	if framework != nil {
		kb.cache.Store(cacheKey, framework)
	}
	return framework
}

// GetSequenceTemplate retrieves a pre-built sequence template
func (kb *KnowledgeBase) GetSequenceTemplate(outcome, vertical string) *SequenceTemplate {
	cacheKey := fmt.Sprintf("seq:%s:%s", outcome, vertical)

	if cached, found := kb.cache.Load(cacheKey); found {
		return cached.(*SequenceTemplate)
	}

	key := fmt.Sprintf("%s_%s", strings.ToLower(outcome), strings.ToLower(vertical))
	template := kb.SequenceTemplates[key]
	if template != nil {
		kb.cache.Store(cacheKey, template)
	}
	return template
}

// GetVerticalGuidance returns guidance for a specific vertical
func (kb *KnowledgeBase) GetVerticalGuidance(vertical string) VerticalGuidance {
	return kb.VerticalGuides[strings.ToLower(vertical)]
}

// ExtractRelevantContext builds a minimal context string from KB
// This is called by the instruction composer to inject only relevant knowledge
func (kb *KnowledgeBase) ExtractRelevantContext(outcome, vertical, currentStep string) string {
	var sb strings.Builder

	// Add relevant frameworks (only at StepExecution when actually generating sequence)
	recommendedFrameworks := kb.getFrameworksForStep(currentStep)
	if len(recommendedFrameworks) > 0 {
		sb.WriteString("## APPLICABLE COPYWRITING FRAMEWORKS\n\n")
		for _, fw := range recommendedFrameworks {
			if framework := kb.GetFramework(fw); framework != nil {
				sb.WriteString(fmt.Sprintf("**%s (%s):** %s\n",
					framework.Name,
					framework.Acronym,
					framework.BestFor[0]))
				if framework.EmotionalTone != "" {
					sb.WriteString(fmt.Sprintf("**Tone:** %s\n", framework.EmotionalTone))
				}
				if len(framework.Components) > 0 {
					sb.WriteString(fmt.Sprintf("**Components:** %s\n", strings.Join(framework.Components, ", ")))
				}
			}
		}
		sb.WriteString("\n")
	}

	// Add sequence template if available
	if outcome != "" && vertical != "" {
		if template := kb.GetSequenceTemplate(outcome, vertical); template != nil {
			sb.WriteString("\n## SEQUENCE TEMPLATE\n\n")
			sb.WriteString(fmt.Sprintf("Outcome: %s\n", template.Outcome))
			sb.WriteString(fmt.Sprintf("Duration: %s\n", template.Duration))
			sb.WriteString(fmt.Sprintf("Touch Points: %d\n", template.TouchPoints))
			sb.WriteString(fmt.Sprintf("Cadence: %s\n", template.CadenceString))
			sb.WriteString(fmt.Sprintf("Key Messages: %s\n", strings.Join(template.KeyMessages, ", ")))
		}
	}

	// Add vertical guidance if detected
	if vertical != "" {
		if guidance := kb.GetVerticalGuidance(vertical); guidance.VerticalName != "" {
			sb.WriteString("\n## VERTICAL GUIDANCE\n\n")
			sb.WriteString(fmt.Sprintf("Characteristics: %s\n", strings.Join(guidance.Characteristics, "; ")))
			sb.WriteString(fmt.Sprintf("Key Principles: %s\n", strings.Join(guidance.KeyPrinciples, "; ")))
		}
	}

	return sb.String()
}

// getFrameworksForStep returns recommended frameworks for a workflow step
func (kb *KnowledgeBase) getFrameworksForStep(step string) []string {
	frameworkMap := map[string][]string{
		"StepDiscovery":            {},             // No frameworks needed - just asking for USP/ICP
		"StepValidation":           {},             // No frameworks needed - just validating understanding
		"StepFrameworkApplication": {},             // No frameworks needed - just asking about Circles of Trust
		"StepCircleConfirmation":   {},             // No frameworks needed - just confirming Circle
		"StepGoalSetting":          {},             // No frameworks needed - just asking for outcome
		"StepAnalysis":             {},             // No frameworks needed - just analyzing appropriateness
		"StepExecution":            {"4Ps", "FAB"}, // Only inject frameworks when actually generating sequence
	}

	if frameworks, found := frameworkMap[step]; found {
		return frameworks
	}
	return []string{} // Default to empty - frameworks only needed at execution
}

func loadJSON(path string, target interface{}) error {
	// In production, load from file system or embedded assets
	data, err := readFileOrAsset(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func readFileOrAsset(path string) ([]byte, error) {
	// Simplified: in production use embed.FS or ioutil.ReadFile
	return []byte("{}"), nil
}
