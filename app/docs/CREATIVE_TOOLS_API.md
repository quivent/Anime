# Creative Tools API Reference

## Quick Start

All creative tools are accessible through the sidebar navigation. Each tab provides full backend integration with the Rust backend.

---

## Writing Tab API

### Frontend Component

```tsx
import WritingView from './components/WritingView'
```

### Backend Commands

#### List Documents

```typescript
import { invoke } from '@tauri-apps/api/core'
import type { Document } from './types/creative'

const docs = await invoke<Document[]>('list_documents')
```

#### Create Document

```typescript
const doc = await invoke<Document>('create_document', {
  title: 'My Story'
})
```

#### Save Document

```typescript
const updated = await invoke<Document>('save_document', {
  id: docId,
  content: 'Story content here...'
})
```

#### Delete Document

```typescript
await invoke('delete_document', { id: docId })
```

#### Export Document

```typescript
await invoke('export_document', { id: docId })
// Exports to ~/Downloads/{title}.txt
```

#### Generate Text

```typescript
import type { WritingResponse } from './types/creative'

const response = await invoke<WritingResponse>('generate_text', {
  mode: 'continuation', // 'continuation' | 'dialogue' | 'scene' | 'outline'
  context: 'Previous story content...',
  prompt: 'Continue with...',
  maxTokens: 500
})

console.log(response.generated_text)
console.log(response.usage) // Token usage stats
```

### Types

```typescript
interface Document {
  id: string
  title: string
  content: string
  created_at: string
  updated_at: string
  word_count: number
  tags: string[]
}

interface WritingResponse {
  generated_text: string
  usage: {
    prompt_tokens: number
    completion_tokens: number
    total_tokens: number
  }
}
```

---

## Analysis Tab API

### Frontend Component

```tsx
import AnalysisView from './components/AnalysisView'
```

### Backend Commands

#### Read File

```typescript
const content = await invoke<string>('read_file', {
  path: '/path/to/script.txt'
})
```

#### Analyze Content

```typescript
import type { AnalysisResult } from './types/creative'

const result = await invoke<AnalysisResult>('analyze_content', {
  type: 'character', // 'character' | 'plot' | 'dialogue' | 'pacing' | 'theme'
  content: scriptContent
})

// Result is tagged union
if (result.type === 'character') {
  console.log(result.data.characters) // Character[]
}
```

#### Export Analysis

```typescript
await invoke('export_analysis', {
  result: analysisResult,
  fileName: 'my-script'
})
// Exports to ~/Downloads/my-script_analysis.json
```

### Types

```typescript
type AnalysisType = 'character' | 'plot' | 'dialogue' | 'pacing' | 'theme'

// Character Analysis
interface CharacterAnalysis {
  characters: Array<{
    name: string
    role: string
    traits: string[]
    arc: string
    dialogue_count: number
    importance_score: number
  }>
}

// Plot Analysis
interface PlotAnalysis {
  structure: {
    act_breakdown: string[]
    turning_points: string[]
    conflicts: string[]
  }
  pacing: {
    slow_sections: number[]
    fast_sections: number[]
    overall_score: number
  }
}

// Dialogue Analysis
interface DialogueAnalysis {
  total_lines: number
  avg_length: number
  unique_voices: number
  readability_score: number
  suggestions: string[]
}

// Pacing Analysis
interface PacingAnalysis {
  beats_per_scene: number[]
  scene_lengths: number[]
  tension_curve: number[]
  recommendations: string[]
}

// Theme Analysis
interface ThemeAnalysis {
  primary_themes: string[]
  recurring_motifs: string[]
  symbolism: Array<{
    symbol: string
    meaning: string
    occurrences: number
  }>
}
```

---

## Storyboards Tab API

### Frontend Component

```tsx
import StoryboardsView from './components/StoryboardsView'
```

### Backend Commands

#### Create Storyboard Project

```typescript
import type { StoryboardProject } from './types/creative'

const project = await invoke<StoryboardProject>('create_storyboard_project', {
  name: 'My Film Project'
})
```

#### Parse Script

```typescript
import type { ScriptParseResult } from './types/creative'

const result = await invoke<ScriptParseResult>('parse_script', {
  script: scriptContent
})

console.log(result.scenes) // StoryboardScene[]
console.log(result.total_scenes) // number
```

#### Generate Shot Suggestions

