import { observer } from 'mobx-react-lite'
import { useState, useEffect } from 'react'
import { useRouter, useParams } from 'next/navigation'
import { createSection, getComicSections } from '@/apis/comic'

const ComicSectionAddPC = observer(() => {
  const router = useRouter()
  const params = useParams()
  const comicId = Number(params.id)

  const [detail, setDetail] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [nextIndex, setNextIndex] = useState(1)
  const [fetchingIndex, setFetchingIndex] = useState(true)

  // 获取当前漫画已有的章节数量，确定新章节的 index
  useEffect(() => {
    const fetchSections = async () => {
      try {
        const response = await getComicSections(comicId)
        if (response.code === 0 && response.data) {
          // 找到最大的 index，新章节的 index 为最大值 + 1
          const maxIndex = response.data.reduce((max, section) =>
            Math.max(max, section.index), 0)
          setNextIndex(maxIndex + 1)
        }
      } catch (err) {
        console.error('获取章节信息错误:', err)
        setError('获取章节信息失败')
      } finally {
        setFetchingIndex(false)
      }
    }

    if (comicId) {
      fetchSections()
    }
  }, [comicId])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    // 验证
    if (!detail.trim()) {
      setError('请输入章节文本')
      return
    }

    setLoading(true)

    try {
      const response = await createSection(comicId, {
        index: nextIndex,
        detail: detail.trim(),
      })

      if (response.code === 0 && response.data) {
        // 成功后跳转到漫画阅读页面，并跳转到对应章节
        router.push(`/comic/read/${comicId}?section=${response.data.id}`)
      } else {
        setError(response.message || '新增章节失败，请重试')
        setLoading(false)
      }
    } catch (err) {
      setError('新增章节失败，请重试')
      setLoading(false)
      console.error('新增章节错误:', err)
    }
  }

  if (fetchingIndex) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-gray-600">加载中...</div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-2xl mx-auto">
        <div className="bg-white rounded-lg shadow-md p-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-8 text-center">
            新增漫画章节
          </h1>

          <form onSubmit={handleSubmit} className="space-y-6">
            {/* 章节序号提示 */}
            <div className="bg-blue-50 border border-blue-200 text-blue-700 px-4 py-3 rounded-lg">
              <p className="text-sm">
                即将创建第 <span className="font-bold">{nextIndex}</span> 章
              </p>
            </div>

            {/* 章节文本 */}
            <div>
              <label htmlFor="detail" className="block text-sm font-medium text-gray-700 mb-2">
                章节文本 <span className="text-red-500">*</span>
              </label>
              <textarea
                id="detail"
                value={detail}
                onChange={(e) => setDetail(e.target.value)}
                placeholder="请输入章节文本内容，后端将根据此文本生成漫画章节"
                rows={12}
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition resize-none"
                disabled={loading}
              />
              <p className="mt-2 text-sm text-gray-500">
                当前字数: {detail.length}
              </p>
            </div>

            {/* 错误提示 */}
            {error && (
              <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">
                {error}
              </div>
            )}

            {/* 提交按钮 */}
            <div className="flex space-x-4">
              <button
                type="submit"
                disabled={loading}
                className="flex-1 bg-blue-600 text-white py-3 px-6 rounded-lg font-medium hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:bg-gray-400 disabled:cursor-not-allowed transition"
              >
                {loading ? (
                  <span className="flex items-center justify-center">
                    <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    生成中...
                  </span>
                ) : (
                  '提交生成'
                )}
              </button>
              <button
                type="button"
                onClick={() => router.back()}
                disabled={loading}
                className="px-6 py-3 border border-gray-300 rounded-lg font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 disabled:bg-gray-100 disabled:cursor-not-allowed transition"
              >
                取消
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  )
})

export default ComicSectionAddPC
