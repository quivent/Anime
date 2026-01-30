import type { ProfessionalCoverage } from '../../types/coverage-professional'

export default function StructureView({ coverage }: { coverage: ProfessionalCoverage }) {
  const { structural_analysis } = coverage

  return (
    <div className="p-8 space-y-8">
      <div>
        <h1 className="text-3xl font-bold text-gray-100 mb-2">Structural Analysis</h1>
        <p className="text-gray-400">{structural_analysis.scene_count} scenes across {structural_analysis.page_count} pages</p>
      </div>

      {/* Overview Stats */}
      <div className="grid grid-cols-4 gap-4">
        <div className="bg-gray-800/30 border border-gray-700 rounded-xl p-4">
          <div className="text-sm text-gray-500 mb-1">Structure</div>
          <div className="text-2xl font-bold text-electric-400 capitalize">{structural_analysis.act_structure}</div>
        </div>
        <div className="bg-gray-800/30 border border-gray-700 rounded-xl p-4">
          <div className="text-sm text-gray-500 mb-1">Scene Count</div>
          <div className="text-2xl font-bold text-mint-400">{structural_analysis.scene_count}</div>
        </div>
        <div className="bg-gray-800/30 border border-gray-700 rounded-xl p-4">
          <div className="text-sm text-gray-500 mb-1">Pacing</div>
          <div className="text-sm font-semibold text-sakura-400">{structural_analysis.pacing_rhythm}</div>
        </div>
        <div className="bg-gray-800/30 border border-gray-700 rounded-xl p-4">
          <div className="text-sm text-gray-500 mb-1">Dialogue/Action</div>
          <div className="text-sm font-semibold text-sunset-400">{structural_analysis.dialogue_to_action_ratio}</div>
        </div>
      </div>

      {/* Act Breakdowns */}
      {structural_analysis.act_breakdowns.map((act) => (
        <div key={act.act_number} className="bg-gray-800/30 border border-gray-700 rounded-xl p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-2xl font-bold text-sunset-400">Act {act.act_number}</h2>
            <span className="px-4 py-1 bg-electric-500/10 border border-electric-500/30 text-electric-400 text-sm rounded-full">
              Pages {act.page_range}
            </span>
          </div>

          {/* Act Milestones */}
          <div className="grid md:grid-cols-2 gap-4 mb-6">
            {act.opening_image && (
              <div className="bg-gray-800/50 border border-gray-700 rounded-lg p-4">
                <div className="text-xs text-gray-500 mb-1 font-semibold">OPENING IMAGE</div>
                <div className="text-sm text-gray-300">{act.opening_image}</div>
              </div>
            )}
            {act.inciting_incident && (
              <div className="bg-gray-800/50 border border-gray-700 rounded-lg p-4">
                <div className="text-xs text-gray-500 mb-1 font-semibold">INCITING INCIDENT</div>
                <div className="text-sm text-gray-300">{act.inciting_incident}</div>
              </div>
            )}
            {act.midpoint && (
              <div className="bg-gray-800/50 border border-gray-700 rounded-lg p-4">
                <div className="text-xs text-gray-500 mb-1 font-semibold">MIDPOINT</div>
                <div className="text-sm text-gray-300">{act.midpoint}</div>
              </div>
            )}
            {act.turning_point && (
              <div className="bg-gray-800/50 border border-gray-700 rounded-lg p-4">
                <div className="text-xs text-gray-500 mb-1 font-semibold">TURNING POINT</div>
                <div className="text-sm text-gray-300">{act.turning_point}</div>
              </div>
            )}
            {act.climax && (
              <div className="bg-gray-800/50 border border-gray-700 rounded-lg p-4">
                <div className="text-xs text-gray-500 mb-1 font-semibold">CLIMAX</div>
                <div className="text-sm text-gray-300">{act.climax}</div>
              </div>
            )}
          </div>

          {/* Structural Strengths */}
          {act.structural_strengths.length > 0 && (
            <div className="mb-4">
              <h3 className="text-sm font-bold text-mint-400 mb-2">Structural Strengths</h3>
              <ul className="space-y-1">
                {act.structural_strengths.map((strength, i) => (
                  <li key={i} className="flex items-start gap-2 text-sm text-gray-300">
                    <span className="text-mint-400 mt-0.5">✓</span>
                    <span>{strength}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}

          {/* Pacing Observations */}
          {act.pacing_observations.length > 0 && (
            <div className="mb-4">
              <h3 className="text-sm font-bold text-electric-400 mb-2">Pacing Observations</h3>
              <ul className="space-y-1">
                {act.pacing_observations.map((obs, i) => (
                  <li key={i} className="flex items-start gap-2 text-sm text-gray-300">
                    <span className="text-electric-400 mt-0.5">→</span>
                    <span>{obs}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}

          {/* Trim Recommendations */}
          {act.trim_recommendations && act.trim_recommendations.length > 0 && (
            <div>
              <h3 className="text-sm font-bold text-sunset-400 mb-2">Trim Recommendations</h3>
              <ul className="space-y-1">
                {act.trim_recommendations.map((rec, i) => (
                  <li key={i} className="flex items-start gap-2 text-sm text-gray-300">
                    <span className="text-sunset-400 mt-0.5">✂</span>
                    <span>{rec}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
      ))}

      {/* Structural Innovations */}
      {structural_analysis.structural_innovations.length > 0 && (
        <div className="bg-electric-500/5 border border-electric-500/30 rounded-xl p-6">
          <h2 className="text-xl font-bold text-electric-400 mb-4">Structural Innovations</h2>
          <ul className="space-y-2">
            {structural_analysis.structural_innovations.map((innovation, i) => (
              <li key={i} className="flex items-start gap-2 text-gray-300">
                <span className="text-electric-400 mt-0.5">★</span>
                <span>{innovation}</span>
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* Overall Trim Recommendations */}
      {structural_analysis.trim_recommendations.length > 0 && (
        <div className="bg-sunset-500/5 border border-sunset-500/30 rounded-xl p-6">
          <h2 className="text-xl font-bold text-sunset-400 mb-4">Overall Trim Recommendations</h2>
          <div className="space-y-4">
            {structural_analysis.trim_recommendations.map((rec, i) => (
              <div key={i} className="bg-gray-800/50 border border-gray-700 rounded-lg p-4">
                <div className="flex items-center justify-between mb-2">
                  <div className="font-semibold text-gray-200">{rec.section}</div>
                  <div className="px-3 py-1 bg-sunset-500/10 border border-sunset-500/30 text-sunset-400 text-xs rounded-full">
                    {rec.pages}
                  </div>
                </div>
                <div className="text-sm text-gray-400">{rec.rationale}</div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
