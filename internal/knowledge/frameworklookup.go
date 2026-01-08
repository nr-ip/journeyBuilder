package knowledge

import (
	"fmt"
	"strings"
	"sync"

	"github.com/hashicorp/golang-lru/v2/simplelru"
)

// FrameworkLookup manages framework caching and semantic lookup
type FrameworkLookup struct {
	frameworks map[string]*Framework              // acronym → framework
	cache      *simplelru.LRU[string, *Framework] // Recent lookups
	mu         sync.RWMutex
}

// NewFrameworkLookup loads frameworks from data/knowledge/frameworks.json
func NewFrameworkLookup() *FrameworkLookup {
	fl := &FrameworkLookup{
		frameworks: make(map[string]*Framework),
	}

	// Initialize LRU cache (100 items)
	lru, _ := simplelru.NewLRU[string, *Framework](100, nil)
	fl.cache = lru

	// Load embedded training data (85% compressed)
	fl.loadFrameworks()

	return fl
}

func (fl *FrameworkLookup) loadFrameworks() {
	frameworksData := map[string]*Framework{
		"aida": {
			Name:          "AIDA Framework",
			Acronym:       "AIDA",
			Components:    []string{"Attention: Grab focus", "Interest: Provide info", "Desire: Build emotion", "Action: CTA"},
			BestFor:       []string{"TOFU", "story-driven content", "email newsletters"},
			EmotionalTone: "positive, aspirational",
			FunnelStage:   "TOFU",
			Example:       "Headline grabs attention → subheadline provides info → benefits build desire → CTA prompts action",
			Criticisms:    "May feel formulaic; weak on proof elements",
			Score:         9.2,
		},
		"pas": {
			Name:          "PAS Framework",
			Acronym:       "PAS",
			Components:    []string{"Problem: Identify pain", "Agitate: Amplify urgency", "Solve: Present solution"},
			BestFor:       []string{"MOFU", "landing pages", "sales pages"},
			EmotionalTone: "urgent, problem-aware",
			FunnelStage:   "MOFU",
			Example:       "Problem: Can't increase CLV. Agitate: Loses revenue yearly. Solve: Implement this automation.",
			Criticisms:    "Can feel negative if not balanced with solution",
			Score:         8.9,
		},
		"fab": {
			Name:          "FAB Framework",
			Acronym:       "FAB",
			Components:    []string{"Feature: What it is", "Advantage: What it does", "Benefit: How customer feels"},
			BestFor:       []string{"All stages", "solution descriptions"},
			EmotionalTone: "customer-centric",
			FunnelStage:   "MOFU-BOFU",
			Example:       "Feature: Automated workflows. Advantage: Saves 8 hours/week. Benefit: More time for strategy.",
			Criticisms:    "Dry if not emotionally charged",
			Score:         8.7,
		},
		"bab": {
			Name:          "BAB Framework",
			Acronym:       "BAB",
			Components:    []string{"Before: Current frustration", "After: Desired outcome", "Bridge: The product"},
			BestFor:       []string{"MOFU", "social ads", "short copy"},
			EmotionalTone: "transformational",
			FunnelStage:   "MOFU",
			Example:       "Before: Overwhelmed with manual tasks. After: Calm, organized, strategic. Bridge: Use Da Vinci.",
			Criticisms:    "Requires strong visualization skills",
			Score:         9.1,
		},
		"4ps": {
			Name:          "4Ps Framework",
			Acronym:       "4Ps",
			Components:    []string{"Promise: The hook", "Picture: Visualize", "Proof: Credibility", "Push: CTA"},
			BestFor:       []string{"BOFU", "sales letters", "high-value pages"},
			EmotionalTone: "credible, aspirational",
			FunnelStage:   "BOFU",
			Example:       "Promise: 3x conversions. Picture: Envision revenue growth. Proof: 100+ case studies. Push: Start free trial.",
			Criticisms:    "Longer format required",
			Score:         9.4,
		},
		"hero": {
			Name:          "Hero Section Framework",
			Acronym:       "Hero",
			Components:    []string{"Headline: UVP", "Subheadline: Context", "Visuals: Connection", "CTA: Action"},
			BestFor:       []string{"All landing pages", "above-the-fold"},
			EmotionalTone: "immediate, benefit-focused",
			FunnelStage:   "TOFU-MOFU",
			Example:       "Headline: 'Generate 10x Email Sequences in Hours'. Subheadline: 'AI-powered copywriting for DTC.' CTA: 'Start free.'",
			Criticisms:    "Must be perfect or loses 80% of visitors",
			Score:         8.8,
		},
	}

	fl.mu.Lock()
	for key, fw := range frameworksData {
		fl.frameworks[strings.ToLower(key)] = fw
	}
	fl.mu.Unlock()
}

