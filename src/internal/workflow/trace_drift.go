package workflow

import (
	"context"
	"fmt"
	"strings"

	"speckeep/src/internal/config"
	"speckeep/src/internal/trace"
)

// checkTraceDrift warns when actual changes do not match declared Touches.
// It never blocks: drift is a review signal, not a correctness gate.
func checkTraceDrift(ctx context.Context, result *CheckResult, root, slug string) {
	cfg, err := config.Load(ctx, root)
	if err != nil {
		return
	}
	specsDir, err := cfg.SpecsDir(root)
	if err != nil {
		return
	}
	if _, ok, err := trace.LoadMeta(specsDir, slug); err != nil || !ok {
		return
	}

	changes, err := trace.BuildChanges(ctx, root, slug)
	if err != nil {
		result.AddWarn(fmt.Sprintf("trace: could not compute change attribution: %v", err))
		return
	}
	if len(changes.Unmapped) > 0 {
		preview := changes.Unmapped
		if len(preview) > 5 {
			preview = preview[:5]
		}
		result.AddWarn(fmt.Sprintf("trace: %d changed file(s) are outside declared Touches: %s — review with `speckeep trace %s --files`",
			len(changes.Unmapped), strings.Join(preview, ", "), slug))
	}
	if len(changes.NotChanged) > 0 {
		preview := changes.NotChanged
		if len(preview) > 5 {
			preview = preview[:5]
		}
		result.AddWarn(fmt.Sprintf("trace: %d declared surface(s) were not changed: %s — update Touches or finish the work",
			len(changes.NotChanged), strings.Join(preview, ", ")))
	}
	if len(changes.Unmapped) == 0 && len(changes.NotChanged) == 0 {
		result.AddOK("trace: change attribution matches declared surfaces")
	}
}
