Create templates with different composition modes.

Usage: 
- `/user:create-template NAME [--raw] CONTENT` (default: raw content)
- `/user:create-template NAME --compose COMP1 COMP2...` (compose from existing components)

Modes:

**--raw (default)**: Same as create-component but marked as template in metadata. Stores raw content with format detection, preview, and confirmation.

**--compose**: Concatenates existing components/templates in sequence order. Validates all components exist, reads and combines content with separators, previews result, and saves with composition metadata.

Examples:
- `/user:create-template progress-header "$MARKDOWN"`
- `/user:create-template full-spec --compose header.md progress.md footer.md`

Process:
1. Parse mode flag (--raw default, --compose explicit)
2. Raw mode: identical to create-component workflow but marked as template
3. Compose mode: validate components exist → read and concatenate → preview → save with composition metadata
4. All modes use same preview/confirmation UX

Storage: Same as create-component but metadata includes template type and composition details.

$ARGUMENTS