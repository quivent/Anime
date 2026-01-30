import { useState } from 'react'
import type { ProfessionalCoverage, SceneAnalysis } from '../../types/coverage-professional'

export default function SceneAnalysisView({ coverage }: { coverage: ProfessionalCoverage }) {
  const [expandedScene, setExpandedScene] = useState<number | null>(null)
  const [filter, setFilter] = useState<'all' | 'high-value' | 'trim-candidates'>('all')

  const filteredScenes = coverage.scene_analyses.filter((scene) => {
    if (filter === 'high-value') return scene.dialogue_quality_rating >= 8
    if (filter === 'trim-candidates') return scene.production_considerations.budget_impact === 'LOW'
    return true
  })

  return (
    <div className="p-8 space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-100 mb-2">Scene-by-Scene Analysis</h1>
          <p className="text-gray-400">{coverage.scene_analyses.length} scenes analyzed</p>
        </div>

        {/* Filter */}
        <div className="flex gap-2">
          {(['all', 'high-value', 'trim-candidates'] as const).map((f) => (
            <button
              key={f}
              onClick={() => setFilter(f)}
              className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${
                filter === f
                  ? 'bg-sunset-500/20 border border-sunset-500/50 text-sunset-400'
                  : 'bg-gray-800/50 border border-gray-700 text-gray-400 hover:text-gray-300'
              }`}
            >
              {f === 'all' ? 'All Scenes' : f === 'high-value' ? 'High Value (8+)' : 'Trim Candidates'}
            </button>
          ))}
        </div>
      </div>

      {/* Scene Cards */}
      <div className="space-y-3">
        {filteredScenes.map((scene) => (
          <SceneCard
            key={scene.scene_number}
            scene={scene}
            expanded={expandedScene === scene.scene_number}
            onToggle={() => setExpandedScene(expandedScene === scene.scene_number ? null : scene.scene_number)}
          />
        ))}
      </div>
    </div>
  )
}

function SceneCard({ scene, expanded, onToggle }: { scene: SceneAnalysis; expanded: boolean; onToggle: () => void }) {
  const getRatingColor = (rating: number) => {
    if (rating >= 9) return 'text-mint-400 bg-mint-500/10 border-mint-500/30'
    if (rating >= 7) return 'text-electric-400 bg-electric-500/10 border-electric-500/30'
    if (rating >= 5) return 'text-sakura-400 bg-sakura-500/10 border-sakura-500/30'
    return 'text-sunset-400 bg-sunset-500/10 border-sunset-500/30'
  }

  const getBudgetColor = (impact: string) => {
    if (impact === 'LOW') return 'text-mint-400'
    if (impact === 'MEDIUM') return 'text-electric-400'
    return 'text-sunset-400'
  }

  return (
    <div className="bg-gray-800/30 border border-gray-700 rounded-xl overflow-hidden transition-all">
      {/* Header - Always Visible */}
      <div
        className="p-4 cursor-pointer hover:bg-gray-800/50 transition-all"
        onClick={onToggle}
      >
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4 flex-1">
            <div className="flex items-center gap-3">
              <span className="text-2xl font-bold text-gray-500">#{scene.scene_number}</span>
              <div>
                <div className="font-semibold text-gray-200">{scene.scene_heading}</div>
                <div className="text-xs text-gray-500">Pages {scene.page_start}-{scene.page_end}</div>
              </div>
            </div>
          </div>

          <div className="flex items-center gap-3">
            <span className={`px-3 py-1 rounded-full text-xs font-bold border ${getRatingColor(scene.dialogue_quality_rating)}`}>
              {scene.dialogue_quality_rating}/10
            </span>
            <span className={`text-xs font-semibold ${getBudgetColor(scene.production_considerations.budget_impact)}`}>
              {scene.production_considerations.budget_impact}
            </span>
            <span className="text-gray-500">{expanded ? '▼' : '▶'}</span>
          </div>
        </div>
      </div>

      {/* Expanded Content */}
      {expanded && (
        <div className="p-6 pt-0 space-y-4 border-t border-gray-700">
          {/* Scene Function */}
          <div>
            <h3 className="text-sm font-bold text-sunset-400 mb-2">Scene Function</h3>
            <p className="text-sm text-gray-300">{scene.scene_function}</p>
          </div>

          {/* Commercial Appeal */}
          <div className="grid md:grid-cols-3 gap-4">
            <div className="bg-gray-800/50 border border-gray-700 rounded-lg p-3">
              <div className="text-xs text-gray-500 mb-1 font-semibold">Festival Circuit</div>
              <div className="text-xs text-gray-300">{scene.commercial_appeal.festival_circuit}</div>
            </div>
            <div className="bg-gray-800/50 border border-gray-700 rounded-lg p-3">
              <div className="text-xs text-gray-500 mb-1 font-semibold">Actor Showcase</div>
              <div className="text-xs text-gray-300">{scene.commercial_appeal.actor_showcase}</div>
            </div>
            <div className="bg-gray-800/50 border border-gray-700 rounded-lg p-3">
              <div className="text-xs text-gray-500 mb-1 font-semibold">International</div>
              <div className="text-xs text-gray-300">{scene.commercial_appeal.international}</div>
            </div>
          </div>

          {/* Actor Appeal */}
          <div>
            <h3 className="text-sm font-bold text-electric-400 mb-2">Actor Appeal</h3>
            <div className="space-y-2">
              <div className="text-sm text-gray-300">{scene.actor_appeal.role_significance}</div>
              {scene.actor_appeal.comp_casting.length > 0 && (
                <div>
                  <div className="text-xs text-gray-500 mb-1">Comp Casting:</div>
                  <div className="flex flex-wrap gap-2">
                    {scene.actor_appeal.comp_casting.map((actor) => (
                      <span key={actor} className="px-2 py-1 bg-electric-500/10 text-electric-400 text-xs rounded">
                        {actor}
                      </span>
                    ))}
                  </div>
                </div>
              )}
              {scene.actor_appeal.oscar_potential && (
                <div className="text-xs text-mint-400 font-semibold">⭐ {scene.actor_appeal.oscar_potential}</div>
              )}
            </div>
          </div>

          {/* Key Exchanges */}
          {scene.key_exchanges.length > 0 && (
            <div>
              <h3 className="text-sm font-bold text-sakura-400 mb-2">Key Exchanges (Trailer/Marketing)</h3>
              <div className="space-y-2">
                {scene.key_exchanges.map((exchange, i) => (
                  <div key={i} className="bg-sakura-500/5 border border-sakura-500/20 rounded p-2 text-xs text-gray-300 font-mono">
                    {exchange}
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Strengths & Concerns */}
          <div className="grid md:grid-cols-2 gap-4">
            {scene.strengths.length > 0 && (
              <div>
                <h3 className="text-sm font-bold text-mint-400 mb-2">Strengths</h3>
                <ul className="space-y-1">
                  {scene.strengths.map((strength, i) => (
                    <li key={i} className="flex items-start gap-2 text-xs text-gray-300">
                      <span className="text-mint-400 mt-0.5">✓</span>
                      <span>{strength}</span>
                    </li>
                  ))}
                </ul>
              </div>
            )}
            {scene.concerns.length > 0 && (
              <div>
                <h3 className="text-sm font-bold text-sunset-400 mb-2">Concerns</h3>
                <ul className="space-y-1">
                  {scene.concerns.map((concern, i) => (
                    <li key={i} className="flex items-start gap-2 text-xs text-gray-300">
                      <span className="text-sunset-400 mt-0.5">!</span>
                      <span>{concern}</span>
                    </li>
                  ))}
                </ul>
              </div>
            )}
          </div>

          {/* Production Considerations */}
          <div className="bg-gray-800/50 border border-gray-700 rounded-lg p-3">
            <h3 className="text-xs font-bold text-gray-400 mb-2">Production Considerations</h3>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-3 text-xs">
              <div>
                <div className="text-gray-500 mb-1">Budget Impact</div>
                <div className={`font-semibold ${getBudgetColor(scene.production_considerations.budget_impact)}`}>
                  {scene.production_considerations.budget_impact}
                </div>
              </div>
              <div>
                <div className="text-gray-500 mb-1">Shooting Days</div>
                <div className="text-gray-300">{scene.production_considerations.shooting_days}</div>
              </div>
              <div>
                <div className="text-gray-500 mb-1">Location</div>
                <div className="text-gray-300">{scene.production_considerations.location_requirements}</div>
              </div>
              <div>
                <div className="text-gray-500 mb-1">VFX</div>
                <div className="text-gray-300">{scene.production_considerations.vfx_requirements}</div>
              </div>
            </div>
          </div>

          {/* Notes */}
          {(scene.dialogue_notes || scene.pacing_notes || scene.marketing_implications) && (
            <div className="space-y-2">
              {scene.dialogue_notes && (
                <div className="text-xs">
                  <span className="text-gray-500 font-semibold">Dialogue Notes: </span>
                  <span className="text-gray-300">{scene.dialogue_notes}</span>
                </div>
              )}
              {scene.pacing_notes && (
                <div className="text-xs">
                  <span className="text-gray-500 font-semibold">Pacing: </span>
                  <span className="text-gray-300">{scene.pacing_notes}</span>
                </div>
              )}
              {scene.marketing_implications && (
                <div className="text-xs">
                  <span className="text-gray-500 font-semibold">Marketing: </span>
                  <span className="text-gray-300">{scene.marketing_implications}</span>
                </div>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