```typescript
import type { ShotSuggestion } from './types/creative'

const suggestions = await invoke<ShotSuggestion[]>('generate_shot_suggestions', {
  projectId: project.id,
  sceneId: scene.id
})

suggestions.forEach(shot => {
  console.log(shot.type) // "Wide Shot", "Close-up", etc.
  console.log(shot.composition) // "Rule of thirds", etc.
  console.log(shot.camera_angle) // "Eye level", etc.
  console.log(shot.lighting) // "Natural daylight", etc.
})
```

#### Generate Storyboard Image

```typescript
const imageUrl = await invoke<string>('generate_storyboard_image', {
  description: panel.description,
  shotType: panel.shot_type,
  composition: panel.composition
})

// Currently returns placeholder, ready for ComfyUI integration
```

#### Export Storyboard

```typescript
await invoke('export_storyboard', {
  projectId: project.id,
  format: 'pdf' // 'pdf' | 'images'
})
```

### Types

```typescript
interface StoryboardProject {
  id: string
  name: string
  created_at: string
  updated_at: string
  script: string
  scenes: StoryboardScene[]
  panels: StoryboardPanel[]
}

interface StoryboardScene {
  id: string
  scene_number: number
  title: string
  description: string
  duration?: string
  location?: string
  time?: string
}

interface StoryboardPanel {
  id: string
  scene_id: string
  panel_number: number
  shot_type: string
  composition: string
  description: string
  dialogue?: string
  image_url?: string
  notes?: string
}

interface ShotSuggestion {
  type: string
  description: string
  composition: string
  camera_angle: string
  lighting: string
}
```

---

## File Dialog API

### Using Tauri Dialog Plugin

```typescript
import { open } from '@tauri-apps/plugin-dialog'

// Single file
const selected = await open({
  multiple: false,
  filters: [{
    name: 'Text Files',
    extensions: ['txt', 'md', 'pdf', 'docx']
  }]
})

if (selected && typeof selected === 'string') {
  const content = await invoke<string>('read_file', { path: selected })
}

// Multiple files
const files = await open({
  multiple: true,
  filters: [{
    name: 'Scripts',
    extensions: ['fountain', 'txt']
  }]
})
```

---

## Storage Locations

### Documents

- Path: `~/.anime-desktop/documents/`
- Format: `{document-id}.json`
- Structure: Full `Document` object

### Storyboard Projects

- Path: `~/.anime-desktop/storyboards/`
- Format: `{project-id}.json`
- Structure: Full `StoryboardProject` object

### Exports

- Path: `~/Downloads/`
- Formats:
  - Documents: `{title}.txt`
  - Analysis: `{fileName}_analysis.json`
  - Storyboards: `{projectName}_storyboard.pdf`

---

## Error Handling

All commands can throw errors. Always use try-catch:

```typescript
try {
  const doc = await invoke<Document>('create_document', {
    title: 'My Story'
  })
} catch (error) {
  console.error('Failed to create document:', error)
  // Show error to user
}
```

---

## Integrating with Real AI

### Writing Generation

Replace mock implementation in `creative.rs`:

```rust
#[tauri::command]
pub async fn generate_text(
    mode: String,
    context: String,
    prompt: String,
    max_tokens: Option<usize>,
) -> Result<WritingResponse, String> {
    // Call your LLM API here
    let client = reqwest::Client::new();
    let response = client
        .post("https://api.anthropic.com/v1/messages")
        .header("x-api-key", env::var("ANTHROPIC_API_KEY")?)
        .json(&json!({
            "model": "claude-3-5-sonnet-20241022",
            "max_tokens": max_tokens.unwrap_or(500),
            "messages": [{
                "role": "user",
                "content": format!("{}\n\n{}", context, prompt)
            }]
        }))
        .send()
        .await?;

    // Parse and return
}
```

### Content Analysis

Replace mock implementation:

```rust
#[tauri::command]
pub async fn analyze_content(
    r#type: String,
    content: String,
) -> Result<serde_json::Value, String> {
    // Use NLP library or LLM API
    match r#type.as_str() {
        "character" => {
            // Extract characters using NLP
        }
        "plot" => {
            // Analyze structure
        }
        // ...
    }
}
```

### Image Generation

Replace placeholder in `creative.rs`:

```rust
#[tauri::command]
pub async fn generate_storyboard_image(
    description: String,
    shot_type: String,
    composition: String,
) -> Result<String, String> {
    // Call ComfyUI workflow
    let workflow = create_storyboard_workflow(
        &description,
        &shot_type,
        &composition
    )?;

    let image_path = execute_comfyui_workflow(workflow).await?;
    Ok(image_path)
}
```

