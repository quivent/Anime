Open a template from ~/.claude/templates/ for preview using the system default application.

Usage: `/user:template-preview TEMPLATE_NAME`

Examples:
- `/user:template-preview progress-header.md`
- `/user:template-preview full-spec.md`

Process:
1. Look for the template in ~/.claude/templates/
2. Use bash `open` command to launch with system default app
3. If template doesn't exist, show available templates

This allows quick preview of stored templates without editing or recreating them.

$ARGUMENTS