import type { ProfessionalCoverage, CharacterAnalysis } from '../../types/coverage-professional'

export default function CharacterArcsView({ coverage }: { coverage: ProfessionalCoverage }) {
  return (
    <div className="p-8 space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-100 mb-2">Character Analysis</h1>
        <p className="text-gray-400">{coverage.character_analyses.length} characters analyzed</p>
      </div>

      {coverage.character_analyses.map((character) => (
        <CharacterCard key={character.name} character={character} />
      ))}
    </div>
  )
}

function CharacterCard({ character }: { character: CharacterAnalysis }) {
  const getRoleColor = (role: string) => {
    if (role === 'lead') return 'text-sunset-400 bg-sunset-500/10 border-sunset-500/30'
    if (role === 'supporting') return 'text-electric-400 bg-electric-500/10 border-electric-500/30'
    return 'text-gray-400 bg-gray-500/10 border-gray-500/30'
  }

  return (
    <div className="bg-gray-800/30 border border-gray-700 rounded-xl p-6 space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-100 mb-2">{character.name}</h2>
          <p className="text-gray-400">{character.arc_description}</p>
        </div>
        <div className="flex items-center gap-3">
          <span className={`px-3 py-1 rounded-full text-xs font-bold border capitalize ${getRoleColor(character.role)}`}>
            {character.role}
          </span>
          <span className="px-3 py-1 bg-electric-500/10 border border-electric-500/30 text-electric-400 text-xs rounded-full">
            {character.screen_time_percentage}% Screen Time
          </span>
        </div>
      </div>

      {/* Complexity Notes */}
      <div className="bg-gray-800/50 border border-gray-700 rounded-lg p-4">
        <h3 className="text-sm font-bold text-sunset-400 mb-2">Character Complexity</h3>
        <p className="text-sm text-gray-300">{character.complexity_notes}</p>
      </div>

      {/* Voice Evolution Timeline */}
      <div>
        <h3 className="text-lg font-bold text-electric-400 mb-4">Voice Evolution</h3>
        <div className="space-y-4">
          {/* Act One */}
          <div className="relative pl-8 border-l-2 border-electric-500/30">
            <div className="absolute -left-2 top-0 w-4 h-4 bg-electric-500 rounded-full"></div>
            <div className="mb-1">
              <span className="text-sm font-bold text-electric-400">Act One</span>
            </div>
            <ul className="space-y-1">
              {character.voice_evolution.act_one.map((line, i) => (
                <li key={i} className="text-sm text-gray-300 italic">"{line}"</li>
              ))}
            </ul>
          </div>

          {/* Act Two */}
          <div className="relative pl-8 border-l-2 border-sakura-500/30">
            <div className="absolute -left-2 top-0 w-4 h-4 bg-sakura-500 rounded-full"></div>
            <div className="mb-1">
              <span className="text-sm font-bold text-sakura-400">Act Two</span>
            </div>
            <ul className="space-y-1">
              {character.voice_evolution.act_two.map((line, i) => (
                <li key={i} className="text-sm text-gray-300 italic">"{line}"</li>
              ))}
            </ul>
          </div>

          {/* Act Three */}
          <div className="relative pl-8 border-l-2 border-mint-500/30">
            <div className="absolute -left-2 top-0 w-4 h-4 bg-mint-500 rounded-full"></div>
            <div className="mb-1">
              <span className="text-sm font-bold text-mint-400">Act Three</span>
            </div>
            <ul className="space-y-1">
              {character.voice_evolution.act_three.map((line, i) => (
                <li key={i} className="text-sm text-gray-300 italic">"{line}"</li>
              ))}
            </ul>
          </div>
        </div>
      </div>

      {/* Performance Demands */}
      <div>
        <h3 className="text-sm font-bold text-sakura-400 mb-3">Performance Demands</h3>
        <ul className="space-y-2">
          {character.performance_demands.map((demand, i) => (
            <li key={i} className="flex items-start gap-2 text-sm text-gray-300">
              <span className="text-sakura-400 mt-0.5">★</span>
              <span>{demand}</span>
            </li>
          ))}
        </ul>
      </div>

      {/* Casting Recommendations */}
      <div>
        <h3 className="text-sm font-bold text-mint-400 mb-3">Casting Recommendations</h3>
        <div className="flex flex-wrap gap-2">
          {character.casting_recommendations.map((actor) => (
            <span
              key={actor}
              className="px-3 py-1.5 bg-mint-500/10 border border-mint-500/30 text-mint-400 rounded-lg text-sm font-medium"
            >
              {actor}
            </span>
          ))}
        </div>
      </div>

      {/* Development Opportunities */}
      {character.development_opportunities && character.development_opportunities.length > 0 && (
        <div className="bg-sunset-500/5 border border-sunset-500/30 rounded-lg p-4">
          <h3 className="text-sm font-bold text-sunset-400 mb-2">Development Opportunities</h3>
          <ul className="space-y-1">
            {character.development_opportunities.map((opp, i) => (
              <li key={i} className="flex items-start gap-2 text-sm text-gray-300">
                <span className="text-sunset-400 mt-0.5">→</span>
                <span>{opp}</span>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  )
}
