// Professional screenplay analysis data structures based on real analysis standards

export interface SceneAnalysis {
  scene_number: number
  scene_heading: string
  page_start: number
  page_end: number
  scene_function: string
  dialogue_quality_rating: number // 1-10
  dialogue_notes: string
  commercial_appeal: {
    festival_circuit: string
    actor_showcase: string
    international: string
  }
  actor_appeal: {
    role_significance: string
    comp_casting: string[]
    oscar_potential: string
  }
  production_considerations: {
    budget_impact: 'LOW' | 'MEDIUM' | 'HIGH'
    shooting_days: string
    location_requirements: string
    vfx_requirements: string
  }
  pacing_notes: string
  marketing_implications: string
  key_exchanges: string[]
  strengths: string[]
  concerns: string[]
}

export interface ActAnalysis {
  act_number: 1 | 2 | 3
  page_range: string
  scenes: number[]
  opening_image?: string
  inciting_incident?: string
  turning_point?: string
  midpoint?: string
  climax?: string
  structural_strengths: string[]
  pacing_observations: string[]
  trim_recommendations?: string[]
}

export interface CharacterAnalysis {
  name: string
  role: 'lead' | 'supporting' | 'minor'
  screen_time_percentage: number
  arc_description: string
  complexity_notes: string
  voice_evolution: {
    act_one: string[]
    act_two: string[]
    act_three: string[]
  }
  performance_demands: string[]
  casting_recommendations: string[]
  development_opportunities?: string[]
}

export interface ThematicAnalysis {
  primary_themes: {
    theme: string
    analysis: string
    sophistication_notes: string
  }[]
  visual_motifs: string[]
  symbolic_elements: string[]
  philosophical_position: string
}

export interface IndustryIntelligence {
  comp_titles: {
    title: string
    year: number
    similarity_percentage: number
    comparison_notes: string
  }[]
  market_position: string
  budget_range: string
  revenue_projection: string
  awards_potential: 'VERY HIGH' | 'HIGH' | 'MODERATE' | 'LOW'
  festival_strategy: string
  target_distributors: string[]
  casting_tier: string
}

export interface StructuralAnalysis {
  page_count: number
  act_structure: '3-act' | '4-act' | '5-act' | 'non-traditional'
  act_breakdowns: ActAnalysis[]
  pacing_rhythm: string
  scene_count: number
  dialogue_to_action_ratio: string
  structural_innovations: string[]
  trim_recommendations: {
    section: string
    pages: string
    rationale: string
  }[]
}

export interface ProfessionalCoverage {
  // Header metadata
  id: string
  title: string
  author: string
  genre: string[]
  setting: string
  page_count: number
  format: 'feature' | 'pilot' | 'limited-series'
  analysis_date: string
  analyst_names: string[]

  // Core analysis
  logline: string
  synopsis: {
    act_one: string
    act_two: string
    act_three: string
    resolution: string
  }

  // Executive summary
  consensus_rating: number // 1-10
  executive_summary: string
  strengths: string[]
  areas_for_development: string[]

  // Deep analysis sections
  structural_analysis: StructuralAnalysis
  character_analyses: CharacterAnalysis[]
  thematic_analysis: ThematicAnalysis
  scene_analyses: SceneAnalysis[]

  // Industry intelligence
  industry_intelligence: IndustryIntelligence

  // Dialogue analysis
  dialogue_analysis: {
    overall_quality: string
    character_voices: {
      character: string
      voice_description: string
      examples: string[]
    }[]
    subtext_examples: {
      scene: number
      exchange: string
      surface: string
      subtext: string
    }[]
  }

  // Visual storytelling
  visual_storytelling: {
    overall_assessment: string
    techniques: string[]
    environment_as_emotion: string[]
    camera_aware_notes: string[]
  }

  // Author analysis
  author_profile?: {
    estimated_age_range: string
    education_indicators: string
    sophistication_level: string
    philosophical_position: string
    industry_readiness: string
  }

  // Final recommendation
  recommendation: {
    verdict: 'PASS' | 'CONSIDER' | 'RECOMMEND'
    summary: string
    next_steps: string[]
  }

  // Metadata
  created_at: string
  updated_at: string
  version: number
  status: 'draft' | 'in_review' | 'completed' | 'archived'
}
