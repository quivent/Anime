import { useState } from 'react'
import { useCoverageStore } from '../store/coverageStore'
import type { CoverageReport } from '../types/coverage'
import ScriptUpload from './ScriptUpload'
import ParsingProgress from './ParsingProgress'
import CoverageEditor from './CoverageEditor'
import CoverageDetail from './CoverageDetail'
import CoverageDetailProfessional from './coverage/CoverageDetailProfessional'
import { EXAMPLE_PROFESSIONAL_COVERAGE } from '../data/example-coverage'

interface ParsingStep {
  id: string
  name: string
  status: 'pending' | 'in_progress' | 'completed' | 'error'
  progress?: number
  details?: string
}

export default function CoverageView() {
  const { coverages, deleteCoverage } = useCoverageStore()
  const [selectedCoverage, setSelectedCoverage] = useState<CoverageReport | null>(null)
  const [showEditor, setShowEditor] = useState(false)
  const [isParsing, setIsParsing] = useState(false)
  const [parsingSteps, setParsingSteps] = useState<ParsingStep[]>([])
  const [currentFile, setCurrentFile] = useState<string>('')
  const [viewMode, setViewMode] = useState<'list' | 'detail' | 'editor' | 'example'>('list')

  const exampleCoverage: CoverageReport = {
    id: 'example',
    title: 'Coverage: The Last Stand',
    script_title: 'The Last Stand',
    submitted_by: 'Sarah Chen',
    submitted_date: '2025-01-15',
    template_id: 'default',
    template_version: '1.0',
    created_at: '2025-01-15',
    updated_at: '2025-01-15',
    logline: 'A disillusioned detective must team up with his estranged daughter to stop a terrorist organization from releasing a deadly bioweapon in downtown Los Angeles.',
    synopsis: `ACT I:\n\nDetective Marcus Kane is a burned-out LAPD veteran haunted by the death of his partner two years ago. His marriage has fallen apart, and his relationship with his 23-year-old daughter Emma, a CDC scientist, is virtually nonexistent.\n\nWhen a series of mysterious deaths occur across the city, Emma discovers a pattern that suggests a weaponized pathogen. She tries to warn her father, but he dismisses her concerns.\n\nACT II:\n\nThe situation escalates when Emma is kidnapped by a sophisticated terrorist cell led by the enigmatic Victor Kaine, a former bioweapons researcher with a vendetta against the government. Marcus is forced to confront his failures as both a detective and a father.\n\nWorking with Emma's research partner and an FBI task force, Marcus uncovers Kaine's plan to release the pathogen during a major public event. The clock is ticking, and Marcus must navigate both the criminal underworld and his own demons to find his daughter.\n\nACT III:\n\nMarcus infiltrates Kaine's compound and rescues Emma, but not before learning that she's been exposed to the pathogen. With only hours before symptoms appear, they must work together to stop Kaine's plan and find the antidote.\n\nThe climax takes place at the event venue, where Marcus and Emma confront Kaine in a tense standoff that tests both their courage and their rekindled bond as father and daughter.`,
    ratings: {
      overall: 7,
      premise: 8,
      character: 7,
      dialogue: 6,
      structure: 8,
      pacing: 7,
      marketability: 8,
      originality: 5,
      execution: 7,
    },
    analysis: {
      strengths: [
        'Strong high-concept premise with clear commercial appeal',
        'Well-structured three-act progression with clear turning points',
        'Compelling emotional core in the father-daughter relationship',
        'Effective ticking-clock tension in the third act',
        'Good balance of action and character development',
      ],
      weaknesses: [
        'Familiar genre tropes and predictable plot beats',
        'Villain lacks depth and clear motivation beyond revenge',
        'Some dialogue feels expository, particularly in technical scenes',
        'Supporting characters are underdeveloped',
        'Third act resolution feels rushed',
      ],
      premise_analysis: 'The bioterrorism angle combined with the estranged father-daughter dynamic creates a compelling dual-threat narrative.',
      character_analysis: 'Marcus is a well-drawn protagonist with clear flaws and arc. Emma needs more development beyond her role as victim/scientist.',
      dialogue_analysis: 'Generally solid with some strong character moments, but technical exposition can feel clunky.',
      structure_analysis: 'Classic three-act structure executed competently with strong act breaks and escalating stakes.',
      pacing_analysis: 'Good momentum through the first two acts. Third act feels compressed and could use additional beats.',
      theme_analysis: 'Themes of redemption and second chances are clear but could be woven more subtly throughout.',
      page_notes: [],
      character_breakdowns: [],
    },
    recommendation: {
      type: 'consider',
      summary: `THE LAST STAND is a competent action-thriller with strong commercial potential. While it doesn't break new ground in the genre, it executes familiar elements well and has a solid emotional core.\n\nThe father-daughter relationship provides genuine stakes beyond the ticking-clock thriller elements, and the script demonstrates professional craft in structure and pacing.\n\nWith revisions to deepen the villain, develop supporting characters, and expand the third act, this could be a solid mid-budget action film. The concept is marketable and the execution is professional, making it worth further development consideration.\n\nRecommend: CONSIDER with revisions`,
      commercial_appeal: 8,
      target_audience: 'Adult audiences 25-54, action-thriller fans',
      comparable_titles: ['Taken', 'The Fugitive', 'White House Down'],
      notes: 'Strong potential for a name actor in the lead role. Budget estimate: $30-40M.',
    },
    genre: ['Action', 'Thriller', 'Drama'],
    budget_estimate: 'medium',
    status: 'completed',
    tags: ['bioterrorism', 'detective', 'father-daughter', 'los-angeles'],
    version: 1,
  }

  const handleCreateNew = () => {
    setSelectedCoverage(null)
    setViewMode('editor')
    setShowEditor(true)
  }

  const handleViewDetail = (coverage: CoverageReport) => {
    setSelectedCoverage(coverage)
    setViewMode('detail')
  }

  const handleEditFromDetail = () => {
    setViewMode('editor')
    setShowEditor(true)
  }

  const handleFileUpload = (file: File) => {
    setCurrentFile(file.name)
    setIsParsing(true)

    setParsingSteps([
      { id: '1', name: 'Extracting PDF Content', status: 'in_progress', progress: 0 },
      { id: '2', name: 'Identifying Script Structure', status: 'pending' },
      { id: '3', name: 'Parsing Acts', status: 'pending' },
      { id: '4', name: 'Parsing Scenes', status: 'pending' },
      { id: '5', name: 'Generating Coverage Analysis', status: 'pending' },
    ])

    simulateParsing()
  }

  const simulateParsing = () => {
    let currentStep = 0
    const steps = ['1', '2', '3', '4', '5']

    const interval = setInterval(() => {
      setParsingSteps(prev => {
        const newSteps = [...prev]
        const step = newSteps.find(s => s.id === steps[currentStep])

        if (step) {
          if (step.progress !== undefined && step.progress < 100) {
            step.progress += 20
          } else if (step.progress === 100 || step.progress === undefined) {
            step.status = 'completed'
            currentStep++

            if (currentStep < steps.length) {
              const nextStep = newSteps.find(s => s.id === steps[currentStep])
              if (nextStep) {
                nextStep.status = 'in_progress'
                nextStep.progress = 0
              }
            } else {
              clearInterval(interval)
              setTimeout(() => {
                setIsParsing(false)
                setViewMode('editor')
                setShowEditor(true)
              }, 1000)
            }
          }
        }

        return newSteps
      })
    }, 500)
  }

  const handleDelete = (id: string) => {
    if (confirm('Delete this coverage report?')) {
      deleteCoverage(id)
    }
  }

  const getRecommendationColor = (type: string) => {
    switch (type) {
      case 'recommend':
        return 'text-mint-400 bg-mint-500/10 border-mint-500/30'
      case 'consider':
        return 'text-electric-400 bg-electric-500/10 border-electric-500/30'
      case 'pass':
        return 'text-sunset-400 bg-sunset-500/10 border-sunset-500/30'
      default:
        return 'text-gray-400 bg-gray-500/10 border-gray-500/30'
    }
  }

  return (
    <div className="flex flex-col h-full overflow-hidden">
      {/* Header */}
      <div className="p-6 border-b border-gray-800 bg-gray-900/50">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold text-sunset-400 flex items-center gap-3">
              <span className="text-3xl">🔍</span>
              Script Coverage
            </h2>
            <p className="text-gray-400 mt-1 text-sm">
              Professional script analysis and evaluation
            </p>
          </div>

          <div className="flex items-center gap-3">
            <button
              onClick={() => setViewMode('example')}
              className="px-6 py-3 bg-electric-500/20 hover:bg-electric-500/30 border border-electric-500/50 rounded-lg text-electric-400 font-medium transition-all"
            >
              View Example
            </button>
            <button
              onClick={handleCreateNew}
              className="px-6 py-3 bg-sunset-500/20 hover:bg-sunset-500/30 border border-sunset-500/50 rounded-lg text-sunset-400 font-medium transition-all flex items-center gap-2 anime-glow-sunset"
            >
              <span className="text-xl">+</span>
              <span>New Coverage</span>
            </button>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-auto">
        {isParsing ? (
          <ParsingProgress steps={parsingSteps} currentFile={currentFile} />
        ) : viewMode === 'example' ? (
          <CoverageDetailProfessional
            coverage={EXAMPLE_PROFESSIONAL_COVERAGE}
            onClose={() => setViewMode('list')}
            onEdit={() => {}}
          />
        ) : viewMode === 'detail' && selectedCoverage ? (
          <CoverageDetail
            coverage={selectedCoverage}
            onClose={() => setViewMode('list')}
            onEdit={handleEditFromDetail}
          />
        ) : viewMode === 'editor' ? (
          <CoverageEditor
            coverage={selectedCoverage || undefined}
            onClose={() => {
              setViewMode('list')
              setShowEditor(false)
            }}
          />
        ) : coverages.length === 0 ? (
          <ScriptUpload onUpload={handleFileUpload} />
        ) : (
          <div className="p-6">
            <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-4">
            {coverages.map((coverage) => (
              <div
                key={coverage.id}
                className="bg-gray-800/50 border border-gray-700 rounded-xl p-5 hover:bg-gray-800/70 hover:border-gray-600 transition-all cursor-pointer"
                onClick={() => handleViewDetail(coverage)}
              >
                <div className="flex items-start justify-between mb-3">
                  <div className="flex-1">
                    <h3 className="font-bold text-gray-200 mb-1">{coverage.script_title}</h3>
                    <p className="text-xs text-gray-500">by {coverage.submitted_by}</p>
                  </div>
                  <button
                    onClick={(e) => {
                      e.stopPropagation()
                      handleDelete(coverage.id)
                    }}
                    className="p-1.5 rounded text-gray-400 hover:text-sunset-400"
                  >
                    🗑️
                  </button>
                </div>

                <p className="text-sm text-gray-400 mb-3 line-clamp-2">{coverage.logline}</p>

                <div className="mb-3">
                  <div className="flex items-center justify-between mb-1">
                    <span className="text-xs text-gray-500">Overall Rating</span>
                    <span className="text-sm font-bold text-sunset-400">{coverage.ratings.overall}/10</span>
                  </div>
                  <div className="h-2 bg-gray-700 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-gradient-to-r from-sunset-500 to-sakura-500 transition-all"
                      style={{ width: `${(coverage.ratings.overall / 10) * 100}%` }}
                    />
                  </div>
                </div>

                <div className="flex items-center justify-between mb-3">
                  <span
                    className={`px-3 py-1 rounded-full text-xs font-medium border ${getRecommendationColor(
                      coverage.recommendation.type
                    )}`}
                  >
                    {coverage.recommendation.type.toUpperCase()}
                  </span>
                  <span className="text-xs text-gray-500">
                    {new Date(coverage.submitted_date).toLocaleDateString()}
                  </span>
                </div>

                {coverage.genre.length > 0 && (
                  <div className="flex flex-wrap gap-1">
                    {coverage.genre.slice(0, 3).map((g) => (
                      <span key={g} className="px-2 py-0.5 bg-gray-700/50 text-gray-400 text-xs rounded">
                        {g}
                      </span>
                    ))}
                    {coverage.genre.length > 3 && (
                      <span className="px-2 py-0.5 bg-gray-700/50 text-gray-400 text-xs rounded">
                        +{coverage.genre.length - 3}
                      </span>
                    )}
                  </div>
                )}
              </div>
            ))}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
