import type { ProfessionalCoverage } from '../../types/coverage-professional'

export default function ThematicAnalysisView({ coverage }: { coverage: ProfessionalCoverage }) {
  const { thematic_analysis } = coverage

  return (
    <div className="p-8 space-y-8">
      <div>
        <h1 className="text-3xl font-bold text-gray-100 mb-2">Thematic Analysis</h1>
        <p className="text-gray-400">Deep dive into themes, motifs, and philosophical positioning</p>
      </div>

      {/* Philosophical Position */}
      <div className="bg-sunset-500/5 border border-sunset-500/30 rounded-xl p-6">
        <h2 className="text-xl font-bold text-sunset-400 mb-3">Philosophical Position</h2>
        <p className="text-gray-300 leading-relaxed">{thematic_analysis.philosophical_position}</p>
      </div>

      {/* Primary Themes */}
      <div>
        <h2 className="text-2xl font-bold text-electric-400 mb-4">Primary Themes</h2>
        <div className="space-y-4">
          {thematic_analysis.primary_themes.map((theme, i) => (
            <div
              key={i}
              className="bg-gray-800/30 border border-gray-700 rounded-xl p-6"
            >
              <h3 className="text-xl font-bold text-gray-100 mb-3">{theme.theme}</h3>

              <div className="space-y-3">
                <div>
                  <div className="text-sm font-semibold text-electric-400 mb-1">Analysis</div>
                  <p className="text-sm text-gray-300 leading-relaxed">{theme.analysis}</p>
                </div>

                <div className="bg-electric-500/5 border border-electric-500/20 rounded-lg p-3">
                  <div className="text-xs font-semibold text-electric-400 mb-1">Sophistication Notes</div>
                  <p className="text-xs text-gray-300">{theme.sophistication_notes}</p>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Visual Motifs */}
      <div className="bg-sakura-500/5 border border-sakura-500/30 rounded-xl p-6">
        <h2 className="text-xl font-bold text-sakura-400 mb-4">Visual Motifs</h2>
        <div className="grid md:grid-cols-2 gap-3">
          {thematic_analysis.visual_motifs.map((motif, i) => (
            <div
              key={i}
              className="bg-gray-800/50 border border-gray-700 rounded-lg p-3 flex items-center gap-2"
            >
              <span className="text-sakura-400">◆</span>
              <span className="text-sm text-gray-300">{motif}</span>
            </div>
          ))}
        </div>
      </div>

      {/* Symbolic Elements */}
      <div className="bg-mint-500/5 border border-mint-500/30 rounded-xl p-6">
        <h2 className="text-xl font-bold text-mint-400 mb-4">Symbolic Elements</h2>
        <div className="space-y-2">
          {thematic_analysis.symbolic_elements.map((element, i) => (
            <div
              key={i}
              className="flex items-start gap-3 p-3 bg-gray-800/50 border border-gray-700 rounded-lg"
            >
              <span className="text-mint-400 mt-0.5">✦</span>
              <span className="text-sm text-gray-300 flex-1">{element}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
