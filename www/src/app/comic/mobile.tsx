'use client'

import { useState, useEffect, useRef, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { getComics, type Comic } from '@/apis/comic'

const ComicListMobile = () => {
  const router = useRouter()
  const [comics, setComics] = useState<Comic[]>([])
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [hasMore, setHasMore] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [shakeId, setShakeId] = useState<number | null>(null)
  const scrollContainerRef = useRef<HTMLDivElement>(null)
  const touchStartY = useRef(0)
  const pullDistance = useRef(0)

  const fetchComics = useCallback(async (pageNum: number, isRefresh = false) => {
    if (loading && !isRefresh) return
    
    setLoading(true)
    try {
      const response = await getComics({ page: pageNum, limit: 10 })
      if (response.code === 200) {
        const newComics = response.data.comics
        if (isRefresh) {
          setComics(newComics)
        } else {
          setComics(prev => [...prev, ...newComics])
        }
        setHasMore(newComics.length === 10)
      }
    } catch (error) {
      console.error('Failed to fetch comics:', error)
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }, [loading])

  useEffect(() => {
    fetchComics(1, true)
  }, [])

  const handleRefresh = useCallback(async () => {
    setRefreshing(true)
    setPage(1)
    await fetchComics(1, true)
  }, [fetchComics])

  const handleScroll = useCallback(() => {
    if (!scrollContainerRef.current || loading || !hasMore) return

    const { scrollTop, scrollHeight, clientHeight } = scrollContainerRef.current
    if (scrollHeight - scrollTop - clientHeight < 100) {
      const nextPage = page + 1
      setPage(nextPage)
      fetchComics(nextPage)
    }
  }, [page, loading, hasMore, fetchComics])

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

  const handleComicClick = useCallback((comic: Comic) => {
    if (comic.status === 'completed') {
      router.push(`/comic/detail/${comic.id}`)
    } else {
      setShakeId(comic.id)
      setTimeout(() => setShakeId(null), 500)
    }
  }, [router])

  const handleAddClick = useCallback(() => {
    router.push('/comic/add')
  }, [router])

  const getStatusIcon = (status: Comic['status']) => {
    if (status === 'pending') {
      return (
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
      )
    }
    
    if (status === 'failed') {
      return (
        <svg
          className="w-6 h-6 text-red-400"
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
      )
    }
    
    return null
  }

  return (
    <div 
      ref={scrollContainerRef}
      className="min-h-screen text-white overflow-y-auto"
      onScroll={handleScroll}
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

      <div className="fixed top-0 left-0 right-0 bg-slate-900/30 backdrop-blur-lg z-40">
        <div className="px-6 py-2 pt-10 text-center flex items-center justify-center">
          <h1 className="text-xl font-bold">欢迎回家..</h1>
        </div>
      </div>

      <div className="px-6 mt-20 space-y-4 pb-24">
        {comics.map((comic) => (
          <div
            key={comic.id}
            onClick={() => handleComicClick(comic)}
            className={`bg-white/10 backdrop-blur-md rounded-3xl p-4 m-6 transition-transform ${
              shakeId === comic.id ? 'animate-shake' : ''
            } ${
              comic.status === 'completed' ? 'cursor-pointer' : 'cursor-not-allowed'
            }`}
            style={{
              animation: shakeId === comic.id ? 'shake 0.5s' : undefined
            }}
          >
            <div className="flex gap-4 items-center">
              <div className="w-16 h-16 flex items-center justify-center bg-gradient-to-br from-cyan-400 via-blue-500 to-purple-600 rounded-full overflow-hidden flex-shrink-0">
                {comic.icon ? (
                  <img src={comic.icon} alt={comic.title} className="w-full h-full object-cover" />
                ) : (
                  <span className="text-3xl">📚</span>
                )}
              </div>

              <div className="flex-1 min-w-0">
                <h3 className="text-cyan-400 font-bold text-lg mb-1">
                  {comic.title}
                </h3>
                <p className="text-gray-300 text-sm line-clamp-3">
                  {comic.brief}
                </p>
              </div>

              {getStatusIcon(comic.status) && (
                <div className="flex-shrink-0">
                  {getStatusIcon(comic.status)}
                </div>
              )}
            </div>
          </div>
        ))}
        
        {loading && !refreshing && (
          <div className="flex justify-center py-4">
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

      <div className="fixed bottom-0 left-0 right-0 bg-slate-900/30 backdrop-blur-lg">
        <div className="flex justify-around py-3 px-6">
          <button
            onClick={handleAddClick}
            className="w-12 h-12 flex items-center justify-center bg-cyan-400/20 backdrop-blur-sm rounded-full cursor-pointer"
            aria-label="添加"
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

export default ComicListMobile
