package domain

import "fmt"

type RiskLevel string

const (
	RiskLow      RiskLevel = "low"
	RiskMedium   RiskLevel = "medium"
	RiskHigh     RiskLevel = "high"
	RiskCritical RiskLevel = "critical"
)

var riskOrder = map[RiskLevel]int{
	RiskLow:      0,
	RiskMedium:   1,
	RiskHigh:     2,
	RiskCritical: 3,
}

func (r RiskLevel) Exceeds(ceiling RiskLevel) bool {
	return riskOrder[r] > riskOrder[ceiling]
}

func ParseRiskLevel(s string) (RiskLevel, error) {
	switch RiskLevel(s) {
	case RiskLow, RiskMedium, RiskHigh, RiskCritical:
		return RiskLevel(s), nil
	default:
		return "", fmt.Errorf("unknown risk level: %q", s)
	}
}
