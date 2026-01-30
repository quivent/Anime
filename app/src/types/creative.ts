// Writing Tab Types
export interface Document {
  id: string
  title: string
  content: string
  created_at: string
  updated_at: string
  word_count: number
  tags: string[]
}

export type WritingMode = 'continuation' | 'dialogue' | 'scene' | 'outline'

export interface WritingRequest {
  mode: WritingMode
  context: string
  prompt: string
  max_tokens?: number
}

export interface WritingResponse {
  generated_text: string
  usage: {
    prompt_tokens: number
    completion_tokens: number
    total_tokens: number
  }
}

// Analysis Tab Types
export type AnalysisType = 'character' | 'plot' | 'dialogue' | 'pacing' | 'theme'

export interface AnalysisRequest {
  type: AnalysisType
  content: string
  options?: Record<string, any>
}

export interface CharacterAnalysis {
  characters: Array<{
    name: string
    role: string
    traits: string[]
    arc: string
    dialogue_count: number
    importance_score: number
  }>
}

export interface PlotAnalysis {
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

export interface DialogueAnalysis {
  total_lines: number
  avg_length: number
  unique_voices: number
  readability_score: number
  suggestions: string[]
}

export interface PacingAnalysis {
  beats_per_scene: number[]
  scene_lengths: number[]
  tension_curve: number[]
  recommendations: string[]
}

export interface ThemeAnalysis {
  primary_themes: string[]
  recurring_motifs: string[]
  symbolism: Array<{
    symbol: string
    meaning: string
    occurrences: number
  }>
}

export type AnalysisResult =
  | { type: 'character'; data: CharacterAnalysis }
  | { type: 'plot'; data: PlotAnalysis }
  | { type: 'dialogue'; data: DialogueAnalysis }
  | { type: 'pacing'; data: PacingAnalysis }
  | { type: 'theme'; data: ThemeAnalysis }

// Storyboard Tab Types
export interface StoryboardScene {
  id: string
  scene_number: number
  title: string
  description: string
  duration?: string
  location?: string
  time?: string
}

export interface StoryboardPanel {
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

export interface ScriptParseResult {
  scenes: StoryboardScene[]
  total_scenes: number
}

export interface StoryboardProject {
  id: string
  name: string
  created_at: string
  updated_at: string
  script: string
  scenes: StoryboardScene[]
  panels: StoryboardPanel[]
}

export interface ShotSuggestion {
  type: string
  description: string
  composition: string
  camera_angle: string
  lighting: string
}
