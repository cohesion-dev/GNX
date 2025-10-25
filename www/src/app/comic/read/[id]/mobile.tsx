'use client'

import { useParams, useSearchParams } from 'next/navigation'

const ComicReadMobile = () => {
  const params = useParams<{ id: string }>()
  const searchParams = useSearchParams()
  const sectionIndex = searchParams.get('section-index')

  return (
    <div className="min-h-screen text-white p-6">
      <h1 className="text-2xl font-bold mb-4">Comic Read Page</h1>
      <div className="space-y-2">
        <p>Comic ID: {params?.id || 'N/A'}</p>
        <p>Section Index: {sectionIndex || 'N/A'}</p>
      </div>
    </div>
  )
}

export default ComicReadMobile
