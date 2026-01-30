Create a reusable component/template with intelligent format detection and preview.

Usage: `/user:create-component NAME [--EXT] CONTENT`

Examples:
- `/user:create-component Header.tsx "$CODE"`
- `/user:create-component progress --md "$MARKDOWN"`
- `/user:create-component config "$JSON"`

Process:
1. Parse command arguments for name, optional extension flag, and content
2. Detect file format using priority: explicit flag > file extension > content inference > fallback to .txt
3. Create temporary preview file and open it with system default application
4. Wait for user confirmation to save
5. If confirmed, save component to ~/.claude/templates/ and create metadata file
6. Clean up temporary files

Format detection:
- Explicit flags: --js, --ts, --tsx, --md, --json, --sh, --py, --rs, etc.
- Extension inference: analyze filename for .ext
- Content inference: look for syntax markers (```lang, import, #, {, etc.)
- Fallback: save as .txt if unclear

Storage structure:
- Component file: ~/.claude/templates/ComponentName.ext
- Metadata file: ~/.claude/templates/ComponentName.ext.meta.yaml

The metadata includes creation date, detected placeholders, file type, and description for future organization and search.

$ARGUMENTS