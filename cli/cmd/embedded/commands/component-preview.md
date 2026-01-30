Open a component/template from ~/.claude/templates/ for preview using the system default application.

Usage: `/user:component-preview COMPONENT_NAME`

Examples:
- `/user:component-preview Header.tsx`
- `/user:component-preview progress-spec.md`

Process:
1. Look for the component in ~/.claude/templates/
2. Use bash `open` command to launch with system default app
3. If component doesn't exist, show available components

This allows quick preview of stored components without editing or recreating them.

$ARGUMENTS