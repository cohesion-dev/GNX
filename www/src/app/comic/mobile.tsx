'use client'

const testList = [
  {
    id: 1,
    title: '示例漫画 1',
    description: '这是一个精彩的漫画故事，讲述了英雄们的冒险旅程，充满了惊险刺激的情节和动人心弦的情感...',
    image: '',
    isLoading: false
  },
  {
    id: 2,
    title: '示例漫画 2',
    description: '另一个有趣的故事，主角在神秘的世界中探索未知的领域，遇到了各种各样的挑战和机遇...',
    image: '',
    isLoading: true
  },
  {
    id: 3,
    title: '示例漫画 3',
    description: '一段温馨的日常故事，讲述了朋友之间的友谊和成长，让人感受到生活中的美好瞬间...',
    image: '',
    isLoading: false
  }
]

const ComicListMobile = () => {
  return (
    <div className="min-h-screen bg-gradient-to-b from-slate-800 to-purple-900 text-white">
      <div className="px-6 py-4 pt-12 text-center">
        <h1 className="text-xl font-bold">列表</h1>
      </div>

      <div className="px-6 mt-6 space-y-4">
        {testList.map((item) => (
          <div
            key={item.id}
            className="bg-white/10 backdrop-blur-md rounded-3xl p-4 m-6"
          >
            <div className="flex gap-4">
              <div className="w-16 h-16 flex items-center justify-center bg-gradient-to-br from-cyan-400 via-blue-500 to-purple-600 rounded-full overflow-hidden flex-shrink-0">
                <span className="text-3xl">📚</span>
              </div>

              <div className="flex-1 min-w-0">
                <h3 className="text-cyan-400 font-bold text-lg mb-1">
                  {item.title}
                </h3>
                <p className="text-gray-300 text-sm line-clamp-3">
                  {item.description}
                </p>
              </div>

              {item.isLoading && (
                <div className="flex-shrink-0">
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
            </div>
          </div>
        ))}
      </div>

      <div className="fixed bottom-0 left-0 right-0 bg-slate-900/80 backdrop-blur-lg border-t border-white/10">
        <div className="flex justify-around py-3 px-6">
          <button
            className="w-12 h-12 flex items-center justify-center bg-cyan-400/20 backdrop-blur-sm rounded-full"
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
    </div>
  )
}

export default ComicListMobile
