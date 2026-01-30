import { useState } from 'react'
import { useCoverageStore } from '../store/coverageStore'
import type { CoverageReport, CoverageRatings } from '../types/coverage'

interface CoverageEditorProps {
  coverage?: CoverageReport
  onClose: () => void
}

export default function CoverageEditor({ coverage, onClose }: CoverageEditorProps) {
  const { addCoverage, updateCoverage } = useCoverageStore()
  const isEditing = !!coverage

  const [formData, setFormData] = useState({
    script_title: coverage?.script_title || '',
    submitted_by: coverage?.submitted_by || '',
    logline: coverage?.logline || '',
    synopsis: coverage?.synopsis || '',
    genre: coverage?.genre || [],
    ratings: coverage?.ratings || {
      overall: 5,
      premise: 5,
      character: 5,
      dialogue: 5,
      structure: 5,
      pacing: 5,
      marketability: 5,
      originality: 5,
      execution: 5,
    },
    strengths: coverage?.analysis.strengths || [],
    weaknesses: coverage?.analysis.weaknesses || [],
    recommendation_type: coverage?.recommendation.type || 'consider' as const,
    recommendation_summary: coverage?.recommendation.summary || '',
  })

  const [strengthInput, setStrengthInput] = useState('')
  const [weaknessInput, setWeaknessInput] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    const coverageData = {
      title: `Coverage: ${formData.script_title}`,
      script_title: formData.script_title,
      submitted_by: formData.submitted_by,
      submitted_date: new Date().toISOString(),
      template_id: 'default',
      template_version: '1.0',
      logline: formData.logline,
      synopsis: formData.synopsis,
      ratings: formData.ratings,
      analysis: {
        strengths: formData.strengths,
        weaknesses: formData.weaknesses,
        premise_analysis: '',
        character_analysis: '',
        dialogue_analysis: '',
        structure_analysis: '',
        pacing_analysis: '',
        theme_analysis: '',
        page_notes: [],
        character_breakdowns: [],
      },
      recommendation: {
        type: formData.recommendation_type,
        summary: formData.recommendation_summary,
        commercial_appeal: formData.ratings.marketability,
        target_audience: '',
        comparable_titles: [],
        notes: '',
      },
      genre: formData.genre,
      status: 'completed' as const,
      tags: [],
    }

    if (isEditing) {
      updateCoverage(coverage.id, coverageData)
    } else {
      addCoverage(coverageData)
    }

    onClose()
  }

  const handleRatingChange = (key: keyof CoverageRatings, value: number) => {
    setFormData(prev => ({
      ...prev,
      ratings: { ...prev.ratings, [key]: value }
    }))
  }

  const addStrength = () => {
    if (strengthInput.trim()) {
      setFormData(prev => ({
        ...prev,
        strengths: [...prev.strengths, strengthInput.trim()]
      }))
      setStrengthInput('')
    }
  }

  const addWeakness = () => {
    if (weaknessInput.trim()) {
      setFormData(prev => ({
        ...prev,
        weaknesses: [...prev.weaknesses, weaknessInput.trim()]
      }))
      setWeaknessInput('')
    }
  }

  const removeStrength = (index: number) => {
    setFormData(prev => ({
      ...prev,
      strengths: prev.strengths.filter((_, i) => i !== index)
    }))
  }

  const removeWeakness = (index: number) => {
    setFormData(prev => ({
      ...prev,
      weaknesses: prev.weaknesses.filter((_, i) => i !== index)
    }))
  }

  return (
    <div className="flex flex-col h-full overflow-hidden bg-gray-900">
      {/* Header */}
      <div className="p-6 border-b border-gray-800">
        <div className="flex items-center justify-between">
          <h2 className="text-2xl font-bold text-sunset-400">
            {isEditing ? 'Edit Coverage' : 'New Coverage Report'}
          </h2>
          <button
            onClick={onClose}
            className="px-4 py-2 bg-gray-700/50 hover:bg-gray-700 border border-gray-600 rounded-lg text-gray-300 transition-all"
          >
            ← Back
          </button>
        </div>
      </div>

      {/* Form */}
      <form onSubmit={handleSubmit} className="flex-1 overflow-auto p-6 space-y-6">
        {/* Basic Info */}
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-2">
              Script Title <span className="text-sunset-400">*</span>
            </label>
            <input
              type="text"
              value={formData.script_title}
              onChange={(e) => setFormData({ ...formData, script_title: e.target.value })}
              className="w-full px-4 py-3 bg-gray-800/50 border border-gray-700 rounded-lg focus:outline-none focus:border-sunset-500 text-white"
              required
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-2">
              Submitted By <span className="text-sunset-400">*</span>
            </label>
            <input
              type="text"
              value={formData.submitted_by}
              onChange={(e) => setFormData({ ...formData, submitted_by: e.target.value })}
              className="w-full px-4 py-3 bg-gray-800/50 border border-gray-700 rounded-lg focus:outline-none focus:border-sunset-500 text-white"
              required
            />
          </div>
        </div>

        {/* Logline */}
        <div>
          <label className="block text-sm font-medium text-gray-300 mb-2">
            Logline <span className="text-sunset-400">*</span>
          </label>
          <input
            type="text"
            value={formData.logline}
            onChange={(e) => setFormData({ ...formData, logline: e.target.value })}
            placeholder="One sentence summary of the script..."
            className="w-full px-4 py-3 bg-gray-800/50 border border-gray-700 rounded-lg focus:outline-none focus:border-sunset-500 text-white"
            required
          />
        </div>

        {/* Synopsis */}
        <div>
          <label className="block text-sm font-medium text-gray-300 mb-2">
            Synopsis
          </label>
          <textarea
            value={formData.synopsis}
            onChange={(e) => setFormData({ ...formData, synopsis: e.target.value })}
            placeholder="Detailed story summary..."
            rows={6}
            className="w-full px-4 py-3 bg-gray-800/50 border border-gray-700 rounded-lg focus:outline-none focus:border-sunset-500 text-white resize-none"
          />
        </div>

        {/* Ratings */}
        <div className="bg-gray-800/30 border border-gray-700 rounded-xl p-6">
          <h3 className="text-lg font-semibold text-gray-200 mb-4">Ratings (1-10)</h3>
          <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
            {Object.entries(formData.ratings).map(([key, value]) => (
              <div key={key}>
                <div className="flex items-center justify-between mb-2">
                  <label className="text-sm text-gray-400 capitalize">{key}</label>
                  <span className="text-sm font-bold text-sunset-400">{value}</span>
                </div>
                <input
                  type="range"
                  min="1"
                  max="10"
                  value={value}
                  onChange={(e) => handleRatingChange(key as keyof CoverageRatings, parseInt(e.target.value))}
                  className="w-full"
                />
              </div>
            ))}
          </div>
        </div>

        {/* Strengths & Weaknesses */}
        <div className="grid grid-cols-2 gap-6">
          {/* Strengths */}
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-2">Strengths</label>
            <div className="flex gap-2 mb-3">
              <input
                type="text"
                value={strengthInput}
                onChange={(e) => setStrengthInput(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && (e.preventDefault(), addStrength())}
                placeholder="Add a strength..."
                className="flex-1 px-4 py-2 bg-gray-800/50 border border-gray-700 rounded-lg focus:outline-none focus:border-mint-500 text-white text-sm"
              />
              <button
                type="button"
                onClick={addStrength}
                className="px-4 py-2 bg-mint-500/20 border border-mint-500/50 rounded-lg text-mint-400 hover:bg-mint-500/30 transition-all"
              >
                +
              </button>
            </div>
            <div className="space-y-2">
              {formData.strengths.map((strength, i) => (
                <div key={i} className="flex items-center gap-2 p-2 bg-mint-500/10 border border-mint-500/30 rounded-lg">
                  <span className="flex-1 text-sm text-gray-300">{strength}</span>
                  <button
                    type="button"
                    onClick={() => removeStrength(i)}
                    className="text-gray-400 hover:text-sunset-400"
                  >
                    ×
                  </button>
                </div>
              ))}
            </div>
          </div>

          {/* Weaknesses */}
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-2">Weaknesses</label>
            <div className="flex gap-2 mb-3">
              <input
                type="text"
                value={weaknessInput}
                onChange={(e) => setWeaknessInput(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && (e.preventDefault(), addWeakness())}
                placeholder="Add a weakness..."
                className="flex-1 px-4 py-2 bg-gray-800/50 border border-gray-700 rounded-lg focus:outline-none focus:border-sunset-500 text-white text-sm"
              />
              <button
                type="button"
                onClick={addWeakness}
                className="px-4 py-2 bg-sunset-500/20 border border-sunset-500/50 rounded-lg text-sunset-400 hover:bg-sunset-500/30 transition-all"
              >
                +
              </button>
            </div>
            <div className="space-y-2">
              {formData.weaknesses.map((weakness, i) => (
                <div key={i} className="flex items-center gap-2 p-2 bg-sunset-500/10 border border-sunset-500/30 rounded-lg">
                  <span className="flex-1 text-sm text-gray-300">{weakness}</span>
                  <button
                    type="button"
                    onClick={() => removeWeakness(i)}
                    className="text-gray-400 hover:text-sunset-400"
                  >
                    ×
                  </button>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Recommendation */}
        <div className="bg-gray-800/30 border border-gray-700 rounded-xl p-6">
          <h3 className="text-lg font-semibold text-gray-200 mb-4">Recommendation</h3>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">Type</label>
              <select
                value={formData.recommendation_type}
                onChange={(e) => setFormData({ ...formData, recommendation_type: e.target.value as any })}
                className="w-full px-4 py-3 bg-gray-800/50 border border-gray-700 rounded-lg focus:outline-none focus:border-sunset-500 text-white"
              >
                <option value="pass">Pass</option>
                <option value="consider">Consider</option>
                <option value="recommend">Recommend</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">Summary</label>
              <textarea
                value={formData.recommendation_summary}
                onChange={(e) => setFormData({ ...formData, recommendation_summary: e.target.value })}
                placeholder="Overall recommendation and notes..."
                rows={4}
                className="w-full px-4 py-3 bg-gray-800/50 border border-gray-700 rounded-lg focus:outline-none focus:border-sunset-500 text-white resize-none"
              />
            </div>
          </div>
        </div>
      </form>

      {/* Footer */}
      <div className="p-6 border-t border-gray-800 flex gap-3">
        <button
          type="button"
          onClick={onClose}
          className="flex-1 px-6 py-3 bg-gray-800/50 hover:bg-gray-700/50 border border-gray-700 rounded-lg text-gray-300 transition-all"
        >
          Cancel
        </button>
        <button
          type="submit"
          onClick={handleSubmit}
          className="flex-1 px-6 py-3 bg-sunset-500/20 hover:bg-sunset-500/30 border border-sunset-500/50 rounded-lg text-sunset-400 font-medium transition-all"
        >
          {isEditing ? 'Update Coverage' : 'Create Coverage'}
        </button>
      </div>
    </div>
  )
}
