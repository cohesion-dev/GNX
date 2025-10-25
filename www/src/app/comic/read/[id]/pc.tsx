'use client'

import { observer } from 'mobx-react-lite'
import { useParams, useSearchParams } from 'next/navigation'
import { useEffect, useState, useRef } from 'react'
import { getSectionContent, SectionContent } from '@/apis/comic'
import { getTTSAudio } from '@/apis/tts'

interface LoadedStoryboard {
  id: number
  imageUrl: string
  audioUrls: { detailId: number; url: string }[]
}

const ComicReadPC = observer(() => {
  const params = useParams()
  const searchParams = useSearchParams()
  const comicId = params.id as string
  const sectionId = searchParams.get('sectionId')

  const [sectionContent, setSectionContent] = useState<SectionContent | null>(null)
  const [loadedStoryboards, setLoadedStoryboards] = useState<LoadedStoryboard[]>([])
  const [currentStoryboardIndex, setCurrentStoryboardIndex] = useState(0)
  const [currentDetailIndex, setCurrentDetailIndex] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isPlaying, setIsPlaying] = useState(false)

  const audioRef = useRef<HTMLAudioElement>(null)

  // 加载 section 内容
  useEffect(() => {
    if (!comicId || !sectionId) {
      setError('缺少 comic ID 或 section ID')
      setLoading(false)
      return
    }

    const loadSectionContent = async () => {
      try {
        setLoading(true)
        setError(null)
        const response = await getSectionContent(Number(comicId), Number(sectionId))

        if (response.code !== 200) {
          throw new Error(response.message || '加载 section 内容失败')
        }

        setSectionContent(response.data)

        // 开始加载图片和音频
        await loadStoryboardsData(response.data)

        setLoading(false)
        // 加载完成后自动开始播放
        setIsPlaying(true)
      } catch (err) {
        console.error('加载 section 内容失败:', err)
        setError(err instanceof Error ? err.message : '加载失败')
        setLoading(false)
      }
    }

    loadSectionContent()
  }, [comicId, sectionId])

  // 加载所有 storyboards 的图片和音频
  const loadStoryboardsData = async (content: SectionContent) => {
    const loaded: LoadedStoryboard[] = []

    for (const storyboard of content.storyboards) {
      try {
        // 加载音频
        const audioUrls: { detailId: number; url: string }[] = []

        for (const detail of storyboard.details) {
          try {
            const audioBlob = await getTTSAudio(detail.id)
            const audioUrl = URL.createObjectURL(audioBlob)
            audioUrls.push({ detailId: detail.id, url: audioUrl })
          } catch (err) {
            console.error(`加载音频失败 (detail ${detail.id}):`, err)
          }
        }

        loaded.push({
          id: storyboard.id,
          imageUrl: storyboard.image_url,
          audioUrls
        })
      } catch (err) {
        console.error(`加载 storyboard ${storyboard.id} 失败:`, err)
      }
    }

    setLoadedStoryboards(loaded)
  }

  // 播放当前音频
  useEffect(() => {
    if (!isPlaying || !audioRef.current) return

    const currentStoryboard = loadedStoryboards[currentStoryboardIndex]
    if (!currentStoryboard) return

    const currentAudio = currentStoryboard.audioUrls[currentDetailIndex]
    if (!currentAudio) {
      // 当前 storyboard 的所有音频播放完毕，切换到下一个 storyboard
      if (currentStoryboardIndex < loadedStoryboards.length - 1) {
        setCurrentStoryboardIndex(currentStoryboardIndex + 1)
        setCurrentDetailIndex(0)
      } else {
        // 全部播放完毕
        setIsPlaying(false)
      }
      return
    }

    // 设置音频源并播放
    audioRef.current.src = currentAudio.url
    audioRef.current.play().catch(err => {
      console.error('播放音频失败:', err)
    })
  }, [isPlaying, currentStoryboardIndex, currentDetailIndex, loadedStoryboards])

  // 音频播放结束时自动播放下一个
  const handleAudioEnded = () => {
    const currentStoryboard = loadedStoryboards[currentStoryboardIndex]
    if (!currentStoryboard) return

    // 检查是否还有更多音频在当前 storyboard
    if (currentDetailIndex < currentStoryboard.audioUrls.length - 1) {
      setCurrentDetailIndex(currentDetailIndex + 1)
    } else if (currentStoryboardIndex < loadedStoryboards.length - 1) {
      // 切换到下一个 storyboard
      setCurrentStoryboardIndex(currentStoryboardIndex + 1)
      setCurrentDetailIndex(0)
    } else {
      // 全部播放完毕
      setIsPlaying(false)
    }
  }

  // 切换播放/暂停
  const togglePlayPause = () => {
    if (isPlaying) {
      audioRef.current?.pause()
      setIsPlaying(false)
    } else {
      audioRef.current?.play().catch(err => console.error('播放失败:', err))
      setIsPlaying(true)
    }
  }

  // 清理音频 URL
  useEffect(() => {
    return () => {
      loadedStoryboards.forEach(storyboard => {
        storyboard.audioUrls.forEach(audio => {
          URL.revokeObjectURL(audio.url)
        })
      })
    }
  }, [loadedStoryboards])

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="text-xl font-semibold mb-2">加载中...</div>
          <div className="text-gray-600">正在加载漫画内容</div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center text-red-600">
          <div className="text-xl font-semibold mb-2">加载失败</div>
          <div>{error}</div>
        </div>
      </div>
    )
  }

  if (!sectionContent || loadedStoryboards.length === 0) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center text-gray-600">
          <div className="text-xl font-semibold mb-2">暂无内容</div>
        </div>
      </div>
    )
  }

  const currentStoryboard = loadedStoryboards[currentStoryboardIndex]

  return (
    <div className="min-h-screen bg-gray-900 flex flex-col">
      {/* 顶部信息栏 */}
      <div className="bg-gray-800 text-white p-4 shadow-lg">
        <div className="max-w-6xl mx-auto flex justify-between items-center">
          <div>
            <h1 className="text-2xl font-bold">漫画阅读</h1>
            <p className="text-gray-400 text-sm mt-1">
              Section {sectionId} - 进度: {currentStoryboardIndex + 1} / {loadedStoryboards.length}
            </p>
          </div>
          <div className="flex items-center gap-4">
            <button
              onClick={togglePlayPause}
              className="px-6 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg font-semibold transition-colors"
            >
              {isPlaying ? '暂停' : '播放'}
            </button>
          </div>
        </div>
      </div>

      {/* 主内容区域 */}
      <div className="flex-1 flex items-center justify-center p-8">
        <div className="max-w-4xl w-full">
          {currentStoryboard && (
            <div className="bg-gray-800 rounded-lg overflow-hidden shadow-2xl">
              <img
                src={currentStoryboard.imageUrl}
                alt={`Storyboard ${currentStoryboard.id}`}
                className="w-full h-auto"
              />
            </div>
          )}
        </div>
      </div>

      {/* 底部进度指示器 */}
      <div className="bg-gray-800 text-white p-4">
        <div className="max-w-6xl mx-auto">
          <div className="flex gap-2 justify-center">
            {loadedStoryboards.map((_, index) => (
              <div
                key={index}
                className={`h-2 flex-1 rounded-full transition-colors ${
                  index === currentStoryboardIndex
                    ? 'bg-blue-600'
                    : index < currentStoryboardIndex
                    ? 'bg-green-600'
                    : 'bg-gray-600'
                }`}
              />
            ))}
          </div>
          <div className="text-center text-sm text-gray-400 mt-2">
            {currentStoryboard && (
              <>音频: {currentDetailIndex + 1} / {currentStoryboard.audioUrls.length}</>
            )}
          </div>
        </div>
      </div>

      {/* 音频播放器 */}
      <audio
        ref={audioRef}
        onEnded={handleAudioEnded}
        className="hidden"
      />
    </div>
  )
})

export default ComicReadPC
