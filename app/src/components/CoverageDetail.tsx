import type { CoverageReport } from '../types/coverage'

interface CoverageDetailProps {
  coverage: CoverageReport
  onClose: () => void
  onEdit: () => void
}

export default function CoverageDetail({ coverage, onClose, onEdit }: CoverageDetailProps) {
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

  const getRatingColor = (rating: number) => {
    if (rating >= 8) return 'text-mint-400'
    if (rating >= 6) return 'text-electric-400'
    if (rating >= 4) return 'text-sakura-400'
    return 'text-sunset-400'
  }

  return (
    <div className="flex flex-col h-full overflow-hidden bg-gray-900">
      {/* Header */}
      <div className="p-6 border-b border-gray-800 bg-gray-900/50">
        <div className="flex items-center justify-between mb-4">
          <div className="flex-1">
            <h1 className="text-3xl font-bold text-gray-100 mb-2">{coverage.script_title}</h1>
            <div className="flex items-center gap-4 text-sm text-gray-400">
              <span>Coverage by {coverage.submitted_by}</span>
              <span>•</span>
              <span>{new Date(coverage.submitted_date).toLocaleDateString()}</span>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <button
              onClick={onEdit}
              className="px-4 py-2 bg-electric-500/20 hover:bg-electric-500/30 border border-electric-500/50 rounded-lg text-electric-400 transition-all"
            >
              Edit
            </button>
            <button
              onClick={onClose}
              className="px-4 py-2 bg-gray-700/50 hover:bg-gray-700 border border-gray-600 rounded-lg text-gray-300 transition-all"
            >
              ← Back
            </button>
          </div>
        </div>

        {/* Recommendation Badge */}
        <div className="flex items-center gap-4">
          <span
            className={`px-4 py-2 rounded-lg text-sm font-bold border ${getRecommendationColor(
              coverage.recommendation.type
            )}`}
          >
            {coverage.recommendation.type.toUpperCase()}
          </span>
          <div className="flex items-center gap-2">
            <span className="text-sm text-gray-400">Overall Rating:</span>
            <span className={`text-2xl font-bold ${getRatingColor(coverage.ratings.overall)}`}>
              {coverage.ratings.overall}/10
            </span>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-auto p-6 space-y-6">
        {/* Logline */}
        <div className="bg-gray-800/30 border border-gray-700 rounded-xl p-6">
          <h2 className="text-lg font-bold text-sunset-400 mb-3">Logline</h2>
          <p className="text-gray-300 leading-relaxed">{coverage.logline}</p>
        </div>

        {/* Synopsis */}
        {coverage.synopsis && (
          <div className="bg-gray-800/30 border border-gray-700 rounded-xl p-6">
            <h2 className="text-lg font-bold text-sunset-400 mb-3">Synopsis</h2>
            <p className="text-gray-300 leading-relaxed whitespace-pre-line">{coverage.synopsis}</p>
          </div>
        )}

        {/* Ratings Grid */}
        <div className="bg-gray-800/30 border border-gray-700 rounded-xl p-6">
          <h2 className="text-lg font-bold text-sunset-400 mb-4">Category Ratings</h2>
          <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
            {Object.entries(coverage.ratings).map(([key, value]) => (
              <div key={key} className="bg-gray-800/50 border border-gray-700 rounded-lg p-4">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm text-gray-400 capitalize">{key}</span>
                  <span className={`text-xl font-bold ${getRatingColor(value)}`}>{value}</span>
                </div>
                <div className="h-2 bg-gray-700 rounded-full overflow-hidden">
                  <div
                    className={`h-full transition-all ${
                      value >= 8
                        ? 'bg-mint-500'
                        : value >= 6
                        ? 'bg-electric-500'
                        : value >= 4
                        ? 'bg-sakura-500'
                        : 'bg-sunset-500'
                    }`}
                    style={{ width: `${(value / 10) * 100}%` }}
                  />
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Strengths & Weaknesses */}
        <div className="grid md:grid-cols-2 gap-6">
          {/* Strengths */}
          <div className="bg-mint-500/5 border border-mint-500/30 rounded-xl p-6">
            <h2 className="text-lg font-bold text-mint-400 mb-4">Strengths</h2>
            {coverage.analysis.strengths.length > 0 ? (
              <ul className="space-y-2">
                {coverage.analysis.strengths.map((strength, i) => (
                  <li key={i} className="flex items-start gap-2 text-gray-300">
                    <span className="text-mint-400 mt-1">✓</span>
                    <span>{strength}</span>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="text-gray-500 italic">No strengths listed</p>
            )}
          </div>

          {/* Weaknesses */}
          <div className="bg-sunset-500/5 border border-sunset-500/30 rounded-xl p-6">
            <h2 className="text-lg font-bold text-sunset-400 mb-4">Weaknesses</h2>
            {coverage.analysis.weaknesses.length > 0 ? (
              <ul className="space-y-2">
                {coverage.analysis.weaknesses.map((weakness, i) => (
                  <li key={i} className="flex items-start gap-2 text-gray-300">
                    <span className="text-sunset-400 mt-1">✗</span>
                    <span>{weakness}</span>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="text-gray-500 italic">No weaknesses listed</p>
            )}
          </div>
        </div>

        {/* Recommendation Summary */}
        <div className="bg-gray-800/30 border border-gray-700 rounded-xl p-6">
          <h2 className="text-lg font-bold text-sunset-400 mb-3">Recommendation</h2>
          <div className="mb-4">
            <span
              className={`inline-block px-4 py-2 rounded-lg text-sm font-bold border ${getRecommendationColor(
                coverage.recommendation.type
              )}`}
            >
              {coverage.recommendation.type.toUpperCase()}
            </span>
          </div>
          {coverage.recommendation.summary && (
            <p className="text-gray-300 leading-relaxed whitespace-pre-line">
              {coverage.recommendation.summary}
            </p>
          )}
        </div>

        {/* Genres */}
        {coverage.genre.length > 0 && (
          <div className="bg-gray-800/30 border border-gray-700 rounded-xl p-6">
            <h2 className="text-lg font-bold text-sunset-400 mb-3">Genre</h2>
            <div className="flex flex-wrap gap-2">
              {coverage.genre.map((g) => (
                <span
                  key={g}
                  className="px-3 py-1 bg-electric-500/10 border border-electric-500/30 text-electric-400 rounded-full text-sm"
                >
                  {g}
                </span>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
