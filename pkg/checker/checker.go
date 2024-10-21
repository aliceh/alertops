package checker

import (
	tea "github.com/charmbracelet/bubbletea"
)

type choices struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
}

func InitialModel() choices {
	return choices{

		choices:  []string{"PruningCronjobErrorSRE", "ClusterMonitoringErrorBudgetBurnSRE", "UpgradeNodeUpgradeTimeoutSRE"},
		selected: make(map[int]struct{}),
	}
}

func (c choices) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}
