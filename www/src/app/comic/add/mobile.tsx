'use client'

import { useState, useEffect, useCallback, useRef } from 'react'
import { useRouter } from 'next/navigation'
import { getComics, createComic } from '@/apis/comic'

const ComicAddMobile = () => {
  const router = useRouter()
  const [hasComics, setHasComics] = useState(false)
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [title, setTitle] = useState('')
  const [userPrompt, setUserPrompt] = useState('')
  const [file, setFile] = useState<File | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    const checkComics = async () => {
      setLoading(true)
      try {
        const response = await getComics({ page: 1, limit: 1 })
        if (response.code === 200) {
          setHasComics(response.data.comics.length > 0)
        }
      } catch (error) {
        console.error('Failed to fetch comics:', error)
      } finally {
        setLoading(false)
      }
    }
    checkComics()
  }, [])

  const handleBackClick = useCallback(() => {
    router.push('/comic')
  }, [router])

  const handleFileSelect = useCallback(() => {
    fileInputRef.current?.click()
  }, [])

  const handleFileChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFile = e.target.files?.[0]
    if (selectedFile) {
      setFile(selectedFile)
    }
  }, [])

  const handleSubmit = useCallback(async () => {
    if (!title.trim() || !userPrompt.trim() || !file) {
      return
    }

    setSubmitting(true)
    try {
      const response = await createComic({
        title: title.trim(),
        user_prompt: userPrompt.trim(),
        file
      })

      if (response.code === 200) {
        router.push(`/comic/read/${response.data.id}`)
      }
    } catch (error) {
      console.error('Failed to create comic:', error)
    } finally {
      setSubmitting(false)
    }
  }, [title, userPrompt, file, router])

  const isFormValid = title.trim() && userPrompt.trim() && file

  return (
    <div className="min-h-screen text-white flex flex-col">
      {hasComics && (
        <div className="fixed top-0 left-0 right-0 bg-slate-900/20 backdrop-blur-lg z-40">
          <div className="px-6 pb-3 flex items-center" style={{ paddingTop: '16px' }}>
            <button
              onClick={handleBackClick}
              className="mr-3 p-2 -ml-2 cursor-pointer"
              aria-label="返回"
            >
              <svg
                className="w-6 h-6 text-cyan-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M15 19l-7-7 7-7"
                />
              </svg>
            </button>
            <h1 className="text-xl font-bold flex-1 text-center mr-8">
              新的开始..
            </h1>
          </div>
        </div>
      )}

      <div className={`flex-1 flex flex-col justify-center px-6 ${hasComics ? 'mt-16' : ''}`}>
        {!hasComics && (
          <h1 className="text-2xl font-bold text-center mb-12">
            沉浸式有声漫画群像剧
          </h1>
        )}

        <div className="space-y-6 max-w-md mx-auto w-full">
          <div>
            <input
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="起个名字"
              className="w-full bg-white/10 backdrop-blur-md rounded-2xl px-6 py-4 text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-cyan-400"
            />
          </div>

          <div>
            <textarea
              value={userPrompt}
              onChange={(e) => setUserPrompt(e.target.value)}
              placeholder="这次想看点什么? 美型还是水墨, 软萌还是硬朗.."
              rows={4}
              className="w-full bg-white/10 backdrop-blur-md rounded-2xl px-6 py-4 text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-cyan-400 resize-none"
            />
          </div>

          <div>
            <input
              ref={fileInputRef}
              type="file"
              onChange={handleFileChange}
              accept=".txt,.pdf,.doc,.docx"
              className="hidden"
            />
            <button
              onClick={handleFileSelect}
              className="w-full bg-white/10 backdrop-blur-md rounded-2xl px-6 py-4 text-left cursor-pointer focus:outline-none focus:ring-2 focus:ring-cyan-400"
            >
              {file ? (
                <span className="text-white">{file.name}</span>
              ) : (
                <span className="text-gray-400">选择本地小说文件</span>
              )}
            </button>
          </div>

          <button
            onClick={handleSubmit}
            disabled={!isFormValid || submitting}
            className={`w-full rounded-2xl px-6 py-4 font-bold transition-all ${
              isFormValid && !submitting
                ? 'text-white cursor-pointer hover:opacity-90'
                : 'bg-white/10 text-gray-500 cursor-not-allowed'
            }`}
            style={{
              backgroundImage: isFormValid && !submitting ? 'linear-gradient(135deg, #00E5E5 0%, #72A5F2 51.04%, #E961FF 100%)' : undefined
            }}
          >
            {submitting ? (
              <div className="flex items-center justify-center gap-2">
                <svg
                  className="w-5 h-5 animate-spin"
                  viewBox="0 0 24 24"
                  fill="none"
                >
                  <circle
                    className="opacity-25"
                    cx="12"
                    cy="12"
                    r="10"
                    stroke="currentColor"
                    strokeWidth="4"
                  />
                  <path
                    className="opacity-75"
                    fill="currentColor"
                    d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                  />
                </svg>
                <span>处理中...</span>
              </div>
            ) : (
              '准备好了'
            )}
          </button>
        </div>
      </div>
    </div>
  )
}

export default ComicAddMobile
