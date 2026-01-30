import { useState, useCallback } from 'react'

interface ScriptUploadProps {
  onUpload: (file: File) => void
}

export default function ScriptUpload({ onUpload }: ScriptUploadProps) {
  const [isDragging, setIsDragging] = useState(false)
  const [selectedFile, setSelectedFile] = useState<File | null>(null)

  const handleDragEnter = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setIsDragging(true)
  }, [])

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setIsDragging(false)
  }, [])

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
  }, [])

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setIsDragging(false)

    const files = Array.from(e.dataTransfer.files)
    const pdfFile = files.find(f => f.type === 'application/pdf')

    if (pdfFile) {
      setSelectedFile(pdfFile)
    }
  }, [])

  const handleFileSelect = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file && file.type === 'application/pdf') {
      setSelectedFile(file)
    }
  }, [])

  const handleUpload = () => {
    if (selectedFile) {
      onUpload(selectedFile)
      setSelectedFile(null)
    }
  }

  const formatFileSize = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  }

  return (
    <div className="flex flex-col items-center justify-center h-full p-8">
      <div
        onDragEnter={handleDragEnter}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        className={`
          relative w-full max-w-2xl border-2 border-dashed rounded-2xl p-12
          transition-all duration-300 ease-out
          ${isDragging
            ? 'border-sunset-500 bg-sunset-500/10 scale-105'
            : 'border-gray-700 bg-gray-800/30 hover:border-gray-600 hover:bg-gray-800/50'
          }
        `}
      >
        <input
          type="file"
          accept=".pdf"
          onChange={handleFileSelect}
          className="hidden"
          id="file-upload"
        />

        <div className="text-center">
          {selectedFile ? (
            <>
              <div className="text-6xl mb-4 animate-bounce">📄</div>
              <h3 className="text-xl font-bold text-gray-200 mb-2">{selectedFile.name}</h3>
              <p className="text-gray-400 mb-6">{formatFileSize(selectedFile.size)}</p>
              <div className="flex gap-3 justify-center">
                <button
                  onClick={() => setSelectedFile(null)}
                  className="px-6 py-3 bg-gray-700/50 hover:bg-gray-700 border border-gray-600 rounded-lg text-gray-300 font-medium transition-all"
                >
                  Change File
                </button>
                <button
                  onClick={handleUpload}
                  className="px-6 py-3 bg-sunset-500/20 hover:bg-sunset-500/30 border border-sunset-500/50 rounded-lg text-sunset-400 font-medium transition-all anime-glow-sunset"
                >
                  Parse Script →
                </button>
              </div>
            </>
          ) : (
            <>
              <div className="text-6xl mb-4">🎬</div>
              <h3 className="text-2xl font-bold text-gray-200 mb-3">Upload Your Script</h3>
              <p className="text-gray-400 mb-6 max-w-md mx-auto">
                Drag and drop your screenplay PDF here, or click to browse
              </p>
              <label
                htmlFor="file-upload"
                className="inline-block px-8 py-4 bg-sunset-500/20 hover:bg-sunset-500/30 border border-sunset-500/50 rounded-lg text-sunset-400 font-medium transition-all cursor-pointer anime-glow-sunset"
              >
                Choose PDF File
              </label>
              <p className="text-xs text-gray-500 mt-4">Supports PDF files up to 50MB</p>
            </>
          )}
        </div>

        {isDragging && (
          <div className="absolute inset-0 bg-sunset-500/5 rounded-2xl pointer-events-none" />
        )}
      </div>

      <div className="mt-8 text-center text-sm text-gray-500 max-w-lg">
        <p className="mb-2">Automatic parsing into structured format:</p>
        <div className="flex items-center justify-center gap-4 text-gray-400">
          <span>Acts</span>
          <span>→</span>
          <span>Scenes</span>
          <span>→</span>
          <span>Analysis</span>
        </div>
      </div>
    </div>
  )
}
