import { useState } from 'react'
import type { ProfessionalCoverage } from '../../types/coverage-professional'
import StructureView from './StructureView'
import SceneAnalysisView from './SceneAnalysisView'
import CharacterArcsView from './CharacterArcsView'
import ThematicAnalysisView from './ThematicAnalysisView'
import IndustryIntelView from './IndustryIntelView'
import DialogueAnalysisView from './DialogueAnalysisView'
import VisualStorytellingView from './VisualStorytellingView'

interface Props {
  coverage: ProfessionalCoverage
  onClose: () => void
  onEdit: () => void
}

type ViewMode = 'executive' | 'structure' | 'scenes' | 'characters' | 'themes' | 'industry' | 'dialogue' | 'visual'

export default function CoverageDetailProfessional({ coverage, onClose, onEdit }: Props) {
  const [viewMode, setViewMode] = useState<ViewMode>('executive')

  const navItems: { id: ViewMode; label: string; icon: string }[] = [
    { id: 'executive', label: 'Executive Summary', icon: '📊' },
    { id: 'structure', label: 'Structure & Acts', icon: '🏗️' },
    { id: 'scenes', label: 'Scene Analysis', icon: '🎬' },
    { id: 'characters', label: 'Character Arcs', icon: '👥' },
    { id: 'themes', label: 'Thematic Analysis', icon: '💭' },
    { id: 'industry', label: 'Industry Intel', icon: '🎯' },
    { id: 'dialogue', label: 'Dialogue', icon: '💬' },
    { id: 'visual', label: 'Visual Storytelling', icon: '📹' },
  ]

  return (
    <div className="flex h-full overflow-hidden bg-gray-900">
      {/* Sidebar Navigation */}
      <div className="w-64 border-r border-gray-800 bg-gray-900/50 flex flex-col">
        {/* Header */}
        <div className="p-4 border-b border-gray-800">
          <h3 className="font-bold text-gray-200 truncate">{coverage.title}</h3>
          <p className="text-xs text-gray-500 mt-1">by {coverage.author}</p>
          <div className="mt-2 flex items-center gap-2">
            <span className="px-2 py-1 bg-electric-500/10 border border-electric-500/30 text-electric-400 text-xs rounded">
              {coverage.consensus_rating}/10
            </span>
            <span className="text-xs text-gray-500">{coverage.page_count}pp</span>
          </div>
        </div>

        {/* Navigation */}
        <nav className="flex-1 p-2 overflow-auto">
          {navItems.map((item) => (
            <button
              key={item.id}
              onClick={() => setViewMode(item.id)}
              className={`
                w-full flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-all mb-1
                ${viewMode === item.id
                  ? 'bg-sunset-500/20 border border-sunset-500/50 text-sunset-400'
                  : 'hover:bg-gray-800/50 text-gray-400 hover:text-gray-300'
                }
              `}
            >
              <span>{item.icon}</span>
              <span>{item.label}</span>
            </button>
          ))}
        </nav>

        {/* Actions */}
        <div className="p-4 border-t border-gray-800 space-y-2">
          <button
            onClick={onEdit}
            className="w-full px-4 py-2 bg-electric-500/20 hover:bg-electric-500/30 border border-electric-500/50 rounded-lg text-electric-400 text-sm transition-all"
          >
            Edit Analysis
          </button>
          <button
            onClick={onClose}
            className="w-full px-4 py-2 bg-gray-700/50 hover:bg-gray-700 border border-gray-600 rounded-lg text-gray-300 text-sm transition-all"
          >
            ← Back
          </button>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 overflow-auto">
        {viewMode === 'executive' && <ExecutiveSummaryView coverage={coverage} />}
        {viewMode === 'structure' && <StructureView coverage={coverage} />}
        {viewMode === 'scenes' && <SceneAnalysisView coverage={coverage} />}
        {viewMode === 'characters' && <CharacterArcsView coverage={coverage} />}
        {viewMode === 'themes' && <ThematicAnalysisView coverage={coverage} />}
        {viewMode === 'industry' && <IndustryIntelView coverage={coverage} />}
        {viewMode === 'dialogue' && <DialogueAnalysisView coverage={coverage} />}
        {viewMode === 'visual' && <VisualStorytellingView coverage={coverage} />}
      </div>
    </div>
  )
}

