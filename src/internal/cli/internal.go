package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"speckeep/src/internal/workflow"
)

func newInternalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "__internal",
		Short:         "Internal SpecKeep helpers",
		Hidden:        true,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(newInternalCheckConstitutionCmd())
	cmd.AddCommand(newInternalCheckSpecReadyCmd())
	cmd.AddCommand(newInternalCheckInspectReadyCmd())
	cmd.AddCommand(newInternalCheckPlanReadyCmd())
	cmd.AddCommand(newInternalCheckTasksReadyCmd())
	cmd.AddCommand(newInternalCheckImplementReadyCmd())
	cmd.AddCommand(newInternalCheckVerifyReadyCmd())
	cmd.AddCommand(newInternalCheckArchiveReadyCmd())
	cmd.AddCommand(newInternalInspectSpecCmd())
	cmd.AddCommand(newInternalVerifyTaskStateCmd())
	cmd.AddCommand(newInternalListOpenTasksCmd())
	cmd.AddCommand(newInternalListSpecsCmd())
	cmd.AddCommand(newInternalShowSpecCmd())
	cmd.AddCommand(newInternalLinkAgentsCmd())

	return cmd
}

func newInternalCheckConstitutionCmd() *cobra.Command {
	var root string
	cmd := &cobra.Command{
		Use:           "check-constitution [constitution-file]",
		Hidden:        true,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			constitutionPath := ""
			if len(args) == 1 {
				constitutionPath = args[0]
			}
			result, err := workflow.CheckConstitution(root, constitutionPath)
			return renderCheckResult(cmd, result, err)
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "SpecKeep project root")
	return cmd
}

func newInternalCheckSpecReadyCmd() *cobra.Command {
	var root string
	cmd := &cobra.Command{
		Use:           "check-spec-ready [slug]",
		Hidden:        true,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := ""
			if len(args) == 1 {
				slug = args[0]
			}
			result, err := workflow.CheckSpecReadyForSlug(root, slug)
			return renderCheckResult(cmd, result, err)
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "SpecKeep project root")
	return cmd
}

func newInternalCheckInspectReadyCmd() *cobra.Command {
	var root string
	cmd := &cobra.Command{
		Use:           "check-inspect-ready <slug>",
		Hidden:        true,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := workflow.CheckInspectReady(root, args[0])
			return renderCheckResult(cmd, result, err)
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "SpecKeep project root")
	return cmd
}

func newInternalCheckPlanReadyCmd() *cobra.Command {
	var root string
	cmd := &cobra.Command{
		Use:           "check-plan-ready <slug>",
		Hidden:        true,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := workflow.CheckPlanReady(root, args[0])
			return renderCheckResult(cmd, result, err)
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "SpecKeep project root")
	return cmd
}

func newInternalCheckTasksReadyCmd() *cobra.Command {
	var root string
	cmd := &cobra.Command{
		Use:           "check-tasks-ready <slug>",
		Hidden:        true,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := workflow.CheckTasksReady(root, args[0])
			return renderCheckResult(cmd, result, err)
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "SpecKeep project root")
	return cmd
}

func newInternalCheckImplementReadyCmd() *cobra.Command {
	var root string
	cmd := &cobra.Command{
		Use:           "check-implement-ready <slug>",
		Hidden:        true,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := workflow.CheckImplementReady(root, args[0])
			return renderCheckResult(cmd, result, err)
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "SpecKeep project root")
	return cmd
}

func newInternalCheckVerifyReadyCmd() *cobra.Command {
	var root string
	cmd := &cobra.Command{
		Use:           "check-verify-ready <slug>",
		Hidden:        true,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := workflow.CheckVerifyReady(root, args[0])
			return renderCheckResult(cmd, result, err)
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "SpecKeep project root")
	return cmd
}

func newInternalCheckArchiveReadyCmd() *cobra.Command {
	var root string
	cmd := &cobra.Command{
		Use:           "check-archive-ready <slug> <status> [reason]",
		Hidden:        true,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			reason := ""
			if len(args) == 3 {
				reason = args[2]
			}
			result, err := workflow.CheckArchiveReady(root, args[0], args[1], reason)
			return renderCheckResult(cmd, result, err)
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "SpecKeep project root")
	return cmd
}

func newInternalInspectSpecCmd() *cobra.Command {
	var root string
	cmd := &cobra.Command{
		Use:           "inspect-spec <spec-file> [tasks-file]",
		Hidden:        true,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			tasksPath := ""
			if len(args) == 2 {
				tasksPath = args[1]
			}
			result, err := workflow.InspectSpec(root, args[0], tasksPath)
			return renderCheckResult(cmd, result, err)
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "SpecKeep project root")
	return cmd
}

func newInternalVerifyTaskStateCmd() *cobra.Command {
	var root string
	cmd := &cobra.Command{
		Use:           "verify-task-state <slug>",
		Hidden:        true,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, _, err := workflow.VerifyTaskState(root, args[0])
			return renderCheckResult(cmd, result, err)
		},
	}
	cmd.Flags().StringVar(&root, "root", ".", "SpecKeep project root")
	return cmd
}

func renderCheckResult(cmd *cobra.Command, result workflow.CheckResult, err error) error {
	if err != nil {
		return err
	}
	for _, line := range result.Lines {
		fmt.Fprintln(cmd.OutOrStdout(), line)
	}
	if result.Failed {
		return newExitError(1, "")
	}
	return nil
}
