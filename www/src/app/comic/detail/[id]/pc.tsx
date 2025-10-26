'use client'

import { useState, useEffect, useCallback } from 'react'
import { useRouter, useParams } from 'next/navigation'
import { getComic, type ComicDetail, type Section } from '@/apis'
import ComicBackground from '@/components/ComicBackground'

const ComicDetailPC = () => {
  const router = useRouter()
  const params = useParams<{ id: string }>()
  const [comic, setComic] = useState<ComicDetail | null>(null)
  const [sections, setSections] = useState<Section[]>([])
  const [loading, setLoading] = useState(false)
  const [shakeId, setShakeId] = useState<string | null>(null)

  const fetchComicData = useCallback(async () => {
    if (!params?.id || loading) return
    
    setLoading(true)
    try {
      const comicResponse = await getComic(params.id)
      
      if (comicResponse.code === 200) {
        setComic(comicResponse.data)
        setSections(comicResponse.data.sections || [])
      }
    } catch (error) {
      console.error('Failed to fetch comic data:', error)
    } finally {
      setLoading(false)
    }
  }, [params?.id, loading])

  useEffect(() => {
    fetchComicData()
  }, [])

  const handleBackClick = useCallback(() => {
    router.push('/comic')
  }, [router])

  const handleSectionClick = useCallback((section: Section) => {
    if (section.status === 'completed' && params?.id) {
      router.push(`/comic/read/${params.id}?section-index=${section.index}`)
    } else {
      setShakeId(section.id)
      setTimeout(() => setShakeId(null), 500)
    }
  }, [router, params?.id])

  const handleAddSectionClick = useCallback(() => {
    if (params?.id) {
      router.push(`/comic/detail/${params.id}/section/add`)
    }
  }, [router, params?.id])

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900 text-white">
      <div className="max-w-6xl mx-auto px-8 py-8">
        <div className="flex items-center mb-8">
          <button
            onClick={handleBackClick}
            className="mr-4 p-2 hover:bg-white/10 rounded-lg transition-colors cursor-pointer"
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
          <h1 className="text-3xl font-bold">
            {comic?.title || '加载中...'}
          </h1>
        </div>

        {comic && (
          <div className="bg-white/10 backdrop-blur-md rounded-3xl p-8 mb-8">
            <div className="flex gap-8">
              <ComicBackground
                imageId={comic.background_image_id}
                alt={comic.title}
                className="w-48 h-64 flex-shrink-0 rounded-lg overflow-hidden object-cover"
              />
              <div className="flex-1 min-w-0">
                <h2 className="text-xl font-semibold mb-4 text-cyan-400">简介</h2>
                <p className="text-gray-300 text-base leading-relaxed">
                  {comic.brief}
                </p>
              </div>
            </div>
          </div>
        )}

        <div className="flex items-center justify-between mb-6">
          <h2 className="text-2xl font-bold">章节列表</h2>
          <button
            onClick={handleAddSectionClick}
            className="px-6 py-3 bg-cyan-400/20 hover:bg-cyan-400/30 backdrop-blur-sm rounded-xl flex items-center gap-2 transition-colors cursor-pointer"
            aria-label="添加章节"
          >
            <svg
              className="w-5 h-5 text-cyan-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2.5}
                d="M12 4v16m8-8H4"
              />
            </svg>
            <span className="text-cyan-400 font-medium">添加章节</span>
          </button>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {sections.map((section) => (
            <div
              key={section.id}
              onClick={() => handleSectionClick(section)}
              className={`bg-white/10 backdrop-blur-md rounded-2xl p-6 transition-all ${
                shakeId === section.id ? 'animate-shake' : ''
              } ${
                section.status === 'completed' 
                  ? 'cursor-pointer hover:bg-white/15 hover:scale-105' 
                  : 'cursor-not-allowed opacity-60'
              }`}
              style={{
                animation: shakeId === section.id ? 'shake 0.5s' : undefined
              }}
            >
              <div className="flex items-center justify-between">
                <div className="flex-1">
                  <h3 className="text-lg">
                    <span className="text-cyan-400 font-semibold">第 {section.index} 章</span>
                  </h3>
                  <p className="text-white mt-2">{section.title}</p>
                </div>
                {section.status === 'pending' && (
                  <svg
                    className="w-6 h-6 animate-spin ml-3 flex-shrink-0"
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
                )}
                {section.status === 'failed' && (
                  <svg
                    className="w-6 h-6 text-red-400 ml-3 flex-shrink-0"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                  >
                    <circle cx="12" cy="12" r="10" strokeWidth="2" />
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth="2"
                      d="M12 8v4m0 4h.01"
                    />
                  </svg>
                )}
              </div>
            </div>
          ))}
        </div>
        
        {loading && (
          <div className="flex justify-center py-12">
            <svg
              className="w-12 h-12 animate-spin"
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
          </div>
        )}
      </div>

      <style jsx>{`
        @keyframes shake {
          0%, 100% { transform: translateX(0); }
          25% { transform: translateX(-10px); }
          75% { transform: translateX(10px); }
        }
      `}</style>
    </div>
  )
}

export default ComicDetailPC