---

## Usage Examples

### Complete Writing Flow

```typescript
// 1. Create document
const doc = await invoke<Document>('create_document', {
  title: 'Epic Fantasy Novel'
})

// 2. Generate opening
const opening = await invoke<WritingResponse>('generate_text', {
  mode: 'scene',
  context: '',
  prompt: 'A mysterious traveler arrives in a medieval village at sunset'
})

// 3. Save content
await invoke<Document>('save_document', {
  id: doc.id,
  content: opening.generated_text
})

// 4. Continue story
const continuation = await invoke<WritingResponse>('generate_text', {
  mode: 'continuation',
  context: opening.generated_text,
  prompt: 'The traveler asks about the old castle'
})

// 5. Export final version
await invoke('export_document', { id: doc.id })
```

### Complete Analysis Flow

```typescript
// 1. Upload file
const content = await invoke<string>('read_file', {
  path: selectedFilePath
})

// 2. Run multiple analyses
const charAnalysis = await invoke('analyze_content', {
  type: 'character',
  content
})

const plotAnalysis = await invoke('analyze_content', {
  type: 'plot',
  content
})

// 3. Export combined report
const report = {
  characters: charAnalysis,
  plot: plotAnalysis,
  timestamp: new Date().toISOString()
}

await invoke('export_analysis', {
  result: report,
  fileName: 'full-analysis'
})
```

### Complete Storyboard Flow

```typescript
// 1. Create project
const project = await invoke<StoryboardProject>('create_storyboard_project', {
  name: 'Short Film - The Journey'
})

// 2. Parse script
const parsed = await invoke<ScriptParseResult>('parse_script', {
  script: uploadedScript
})

// 3. For each scene, generate shots
for (const scene of parsed.scenes) {
  const shots = await invoke<ShotSuggestion[]>('generate_shot_suggestions', {
    projectId: project.id,
    sceneId: scene.id
  })

  // 4. Generate images for each shot
  for (const shot of shots) {
    const imageUrl = await invoke<string>('generate_storyboard_image', {
      description: shot.description,
      shotType: shot.type,
      composition: shot.composition
    })

    // Store panel with image
  }
}

// 5. Export final storyboard
await invoke('export_storyboard', {
  projectId: project.id,
  format: 'pdf'
})
```

---

## Design System Integration

All components use ANIME design tokens:

```tsx
// Color classes
className="text-electric-400"     // Primary accent
className="text-mint-400"          // Success
className="text-sakura-400"        // Highlight
className="text-sunset-400"        // Warning
className="text-neon-400"          // Secondary

// Background classes
className="bg-electric-500/20"     // Subtle background
className="bg-mint-500/10"         // Success background

// Border classes
className="border-electric-500/50" // Accent border
className="border-gray-700"        // Neutral border

// Effects
className="anime-glow"             // Glow animation
className="backdrop-blur-md"       // Blur effect
```

---

## Best Practices

### Performance

- Use `useMemo` for expensive computations
- Debounce AI generation requests
- Show loading states for all async operations
- Stream large file uploads

### UX

- Always show loading indicators
- Provide clear error messages
- Enable keyboard shortcuts
- Auto-save drafts
- Confirm destructive actions

### Security

- Validate all file paths
- Sanitize user input
- Never expose API keys in frontend
- Use environment variables for secrets

---

## Troubleshooting

### Document Not Saving

```typescript
// Check app data directory exists
const docs = await invoke<Document[]>('list_documents')
if (docs.length === 0) {
  // First time - directory will be created
}
```

### File Upload Failing

```typescript
// Ensure dialog plugin is initialized
import { open } from '@tauri-apps/plugin-dialog'

// Check file permissions
const content = await invoke<string>('read_file', { path })
  .catch(err => {
    console.error('Permission denied or file not found')
  })
```

### Image Generation Not Working

```typescript
// Placeholder URLs currently used
// To enable real generation:
// 1. Set up ComfyUI backend
// 2. Update generate_storyboard_image in creative.rs
// 3. Configure image model paths
```

---

## Next Steps

1. **Replace Mock AI**: Integrate real LLM APIs
2. **Add Rich Text**: Upgrade to TipTap editor
3. **ComfyUI Integration**: Connect image generation
4. **Cloud Sync**: Optional cloud backup
5. **Collaboration**: Real-time co-editing

---

## Support

For issues or questions:
- Check implementation: `/src-tauri/src/creative.rs`
- Review types: `/src/types/creative.ts`
- See examples: Component source files
