package status

import (
	"context"

	"speckeep/src/internal/workflow"
)

type Service interface {
	Check(ctx context.Context, root, slug string) (Result, error)
	List(ctx context.Context, root string) ([]Result, error)
}

type service struct{}

func NewService() Service {
	return &service{}
}

func (s *service) Check(ctx context.Context, root, slug string) (Result, error) {
	return Check(ctx, root, slug)
}

func (s *service) List(ctx context.Context, root string) ([]Result, error) {
	return List(ctx, root)
}

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

func Check(ctx context.Context, root, slug string) (Result, error) {
	state, err := workflow.State(ctx, root, slug)
	if err != nil {
		return Result{}, err
	}
	return fromFeatureState(state), nil
}

func List(ctx context.Context, root string) ([]Result, error) {
	states, err := workflow.States(ctx, root)
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
