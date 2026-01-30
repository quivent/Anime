# Coverage Analysis Architecture

## Overview

The Coverage Analysis system is designed to help filmmakers and screenwriters track script coverage, organize feedback, and analyze scripts across multiple dimensions. This document outlines the complete architecture for integrating coverage analysis into the ANIME desktop application.

---

## Table of Contents

1. [Data Model](#data-model)
2. [Component Structure](#component-structure)
3. [User Flow](#user-flow)
4. [Visual Design](#visual-design)
5. [Integration Points](#integration-points)
6. [Technical Requirements](#technical-requirements)
7. [Implementation Phases](#implementation-phases)

---

## Data Model

### Core Entities

#### Coverage Report

The primary entity representing a complete coverage analysis of a script.

```typescript
interface CoverageReport {
  id: string
  title: string
  script_id?: string // Optional link to a script document
  script_title: string
  script_path?: string // Path to uploaded script file

  // Metadata
  created_at: string
  updated_at: string
  submitted_by: string // Reader/analyst name
  submitted_date: string

  // Coverage Template
  template_id: string // Which coverage template was used
  template_version: string

  // Core Analysis
  logline: string
  synopsis: string // 1-2 paragraph summary

  // Ratings (1-10 scale)
  ratings: CoverageRatings

  // Detailed Analysis
  analysis: CoverageAnalysis

  // Recommendation
  recommendation: CoverageRecommendation

  // Additional Fields
  genre: string[]
  comparable_titles: string[] // "It's like X meets Y"
  target_audience: string
  budget_estimate?: BudgetRange

  // Tracking
  status: CoverageStatus
  tags: string[]
  version: number // For tracking revisions
}

interface CoverageRatings {
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

interface CoverageAnalysis {
  // Strengths and Weaknesses
  strengths: string[]
  weaknesses: string[]

  // Detailed Sections
  premise_analysis: string
  character_analysis: string
  dialogue_analysis: string
  structure_analysis: string
  pacing_analysis: string
  theme_analysis: string

  // Page-specific notes
  page_notes: PageNote[]

  // Character breakdowns
  characters: CharacterBreakdown[]
}

interface PageNote {
  id: string
  page_number: number
  timestamp?: string // Scene timestamp if applicable
  note_type: 'praise' | 'critique' | 'question' | 'suggestion'
  content: string
  category: string // 'dialogue', 'action', 'structure', etc.
}

interface CharacterBreakdown {
  name: string
  role: 'protagonist' | 'antagonist' | 'supporting' | 'minor'
  description: string
  arc_rating: number // 1-10
  arc_description: string
  strengths: string[]
  weaknesses: string[]
  notes: string
}

interface CoverageRecommendation {
  decision: 'pass' | 'consider' | 'recommend'
  confidence: number // 1-10, how confident in the recommendation
  summary: string // 2-3 sentences explaining the decision
  next_steps: string[] // What should happen next with this script
  revision_notes?: string // If "consider", what needs to change
}

type CoverageStatus = 'draft' | 'in_review' | 'completed' | 'archived'

type BudgetRange =
  | 'micro' // < $1M
  | 'low' // $1M-5M
  | 'medium' // $5M-20M
  | 'high' // $20M-50M
  | 'studio' // > $50M
```

#### Coverage Template

Templates define different coverage formats (e.g., studio coverage, contest coverage, development notes).

```typescript
interface CoverageTemplate {
  id: string
  name: string
  description: string
  version: string

  // Template Configuration
  sections: TemplateSectionConfig[]
  required_fields: string[]
  rating_categories: RatingCategory[]

  // Metadata
  created_at: string
  is_default: boolean
  is_custom: boolean
}

interface TemplateSectionConfig {
  id: string
  title: string
  description: string
  field_name: keyof CoverageAnalysis
  is_required: boolean
  placeholder?: string
  max_length?: number
  guidelines?: string // Help text for this section
}

interface RatingCategory {
  id: string
  name: string
  field_name: keyof CoverageRatings
  description: string
  weight?: number // For weighted overall score
}
```

#### Coverage Comparison

For comparing multiple coverage reports or script revisions.

```typescript
interface CoverageComparison {
  id: string
  name: string
  created_at: string

  // Reports being compared
  reports: string[] // Coverage report IDs

  // Comparison metrics
  rating_changes: RatingChange[]
  common_strengths: string[]
  common_weaknesses: string[]
  divergent_opinions: DivergentOpinion[]

  // Revision tracking (if comparing versions)
  is_revision_comparison: boolean
  revision_summary?: string
}

interface RatingChange {
  report_id: string
  report_title: string
  category: keyof CoverageRatings
  old_value?: number
  new_value: number
  change: number // delta
}

interface DivergentOpinion {
  category: string
  opinions: Array<{
    report_id: string
    report_title: string
    content: string
    rating?: number
  }>
}
```

#### Historical Tracking

```typescript
interface CoverageHistory {
  script_id: string
  script_title: string

  // All coverage reports for this script
  reports: CoverageReport[]

  // Aggregate stats
  stats: {
    total_reports: number
    average_overall_rating: number
    recommendation_breakdown: Record<CoverageRecommendation['decision'], number>
    common_strengths: string[]
    common_weaknesses: string[]
  }

  // Timeline
  timeline: HistoryEvent[]
}

interface HistoryEvent {
  id: string
  timestamp: string
  event_type: 'coverage_submitted' | 'revision_uploaded' | 'status_changed'
  description: string
  metadata?: Record<string, any>
}
```

---

## Component Structure

### Component Hierarchy

```
CoverageView (Main Container)
├── CoverageHeader
│   ├── ViewToggle (List/Grid/Board)
│   ├── FilterBar
│   └── ActionButtons (New Coverage, Import, Export)
│
├── CoverageList (List/Grid View)
│   ├── CoverageCard (for each report)
│   │   ├── CoverageCardHeader
│   │   ├── CoverageCardRatings
│   │   ├── CoverageCardSummary
│   │   └── CoverageCardActions
│   └── EmptyState
│
├── CoverageKanbanBoard (Board View)
│   ├── KanbanColumn (for each status)
│   │   └── CoverageCard[]
│   └── KanbanActions
│
├── CoverageEditor (Create/Edit Form)
│   ├── EditorHeader
│   ├── EditorSidebar
│   │   ├── TemplateSelector
│   │   ├── ScriptUploader
│   │   └── NavigationMenu
│   ├── EditorContent
│   │   ├── BasicInfoSection
│   │   ├── LoglineSynopsisSection
│   │   ├── RatingsSection
│   │   │   └── RatingSlider (for each category)
│   │   ├── AnalysisSection
│   │   │   ├── StrengthsWeaknessesInput
│   │   │   ├── DetailedAnalysisFields
│   │   │   └── PageNotesManager
│   │   ├── CharacterSection
│   │   │   └── CharacterBreakdownCard[]
│   │   └── RecommendationSection
│   └── EditorFooter (Save, Preview, Submit)
│
├── CoverageDetailView (Read-only view)
│   ├── DetailHeader
│   │   ├── ScriptInfo
│   │   ├── SubmissionInfo
│   │   └── ActionMenu (Edit, Export, Compare, Archive)
│   ├── DetailSidebar
│   │   ├── RatingsVisualization
│   │   ├── RecommendationBadge
│   │   └── QuickStats
│   ├── DetailContent
│   │   ├── LoglineSynopsisDisplay
│   │   ├── AnalysisTabPanel
│   │   │   ├── OverviewTab
│   │   │   ├── CharactersTab
│   │   │   ├── StructureTab
│   │   │   └── PageNotesTab
│   │   └── ExportPanel
│   └── RelatedCoverage (other reports on same script)
│
├── CoverageComparisonView
│   ├── ComparisonHeader
│   ├── ComparisonSidebar (Select reports)
│   ├── ComparisonContent
│   │   ├── RatingsComparison (chart)
│   │   ├── SideBySideAnalysis
│   │   ├── StrengthsWeaknessesComparison
│   │   └── DivergentOpinionsPanel
│   └── ComparisonExport
│
├── TemplateManager
│   ├── TemplateList
│   ├── TemplateEditor
│   │   ├── SectionConfigurator
│   │   └── RatingCategoryEditor
│   └── TemplatePreview
│
└── AnalyticsView (Historical dashboard)
    ├── AnalyticsFilters (date range, genre, etc.)
    ├── StatsCards (total coverage, avg ratings, etc.)
    ├── TrendsChart (ratings over time)
    ├── InsightsPanel (common strengths/weaknesses)
    └── TopScripts (by rating)
```

### Component Descriptions

#### CoverageView

Main container component. Manages routing between list/grid/board views and detail/edit views.

**State:**
- Current view mode (list/grid/board)
- Selected coverage report
- Filter/sort settings
- View state (list vs detail vs edit)

**Props:** None (top-level)

---

#### CoverageEditor

Complex form for creating/editing coverage reports.

**Features:**
- Auto-save drafts
- Template-driven sections
- Real-time word count
- Inline help/guidelines
- Character limit indicators
- Page notes with inline editing
- Rating sliders with hover descriptions
- Character breakdown mini-forms

**State:**
- Current coverage report (partial)
- Selected template
- Current section (for scroll spy)
- Validation errors
- Save status
- Uploaded script file info

**Props:**
- `coverageId?: string` (for editing)
- `templateId?: string` (default template)
- `onSave: (report: CoverageReport) => void`
- `onCancel: () => void`

---

#### CoverageDetailView

Read-only detailed view of a coverage report.

**Features:**
- Tabbed interface for different analysis sections
- Export to PDF/DOCX
- Print view
- Share/email functionality
- Mark as reviewed
- Add to comparison

**State:**
- Active tab
- Export format
- Related coverage reports

**Props:**
- `coverageId: string`
- `onEdit: () => void`
- `onDelete: () => void`
- `onCompare: (reportIds: string[]) => void`

---

#### CoverageComparisonView

Side-by-side comparison of multiple coverage reports.

**Features:**
- Visual rating differences (bar charts)
- Highlighted text differences
- Common themes extraction
- Consensus recommendations
- Export comparison report

**State:**
- Selected report IDs
- Comparison configuration
- Active comparison type

**Props:**
- `reportIds: string[]`
- `comparisonType: 'multiple_readers' | 'script_revisions'`

---

#### RatingsSection

Interactive rating input with sliders and visual feedback.

**Features:**
- 1-10 sliders with snap points
- Color-coded by value (red→yellow→green)
- Hover tooltips with rating descriptions
- Overall rating auto-calculated (if weighted)
- Radar chart preview

**State:**
- Current ratings
- Show preview chart

**Props:**
- `ratings: CoverageRatings`
- `template: CoverageTemplate`
- `onChange: (ratings: CoverageRatings) => void`
- `readonly?: boolean`

---

#### PageNotesManager

Interface for managing page-specific notes.

**Features:**
- Add notes with page number
- Filter by note type
- Quick jump to page in linked script
- Bulk edit/delete
- Export page notes separately

**State:**
- Notes list
- Filter settings
- Active note (for editing)

**Props:**
- `notes: PageNote[]`
- `onChange: (notes: PageNote[]) => void`
- `readonly?: boolean`

---

#### CharacterBreakdownCard

Individual character analysis form/display.

**Features:**
- Character role selector
- Arc rating slider
- Expandable strengths/weaknesses
- Delete/duplicate character

**Props:**
- `character: CharacterBreakdown`
- `onChange: (character: CharacterBreakdown) => void`
- `onDelete: () => void`
- `readonly?: boolean`

---

#### TemplateManager

Admin interface for managing coverage templates.

**Features:**
- Create custom templates
- Duplicate/modify existing templates
- Set default template
- Import/export templates
- Preview template before using

**State:**
- Template list
- Selected template
- Edit mode

---

#### AnalyticsView

Dashboard for historical coverage analysis.

**Features:**
- Aggregate statistics
- Trend charts (ratings over time)
- Genre breakdowns
- Reader performance metrics (if multi-reader)
- Script performance leaderboard
- Common feedback patterns

**State:**
- Date range
- Filter settings (genre, reader, etc.)
- Active chart type

---

## User Flow

### Primary Workflows

#### 1. Create New Coverage Report

```
User Journey:
1. Click "New Coverage" button
2. Select coverage template (or use default)
3. Upload/link script file (optional)
4. Fill in basic info (title, submitted by, date)
5. Write logline & synopsis
6. Rate script across categories
7. Complete detailed analysis sections
   - Navigate via sidebar menu
   - Auto-save on section complete
8. Add character breakdowns
9. Insert page-specific notes as needed
10. Write recommendation & next steps
11. Preview coverage report
12. Submit (or save as draft)

Edge Cases:
- Resume from draft
- Switch templates mid-way (confirm dialog)
- Upload multiple script versions
- Auto-extract metadata from script
```

**Flow Diagram (ASCII):**

```
[Start] → [Select Template] → [Upload Script (Optional)]
           ↓
    [Basic Info Form]
           ↓
    [Logline & Synopsis]
           ↓
    [Rating Sliders] ← Auto-save every 30s
           ↓
    [Analysis Sections]
    ├─ Premise
    ├─ Characters
    ├─ Dialogue
    ├─ Structure
    └─ Pacing
           ↓
    [Character Breakdowns]
           ↓
    [Page Notes] (Optional)
           ↓
    [Recommendation]
           ↓
    [Preview] → [Edit] (loop back)
           ↓
    [Submit] → [Detail View]
```

---

#### 2. View & Manage Coverage Reports

```
User Journey:
1. Navigate to Coverage tab
2. View list/grid/kanban of all reports
3. Filter by:
   - Status (draft/completed/archived)
   - Date range
   - Script title
   - Reader name
   - Recommendation
   - Genre
   - Rating threshold
4. Sort by:
   - Date submitted
   - Overall rating
   - Script title
   - Recommendation
5. Click card to view details
6. From detail view:
   - Export to PDF/DOCX
   - Edit coverage
   - Archive report
   - Add to comparison
   - View related coverage

Bulk Actions:
- Select multiple reports
- Bulk archive
- Bulk export
- Generate comparison
```

**View Toggle:**

```
[List View]     [Grid View]     [Kanban Board]
   ┃                ┃                  ┃
   ┃ ┌──────┐      ┃  ┌───┐ ┌───┐     ┃  Pass  │Consider│Recommend
   ┃ │ Card │      ┃  │ │ │ │ │ │     ┃  ┌───┐ │ ┌───┐ │ ┌───┐
   ┃ ├──────┤      ┃  └───┘ └───┘     ┃  │   │ │ │   │ │ │   │
   ┃ │ Card │      ┃  ┌───┐ ┌───┐     ┃  └───┘ │ └───┘ │ └───┘
   ┃ └──────┘      ┃  │   │ │   │     ┃        │       │
                   ┃  └───┘ └───┘     ┃  Drag & drop between columns
```

---

#### 3. Compare Coverage Reports

```
User Journey:
1. From coverage list, select 2-5 reports
2. Click "Compare" button
3. Choose comparison type:
   - Multiple readers on same script
   - Script revisions over time
4. View comparison dashboard:
   - Rating differences (chart)
   - Side-by-side analysis text
   - Consensus strengths/weaknesses
   - Divergent opinions highlighted
5. Export comparison report
6. Navigate to individual reports

Use Cases:
- See how multiple readers rated the same script
- Track improvement across script revisions
- Identify consistent feedback patterns
- Make informed development decisions
```

**Comparison Layout:**

```
╔════════════════════════════════════════════════════════════╗
║  Comparing: "Script Title" - 3 Reports                    ║
╠════════════════════════════════════════════════════════════╣
║  [Ratings Chart]              [Recommendation Breakdown]   ║
║                                                            ║
║  Overall: 7.3 avg (±1.2)      Pass: 0                     ║
║  Premise: 8.0 avg (±0.5)      Consider: 1                 ║
║  Character: 6.5 avg (±2.1)    Recommend: 2                ║
╠════════════════════════════════════════════════════════════╣
║  Common Strengths:                                         ║
║  • "Strong visual storytelling"                            ║
║  • "Unique premise with market appeal"                     ║
╠════════════════════════════════════════════════════════════╣
║  Common Weaknesses:                                        ║
║  • "Second act pacing issues"                              ║
║  • "Protagonist arc needs strengthening"                   ║
╠════════════════════════════════════════════════════════════╣
║  Divergent Opinions:                                       ║
║  ┌──────────────────────────────────────────────────┐     ║
║  │ Dialogue Quality                                 │     ║
║  │ Reader A (9/10): "Snappy, authentic dialogue"   │     ║
║  │ Reader B (5/10): "Dialogue feels forced at times"│     ║
║  └──────────────────────────────────────────────────┘     ║
╚════════════════════════════════════════════════════════════╝
```

---

#### 4. Historical Analysis & Trends

```
User Journey:
1. Navigate to Analytics view
2. Set filters (date range, genre, etc.)
3. View aggregate statistics:
   - Total coverage reports submitted
   - Average ratings by category
   - Recommendation distribution
   - Most common genres covered
4. Explore trend charts:
   - Rating trends over time
   - Genre performance
   - Reader consistency metrics
5. Identify patterns:
   - Most praised elements across all coverage
   - Most criticized elements
   - Correlation between ratings and recommendations
6. Export analytics report

Insights Generated:
- "Character development averages 6.8/10 across all reports"
- "Scripts with 8+ premise ratings have 73% recommend rate"
- "Common weakness: Act 2 pacing (mentioned in 42% of reports)"
```

---

#### 5. Template Management

```
User Journey:
1. Open Template Manager
2. View available templates:
   - Studio Coverage (default)
   - Contest Coverage
   - Development Notes
   - Quick Reader Report
   - Custom templates
3. Create new template:
   - Name and description
   - Configure sections (required/optional)
   - Define rating categories
   - Set field lengths and guidelines
4. Preview template
5. Set as default (optional)
6. Export/import templates

Template Types:
- Studio Coverage: Comprehensive 3-5 page report
- Contest Coverage: Focus on originality and marketability
- Development Notes: Revision-focused feedback
- Reader Report: Quick 1-page assessment
```

---

## Visual Design

### Design System Integration

All coverage components use the existing ANIME design system:

**Color Palette:**
- `electric-400/500`: Primary actions, highlights
- `mint-400/500`: Success, positive ratings, "Recommend"
- `sunset-400/500`: Warnings, "Consider"
- `sakura-400/500`: Accents, highlights
- `neon-400/500`: Secondary actions
- `gray-700/800/900`: Backgrounds, borders

**Rating Color Scale:**
```
1-3:   sunset-500   (Poor)      #FF7F00
4-5:   sunset-400   (Below Avg) #FFB86C
6-7:   electric-400 (Average)   #33CFFF
8-9:   mint-400     (Good)      #5BFA67
10:    mint-500     (Excellent) #50FA7B
```

---

### Layout Concepts

#### Coverage List View

```
┌─────────────────────────────────────────────────────────────┐
│ 📋 Coverage Reports                          [+ New Coverage]│
│ ─────────────────────────────────────────────────────────── │
│ [🔍 Search] [Filter ▼] [Sort ▼] [List][Grid][Board]         │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│ ┌────────────────────────────────────────────────────────┐  │
│ │ "The Last Guardian" - Studio Coverage                  │  │
│ │ Submitted by: Jane Doe | Nov 15, 2025                  │  │
│ │ ─────────────────────────────────────────────────────  │  │
│ │ Overall: ████████░░ 8/10                               │  │
│ │ Genre: Sci-Fi, Thriller                                │  │
│ │ ✅ RECOMMEND - Strong premise, excellent characters    │  │
│ │                                                         │  │
│ │ [View] [Edit] [Export] [Compare]                       │  │
│ └────────────────────────────────────────────────────────┘  │
│                                                              │
│ ┌────────────────────────────────────────────────────────┐  │
│ │ "Summer Nights" - Contest Coverage                     │  │
│ │ Submitted by: John Smith | Nov 12, 2025                │  │
│ │ ─────────────────────────────────────────────────────  │  │
│ │ Overall: ████░░░░░░ 4/10                               │  │
│ │ Genre: Romance, Drama                                  │  │
│ │ ⚠️ CONSIDER - Weak structure, needs revision           │  │
│ │                                                         │  │
│ │ [View] [Edit] [Export] [Compare]                       │  │
│ └────────────────────────────────────────────────────────┘  │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

#### Coverage Editor Layout

```
┌─────────────────────────────────────────────────────────────┐
│ ✏️ New Coverage Report                    [💾 Save Draft]   │
│ ─────────────────────────────────────────────────────────── │
├──────────┬──────────────────────────────────────────────────┤
│          │                                                   │
│ SECTIONS │  BASIC INFORMATION                                │
│          │  ────────────────────                             │
│ ✓ Basics │  Script Title: [___________________________]     │
│ ○ Logline│  Submitted By: [___________________________]     │
│ ○ Rating │  Date: [___________]  Genre: [___________]       │
│ ○ Analysis                                                  │
│ ○ Characters                                                │
│ ○ Page Notes                                                │
│ ○ Recommend                                                 │
│          │  Template: [Studio Coverage ▼]                   │
│          │                                                   │
│ [Preview]│  Upload Script: [Choose File] (Optional)         │
│          │                                                   │
│          │  ──────────────────────────────────────────────  │
│          │                                                   │
│          │  [Continue to Logline →]                         │
│          │                                                   │
├──────────┴──────────────────────────────────────────────────┤
│ Auto-saved 2 minutes ago                     [Cancel][Next] │
└─────────────────────────────────────────────────────────────┘
```

---

#### Coverage Detail View

```
┌─────────────────────────────────────────────────────────────┐
│ "The Last Guardian" - Coverage Report                       │
│ ──────────────────────────────────────────────────────────  │
│ [Edit] [Export PDF] [Export DOCX] [Compare] [Archive]      │
├─────────────────────┬───────────────────────────────────────┤
│                     │                                        │
│  RATINGS OVERVIEW   │  LOGLINE                              │
│  ────────────────   │  ──────                               │
│                     │  "A rogue AI guardian must choose..." │
│   ╱────────╲        │                                        │
│  ╱  8.5     ╲       │  SYNOPSIS                             │
│  │          │       │  ─────────                            │
│  │  Overall │       │  In 2087, an advanced AI guardian...  │
│   ╲        ╱        │  [Full synopsis text...]              │
│    ╲──────╱         │                                        │
│                     │  ──────────────────────────────────── │
│  Premise:     9     │                                        │
│  Character:   8     │  [Overview] [Characters] [Structure]  │
│  Dialogue:    7     │  [Page Notes]                         │
│  Structure:   9     │                                        │
│  Pacing:      8     │  STRENGTHS                            │
│  Market:      9     │  • Compelling high-concept premise    │
│  Original:    8     │  • Well-developed protagonist arc     │
│  Execution:   8     │  • Tight pacing in Act 3              │
│                     │                                        │
│  ✅ RECOMMEND       │  WEAKNESSES                           │
│  Confidence: 9/10   │  • Supporting cast underdeveloped     │
│                     │  • Villain motivations unclear        │
│  Submitted by:      │                                        │
│  Jane Doe           │  ──────────────────────────────────── │
│  Nov 15, 2025       │                                        │
│                     │  RECOMMENDATION                        │
│  Genre:             │  This script demonstrates strong...   │
│  Sci-Fi, Thriller   │                                        │
│                     │  Next Steps:                          │
│  Budget: Medium     │  1. Polish villain's backstory        │
│                     │  2. Develop supporting characters     │
│                     │  3. Ready for producer pitch          │
│                     │                                        │
└─────────────────────┴───────────────────────────────────────┘
```

---

#### Rating Input Component

```
┌─────────────────────────────────────────────────────────────┐
│  RATINGS                                                     │
│  ───────                                                     │
│                                                              │
│  Premise                                          9 / 10    │
│  ├────────────────────────────────●──┤ Excellent            │
│  "How unique and compelling is the core concept?"           │
│                                                              │
│  Character Development                            8 / 10    │
│  ├──────────────────────────●────────┤ Good                 │
│  "Are characters well-developed with clear arcs?"           │
│                                                              │
│  Dialogue                                         7 / 10    │
│  ├────────────────────●──────────────┤ Average              │
│  "Is dialogue natural, sharp, and character-specific?"      │
│                                                              │
│  Structure                                        9 / 10    │
│  ├────────────────────────────────●──┤ Excellent            │
│  "Does the story follow a clear, engaging structure?"       │
│                                                              │
│  [+ Add Custom Rating Category]                             │
│                                                              │
│  ──────────────────────────────────────────────────────────│
│                                                              │
│  Overall Rating (Auto-calculated): 8.5 / 10                 │
│  ┌──────────────────────────────────────┐                   │
│  │        Radar Chart Preview           │                   │
│  │     ╱───────────────────╲            │                   │
│  │    │   Premise (9)      │            │                   │
│  │    │                    │            │                   │
│  │    │   Character (8)    │            │                   │
│  │    │                    │            │                   │
│  │     ╲───────────────────╱            │                   │
│  └──────────────────────────────────────┘                   │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

#### Character Breakdown Card

```
┌─────────────────────────────────────────────────────────────┐
│  CHARACTER BREAKDOWN                                         │
│  ───────────────────                                         │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ ALEX CHEN                              [Edit] [Delete] │ │
│  │ ────────────────────────────────────────────────────── │ │
│  │ Role: Protagonist                                      │ │
│  │                                                         │ │
│  │ Description:                                            │ │
│  │ A brilliant AI engineer haunted by her past...         │ │
│  │                                                         │ │
│  │ Character Arc Rating: ████████░░ 8/10                  │ │
│  │                                                         │ │
│  │ Arc Description:                                        │ │
│  │ Alex transforms from a guilt-ridden recluse to...      │ │
│  │                                                         │ │
│  │ Strengths:                                              │ │
│  │ • Clear internal conflict                              │ │
│  │ • Relatable emotional journey                          │ │
│  │ • Active choices drive the plot                        │ │
│  │                                                         │ │
│  │ Weaknesses:                                             │ │
│  │ • Backstory could be revealed more gradually           │ │
│  │ • Some dialogue feels expository                       │ │
│  │                                                         │ │
│  │ Notes:                                                  │ │
│  │ Consider adding a scene showing Alex's past...         │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                              │
│  [+ Add Character]                                           │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

#### Page Notes Panel

```
┌─────────────────────────────────────────────────────────────┐
│  PAGE NOTES                                                  │
│  ──────────                                                  │
│                                                              │
│  [Add Note] [Filter: All ▼] [Sort by Page ▼]               │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Page 23 │ 💚 PRAISE                                     │ │
│  │ ──────────────────────────────────────────────────────│ │
│  │ Category: Dialogue                                     │ │
│  │                                                         │ │
│  │ "The confrontation between Alex and Marcus is          │ │
│  │ brilliantly written. The subtext is clear without      │ │
│  │ being heavy-handed."                                   │ │
│  │                                                         │ │
│  │ [Edit] [Delete]                                        │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Page 47 │ ⚠️ CRITIQUE                                   │ │
│  │ ──────────────────────────────────────────────────────│ │
│  │ Category: Pacing                                       │ │
│  │                                                         │ │
│  │ "This exposition scene slows down the momentum.        │ │
│  │ Consider cutting or integrating into action."          │ │
│  │                                                         │ │
│  │ [Edit] [Delete]                                        │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Page 89 │ ❓ QUESTION                                   │ │
│  │ ──────────────────────────────────────────────────────│ │
│  │ Category: Structure                                    │ │
│  │                                                         │ │
│  │ "Why does the AI reveal this information now?          │ │
│  │ The timing feels arbitrary."                           │ │
│  │                                                         │ │
│  │ [Edit] [Delete]                                        │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                              │
│  Showing 3 of 12 notes                         [Show All]   │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

#### Recommendation Section

```
┌─────────────────────────────────────────────────────────────┐
│  RECOMMENDATION                                              │
│  ──────────────                                              │
│                                                              │
│  Decision:                                                   │
│  ┌──────────┬──────────┬──────────┐                         │
│  │   PASS   │ CONSIDER │RECOMMEND │                         │
│  │          │          │    ✓     │                         │
│  └──────────┴──────────┴──────────┘                         │
│                                                              │
│  Confidence Level:                               9 / 10     │
│  ├────────────────────────────────●──┤                      │
│                                                              │
│  Summary:                                                    │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ This script presents a compelling high-concept premise │ │
│  │ with strong execution in most areas. The protagonist   │ │
│  │ arc is well-developed, and the third act delivers...   │ │
│  │                                                         │ │
│  │ [Character count: 247 / 500]                           │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                              │
│  Next Steps:                                                 │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ 1. [✓] Polish villain's backstory and motivations      │ │
│  │ 2. [✓] Develop supporting cast                         │ │
│  │ 3. [✓] Tighten Act 2 pacing                            │ │
│  │ 4. [ ] Ready for producer pitch                        │ │
│  │                                                         │ │
│  │ [+ Add Step]                                           │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                              │
│  Revision Notes (if CONSIDER):                               │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ [Hidden when RECOMMEND selected]                       │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

#### Comparison View

```
┌─────────────────────────────────────────────────────────────┐
│ 📊 Coverage Comparison: "The Last Guardian"                 │
│ ──────────────────────────────────────────────────────────  │
│ Comparing 3 reports | [Export Comparison] [Print]           │
├──────────────────────┬──────────────────────────────────────┤
│                      │                                       │
│ REPORTS SELECTED     │  RATINGS COMPARISON                   │
│ ────────────────     │  ──────────────────                   │
│                      │                                       │
│ ☑ Reader A (Nov 15)  │  Overall:                             │
│ ☑ Reader B (Nov 16)  │  ├─────────●─────┤ 8.5 (Reader A)    │
│ ☑ Reader C (Nov 17)  │  ├────────●──────┤ 8.0 (Reader B)    │
│                      │  ├────────────●──┤ 9.0 (Reader C)    │
│ [Add Report]         │  Average: 8.5 (±0.5)                  │
│                      │                                       │
│ COMPARISON TYPE      │  Premise:                             │
│ ────────────────     │  ├────────────●──┤ 9.0 (Reader A)    │
│ ● Multiple Readers   │  ├───────────●───┤ 8.5 (Reader B)    │
│ ○ Script Revisions   │  ├──────────────●┤ 9.5 (Reader C)    │
│                      │  Average: 9.0 (±0.5)                  │
│                      │                                       │
│                      │  [Show All Categories ▼]              │
│                      │                                       │
├──────────────────────┴──────────────────────────────────────┤
│                                                              │
│  RECOMMENDATION BREAKDOWN                                    │
│  ────────────────────────                                    │
│  Pass: 0 │ Consider: 0 │ Recommend: 3 (100%)                │
│                                                              │
│  ──────────────────────────────────────────────────────────│
│                                                              │
│  CONSENSUS STRENGTHS                                         │
│  ───────────────────                                         │
│  • "Strong, unique premise" (mentioned by 3 readers)        │
│  • "Well-developed protagonist arc" (mentioned by 3 readers)│
│  • "Tight pacing in Act 3" (mentioned by 2 readers)         │
│                                                              │
│  CONSENSUS WEAKNESSES                                        │
│  ────────────────────                                        │
│  • "Supporting cast underdeveloped" (mentioned by 3 readers)│
│  • "Villain motivations unclear" (mentioned by 2 readers)   │
│                                                              │
│  ──────────────────────────────────────────────────────────│
│                                                              │
│  DIVERGENT OPINIONS                                          │
│  ──────────────────                                          │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Dialogue Quality                                       │ │
│  │ ──────────────────────────────────────────────────────│ │
│  │ Reader A (Rating: 9/10)                                │ │
│  │ "Snappy, authentic dialogue that reveals character"    │ │
│  │                                                         │ │
│  │ Reader B (Rating: 7/10)                                │ │
│  │ "Generally good, but some exposition feels forced"     │ │
│  │                                                         │ │
│  │ Reader C (Rating: 8/10)                                │ │
│  │ "Sharp dialogue, though villain's lines are weak"      │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Act 2 Pacing                                           │ │
│  │ ──────────────────────────────────────────────────────│ │
│  │ Reader A (Rating: 7/10)                                │ │
│  │ "Solid pacing with clear escalation"                   │ │
│  │                                                         │ │
│  │ Reader B (Rating: 6/10)                                │ │
│  │ "Drags in the middle, needs tightening"                │ │
│  │                                                         │ │
│  │ Reader C (Rating: 8/10)                                │ │
│  │ "Well-paced with good rhythm"                          │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

#### Analytics Dashboard

```
┌─────────────────────────────────────────────────────────────┐
│ 📈 Coverage Analytics                                        │
│ ──────────────────────────────────────────────────────────  │
│ Date Range: [Last 30 Days ▼] Genre: [All ▼] Reader: [All ▼]│
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  QUICK STATS                                                 │
│  ───────────                                                 │
│  ┌──────────┬──────────┬──────────┬──────────┐              │
│  │   56     │   7.2    │   68%    │   42     │              │
│  │ Reports  │ Avg Rate │Recommend │  Genres  │              │
│  └──────────┴──────────┴──────────┴──────────┘              │
│                                                              │
│  RATING TRENDS                                               │
│  ─────────────                                               │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ 10├                                            ●        │ │
│  │  9├                                    ●   ●        ●   │ │
│  │  8├               ●        ●   ●   ●              ●     │ │
│  │  7├       ●   ●       ●                                 │ │
│  │  6├   ●                                                 │ │
│  │  5├                                                     │ │
│  │   └────────────────────────────────────────────────────│ │
│  │    Oct   Nov   Dec   Jan   Feb   Mar   Apr   May       │ │
│  │                                                         │ │
│  │  ─── Overall  ··· Premise  ─ ─ Character  ─·─ Dialogue│ │
│  └────────────────────────────────────────────────────────┘ │
│                                                              │
│  RECOMMENDATION DISTRIBUTION                                 │
│  ───────────────────────────                                 │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  RECOMMEND ████████████████████████░░░░░░ 68% (38)     │ │
│  │  CONSIDER  ██████████░░░░░░░░░░░░░░░░░░░░ 21% (12)     │ │
│  │  PASS      ███░░░░░░░░░░░░░░░░░░░░░░░░░░░ 11% (6)      │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                              │
│  TOP GENRES                                                  │
│  ──────────                                                  │
│  1. Sci-Fi (14 reports, 8.1 avg rating)                     │
│  2. Thriller (12 reports, 7.8 avg rating)                   │
│  3. Drama (10 reports, 7.5 avg rating)                      │
│  4. Comedy (8 reports, 6.9 avg rating)                      │
│  5. Action (7 reports, 7.2 avg rating)                      │
│                                                              │
│  COMMON FEEDBACK PATTERNS                                    │
│  ──────────────────────────                                  │
│  Most Praised:                                               │
│  • "Strong premise" (mentioned in 67% of reports)           │
│  • "Well-developed protagonist" (mentioned in 54%)          │
│  • "Unique voice" (mentioned in 48%)                        │
│                                                              │
│  Most Criticized:                                            │
│  • "Act 2 pacing" (mentioned in 42% of reports)             │
│  • "Supporting cast" (mentioned in 38%)                     │
│  • "Predictable plot" (mentioned in 31%)                    │
│                                                              │
│  HIGHEST RATED SCRIPTS                                       │
│  ─────────────────────                                       │
│  1. "The Last Guardian" - 9.2 avg (3 reports)               │
│  2. "Echoes of Tomorrow" - 8.8 avg (2 reports)              │
│  3. "Midnight Protocol" - 8.5 avg (4 reports)               │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

### UI Components Library

Custom components needed for coverage:

1. **RatingSlider**
   - Interactive 1-10 slider
   - Color-coded by value
   - Hover tooltips
   - Snap to integer values

2. **RadarChart**
   - SVG-based rating visualization
   - Animated on data change
   - Responsive sizing

3. **RecommendationBadge**
   - Color-coded by decision (Pass/Consider/Recommend)
   - Includes confidence indicator
   - Animated glow effect

4. **ProgressIndicator**
   - Multi-step form progress
   - Section completion checkmarks
   - Click to navigate

5. **PageNoteMarker**
   - Inline annotation indicator
   - Click to add/edit note
   - Color-coded by note type

6. **ComparisonChart**
   - Side-by-side bar chart
   - Difference highlighting
   - Interactive legends

7. **StrengthsWeaknessesList**
   - Editable bullet list
   - Add/remove items
   - Drag to reorder

---

## Integration Points

### 1. Integration with Existing Todos System

**Use Case:** Convert coverage feedback into actionable tasks.

**Implementation:**

```typescript
// In CoverageDetailView
const handleCreateTodos = (coverageReport: CoverageReport) => {
  const { useTodoStore } = await import('../store/todoStore')
  const { addTodo } = useTodoStore.getState()

  // Create todos from next steps
  coverageReport.recommendation.next_steps.forEach((step, index) => {
    addTodo({
      title: step,
      description: `From coverage: "${coverageReport.title}"`,
      status: 'pending',
      priority: index === 0 ? 'high' : 'medium',
      category: 'feature',
      assignee: 'human',
      tags: ['coverage', 'revision', coverageReport.script_title],
      dueDate: null,
    })
  })

  // Create todos from weaknesses
  coverageReport.analysis.weaknesses.forEach((weakness) => {
    addTodo({
      title: `Address: ${weakness}`,
      description: `Weakness identified in coverage by ${coverageReport.submitted_by}`,
      status: 'pending',
      priority: 'medium',
      category: 'feature',
      assignee: 'human',
      tags: ['coverage', 'weakness', coverageReport.script_title],
      dueDate: null,
    })
  })
}
```

**UI Elements:**
- "Create Todos from Coverage" button in detail view
- Bulk selection of which items to convert
- Auto-tag todos with coverage ID for tracking

---

### 2. Integration with Writing Tab

**Use Case:** Link coverage to the script document, enable quick access.

**Implementation:**

```typescript
// Extend Document type
interface Document {
  // ... existing fields
  coverage_reports?: string[] // Coverage report IDs
}

// In WritingView, add coverage indicator
const CoverageIndicator = ({ documentId }: { documentId: string }) => {
  const coverageReports = useCoverageStore(state =>
    state.reports.filter(r => r.script_id === documentId)
  )

  if (coverageReports.length === 0) return null

  const avgRating = coverageReports.reduce((sum, r) =>
    sum + r.ratings.overall, 0) / coverageReports.length

  return (
    <div className="flex items-center gap-2 px-3 py-1 bg-electric-500/10 border border-electric-500/30 rounded-lg">
      <span className="text-electric-400">📋</span>
      <span className="text-sm text-electric-400">
        {coverageReports.length} Coverage Report{coverageReports.length > 1 ? 's' : ''}
      </span>
      <span className="text-sm text-mint-400">
        {avgRating.toFixed(1)}/10 avg
      </span>
      <button className="text-xs text-electric-400 hover:text-electric-300">
        View →
      </button>
    </div>
  )
}
```

**Features:**
- Badge showing coverage count and avg rating in WritingView
- Quick link from document to coverage reports
- "Request Coverage" button in WritingView
- Auto-link coverage when script title matches

---

### 3. Integration with Analysis Tab

**Use Case:** Coverage can trigger or supplement automated analysis.

**Implementation:**

```typescript
// In CoverageEditor, add option to run automated analysis
const handleRunAnalysis = async (scriptContent: string) => {
  // Run automated character analysis
  const charAnalysis = await invoke<CharacterAnalysis>('analyze_content', {
    type: 'character',
    content: scriptContent,
  })

  // Pre-populate character breakdowns from automated analysis
  const characters: CharacterBreakdown[] = charAnalysis.characters.map(char => ({
    name: char.name,
    role: char.role as any,
    description: `Appears ${char.dialogue_count} times`,
    arc_rating: char.importance_score,
    arc_description: char.arc,
    strengths: char.traits.slice(0, 3),
    weaknesses: [],
    notes: `Auto-generated from analysis`,
  }))

  setCoverageData(prev => ({
    ...prev,
    analysis: {
      ...prev.analysis,
      characters,
    }
  }))
}
```

**Features:**
- "Auto-populate from Analysis" button in CoverageEditor
- Option to run automated analysis when uploading script
- Compare automated analysis vs human coverage
- Use automated analysis as starting point for coverage

---

### 4. Integration with Workflows

**Use Case:** Coverage as part of script development workflow.

**Implementation:**

```typescript
// Extend ComfyUIWorkflow to include coverage milestones
interface ScriptWorkflow {
  id: string
  name: string
  stages: WorkflowStage[]
}

interface WorkflowStage {
  id: string
  name: string
  type: 'writing' | 'coverage' | 'revision' | 'approval'
  status: 'pending' | 'in_progress' | 'completed'

  // For coverage stage
  required_coverage_count?: number
  min_avg_rating?: number
  coverage_template_id?: string
  coverage_reports?: string[]
}

// Example workflow: "Script to Production"
const scriptWorkflow: ScriptWorkflow = {
  id: 'script-dev-1',
  name: 'Script Development Pipeline',
  stages: [
    {
      id: 'draft',
      name: 'First Draft',
      type: 'writing',
      status: 'completed',
    },
    {
      id: 'coverage-1',
      name: 'Initial Coverage',
      type: 'coverage',
      status: 'in_progress',
      required_coverage_count: 2,
      min_avg_rating: 6.0,
      coverage_template_id: 'studio-coverage',
      coverage_reports: ['cov-1', 'cov-2'],
    },
    {
      id: 'revision-1',
      name: 'First Revision',
      type: 'revision',
      status: 'pending',
    },
    // ... more stages
  ]
}
```

**Features:**
- Coverage as workflow stage
- Workflow gates based on coverage results
- Automatic progression when coverage criteria met
- Workflow dashboard showing coverage status

---

### 5. Backend Storage Integration

**Rust Backend Commands:**

```rust
// src-tauri/src/coverage.rs

use serde::{Deserialize, Serialize};
use std::fs;
use std::path::PathBuf;

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct CoverageReport {
    pub id: String,
    pub title: String,
    pub script_id: Option<String>,
    pub script_title: String,
    pub script_path: Option<String>,
    pub created_at: String,
    pub updated_at: String,
    pub submitted_by: String,
    pub submitted_date: String,
    pub template_id: String,
    pub template_version: String,
    pub logline: String,
    pub synopsis: String,
    pub ratings: CoverageRatings,
    pub analysis: CoverageAnalysis,
    pub recommendation: CoverageRecommendation,
    pub genre: Vec<String>,
    pub comparable_titles: Vec<String>,
    pub target_audience: String,
    pub budget_estimate: Option<String>,
    pub status: String,
    pub tags: Vec<String>,
    pub version: i32,
}

// ... other types

#[tauri::command]
pub async fn create_coverage_report(
    title: String,
    template_id: String,
    submitted_by: String,
) -> Result<CoverageReport, String> {
    let report = CoverageReport {
        id: uuid::Uuid::new_v4().to_string(),
        title,
        script_id: None,
        script_title: String::new(),
        script_path: None,
        created_at: chrono::Utc::now().to_rfc3339(),
        updated_at: chrono::Utc::now().to_rfc3339(),
        submitted_by,
        submitted_date: chrono::Utc::now().to_rfc3339(),
        template_id,
        template_version: "1.0".to_string(),
        logline: String::new(),
        synopsis: String::new(),
        ratings: CoverageRatings::default(),
        analysis: CoverageAnalysis::default(),
        recommendation: CoverageRecommendation::default(),
        genre: Vec::new(),
        comparable_titles: Vec::new(),
        target_audience: String::new(),
        budget_estimate: None,
        status: "draft".to_string(),
        tags: Vec::new(),
        version: 1,
    };

    save_coverage_report(&report)?;
    Ok(report)
}

#[tauri::command]
pub async fn list_coverage_reports() -> Result<Vec<CoverageReport>, String> {
    let coverage_dir = get_coverage_dir()?;
    let mut reports = Vec::new();

    if let Ok(entries) = fs::read_dir(&coverage_dir) {
        for entry in entries.flatten() {
            if let Ok(content) = fs::read_to_string(entry.path()) {
                if let Ok(report) = serde_json::from_str::<CoverageReport>(&content) {
                    reports.push(report);
                }
            }
        }
    }

    // Sort by updated_at desc
    reports.sort_by(|a, b| b.updated_at.cmp(&a.updated_at));
    Ok(reports)
}

#[tauri::command]
pub async fn save_coverage_report(report: &CoverageReport) -> Result<(), String> {
    let coverage_dir = get_coverage_dir()?;
    fs::create_dir_all(&coverage_dir)
        .map_err(|e| format!("Failed to create coverage directory: {}", e))?;

    let file_path = coverage_dir.join(format!("{}.json", report.id));
    let json = serde_json::to_string_pretty(report)
        .map_err(|e| format!("Failed to serialize coverage: {}", e))?;

    fs::write(file_path, json)
        .map_err(|e| format!("Failed to write coverage file: {}", e))?;

    Ok(())
}

#[tauri::command]
pub async fn get_coverage_report(id: String) -> Result<CoverageReport, String> {
    let coverage_dir = get_coverage_dir()?;
    let file_path = coverage_dir.join(format!("{}.json", id));

    let content = fs::read_to_string(file_path)
        .map_err(|e| format!("Failed to read coverage: {}", e))?;

    let report = serde_json::from_str::<CoverageReport>(&content)
        .map_err(|e| format!("Failed to parse coverage: {}", e))?;

    Ok(report)
}

#[tauri::command]
pub async fn delete_coverage_report(id: String) -> Result<(), String> {
    let coverage_dir = get_coverage_dir()?;
    let file_path = coverage_dir.join(format!("{}.json", id));

    fs::remove_file(file_path)
        .map_err(|e| format!("Failed to delete coverage: {}", e))?;

    Ok(())
}

#[tauri::command]
pub async fn export_coverage_pdf(id: String) -> Result<String, String> {
    let report = get_coverage_report(id).await?;

    // TODO: Implement PDF generation
    // For now, export as formatted text
    let export_path = dirs::download_dir()
        .ok_or("Failed to get downloads directory")?
        .join(format!("{}_coverage.txt", report.script_title));

    let formatted = format_coverage_report(&report);
    fs::write(&export_path, formatted)
        .map_err(|e| format!("Failed to export coverage: {}", e))?;

    Ok(export_path.to_string_lossy().to_string())
}

fn get_coverage_dir() -> Result<PathBuf, String> {
    let home = dirs::home_dir().ok_or("Failed to get home directory")?;
    Ok(home.join(".anime-desktop").join("coverage"))
}

fn format_coverage_report(report: &CoverageReport) -> String {
    format!(
        r#"COVERAGE REPORT
==================

Script: {}
Submitted by: {}
Date: {}

LOGLINE
-------
{}

SYNOPSIS
--------
{}

RATINGS
-------
Overall: {}/10
Premise: {}/10
Character: {}/10
Dialogue: {}/10
Structure: {}/10
Pacing: {}/10
Marketability: {}/10
Originality: {}/10
Execution: {}/10

ANALYSIS
--------

STRENGTHS:
{}

WEAKNESSES:
{}

RECOMMENDATION
--------------
Decision: {}
Confidence: {}/10

{}

NEXT STEPS:
{}
"#,
        report.script_title,
        report.submitted_by,
        report.submitted_date,
        report.logline,
        report.synopsis,
        report.ratings.overall,
        report.ratings.premise,
        report.ratings.character,
        report.ratings.dialogue,
        report.ratings.structure,
        report.ratings.pacing,
        report.ratings.marketability,
        report.ratings.originality,
        report.ratings.execution,
        report.analysis.strengths.join("\n• "),
        report.analysis.weaknesses.join("\n• "),
        report.recommendation.decision,
        report.recommendation.confidence,
        report.recommendation.summary,
        report.recommendation.next_steps.join("\n"),
    )
}
```

**Storage Structure:**

```
~/.anime-desktop/
├── coverage/
│   ├── {coverage-id-1}.json
│   ├── {coverage-id-2}.json
│   └── ...
├── templates/
│   ├── studio-coverage.json
│   ├── contest-coverage.json
│   └── custom-{id}.json
└── documents/
    └── ... (existing)
```

---

### 6. Zustand Store

```typescript
// src/store/coverageStore.ts

import { create } from 'zustand'
import { invoke } from '@tauri-apps/api/core'
import type { CoverageReport, CoverageTemplate, CoverageComparison } from '../types/coverage'

interface CoverageStore {
  reports: CoverageReport[]
  templates: CoverageTemplate[]
  selectedReportId: string | null
  comparisonReportIds: string[]
  isLoading: boolean
  error: string | null

  // Actions
  loadReports: () => Promise<void>
  loadTemplates: () => Promise<void>
  createReport: (title: string, templateId: string, submittedBy: string) => Promise<CoverageReport>
  updateReport: (report: CoverageReport) => Promise<void>
  deleteReport: (id: string) => Promise<void>
  selectReport: (id: string | null) => void
  addToComparison: (id: string) => void
  removeFromComparison: (id: string) => void
  clearComparison: () => void
  exportReportPDF: (id: string) => Promise<string>

  // Computed
  getReportById: (id: string) => CoverageReport | undefined
  getReportsByStatus: (status: string) => CoverageReport[]
  getComparisonReports: () => CoverageReport[]
}

export const useCoverageStore = create<CoverageStore>((set, get) => ({
  reports: [],
  templates: [],
  selectedReportId: null,
  comparisonReportIds: [],
  isLoading: false,
  error: null,

  loadReports: async () => {
    set({ isLoading: true, error: null })
    try {
      const reports = await invoke<CoverageReport[]>('list_coverage_reports')
      set({ reports, isLoading: false })
    } catch (error) {
      set({ error: String(error), isLoading: false })
    }
  },

  loadTemplates: async () => {
    try {
      const templates = await invoke<CoverageTemplate[]>('list_coverage_templates')
      set({ templates })
    } catch (error) {
      console.error('Failed to load templates:', error)
    }
  },

  createReport: async (title, templateId, submittedBy) => {
    set({ isLoading: true, error: null })
    try {
      const report = await invoke<CoverageReport>('create_coverage_report', {
        title,
        templateId,
        submittedBy,
      })
      set(state => ({
        reports: [report, ...state.reports],
        selectedReportId: report.id,
        isLoading: false,
      }))
      return report
    } catch (error) {
      set({ error: String(error), isLoading: false })
      throw error
    }
  },

  updateReport: async (report) => {
    try {
      await invoke('save_coverage_report', { report })
      set(state => ({
        reports: state.reports.map(r => r.id === report.id ? report : r),
      }))
    } catch (error) {
      set({ error: String(error) })
      throw error
    }
  },

  deleteReport: async (id) => {
    try {
      await invoke('delete_coverage_report', { id })
      set(state => ({
        reports: state.reports.filter(r => r.id !== id),
        selectedReportId: state.selectedReportId === id ? null : state.selectedReportId,
      }))
    } catch (error) {
      set({ error: String(error) })
      throw error
    }
  },

  selectReport: (id) => {
    set({ selectedReportId: id })
  },

  addToComparison: (id) => {
    set(state => ({
      comparisonReportIds: [...new Set([...state.comparisonReportIds, id])]
    }))
  },

  removeFromComparison: (id) => {
    set(state => ({
      comparisonReportIds: state.comparisonReportIds.filter(rid => rid !== id)
    }))
  },

  clearComparison: () => {
    set({ comparisonReportIds: [] })
  },

  exportReportPDF: async (id) => {
    try {
      const path = await invoke<string>('export_coverage_pdf', { id })
      return path
    } catch (error) {
      set({ error: String(error) })
      throw error
    }
  },

  getReportById: (id) => {
    return get().reports.find(r => r.id === id)
  },

  getReportsByStatus: (status) => {
    return get().reports.filter(r => r.status === status)
  },

  getComparisonReports: () => {
    const { reports, comparisonReportIds } = get()
    return reports.filter(r => comparisonReportIds.includes(r.id))
  },
}))
```

---

## Technical Requirements

### Frontend Dependencies

```json
{
  "dependencies": {
    "recharts": "^2.10.0",
    "react-markdown": "^9.0.1",
    "date-fns": "^3.0.0",
    "zustand": "^4.4.0", // Already installed
    "@tauri-apps/api": "^2.0.0", // Already installed
    "@tauri-apps/plugin-dialog": "^2.0.0" // Already installed
  },
  "devDependencies": {
    "@types/react": "^18.0.0" // Already installed
  }
}
```

### Rust Dependencies

```toml
# Cargo.toml
[dependencies]
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
tauri = { version = "2.0", features = [] }
uuid = { version = "1.6", features = ["v4", "serde"] }
chrono = { version = "0.4", features = ["serde"] }
dirs = "5.0"
# For future PDF generation:
# printpdf = "0.7"
```

### TypeScript Types

Create `/src/types/coverage.ts` with all the interfaces defined in the Data Model section.

---

### Performance Considerations

1. **Lazy Loading**
   - Load coverage reports on-demand
   - Virtualized lists for large coverage collections
   - Paginated API for 100+ reports

2. **Caching**
   - Cache frequently accessed reports in memory
   - Cache rendered markdown/charts
   - Debounce auto-save in editor (500ms)

3. **Optimistic Updates**
   - Update UI immediately, sync to backend async
   - Rollback on error with toast notification
   - Queue background saves

4. **Search & Filter**
   - Client-side filtering for <1000 reports
   - Backend filtering for larger datasets
   - Indexed search (future: use SQLite)

---

### Accessibility

1. **Keyboard Navigation**
   - Tab through all interactive elements
   - Enter to submit forms
   - Escape to close modals
   - Arrow keys for slider inputs

2. **Screen Readers**
   - ARIA labels on all controls
   - Semantic HTML structure
   - Skip navigation links
   - Live region announcements for updates

3. **Visual**
   - High contrast mode support
   - Scalable text (support up to 200%)
   - Color is not the only indicator (use icons + text)
   - Focus indicators visible

---

### Security

1. **Input Validation**
   - Sanitize all user input
   - Validate rating ranges (1-10)
   - Prevent XSS in markdown/text fields
   - File upload size limits

2. **Data Privacy**
   - Local storage only (no cloud by default)
   - Optional export to external services
   - Secure file permissions on coverage directory
   - No telemetry/analytics without consent

---

## Implementation Phases

### Phase 1: Core Infrastructure (Week 1-2)

**Goal:** Basic coverage creation and viewing

**Tasks:**
1. Create TypeScript types (`coverage.ts`)
2. Set up Rust backend commands
3. Create Zustand store (`coverageStore.ts`)
4. Implement file storage system
5. Build CoverageView shell component
6. Add "Coverage" to sidebar navigation

**Deliverables:**
- Can create a coverage report
- Can list all coverage reports
- Can view a single report (basic layout)
- Data persists across app restarts

**Testing:**
- Create 5 test coverage reports
- Verify data persistence
- Test error handling

---

### Phase 2: Coverage Editor (Week 3-4)

**Goal:** Full-featured coverage creation form

**Tasks:**
1. Build CoverageEditor component
2. Create template system
3. Implement rating sliders with validation
4. Build character breakdown forms
5. Add page notes manager
6. Implement auto-save functionality
7. Create form validation

**Deliverables:**
- Complete coverage editor with all sections
- Template selector
- Auto-save drafts every 30 seconds
- Form validation with helpful errors
- Character count indicators

**Testing:**
- Fill out complete coverage report
- Test auto-save recovery
- Validate all required fields
- Test with different templates

---

### Phase 3: Visualization & Detail View (Week 5)

**Goal:** Rich display of coverage data

**Tasks:**
1. Build CoverageDetailView component
2. Create rating visualizations (radar chart)
3. Implement tabbed analysis sections
4. Add export to PDF/TXT
5. Create RecommendationBadge component
6. Build page notes display

**Deliverables:**
- Read-only detail view
- Interactive charts
- Export functionality
- Print-friendly view
- Related coverage sidebar

**Testing:**
- View various coverage reports
- Export to different formats
- Test print view
- Validate chart rendering

---

### Phase 4: Comparison & Analytics (Week 6)

**Goal:** Multi-report analysis

**Tasks:**
1. Build CoverageComparisonView
2. Implement comparison algorithms
3. Create comparison visualizations
4. Build AnalyticsView dashboard
5. Implement trend charts
6. Add aggregate statistics

**Deliverables:**
- Side-by-side comparison of 2-5 reports
- Consensus analysis
- Historical analytics dashboard
- Trend visualizations
- Export comparison reports

**Testing:**
- Compare 3 reports on same script
- Compare script revisions
- Generate analytics from 20+ reports
- Test edge cases (1 report, 100 reports)

---

### Phase 5: Integration & Polish (Week 7)

**Goal:** Connect to existing systems

**Tasks:**
1. Integrate with Todos system
2. Link to Writing tab
3. Connect to Analysis tab
4. Add to Workflows
5. Implement search & filtering
6. Add keyboard shortcuts
7. Polish animations and transitions

**Deliverables:**
- "Create Todos from Coverage" feature
- Coverage indicators in Writing tab
- Workflow integration
- Global search
- Keyboard navigation
- Smooth animations

**Testing:**
- Test all integration points
- Verify data sync
- Test keyboard shortcuts
- User acceptance testing

---

### Phase 6: Templates & Customization (Week 8)

**Goal:** Flexible coverage system

**Tasks:**
1. Build TemplateManager
2. Implement template CRUD
3. Add template import/export
4. Create default templates
5. Build template preview
6. Add custom rating categories

**Deliverables:**
- Template management UI
- 4 default templates
- Custom template creation
- Template import/export
- Template versioning

**Testing:**
- Create custom template
- Duplicate and modify template
- Export/import templates
- Use different templates for coverage

---

### Phase 7: Advanced Features (Week 9-10)

**Goal:** Power user features

**Tasks:**
1. Implement bulk operations
2. Add advanced filtering
3. Create coverage history timeline
4. Build revision comparison
5. Add collaboration features (comments)
6. Implement coverage reminders/notifications

**Deliverables:**
- Bulk archive/export
- Advanced filter builder
- Historical timeline view
- Revision tracking
- Optional: Comment system

**Testing:**
- Bulk operations on 50+ reports
- Complex filter combinations
- Track script through multiple revisions
- Performance testing

---

### Phase 8: Optimization & Documentation (Week 11-12)

**Goal:** Production-ready system

**Tasks:**
1. Performance optimization
2. Accessibility audit
3. Write user documentation
4. Create video tutorials
5. Add inline help/tooltips
6. Implement data migration utilities
7. Final bug fixes

**Deliverables:**
- Optimized performance (sub-100ms interactions)
- WCAG 2.1 AA compliance
- Complete user guide
- Video walkthrough
- Migration tools
- Stable release

**Testing:**
- Performance benchmarks
- Accessibility testing
- User testing with non-technical users
- Cross-platform testing (macOS, Windows, Linux)

---

## Success Metrics

### User Adoption
- 80% of users create at least one coverage report
- Average of 5+ coverage reports per user after 30 days
- 90% retention rate for coverage feature

### Usability
- Users can create first coverage in under 10 minutes
- 95% form completion rate (draft to submitted)
- Less than 5% error rate on form submission

### Performance
- Coverage list loads in under 500ms (for 100 reports)
- Editor auto-save completes in under 200ms
- Export to PDF completes in under 2 seconds

### Quality
- Zero critical bugs in production
- User satisfaction score of 4.5/5 or higher
- 90% of coverage reports use detailed analysis sections

---

## Future Enhancements

### Version 2.0 Features

1. **AI-Assisted Coverage**
   - Auto-generate logline from synopsis
   - Suggest character arcs based on script analysis
   - Predict ratings based on historical data
   - Flag common script issues automatically

2. **Collaboration**
   - Multi-user coverage on same script
   - Real-time collaborative editing
   - Comment threads on specific sections
   - @mention notifications

3. **Advanced Analytics**
   - Machine learning for pattern detection
   - Predictive success metrics
   - Genre-specific benchmarking
   - Reader bias detection

4. **Cloud Sync**
   - Optional cloud backup
   - Sync across multiple devices
   - Share coverage with external users
   - Public coverage library (opt-in)

5. **Mobile Companion App**
   - Read coverage on mobile
   - Voice notes for page comments
   - Offline reading mode
   - Push notifications for new coverage

6. **Integration Ecosystem**
   - Final Draft integration
   - WriterDuet sync
   - Celtx compatibility
   - API for third-party tools

---

## Conclusion

This architecture provides a comprehensive, production-ready coverage analysis system that integrates seamlessly with the ANIME desktop application's existing design system and workflows. The phased implementation approach ensures steady progress with regular deliverables and testing milestones.

Key strengths:
- **Flexible data model** supporting multiple coverage formats
- **Rich UI/UX** matching anime-themed design system
- **Deep integration** with existing features (todos, writing, analysis)
- **Scalable architecture** supporting future enhancements
- **User-centric design** optimized for screenwriters and development executives

The coverage system will become a central tool for script development, enabling users to systematically track feedback, identify patterns, and make data-driven decisions about their projects.