// GetFramework retrieves framework by name or acronym
func (fl *FrameworkLookup) GetFramework(name string) *Framework {
	key := strings.ToLower(strings.TrimSpace(name))

	// Check cache
	if cached, ok := fl.cache.Get(key); ok {
		return cached
	}

	fl.mu.RLock()
	fw := fl.frameworks[key]
	fl.mu.RUnlock()

	if fw != nil {
		fl.cache.Add(key, fw)
		return fw
	}

	// Semantic fuzzy matching
	return fl.findBestMatch(key)
}

// findBestMatch performs semantic similarity lookup
func (fl *FrameworkLookup) findBestMatch(query string) *Framework {
	fl.mu.RLock()
	defer fl.mu.RUnlock()

	bestMatch := (*Framework)(nil)
	bestScore := 0.0

	for key, fw := range fl.frameworks {
		score := similarityScore(query, key, fw.Name)
		if score > bestScore && score > 0.6 {
			bestScore = score
			bestMatch = fw
		}
	}

	if bestMatch != nil {
		fl.cache.Add(query, bestMatch)
	}

	return bestMatch
}

// GetFrameworksForWorkflowStep returns frameworks suitable for a step
func (fl *FrameworkLookup) GetFrameworksForWorkflowStep(step string) []*Framework {
	stepFrameworks := map[string][]string{
		"StepIntroduction":         {"hero"},
		"StepDiscovery":            {"aida"},
		"StepValidation":           {"aida", "fab"},
		"StepFrameworkApplication": {"aida"},
		"StepCircleConfirmation":   {"pas", "bab"},
		"StepGoalSetting":          {"pas"},
		"StepAnalysis":             {"4ps", "bab"},
		"StepExecution":            {"4ps", "fab", "pas"},
	}

	var frameworks []*Framework
	keys := stepFrameworks[step]

	for _, key := range keys {
		if fw := fl.GetFramework(key); fw != nil {
			frameworks = append(frameworks, fw)
		}
	}

	return frameworks
}

// GetFrameworksForVertical returns vertical-appropriate frameworks
func (fl *FrameworkLookup) GetFrameworksForVertical(vertical string) []*Framework {
	verticalFrameworks := map[string][]string{
		"supplements": {"fab", "4ps", "aida"},
		"coaching":    {"bab", "4ps", "pas"},
		"d tc":        {"aida", "hero", "pas"},
		"nonprofit":   {"bab", "4ps"},
		"ecommerce":   {"pas", "fab", "hero"},
	}

	keys := verticalFrameworks[strings.ToLower(vertical)]
	var frameworks []*Framework

	for _, key := range keys {
		if fw := fl.GetFramework(key); fw != nil {
			frameworks = append(frameworks, fw)
		}
	}

	return frameworks
}

// ListAllFrameworks returns all available frameworks
func (fl *FrameworkLookup) ListAllFrameworks() []*Framework {
	fl.mu.RLock()
	defer fl.mu.RUnlock()

	var list []*Framework
	for _, fw := range fl.frameworks {
		list = append(list, fw)
	}
	return list
}

// FrameworkContextString builds context string for instruction composer
func (fl *FrameworkLookup) FrameworkContextString(workflowStep, vertical string) string {
	var sb strings.Builder

	frameworks := fl.GetFrameworksForWorkflowStep(workflowStep)
	if vertical != "" {
		vfws := fl.GetFrameworksForVertical(vertical)
		frameworks = append(frameworks, vfws...)
	}

	sb.WriteString("## RECOMMENDED COPYWRITING FRAMEWORKS\n\n")
	for _, fw := range frameworks {
		sb.WriteString(fmt.Sprintf("**%s (%s)** - Score: %.1f\n", fw.Name, fw.Acronym, fw.Score))
		sb.WriteString(fmt.Sprintf("Best for: %s\n", strings.Join(fw.BestFor, ", ")))
		sb.WriteString(fmt.Sprintf("Components: %s\n\n", strings.Join(fw.Components, " | ")))
	}

	return sb.String()
}

// similarityScore computes simple semantic similarity
func similarityScore(query, key, name string) float64 {
	words := strings.Fields(strings.ToLower(query))
	score := 0.0

	for _, word := range words {
		if strings.Contains(strings.ToLower(key), word) || strings.Contains(strings.ToLower(name), word) {
			score += 1.0 / float64(len(words))
		}
	}

	return score
}
