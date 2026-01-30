import type { ProfessionalCoverage } from '../../types/coverage-professional'

export default function IndustryIntelView({ coverage }: { coverage: ProfessionalCoverage }) {
  const { industry_intelligence } = coverage

  const getAwardsPotentialColor = (potential: string) => {
    if (potential === 'VERY HIGH') return 'text-mint-400 bg-mint-500/10 border-mint-500/30'
    if (potential === 'HIGH') return 'text-electric-400 bg-electric-500/10 border-electric-500/30'
    if (potential === 'MODERATE') return 'text-sakura-400 bg-sakura-500/10 border-sakura-500/30'
    return 'text-gray-400 bg-gray-500/10 border-gray-500/30'
  }

  return (
    <div className="p-8 space-y-8">
      <div>
        <h1 className="text-3xl font-bold text-gray-100 mb-2">Industry Intelligence</h1>
        <p className="text-gray-400">Market positioning, comps, and commercial strategy</p>
      </div>

      {/* Key Metrics */}
      <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-gray-800/30 border border-gray-700 rounded-xl p-4">
          <div className="text-sm text-gray-500 mb-1">Budget Range</div>
          <div className="text-xl font-bold text-electric-400">{industry_intelligence.budget_range}</div>
        </div>
        <div className="bg-gray-800/30 border border-gray-700 rounded-xl p-4">
          <div className="text-sm text-gray-500 mb-1">Revenue Projection</div>
          <div className="text-xl font-bold text-mint-400">{industry_intelligence.revenue_projection}</div>
        </div>
        <div className="bg-gray-800/30 border border-gray-700 rounded-xl p-4">
          <div className="text-sm text-gray-500 mb-1">Awards Potential</div>
          <div className={`text-sm font-bold px-2 py-1 rounded inline-block ${getAwardsPotentialColor(industry_intelligence.awards_potential)}`}>
            {industry_intelligence.awards_potential}
          </div>
        </div>
        <div className="bg-gray-800/30 border border-gray-700 rounded-xl p-4">
          <div className="text-sm text-gray-500 mb-1">Casting Tier</div>
          <div className="text-sm font-semibold text-sunset-400">{industry_intelligence.casting_tier}</div>
        </div>
      </div>

      {/* Market Position */}
      <div className="bg-gray-800/30 border border-gray-700 rounded-xl p-6">
        <h2 className="text-xl font-bold text-sunset-400 mb-3">Market Position</h2>
        <p className="text-gray-300 leading-relaxed">{industry_intelligence.market_position}</p>
      </div>

      {/* Comparable Titles */}
      <div>
        <h2 className="text-xl font-bold text-electric-400 mb-4">Comparable Titles</h2>
        <div className="space-y-3">
          {industry_intelligence.comp_titles
            .sort((a, b) => b.similarity_percentage - a.similarity_percentage)
            .map((comp) => (
              <div
                key={comp.title}
                className="bg-gray-800/30 border border-gray-700 rounded-xl p-4"
              >
                <div className="flex items-start justify-between mb-2">
                  <div>
                    <h3 className="font-bold text-gray-200">{comp.title} ({comp.year})</h3>
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="text-xs text-gray-500">Similarity</div>
                    <div className={`text-lg font-bold ${
                      comp.similarity_percentage >= 90
                        ? 'text-mint-400'
                        : comp.similarity_percentage >= 80
                        ? 'text-electric-400'
                        : comp.similarity_percentage >= 70
                        ? 'text-sakura-400'
                        : 'text-sunset-400'
                    }`}>
                      {comp.similarity_percentage}%
                    </div>
                  </div>
                </div>

                {/* Similarity Bar */}
                <div className="h-2 bg-gray-700 rounded-full overflow-hidden mb-3">
                  <div
                    className={`h-full transition-all ${
                      comp.similarity_percentage >= 90
                        ? 'bg-mint-500'
                        : comp.similarity_percentage >= 80
                        ? 'bg-electric-500'
                        : comp.similarity_percentage >= 70
                        ? 'bg-sakura-500'
                        : 'bg-sunset-500'
                    }`}
                    style={{ width: `${comp.similarity_percentage}%` }}
                  />
                </div>

                <p className="text-sm text-gray-400">{comp.comparison_notes}</p>
              </div>
            ))}
        </div>
      </div>

      {/* Festival Strategy */}
      <div className="bg-sunset-500/5 border border-sunset-500/30 rounded-xl p-6">
        <h2 className="text-xl font-bold text-sunset-400 mb-3">Festival Strategy</h2>
        <p className="text-gray-300 leading-relaxed">{industry_intelligence.festival_strategy}</p>
      </div>

      {/* Target Distributors */}
      <div>
        <h2 className="text-xl font-bold text-mint-400 mb-4">Target Distributors</h2>
        <div className="grid md:grid-cols-3 gap-3">
          {industry_intelligence.target_distributors.map((distributor) => (
            <div
              key={distributor}
              className="bg-mint-500/5 border border-mint-500/30 rounded-lg p-4 text-center"
            >
              <div className="font-semibold text-mint-400">{distributor}</div>
            </div>
          ))}
        </div>
      </div>

      {/* Author Profile (if available) */}
      {coverage.author_profile && (
        <div className="bg-electric-500/5 border border-electric-500/30 rounded-xl p-6">
          <h2 className="text-xl font-bold text-electric-400 mb-4">Author Profile Assessment</h2>
          <div className="grid md:grid-cols-2 gap-4">
            <div>
              <div className="text-sm text-gray-500 mb-1">Estimated Age Range</div>
              <div className="text-sm font-semibold text-gray-200">{coverage.author_profile.estimated_age_range}</div>
            </div>
            <div>
              <div className="text-sm text-gray-500 mb-1">Education Indicators</div>
              <div className="text-sm font-semibold text-gray-200">{coverage.author_profile.education_indicators}</div>
            </div>
            <div>
              <div className="text-sm text-gray-500 mb-1">Sophistication Level</div>
              <div className="text-sm font-semibold text-electric-400">{coverage.author_profile.sophistication_level}</div>
            </div>
            <div>
              <div className="text-sm text-gray-500 mb-1">Industry Readiness</div>
              <div className="text-sm font-semibold text-mint-400">{coverage.author_profile.industry_readiness}</div>
            </div>
          </div>
          <div className="mt-4">
            <div className="text-sm text-gray-500 mb-2">Philosophical Position</div>
            <p className="text-sm text-gray-300">{coverage.author_profile.philosophical_position}</p>
          </div>
        </div>
      )}
    </div>
  )
}
