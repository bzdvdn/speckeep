package status

import "speckeep/src/internal/workflow"

type Result struct {
	Slug           string `json:"slug"`
	Phase          string `json:"phase"`
	SpecExists     bool   `json:"spec_exists"`
	InspectExists  bool   `json:"inspect_exists"`
	PlanExists     bool   `json:"plan_exists"`
	TasksExists    bool   `json:"tasks_exists"`
	VerifyExists   bool   `json:"verify_exists"`
	Archived       bool   `json:"archived"`
	InspectPath    string `json:"inspect_path,omitempty"`
	InspectLegacy  bool   `json:"inspect_legacy,omitempty"`
	VerifyPath     string `json:"verify_path,omitempty"`
	TasksTotal     int    `json:"tasks_total"`
	TasksCompleted int    `json:"tasks_completed"`
	TasksOpen      int    `json:"tasks_open"`
	ReadyFor       string `json:"ready_for,omitempty"`
	Blocked        bool   `json:"blocked"`
}

func Check(root, slug string) (Result, error) {
	state, err := workflow.State(root, slug)
	if err != nil {
		return Result{}, err
	}
	return fromFeatureState(state), nil
}

func List(root string) ([]Result, error) {
	states, err := workflow.States(root)
	if err != nil {
		return nil, err
	}

	results := make([]Result, 0, len(states))
	for _, state := range states {
		results = append(results, fromFeatureState(state))
	}
	return results, nil
}

func fromFeatureState(state workflow.FeatureState) Result {
	return Result{
		Slug:           state.Slug,
		Phase:          state.Phase,
		SpecExists:     state.SpecExists,
		InspectExists:  state.InspectExists,
		PlanExists:     state.PlanExists,
		TasksExists:    state.TasksExists,
		VerifyExists:   state.VerifyExists,
		Archived:       state.Archived,
		InspectPath:    state.InspectPath,
		InspectLegacy:  state.InspectLegacy,
		VerifyPath:     state.VerifyPath,
		TasksTotal:     state.TasksTotal,
		TasksCompleted: state.TasksCompleted,
		TasksOpen:      state.TasksOpen,
		ReadyFor:       state.ReadyFor,
		Blocked:        state.Blocked,
	}
}
