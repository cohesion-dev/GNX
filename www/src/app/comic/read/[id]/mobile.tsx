'use client'

import { useParams, useSearchParams, useRouter } from 'next/navigation'
import { useState } from 'react'

const ComicReadMobile = () => {
  const params = useParams<{ id: string }>()
  const searchParams = useSearchParams()
  const router = useRouter()
  const sectionIndex = searchParams.get('section-index')

  const [isPlaying, setIsPlaying] = useState(true)
  const [showOverlay, setShowOverlay] = useState(false)
  const [isLoading, setIsLoading] = useState(false)

  const comicName = '漫画标题'
  const currentChapter = '第1章'
  const chapterTitle = '章节标题'
  const currentPage = 1
  const totalPages = 10

  const handleScreenTap = () => {
    if (isPlaying) {
      setIsPlaying(false)
      setShowOverlay(true)
    }
  }

  const handlePlayButtonClick = () => {
    setIsPlaying(true)
    setShowOverlay(false)
  }

  const handleBackClick = () => {
    router.push(`/comic/detail/${params?.id}`)
  }

  return (
    <div className="relative w-full h-screen bg-black overflow-hidden">
      <div
        className="w-full h-full flex items-center justify-center cursor-pointer"
        onClick={handleScreenTap}
      >
        <div className="w-full h-full bg-gray-800 flex items-center justify-center">
          <span className="text-gray-500 text-sm">漫画图片</span>
        </div>
      </div>

      {isLoading && (
        <div className="fixed bottom-6 right-6 w-10 h-10 flex items-center justify-center">
          <div className="w-8 h-8 border-4 border-white/30 border-t-white rounded-full animate-spin"></div>
        </div>
      )}

      {showOverlay && (
        <div className="fixed inset-0 bg-white/20 backdrop-blur-md z-50 flex flex-col">
          <div className="flex items-center justify-between px-6 py-4 bg-white/10">
            <button
              onClick={handleBackClick}
              className="text-white text-2xl w-10 h-10 flex items-center justify-center cursor-pointer"
            >
              ←
            </button>
            <h1 className="text-white text-lg font-medium flex-1 text-center mx-4">
              {comicName}
            </h1>
            <div className="w-10"></div>
          </div>

          <div className="flex-1 flex flex-col items-center justify-center px-6">
            <div className="text-center mb-12">
              <div className="text-white/60 text-sm mb-1">{currentChapter}</div>
              <div className="text-white text-base">{chapterTitle}</div>
            </div>

            <button
              onClick={handlePlayButtonClick}
              className="w-32 h-32 rounded-full bg-white/20 backdrop-blur-sm flex items-center justify-center hover:bg-white/30 transition-colors cursor-pointer"
            >
              <div className="w-0 h-0 border-l-[24px] border-l-white border-t-[16px] border-t-transparent border-b-[16px] border-b-transparent ml-2"></div>
            </button>
          </div>

          <div className="px-6 py-8 text-center">
            <div className="text-white/60 text-sm">
              {currentPage} / {totalPages}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default ComicReadMobile
