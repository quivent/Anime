import { useState } from 'react'
import { invoke } from '@tauri-apps/api/core'
import { open } from '@tauri-apps/plugin-dialog'
import type {
  AnalysisType,
  AnalysisResult,
  CharacterAnalysis,
  PlotAnalysis,
  DialogueAnalysis,
  PacingAnalysis,
  ThemeAnalysis,
} from '../types/creative'

export default function AnalysisView() {
  const [uploadedContent, setUploadedContent] = useState('')
  const [fileName, setFileName] = useState('')
  const [selectedType, setSelectedType] = useState<AnalysisType>('character')
  const [isAnalyzing, setIsAnalyzing] = useState(false)
  const [analysisResult, setAnalysisResult] = useState<AnalysisResult | null>(null)
  const [error, setError] = useState<string | null>(null)

  const handleFileUpload = async () => {
    try {
      const selected = await open({
        multiple: false,
        filters: [{
          name: 'Text Files',
          extensions: ['txt', 'md', 'pdf', 'docx']
        }]
      })

      if (selected && typeof selected === 'string') {
        const content = await invoke<string>('read_file', { path: selected })
        setUploadedContent(content)
        setFileName(selected.split('/').pop() || 'Uploaded File')
        setError(null)
      }
    } catch (err) {
      setError('Failed to upload file: ' + (err instanceof Error ? err.message : String(err)))
    }
  }

  const runAnalysis = async () => {
    if (!uploadedContent.trim()) {
      setError('Please upload a file first')
      return
    }

    setIsAnalyzing(true)
    setError(null)

    try {
      const result = await invoke<AnalysisResult>('analyze_content', {
        type: selectedType,
        content: uploadedContent,
      })

      setAnalysisResult(result)
    } catch (err) {
      setError('Analysis failed: ' + (err instanceof Error ? err.message : String(err)))
    } finally {
      setIsAnalyzing(false)
    }
  }

  const exportReport = async () => {
    if (!analysisResult) return

    try {
      await invoke('export_analysis', {
        result: analysisResult,
        fileName: fileName,
      })
    } catch (err) {
      setError('Export failed: ' + (err instanceof Error ? err.message : String(err)))
    }
  }

  const getTypeIcon = (type: AnalysisType) => {
    switch (type) {
      case 'character': return '👥'
      case 'plot': return '📖'
      case 'dialogue': return '💬'
      case 'pacing': return '⏱️'
      case 'theme': return '🎭'
    }
  }

  const getTypeLabel = (type: AnalysisType) => {
    switch (type) {
      case 'character': return 'Character Analysis'
      case 'plot': return 'Plot Structure'
      case 'dialogue': return 'Dialogue Assessment'
      case 'pacing': return 'Pacing Analysis'
      case 'theme': return 'Theme Extraction'
    }
  }

  return (
    <div className="h-full flex flex-col p-6 overflow-hidden">
      {/* Mock Feature Warning Banner */}
      <div className="mb-4 p-3 bg-sunset-500/20 border border-sunset-500/50 rounded-lg">
        <div className="flex items-center gap-2 text-sm">
          <span className="text-sunset-400">⚠️</span>
          <span className="text-sunset-300 font-medium">Development Mode:</span>
          <span className="text-gray-300">Analysis features use mock/hardcoded data - not real NLP or AI analysis</span>
        </div>
      </div>

      {/* Header */}
      <div className="mb-6">
        <h2 className="text-3xl font-bold mb-2 text-electric-400">
          Content Analysis
        </h2>
        <p className="text-gray-400">
          Analyze scripts, stories, and creative content with AI-powered insights
        </p>
      </div>

      {/* Upload Section */}
      <div className="mb-6">
        <div className="flex gap-3">
          <button
            onClick={handleFileUpload}
            className="px-6 py-3 bg-electric-500/20 hover:bg-electric-500/30 border border-electric-500/50 rounded-lg text-electric-400 font-medium transition-all"
          >
            📁 Upload File
          </button>
          {fileName && (
            <div className="flex-1 flex items-center px-4 py-3 bg-gray-800/50 border border-gray-700 rounded-lg">
              <span className="text-mint-400 mr-2">✓</span>
              <span className="text-gray-300">{fileName}</span>
              <span className="ml-auto text-gray-500 text-sm">
                {uploadedContent.split(/\s+/).filter(Boolean).length} words
              </span>
            </div>
          )}
        </div>
      </div>

      {/* Analysis Type Selector */}
      <div className="mb-6">
        <h3 className="text-sm font-semibold text-gray-400 mb-3">Analysis Type</h3>
        <div className="grid grid-cols-5 gap-2">
          {(['character', 'plot', 'dialogue', 'pacing', 'theme'] as AnalysisType[]).map(type => (
            <button
              key={type}
              onClick={() => setSelectedType(type)}
              className={`p-4 rounded-lg text-center transition-all ${
                selectedType === type
                  ? 'bg-electric-500/20 border border-electric-500/50 text-electric-400 anime-glow'
                  : 'bg-gray-800/50 border border-gray-700 text-gray-400 hover:border-gray-600'
              }`}
            >
              <div className="text-3xl mb-2">{getTypeIcon(type)}</div>
              <div className="text-sm font-medium">{getTypeLabel(type)}</div>
            </button>
          ))}
        </div>
      </div>

      {/* Action Buttons */}
      <div className="flex gap-3 mb-6">
        <button
          onClick={runAnalysis}
          disabled={!uploadedContent || isAnalyzing}
          className="px-8 py-3 bg-mint-500/20 hover:bg-mint-500/30 border border-mint-500/50 rounded-lg text-mint-400 font-medium transition-all disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isAnalyzing ? (
            <span className="flex items-center gap-2">
              <span className="animate-spin">⚙️</span>
              Analyzing...
            </span>
          ) : (
            '🔍 Run Analysis'
          )}
        </button>
        {analysisResult && (
          <button
            onClick={exportReport}
            className="px-6 py-3 bg-gray-800/50 hover:bg-gray-700/50 border border-gray-700 rounded-lg text-gray-300 transition-all"
          >
            📄 Export Report
          </button>
        )}
      </div>

      {error && (
        <div className="mb-6 p-4 bg-sunset-500/10 border border-sunset-500/30 rounded-lg text-sunset-400">
          {error}
        </div>
      )}

      {/* Results Section */}
      <div className="flex-1 overflow-y-auto">
        {analysisResult ? (
          <div className="space-y-6">
            {analysisResult.type === 'character' && (
              <CharacterAnalysisResults data={analysisResult.data} />
            )}
            {analysisResult.type === 'plot' && (
              <PlotAnalysisResults data={analysisResult.data} />
            )}
            {analysisResult.type === 'dialogue' && (
              <DialogueAnalysisResults data={analysisResult.data} />
            )}
            {analysisResult.type === 'pacing' && (
              <PacingAnalysisResults data={analysisResult.data} />
            )}
            {analysisResult.type === 'theme' && (
              <ThemeAnalysisResults data={analysisResult.data} />
            )}
          </div>
        ) : (
          <div className="flex items-center justify-center h-full">
            <div className="text-center">
              <div className="text-6xl mb-4">📊</div>
              <h3 className="text-2xl font-bold text-gray-300 mb-2">No Analysis Yet</h3>
              <p className="text-gray-500">Upload a file and run an analysis to see results</p>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

function CharacterAnalysisResults({ data }: { data: CharacterAnalysis }) {
  return (
    <div className="space-y-4">
      <h3 className="text-xl font-bold text-electric-400 flex items-center gap-2">
        <span>👥</span>
        Character Analysis
      </h3>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {data.characters.map((char, idx) => (
          <div key={idx} className="p-4 bg-gray-800/50 border border-gray-700 rounded-lg">
            <div className="flex items-start justify-between mb-3">
              <div>
                <h4 className="text-lg font-bold text-gray-200">{char.name}</h4>
                <div className="text-sm text-gray-400">{char.role}</div>
              </div>
              <div className="px-3 py-1 bg-mint-500/20 border border-mint-500/30 rounded-full text-mint-400 text-sm">
                Score: {char.importance_score}
              </div>
            </div>
            <div className="space-y-2">
              <div>
                <div className="text-xs text-gray-500 mb-1">Traits</div>
                <div className="flex flex-wrap gap-1">
                  {char.traits.map((trait, i) => (
                    <span key={i} className="px-2 py-1 bg-electric-500/10 text-electric-400 rounded text-xs">
                      {trait}
                    </span>
                  ))}
                </div>
              </div>
              <div>
                <div className="text-xs text-gray-500 mb-1">Character Arc</div>
                <div className="text-sm text-gray-300">{char.arc}</div>
              </div>
              <div className="text-xs text-gray-500">
                Dialogue Lines: {char.dialogue_count}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

function PlotAnalysisResults({ data }: { data: PlotAnalysis }) {
  return (
    <div className="space-y-6">
      <h3 className="text-xl font-bold text-electric-400 flex items-center gap-2">
        <span>📖</span>
        Plot Structure Analysis
      </h3>

      <div className="p-4 bg-gray-800/50 border border-gray-700 rounded-lg">
        <h4 className="font-bold text-gray-200 mb-3">Act Breakdown</h4>
        <div className="space-y-2">
          {data.structure.act_breakdown.map((act, idx) => (
            <div key={idx} className="p-3 bg-gray-900/50 rounded-lg">
              <div className="text-sm text-gray-300">{act}</div>
            </div>
          ))}
        </div>
      </div>

      <div className="p-4 bg-gray-800/50 border border-gray-700 rounded-lg">
        <h4 className="font-bold text-gray-200 mb-3">Turning Points</h4>
        <ul className="space-y-2">
          {data.structure.turning_points.map((point, idx) => (
            <li key={idx} className="flex items-start gap-2 text-sm text-gray-300">
              <span className="text-mint-400">•</span>
              {point}
            </li>
          ))}
        </ul>
      </div>

      <div className="p-4 bg-gray-800/50 border border-gray-700 rounded-lg">
        <h4 className="font-bold text-gray-200 mb-3">Pacing Overview</h4>
        <div className="flex items-center gap-4">
          <div className="flex-1">
            <div className="text-xs text-gray-500 mb-2">Overall Score</div>
            <div className="h-3 bg-gray-900 rounded-full overflow-hidden">
              <div
                className="h-full bg-mint-500"
                style={{ width: `${data.pacing.overall_score}%` }}
              />
            </div>
          </div>
          <div className="text-2xl font-bold text-mint-400">
            {data.pacing.overall_score}%
          </div>
        </div>
      </div>
    </div>
  )
}

function DialogueAnalysisResults({ data }: { data: DialogueAnalysis }) {
  return (
    <div className="space-y-4">
      <h3 className="text-xl font-bold text-electric-400 flex items-center gap-2">
        <span>💬</span>
        Dialogue Analysis
      </h3>

      <div className="grid grid-cols-4 gap-4">
        <div className="p-4 bg-gray-800/50 border border-gray-700 rounded-lg text-center">
          <div className="text-3xl font-bold text-mint-400">{data.total_lines}</div>
          <div className="text-sm text-gray-500 mt-1">Total Lines</div>
        </div>
        <div className="p-4 bg-gray-800/50 border border-gray-700 rounded-lg text-center">
          <div className="text-3xl font-bold text-electric-400">{data.avg_length}</div>
          <div className="text-sm text-gray-500 mt-1">Avg Length</div>
        </div>
        <div className="p-4 bg-gray-800/50 border border-gray-700 rounded-lg text-center">
          <div className="text-3xl font-bold text-neon-400">{data.unique_voices}</div>
          <div className="text-sm text-gray-500 mt-1">Unique Voices</div>
        </div>
        <div className="p-4 bg-gray-800/50 border border-gray-700 rounded-lg text-center">
          <div className="text-3xl font-bold text-sakura-400">{data.readability_score}</div>
          <div className="text-sm text-gray-500 mt-1">Readability</div>
        </div>
      </div>

      <div className="p-4 bg-gray-800/50 border border-gray-700 rounded-lg">
        <h4 className="font-bold text-gray-200 mb-3">Suggestions</h4>
        <ul className="space-y-2">
          {data.suggestions.map((suggestion, idx) => (
            <li key={idx} className="flex items-start gap-2 text-sm text-gray-300">
              <span className="text-mint-400">✓</span>
              {suggestion}
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}

function PacingAnalysisResults({ data }: { data: PacingAnalysis }) {
  const maxTension = Math.max(...data.tension_curve)

  return (
    <div className="space-y-4">
      <h3 className="text-xl font-bold text-electric-400 flex items-center gap-2">
        <span>⏱️</span>
        Pacing Analysis
      </h3>

      <div className="p-4 bg-gray-800/50 border border-gray-700 rounded-lg">
        <h4 className="font-bold text-gray-200 mb-3">Tension Curve</h4>
        <div className="h-32 flex items-end gap-1">
          {data.tension_curve.map((tension, idx) => (
            <div
              key={idx}
              className="flex-1 bg-electric-500 rounded-t"
              style={{ height: `${(tension / maxTension) * 100}%` }}
              title={`Scene ${idx + 1}: ${tension}`}
            />
          ))}
        </div>
      </div>

      <div className="p-4 bg-gray-800/50 border border-gray-700 rounded-lg">
        <h4 className="font-bold text-gray-200 mb-3">Recommendations</h4>
        <ul className="space-y-2">
          {data.recommendations.map((rec, idx) => (
            <li key={idx} className="flex items-start gap-2 text-sm text-gray-300">
              <span className="text-mint-400">→</span>
              {rec}
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}

function ThemeAnalysisResults({ data }: { data: ThemeAnalysis }) {
  return (
    <div className="space-y-4">
      <h3 className="text-xl font-bold text-electric-400 flex items-center gap-2">
        <span>🎭</span>
        Theme Analysis
      </h3>

      <div className="p-4 bg-gray-800/50 border border-gray-700 rounded-lg">
        <h4 className="font-bold text-gray-200 mb-3">Primary Themes</h4>
        <div className="flex flex-wrap gap-2">
          {data.primary_themes.map((theme, idx) => (
            <span key={idx} className="px-4 py-2 bg-electric-500/20 border border-electric-500/50 text-electric-400 rounded-lg">
              {theme}
            </span>
          ))}
        </div>
      </div>

      <div className="p-4 bg-gray-800/50 border border-gray-700 rounded-lg">
        <h4 className="font-bold text-gray-200 mb-3">Recurring Motifs</h4>
        <div className="flex flex-wrap gap-2">
          {data.recurring_motifs.map((motif, idx) => (
            <span key={idx} className="px-3 py-1 bg-mint-500/10 text-mint-400 rounded text-sm">
              {motif}
            </span>
          ))}
        </div>
      </div>

      <div className="p-4 bg-gray-800/50 border border-gray-700 rounded-lg">
        <h4 className="font-bold text-gray-200 mb-3">Symbolism</h4>
        <div className="space-y-3">
          {data.symbolism.map((symbol, idx) => (
            <div key={idx} className="p-3 bg-gray-900/50 rounded-lg">
              <div className="flex items-center justify-between mb-2">
                <span className="font-medium text-neon-400">{symbol.symbol}</span>
                <span className="text-xs text-gray-500">{symbol.occurrences} occurrences</span>
              </div>
              <div className="text-sm text-gray-400">{symbol.meaning}</div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
