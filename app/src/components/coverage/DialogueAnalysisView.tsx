import type { ProfessionalCoverage } from '../../types/coverage-professional'

export default function DialogueAnalysisView({ coverage }: { coverage: ProfessionalCoverage }) {
  const { dialogue_analysis } = coverage

  return (
    <div className="p-8 space-y-8">
      <div>
        <h1 className="text-3xl font-bold text-gray-100 mb-2">Dialogue Analysis</h1>
        <p className="text-gray-400">Character voices, subtext, and linguistic patterns</p>
      </div>

      {/* Overall Quality */}
      <div className="bg-electric-500/5 border border-electric-500/30 rounded-xl p-6">
        <h2 className="text-xl font-bold text-electric-400 mb-3">Overall Assessment</h2>
        <p className="text-gray-300 leading-relaxed">{dialogue_analysis.overall_quality}</p>
      </div>

      {/* Character Voices */}
      <div>
        <h2 className="text-2xl font-bold text-sunset-400 mb-4">Character Voices</h2>
        <div className="space-y-4">
          {dialogue_analysis.character_voices.map((voice) => (
            <div
              key={voice.character}
              className="bg-gray-800/30 border border-gray-700 rounded-xl p-6"
            >
              <h3 className="text-xl font-bold text-gray-100 mb-3">{voice.character}</h3>

              <div className="mb-4">
                <p className="text-sm text-gray-400 leading-relaxed">{voice.voice_description}</p>
              </div>

              <div>
                <div className="text-sm font-semibold text-sakura-400 mb-2">Examples</div>
                <div className="space-y-2">
                  {voice.examples.map((example, i) => (
                    <div
                      key={i}
                      className="bg-gray-800/50 border border-gray-700 rounded-lg p-3"
                    >
                      <p className="text-sm text-gray-300 font-mono italic">"{example}"</p>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Subtext Examples */}
      <div>
        <h2 className="text-2xl font-bold text-mint-400 mb-4">Subtext Analysis</h2>
        <p className="text-sm text-gray-400 mb-4">
          Key exchanges where subtext diverges from surface meaning - essential for understanding character psychology
        </p>
        <div className="space-y-4">
          {dialogue_analysis.subtext_examples.map((example, i) => (
            <div
              key={i}
              className="bg-gray-800/30 border border-gray-700 rounded-xl p-6"
            >
              <div className="flex items-center gap-2 mb-4">
                <span className="px-3 py-1 bg-mint-500/10 border border-mint-500/30 text-mint-400 text-xs font-bold rounded-full">
                  Scene {example.scene}
                </span>
              </div>

              <div className="space-y-4">
                {/* Exchange */}
                <div>
                  <div className="text-xs font-semibold text-gray-500 mb-2">EXCHANGE</div>
                  <div className="bg-gray-800/50 border border-gray-700 rounded-lg p-4">
                    <p className="text-sm text-gray-200 font-mono whitespace-pre-line">{example.exchange}</p>
                  </div>
                </div>

                {/* Surface vs Subtext */}
                <div className="grid md:grid-cols-2 gap-4">
                  <div>
                    <div className="text-xs font-semibold text-electric-400 mb-2">SURFACE MEANING</div>
                    <div className="bg-electric-500/5 border border-electric-500/20 rounded-lg p-3">
                      <p className="text-sm text-gray-300">{example.surface}</p>
                    </div>
                  </div>
                  <div>
                    <div className="text-xs font-semibold text-sunset-400 mb-2">SUBTEXT</div>
                    <div className="bg-sunset-500/5 border border-sunset-500/20 rounded-lg p-3">
                      <p className="text-sm text-gray-300">{example.subtext}</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
