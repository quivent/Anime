export type CoverageStatus = 'draft' | 'in_review' | 'completed' | 'archived'
export type BudgetRange = 'micro' | 'low' | 'medium' | 'high' | 'blockbuster'
export type RecommendationType = 'pass' | 'consider' | 'recommend'

export interface CoverageRatings {
  overall: number // 1-10
  premise: number
  character: number
  dialogue: number
  structure: number
  pacing: number
  marketability: number
  originality: number
  execution: number
}

export interface PageNote {
  id: string
  page: number
  note: string
  category: 'positive' | 'negative' | 'question' | 'suggestion'
}

export interface CharacterBreakdown {
  name: string
  role: 'protagonist' | 'antagonist' | 'supporting' | 'minor'
  arc_rating: number // 1-10
  notes: string
}

export interface CoverageAnalysis {
  strengths: string[]
  weaknesses: string[]
  premise_analysis: string
  character_analysis: string
  dialogue_analysis: string
  structure_analysis: string
  pacing_analysis: string
  theme_analysis: string
  page_notes: PageNote[]
  character_breakdowns: CharacterBreakdown[]
}

export interface CoverageRecommendation {
  type: RecommendationType
  summary: string
  commercial_appeal: number // 1-10
  target_audience: string
  comparable_titles: string[]
  notes: string
}

export interface CoverageReport {
  id: string
  title: string
  script_id?: string
  script_title: string
  script_path?: string

  created_at: string
  updated_at: string
  submitted_by: string
  submitted_date: string

  template_id: string
  template_version: string

  logline: string
  synopsis: string

  ratings: CoverageRatings
  analysis: CoverageAnalysis
  recommendation: CoverageRecommendation

  genre: string[]
  budget_estimate?: BudgetRange

  status: CoverageStatus
  tags: string[]
  version: number
}
