'use client'

import { useParams, useSearchParams, useRouter } from 'next/navigation'
import { useEffect, useMemo } from 'react'
import { observer } from 'mobx-react-lite'
import { ComicReadStore } from '@/stores'
import { getComic } from '@/apis'

const ComicReadMobile = observer(() => {
  const params = useParams<{ id: string }>()
  const searchParams = useSearchParams()
  const router = useRouter()
  const sectionIndex = searchParams.get('section-index')

  const store = useMemo(() => new ComicReadStore(), [])

  useEffect(() => {
    const initializeStore = async () => {
      if (!params?.id || !sectionIndex) return
      
      try {
        const response = await getComic(params.id)
        if (response.code === 200 && response.data.sections) {
          const section = response.data.sections.find(
            s => s.index === parseInt(sectionIndex)
          )
          
          if (section) {
            await store.initialize(params.id, section.id)
          }
        }
      } catch (error) {
        console.error('Failed to initialize comic reader:', error)
      }
    }
    
    initializeStore()
    
    return () => {
      store.dispose()
    }
  }, [params?.id, sectionIndex, store])

  const handleBackClick = () => {
    router.push(`/comic/detail/${params?.id}`)
  }

  const imageUrl = store.pageManager.imageUrl

  return (
    <div className="relative w-full h-screen bg-black overflow-hidden">
      <div
        className="w-full h-full flex items-center justify-center cursor-pointer"
        onClick={() => store.handleScreenTap()}
      >
        {imageUrl ? (
          <img
            src={imageUrl}
            alt="Comic page"
            className="w-full h-full object-contain"
          />
        ) : (
          <div className="w-full h-full bg-gray-800 flex items-center justify-center">
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

      {store.isLoading && (
        <div className="fixed bottom-6 right-6 w-10 h-10 flex items-center justify-center">
          <div className="w-8 h-8 border-4 border-white/30 border-t-white rounded-full animate-spin"></div>
        </div>
      )}

      {store.showOverlay && (
        <div className="fixed inset-0 bg-white/20 backdrop-blur-md z-50 flex flex-col">
          <div className="flex items-center justify-between px-6 py-4 bg-white/10">
            <button
              onClick={handleBackClick}
              className="text-white text-2xl w-10 h-10 flex items-center justify-center cursor-pointer"
            >
              ←
            </button>
            <h1 className="text-white text-lg font-medium flex-1 text-center mx-4">
              {store.comicTitle || '...'}
            </h1>
            <div className="w-10"></div>
          </div>

          <div className="flex-1 flex flex-col items-center justify-center px-6">
            <div className="text-center mb-12">
              <div className="text-white/60 text-sm mb-1">{store.currentChapter}</div>
              <div className="text-white text-base">{store.chapterTitle}</div>
            </div>

            <button
              onClick={() => store.handlePlayButtonClick()}
              className="w-32 h-32 rounded-full bg-white/20 backdrop-blur-sm flex items-center justify-center hover:bg-white/30 transition-colors cursor-pointer"
            >
              <div className="w-0 h-0 border-l-[24px] border-l-white border-t-[16px] border-t-transparent border-b-[16px] border-b-transparent ml-2"></div>
            </button>
          </div>

          <div className="px-6 py-8 text-center">
            <div className="text-white/60 text-sm">
              {store.currentPageNumber} / {store.totalPages}
            </div>
          </div>
        </div>
      )}
    </div>
  )
})

export default ComicReadMobile
