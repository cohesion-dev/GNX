'use client'

import { useState, useEffect, useRef, useCallback } from 'react'
import { useRouter, useParams } from 'next/navigation'
import { getComic, type ComicDetail, type Section } from '@/apis'

const ComicDetailMobile = () => {
  const router = useRouter()
  const params = useParams<{ id: string }>()
  const [comic, setComic] = useState<ComicDetail | null>(null)
  const [sections, setSections] = useState<Section[]>([])
  const [loading, setLoading] = useState(false)
  const [refreshing, setRefreshing] = useState(false)
  const [shakeId, setShakeId] = useState<string | null>(null)
  const scrollContainerRef = useRef<HTMLDivElement>(null)
  const touchStartY = useRef(0)
  const pullDistance = useRef(0)

  const fetchComicData = useCallback(async (isRefresh = false) => {
    if (!params?.id || (loading && !isRefresh)) return
    
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
      setRefreshing(false)
    }
  }, [params?.id])

  useEffect(() => {
    fetchComicData(true)
  }, [])

  const handleRefresh = useCallback(async () => {
    setRefreshing(true)
    await fetchComicData(true)
  }, [fetchComicData])

  const handleTouchStart = useCallback((e: React.TouchEvent) => {
    if (scrollContainerRef.current && scrollContainerRef.current.scrollTop === 0) {
      touchStartY.current = e.touches[0].clientY
    }
  }, [])

  const handleTouchMove = useCallback((e: React.TouchEvent) => {
    if (touchStartY.current === 0 || refreshing) return

    const currentY = e.touches[0].clientY
    const diff = currentY - touchStartY.current

    if (diff > 0 && scrollContainerRef.current && scrollContainerRef.current.scrollTop === 0) {
      pullDistance.current = Math.min(diff, 100)
    }
  }, [refreshing])

  const handleTouchEnd = useCallback(() => {
    if (pullDistance.current > 60 && !refreshing) {
      handleRefresh()
    }
    touchStartY.current = 0
    pullDistance.current = 0
  }, [refreshing, handleRefresh])

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
    <div 
      ref={scrollContainerRef}
      className="min-h-screen text-white overflow-y-auto"
      onTouchStart={handleTouchStart}
      onTouchMove={handleTouchMove}
      onTouchEnd={handleTouchEnd}
    >
      {refreshing && (
        <div className="fixed top-0 left-0 right-0 h-16 flex items-center justify-center bg-slate-900/50 z-50">
          <svg
            className="w-6 h-6 animate-spin"
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
            {comic?.title || '加载中...'}
          </h1>
        </div>
      </div>

      <div className="px-6 mt-20 pb-24">
        {comic && (
          <div className="bg-white/10 backdrop-blur-md rounded-3xl p-4 mb-6">
            <div className="flex gap-4">
              <div className="w-24 h-32 flex-shrink-0 rounded-lg overflow-hidden bg-gradient-to-br from-cyan-400 via-blue-500 to-purple-600">
                <div className="w-full h-full flex items-center justify-center">
                  <span className="text-4xl">📚</span>
                </div>
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-gray-300 text-sm line-clamp-5">
                  {comic.brief}
                </p>
              </div>
            </div>
          </div>
        )}

        <div className="space-y-3">
          {sections.map((section) => (
            <div
              key={section.id}
              onClick={() => handleSectionClick(section)}
              className={`bg-white/10 backdrop-blur-md rounded-2xl p-4 transition-transform ${
                shakeId === section.id ? 'animate-shake' : ''
              } ${
                section.status === 'completed' ? 'cursor-pointer hover:bg-white/15' : 'cursor-not-allowed opacity-60'
              }`}
              style={{
                animation: shakeId === section.id ? 'shake 0.5s' : undefined
              }}
            >
              <div className="flex items-center justify-between">
                <div className="flex-1">
                  <h3 className="text-base">
                    <span className="text-cyan-400">第 {section.index} 章</span>
                    <span className="text-white ml-3">{section.title}</span>
                  </h3>
                </div>
                {section.status === 'pending' && (
                  <svg
                    className="w-5 h-5 animate-spin ml-3 flex-shrink-0"
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
                    className="w-5 h-5 text-red-400 ml-3 flex-shrink-0"
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
        
        {loading && !refreshing && (
          <div className="flex justify-center py-8">
            <svg
              className="w-8 h-8 animate-spin"
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

      <div className="fixed bottom-0 left-0 right-0 bg-slate-900/20 backdrop-blur-lg">
        <div className="flex justify-around py-3 px-6">
          <button
            onClick={handleAddSectionClick}
            className="w-12 h-12 flex items-center justify-center bg-cyan-400/20 backdrop-blur-sm rounded-full cursor-pointer"
            aria-label="添加章节"
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
                strokeWidth={2.5}
                d="M12 4v16m8-8H4"
              />
            </svg>
          </button>
        </div>
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

export default ComicDetailMobile