// Executive Summary View
function ExecutiveSummaryView({ coverage }: { coverage: ProfessionalCoverage }) {
  return (
    <div className="p-8 space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-100 mb-2">{coverage.title}</h1>
        <p className="text-gray-400">Written by {coverage.author}</p>
        <div className="flex items-center gap-3 mt-3">
          {coverage.genre.map((g) => (
            <span key={g} className="px-3 py-1 bg-electric-500/10 border border-electric-500/30 text-electric-400 text-sm rounded">
              {g}
            </span>
          ))}
        </div>
      </div>

      {/* Rating */}
      <div className="bg-gray-800/30 border border-gray-700 rounded-xl p-6">
        <div className="flex items-center justify-between mb-2">
          <h2 className="text-xl font-bold text-sunset-400">Consensus Rating</h2>
          <span className="text-4xl font-bold text-sunset-400">{coverage.consensus_rating}/10</span>
        </div>
        <div className="h-3 bg-gray-700 rounded-full overflow-hidden">
          <div
            className="h-full bg-gradient-to-r from-sunset-500 to-electric-500 transition-all"
            style={{ width: `${(coverage.consensus_rating / 10) * 100}%` }}
          />
        </div>
      </div>

      {/* Logline */}
      <div className="bg-gray-800/30 border border-gray-700 rounded-xl p-6">
        <h2 className="text-lg font-bold text-sunset-400 mb-3">Logline</h2>
        <p className="text-gray-300 leading-relaxed text-lg">{coverage.logline}</p>
      </div>

      {/* Executive Summary */}
      <div className="bg-gray-800/30 border border-gray-700 rounded-xl p-6">
        <h2 className="text-lg font-bold text-sunset-400 mb-3">Executive Summary</h2>
        <p className="text-gray-300 leading-relaxed whitespace-pre-line">{coverage.executive_summary}</p>
      </div>

      {/* Strengths & Development Areas */}
      <div className="grid md:grid-cols-2 gap-6">
        <div className="bg-mint-500/5 border border-mint-500/30 rounded-xl p-6">
          <h2 className="text-lg font-bold text-mint-400 mb-4">Strengths</h2>
          <ul className="space-y-2">
            {coverage.strengths.map((strength, i) => (
              <li key={i} className="flex items-start gap-2 text-gray-300">
                <span className="text-mint-400 mt-1">✓</span>
                <span>{strength}</span>
              </li>
            ))}
          </ul>
        </div>

        <div className="bg-electric-500/5 border border-electric-500/30 rounded-xl p-6">
          <h2 className="text-lg font-bold text-electric-400 mb-4">Areas for Development</h2>
          <ul className="space-y-2">
            {coverage.areas_for_development.map((area, i) => (
              <li key={i} className="flex items-start gap-2 text-gray-300">
                <span className="text-electric-400 mt-1">→</span>
                <span>{area}</span>
              </li>
            ))}
          </ul>
        </div>
      </div>

      {/* Recommendation */}
      <div className="bg-gray-800/30 border border-gray-700 rounded-xl p-6">
        <div className="flex items-center gap-4 mb-4">
          <h2 className="text-lg font-bold text-sunset-400">Recommendation</h2>
          <span className={`px-4 py-2 rounded-lg text-sm font-bold border ${
            coverage.recommendation.verdict === 'RECOMMEND'
              ? 'bg-mint-500/10 border-mint-500/30 text-mint-400'
              : coverage.recommendation.verdict === 'CONSIDER'
              ? 'bg-electric-500/10 border-electric-500/30 text-electric-400'
              : 'bg-sunset-500/10 border-sunset-500/30 text-sunset-400'
          }`}>
            {coverage.recommendation.verdict}
          </span>
        </div>
        <p className="text-gray-300 leading-relaxed whitespace-pre-line mb-4">{coverage.recommendation.summary}</p>
        {coverage.recommendation.next_steps.length > 0 && (
          <div>
            <h3 className="text-sm font-semibold text-gray-400 mb-2">Next Steps:</h3>
            <ul className="space-y-1">
              {coverage.recommendation.next_steps.map((step, i) => (
                <li key={i} className="text-sm text-gray-400 flex items-start gap-2">
                  <span>•</span>
                  <span>{step}</span>
                </li>
              ))}
            </ul>
          </div>
        )}
      </div>
    </div>
  )
}

