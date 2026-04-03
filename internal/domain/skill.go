package domain

// Skill represents injectable context for AI agents.
// A skill pairs a runbook.yaml (type: skill) with a skill.md markdown file.
type Skill struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Trigger     string `json:"trigger,omitempty"`
	Content     string `json:"-"`     // raw skill.md markdown (not serialized in JSON APIs)
	Catalog     string `json:"catalog"` // parent catalog name
}
