# Quick SpecKeep terminal demo

Output demo/speckeep-demo.gif

Require bash
Require git
Require sed

Set Shell "bash"
Set Theme "Catppuccin Mocha"
Set FontSize 28
Set Width 1200
Set Height 720
Set Padding 20
Set TypingSpeed 40ms

# Clean demo workspace
Type "rm -rf demo/_work && mkdir -p demo/_work && cd demo/_work" Enter
Sleep 700ms

# Initialize a fresh SpecKeep workspace in an empty repo
Type "../../bin/speckeep init . --lang en --shell sh --agents codex --agents claude --agents cursor" Enter
Sleep 1200ms

Type "ls" Enter
Sleep 400ms

Type "ls -la .speckeep | sed -n '1,120p'" Enter
Sleep 700ms

Type "sed -n '1,120p' AGENTS.md" Enter
Sleep 900ms

# Show health checks and a dry-run refresh
Type "../../bin/speckeep doctor ." Enter
Sleep 900ms

Type "../../bin/speckeep refresh . --dry-run" Enter
Sleep 1200ms

Sleep 2s
