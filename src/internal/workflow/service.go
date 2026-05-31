package workflow

import (
	"context"
)

type StateService interface {
	State(ctx context.Context, root, slug string) (FeatureState, error)
	States(ctx context.Context, root string) ([]FeatureState, error)
}

type CheckService interface {
	CheckConstitution(ctx context.Context, root, constitutionPath string) (CheckResult, error)
	CheckSpecReady(ctx context.Context, root string) (CheckResult, error)
	CheckSpecReadyForSlug(ctx context.Context, root, slug string) (CheckResult, error)
	CheckInspectReady(ctx context.Context, root, slug string) (CheckResult, error)
	CheckPlanReady(ctx context.Context, root, slug string) (CheckResult, error)
	CheckTasksReady(ctx context.Context, root, slug string) (CheckResult, error)
	CheckImplementReady(ctx context.Context, root, slug string) (CheckResult, error)
	CheckVerifyReady(ctx context.Context, root, slug string) (CheckResult, error)
	CheckArchiveReady(ctx context.Context, root, slug, status, reason string) (CheckResult, error)
	VerifyTaskState(ctx context.Context, root, slug string) (CheckResult, TaskStateSummary, error)
	InspectSpec(ctx context.Context, root, specPath, tasksPath string) (CheckResult, error)
}

type ValidateService interface {
	ValidateProject(ctx context.Context, root string) ([]Finding, error)
	ValidateFeature(ctx context.Context, root, slug string) ([]Finding, error)
}

type RepairService interface {
	RepairFeature(ctx context.Context, root, slug string, dryRun bool) (RepairResult, error)
	MigrateProject(ctx context.Context, root string, dryRun, copyWorkspace bool) (MigrationResult, error)
}

type ReportService interface {
	ParseReport(ctx context.Context, path string) (Report, error)
}

type Service interface {
	StateService
	CheckService
	ValidateService
	RepairService
	ReportService
}

type service struct{}

func NewService() Service {
	return &service{}
}

func (s *service) State(ctx context.Context, root, slug string) (FeatureState, error) {
	return State(ctx, root, slug)
}

func (s *service) States(ctx context.Context, root string) ([]FeatureState, error) {
	return States(ctx, root)
}

func (s *service) CheckConstitution(ctx context.Context, root, constitutionPath string) (CheckResult, error) {
	return CheckConstitution(ctx, root, constitutionPath)
}

func (s *service) CheckSpecReady(ctx context.Context, root string) (CheckResult, error) {
	return CheckSpecReady(ctx, root)
}

func (s *service) CheckSpecReadyForSlug(ctx context.Context, root, slug string) (CheckResult, error) {
	return CheckSpecReadyForSlug(ctx, root, slug)
}

func (s *service) CheckInspectReady(ctx context.Context, root, slug string) (CheckResult, error) {
	return CheckInspectReady(ctx, root, slug)
}

func (s *service) CheckPlanReady(ctx context.Context, root, slug string) (CheckResult, error) {
	return CheckPlanReady(ctx, root, slug)
}

func (s *service) CheckTasksReady(ctx context.Context, root, slug string) (CheckResult, error) {
	return CheckTasksReady(ctx, root, slug)
}

func (s *service) CheckImplementReady(ctx context.Context, root, slug string) (CheckResult, error) {
	return CheckImplementReady(ctx, root, slug)
}

func (s *service) CheckVerifyReady(ctx context.Context, root, slug string) (CheckResult, error) {
	return CheckVerifyReady(ctx, root, slug)
}

func (s *service) CheckArchiveReady(ctx context.Context, root, slug, status, reason string) (CheckResult, error) {
	return CheckArchiveReady(ctx, root, slug, status, reason)
}

func (s *service) VerifyTaskState(ctx context.Context, root, slug string) (CheckResult, TaskStateSummary, error) {
	return VerifyTaskState(ctx, root, slug)
}

func (s *service) InspectSpec(ctx context.Context, root, specPath, tasksPath string) (CheckResult, error) {
	return InspectSpec(ctx, root, specPath, tasksPath)
}

func (s *service) ValidateProject(ctx context.Context, root string) ([]Finding, error) {
	return ValidateProject(ctx, root)
}

func (s *service) ValidateFeature(ctx context.Context, root, slug string) ([]Finding, error) {
	return ValidateFeature(ctx, root, slug)
}

func (s *service) RepairFeature(ctx context.Context, root, slug string, dryRun bool) (RepairResult, error) {
	return RepairFeature(ctx, root, slug, dryRun)
}

func (s *service) MigrateProject(ctx context.Context, root string, dryRun, copyWorkspace bool) (MigrationResult, error) {
	return MigrateProject(ctx, root, dryRun, copyWorkspace)
}

func (s *service) ParseReport(ctx context.Context, path string) (Report, error) {
	return ParseReport(ctx, path)
}
