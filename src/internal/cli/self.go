package cli

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newSelfCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "self",
		Short: "Manage the speckeep binary (check/upgrade)",
		Long: `Check for the latest release and upgrade speckeep in place.

  speckeep self check    — report current vs latest release (read-only)
  speckeep self upgrade  — download the latest release and replace this binary
                           (verifies the sha256 checksum before replacing)`,
	}
	cmd.AddCommand(newSelfCheckCmd())
	cmd.AddCommand(newSelfUpgradeCmd())
	return cmd
}

func newSelfCheckCmd() *cobra.Command {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Report the installed version vs the latest GitHub release",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			latest, err := fetchLatestTag(context.Background())
			if err != nil {
				return fmt.Errorf("check update: %w", err)
			}
			status := planUpgrade(Version, latest)
			if jsonOutput {
				payload, err := json.MarshalIndent(status, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(payload))
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), status.done())
				if status.Update {
					fmt.Fprintln(cmd.OutOrStdout(), "next: speckeep self upgrade")
				} else if status.IsDev {
					fmt.Fprintln(cmd.OutOrStdout(), "note: dev build — run `speckeep self upgrade` to install a release build")
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

func newSelfUpgradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Replace the installed speckeep binary with the latest release",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			latest, err := fetchLatestTag(context.Background())
			if err != nil {
				return fmt.Errorf("check update: %w", err)
			}
			status := planUpgrade(Version, latest)
			if !status.Update {
				fmt.Fprintf(cmd.OutOrStdout(), "already up to date: %s\n", status.Current)
				return nil
			}
			exe, err := upgradeInPlace(context.Background(), status.Latest)
			if err != nil {
				fmt.Fprintln(cmd.OutOrStdout(), "manual upgrade:")
				fmt.Fprintln(cmd.OutOrStdout(), packageManagerHint())
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "upgraded to %s at %s\n", status.Latest, exe)
			return nil
		},
	}
	return cmd
}

func tagNameFromReleaseJSON(body []byte) (string, bool) {
	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", false
	}
	return strings.TrimSpace(payload.TagName), payload.TagName != ""
}

func fileSHA256(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:]), nil
}
