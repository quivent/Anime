import { useState, useEffect, useRef } from 'react'
import { invoke } from '@tauri-apps/api/core'
import type { Document, WritingMode, WritingResponse } from '../types/creative'

export default function WritingView() {
  const [documents, setDocuments] = useState<Document[]>([])
  const [currentDoc, setCurrentDoc] = useState<Document | null>(null)
  const [content, setContent] = useState('')
  const [isGenerating, setIsGenerating] = useState(false)
  const [selectedMode, setSelectedMode] = useState<WritingMode>('continuation')
  const [showNewDocDialog, setShowNewDocDialog] = useState(false)
  const [newDocTitle, setNewDocTitle] = useState('')
  const [prompt, setPrompt] = useState('')
  const [generatedText, setGeneratedText] = useState('')
  const [showHistory, setShowHistory] = useState(false)
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  useEffect(() => {
    loadDocuments()
  }, [])

  useEffect(() => {
    if (currentDoc) {
      setContent(currentDoc.content)
    }
  }, [currentDoc])

  const loadDocuments = async () => {
    try {
      const docs = await invoke<Document[]>('list_documents')
      setDocuments(docs)
    } catch (error) {

      // Start with empty list if no documents exist
      setDocuments([])
    }
  }

  const createDocument = async () => {
    if (!newDocTitle.trim()) return

    try {
      const doc = await invoke<Document>('create_document', {
        title: newDocTitle,
      })
      setDocuments([...documents, doc])
      setCurrentDoc(doc)
      setNewDocTitle('')
      setShowNewDocDialog(false)
    } catch (error) {

    }
  }

  const saveDocument = async () => {
    if (!currentDoc) return

    try {
      const updated = await invoke<Document>('save_document', {
        id: currentDoc.id,
        content,
      })
      setDocuments(documents.map(d => d.id === updated.id ? updated : d))
      setCurrentDoc(updated)
    } catch (error) {

    }
  }

  const deleteDocument = async (id: string) => {
    try {
      await invoke('delete_document', { id })
      setDocuments(documents.filter(d => d.id !== id))
      if (currentDoc?.id === id) {
        setCurrentDoc(null)
        setContent('')
      }
    } catch (error) {

    }
  }

  const generateText = async () => {
    if (!prompt.trim() && selectedMode !== 'continuation') return

    setIsGenerating(true)
    setGeneratedText('')

    try {
      const response = await invoke<WritingResponse>('generate_text', {
        mode: selectedMode,
        context: content,
        prompt: prompt || '',
        maxTokens: 500,
      })

      setGeneratedText(response.generated_text)
    } catch (error) {

      setGeneratedText('Failed to generate text. Please try again.')
    } finally {
      setIsGenerating(false)
    }
  }

  const insertGenerated = () => {
    if (!generatedText) return

    const textarea = textareaRef.current
    if (textarea) {
      const start = textarea.selectionStart
      const end = textarea.selectionEnd
      const newContent = content.substring(0, start) + '\n' + generatedText + '\n' + content.substring(end)
      setContent(newContent)
      setGeneratedText('')
      setPrompt('')
    }
  }

  const getModeIcon = (mode: WritingMode) => {
    switch (mode) {
      case 'continuation': return '📝'
      case 'dialogue': return '💬'
      case 'scene': return '🎬'
      case 'outline': return '📋'
    }
  }

  const getModeLabel = (mode: WritingMode) => {
    switch (mode) {
      case 'continuation': return 'Story Continuation'
      case 'dialogue': return 'Character Dialogue'
      case 'scene': return 'Scene Description'
      case 'outline': return 'Plot Outline'
    }
  }

  return (
    <div className="h-full flex overflow-hidden">
      {/* Mock Feature Warning Banner */}
      <div className="absolute top-0 left-0 right-0 bg-sunset-500/20 border-b border-sunset-500/50 px-4 py-2 z-10">
        <div className="flex items-center gap-2 text-sm">
          <span className="text-sunset-400">⚠️</span>
          <span className="text-sunset-300 font-medium">Development Mode:</span>
          <span className="text-gray-300">AI text generation uses mock/placeholder data - not real AI</span>
        </div>
      </div>

      {/* Sidebar - Document List */}
      <div className="w-64 border-r border-gray-800 bg-gray-900/50 flex flex-col mt-10">
        <div className="p-4 border-b border-gray-800">
          <button
            onClick={() => setShowNewDocDialog(true)}
            className="w-full px-4 py-2 bg-electric-500/20 hover:bg-electric-500/30 border border-electric-500/50 rounded-lg text-electric-400 font-medium transition-all"
          >
            + New Document
          </button>
        </div>

        <div className="flex-1 overflow-y-auto p-4 space-y-2">
          {documents.map(doc => (
            <div
              key={doc.id}
              className={`p-3 rounded-lg cursor-pointer transition-all ${
                currentDoc?.id === doc.id
                  ? 'bg-electric-500/20 border border-electric-500/50'
                  : 'bg-gray-800/50 hover:bg-gray-800 border border-transparent'
              }`}
              onClick={() => setCurrentDoc(doc)}
            >
              <div className="font-medium text-gray-200 truncate">{doc.title}</div>
              <div className="text-xs text-gray-500 mt-1">
                {doc.word_count} words
              </div>
              <div className="text-xs text-gray-600 mt-1">
                {new Date(doc.updated_at).toLocaleDateString()}
              </div>
              {currentDoc?.id === doc.id && (
                <button
                  onClick={(e) => {
                    e.stopPropagation()
                    deleteDocument(doc.id)
                  }}
                  className="mt-2 text-xs text-sunset-400 hover:text-sunset-300"
                >
                  Delete
                </button>
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Main Editor */}
      <div className="flex-1 flex flex-col overflow-hidden mt-10">
        {currentDoc ? (
          <>
            {/* Toolbar */}
            <div className="border-b border-gray-800 bg-gray-900/80 backdrop-blur-md p-4">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-2xl font-bold text-electric-400">{currentDoc.title}</h2>
                <div className="flex gap-2">
                  <button
                    onClick={() => setShowHistory(!showHistory)}
                    className="px-4 py-2 bg-gray-800/50 hover:bg-gray-700/50 border border-gray-700 rounded-lg text-gray-300 transition-all text-sm"
                  >
                    📜 History
                  </button>
                  <button
                    onClick={saveDocument}
                    className="px-4 py-2 bg-mint-500/20 hover:bg-mint-500/30 border border-mint-500/50 rounded-lg text-mint-400 font-medium transition-all"
                  >
                    💾 Save
                  </button>
                  <button
                    onClick={async () => {
                      try {
                        await invoke('export_document', { id: currentDoc.id })
                      } catch (error) {

                      }
                    }}
                    className="px-4 py-2 bg-gray-800/50 hover:bg-gray-700/50 border border-gray-700 rounded-lg text-gray-300 transition-all"
                  >
                    📤 Export
                  </button>
                </div>
              </div>

              {/* Writing Mode Selector */}
              <div className="flex gap-2">
                {(['continuation', 'dialogue', 'scene', 'outline'] as WritingMode[]).map(mode => (
                  <button
                    key={mode}
                    onClick={() => setSelectedMode(mode)}
                    className={`px-3 py-2 rounded-lg text-sm font-medium transition-all ${
                      selectedMode === mode
                        ? 'bg-electric-500/20 border border-electric-500/50 text-electric-400'
                        : 'bg-gray-800/50 border border-gray-700 text-gray-400 hover:border-gray-600'
                    }`}
                  >
                    {getModeIcon(mode)} {getModeLabel(mode)}
                  </button>
                ))}
              </div>
            </div>

            {/* Editor */}
            <div className="flex-1 overflow-hidden flex">
              <div className="flex-1 p-6 overflow-y-auto">
                <textarea
                  ref={textareaRef}
                  value={content}
                  onChange={(e) => setContent(e.target.value)}
                  className="w-full h-full min-h-[600px] bg-transparent text-gray-200 resize-none focus:outline-none text-lg leading-relaxed font-mono"
                  placeholder="Start writing your story..."
                />
              </div>

              {/* AI Assistant Panel */}
              <div className="w-96 border-l border-gray-800 bg-gray-900/50 p-4 overflow-y-auto">
                <h3 className="text-lg font-bold text-electric-400 mb-4">AI Assistant</h3>

                {/* Mock Warning */}
                <div className="mb-4 p-3 bg-sunset-500/10 border border-sunset-500/30 rounded-lg">
                  <div className="flex items-start gap-2">
                    <span className="text-sunset-400 text-sm">⚠️</span>
                    <div className="text-xs text-sunset-300">
                      <div className="font-semibold mb-1">Mock Feature</div>
                      <div className="text-gray-400">This generates placeholder text, not real AI content</div>
                    </div>
                  </div>
                </div>

                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-400 mb-2">
                      Prompt
                    </label>
                    <textarea
                      value={prompt}
                      onChange={(e) => setPrompt(e.target.value)}
                      placeholder={
                        selectedMode === 'continuation'
                          ? 'Continue the story...'
                          : selectedMode === 'dialogue'
                          ? 'Write dialogue for...'
                          : selectedMode === 'scene'
                          ? 'Describe the scene where...'
                          : 'Create an outline for...'
                      }
                      className="w-full h-24 px-3 py-2 bg-gray-800/50 border border-gray-700 rounded-lg text-gray-200 focus:border-electric-500 focus:outline-none resize-none"
                    />
                  </div>

                  <button
                    onClick={generateText}
                    disabled={isGenerating || (!prompt.trim() && selectedMode !== 'continuation')}
                    className="w-full px-4 py-3 bg-electric-500/20 hover:bg-electric-500/30 border border-electric-500/50 rounded-lg text-electric-400 font-medium transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {isGenerating ? (
                      <span className="flex items-center justify-center gap-2">
                        <span className="animate-spin">⚡</span>
                        Generating...
                      </span>
                    ) : (
                      '✨ Generate Text'
                    )}
                  </button>

                  {generatedText && (
                    <div className="space-y-3">
                      <div className="p-4 bg-mint-500/10 border border-mint-500/30 rounded-lg">
                        <div className="text-sm text-gray-300 whitespace-pre-wrap">
                          {generatedText}
                        </div>
                      </div>
                      <div className="flex gap-2">
                        <button
                          onClick={insertGenerated}
                          className="flex-1 px-4 py-2 bg-mint-500/20 hover:bg-mint-500/30 border border-mint-500/50 rounded-lg text-mint-400 font-medium transition-all"
                        >
                          ✓ Insert
                        </button>
                        <button
                          onClick={() => setGeneratedText('')}
                          className="flex-1 px-4 py-2 bg-gray-800/50 hover:bg-gray-700/50 border border-gray-700 rounded-lg text-gray-300 transition-all"
                        >
                          ✕ Discard
                        </button>
                      </div>
                    </div>
                  )}

                  {/* Quick Stats */}
                  <div className="mt-6 p-4 bg-gray-800/50 rounded-lg space-y-2">
                    <h4 className="text-sm font-semibold text-gray-400 mb-2">Document Stats</h4>
                    <div className="text-xs text-gray-500 space-y-1">
                      <div className="flex justify-between">
                        <span>Words:</span>
                        <span className="text-gray-300">{content.split(/\s+/).filter(Boolean).length}</span>
                      </div>
                      <div className="flex justify-between">
                        <span>Characters:</span>
                        <span className="text-gray-300">{content.length}</span>
                      </div>
                      <div className="flex justify-between">
                        <span>Lines:</span>
                        <span className="text-gray-300">{content.split('\n').length}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center">
            <div className="text-center">
              <div className="text-6xl mb-4">✍️</div>
              <h3 className="text-2xl font-bold text-gray-300 mb-2">No Document Selected</h3>
              <p className="text-gray-500 mb-6">Create a new document or select one from the sidebar</p>
              <button
                onClick={() => setShowNewDocDialog(true)}
                className="px-6 py-3 bg-electric-500/20 hover:bg-electric-500/30 border border-electric-500/50 rounded-lg text-electric-400 font-medium transition-all"
              >
                + Create New Document
              </button>
            </div>
          </div>
        )}
      </div>

      {/* New Document Dialog */}
      {showNewDocDialog && (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4">
          <div className="bg-gray-900 border border-electric-500/30 rounded-xl max-w-md w-full anime-glow">
            <div className="p-6 border-b border-gray-800">
              <h2 className="text-2xl font-bold text-electric-400">Create New Document</h2>
            </div>

            <div className="p-6 space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-2">Document Title</label>
                <input
                  type="text"
                  value={newDocTitle}
                  onChange={(e) => setNewDocTitle(e.target.value)}
                  placeholder="Enter document title..."
                  className="w-full px-4 py-3 bg-gray-800/50 border border-gray-700 rounded-lg focus:outline-none focus:border-electric-500 text-white"
                  onKeyDown={(e) => e.key === 'Enter' && createDocument()}
                  autoFocus
                />
              </div>
            </div>

            <div className="p-6 border-t border-gray-800 flex gap-3">
              <button
                onClick={() => {
                  setShowNewDocDialog(false)
                  setNewDocTitle('')
                }}
                className="flex-1 px-6 py-3 bg-gray-800/50 hover:bg-gray-700/50 border border-gray-700 rounded-lg text-gray-300 transition-all"
              >
                Cancel
              </button>
              <button
                onClick={createDocument}
                disabled={!newDocTitle.trim()}
                className="flex-1 px-6 py-3 bg-electric-500/20 hover:bg-electric-500/30 border border-electric-500/50 rounded-lg text-electric-400 font-medium transition-all disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Create
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
