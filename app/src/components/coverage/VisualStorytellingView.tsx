import type { ProfessionalCoverage } from '../../types/coverage-professional'

export default function VisualStorytellingView({ coverage }: { coverage: ProfessionalCoverage }) {
  const { visual_storytelling } = coverage

  return (
    <div className="p-8 space-y-8">
      <div>
        <h1 className="text-3xl font-bold text-gray-100 mb-2">Visual Storytelling</h1>
        <p className="text-gray-400">Cinematic techniques, visual language, and camera-aware writing</p>
      </div>

      {/* Overall Assessment */}
      <div className="bg-sunset-500/5 border border-sunset-500/30 rounded-xl p-6">
        <h2 className="text-xl font-bold text-sunset-400 mb-3">Overall Assessment</h2>
        <p className="text-gray-300 leading-relaxed">{visual_storytelling.overall_assessment}</p>
      </div>

      {/* Techniques */}
      <div>
        <h2 className="text-2xl font-bold text-electric-400 mb-4">Visual Techniques</h2>
        <div className="space-y-3">
          {visual_storytelling.techniques.map((technique, i) => (
            <div
              key={i}
              className="bg-gray-800/30 border border-gray-700 rounded-xl p-4 flex items-start gap-3"
            >
              <div className="flex-shrink-0 w-8 h-8 bg-electric-500/10 border border-electric-500/30 rounded-lg flex items-center justify-center">
                <span className="text-electric-400 font-bold">{i + 1}</span>
              </div>
              <p className="text-sm text-gray-300 flex-1">{technique}</p>
            </div>
          ))}
        </div>
      </div>

      {/* Environment as Emotion */}
      <div className="bg-sakura-500/5 border border-sakura-500/30 rounded-xl p-6">
        <h2 className="text-xl font-bold text-sakura-400 mb-4">Environment as Emotion</h2>
        <p className="text-xs text-gray-500 mb-3">
          How setting, lighting, and space reflect character psychology
        </p>
        <div className="space-y-2">
          {visual_storytelling.environment_as_emotion.map((env, i) => (
            <div
              key={i}
              className="bg-gray-800/50 border border-gray-700 rounded-lg p-3"
            >
              <p className="text-sm text-gray-300 font-mono">"{env}"</p>
            </div>
          ))}
        </div>
      </div>

      {/* Camera-Aware Notes */}
      <div className="bg-mint-500/5 border border-mint-500/30 rounded-xl p-6">
        <h2 className="text-xl font-bold text-mint-400 mb-4">Camera-Aware Writing</h2>
        <p className="text-xs text-gray-500 mb-3">
          Demonstrates understanding of cinematic language and visual composition
        </p>
        <div className="space-y-2">
          {visual_storytelling.camera_aware_notes.map((note, i) => (
            <div
              key={i}
              className="bg-gray-800/50 border border-gray-700 rounded-lg p-3"
            >
              <p className="text-sm text-gray-300 font-mono italic">"{note}"</p>
            </div>
          ))}
        </div>
      </div>

      {/* Insight Box */}
      <div className="bg-electric-500/5 border border-electric-500/30 rounded-xl p-6">
        <h3 className="text-sm font-bold text-electric-400 mb-2">Screenplay Insight</h3>
        <p className="text-sm text-gray-300 leading-relaxed">
          Strong visual storytelling is essential for prestige cinema. This screenplay demonstrates professional-level
          understanding of how to write for the camera - using action lines, environmental description, and visual motifs
          to convey psychology and theme without dialogue. This level of craft significantly increases the script's
          appeal to directors and cinematographers.
        </p>
      </div>
    </div>
  )
}
