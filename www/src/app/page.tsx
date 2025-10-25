'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { getComics } from '@/apis/comic'

export default function Home() {
  const router = useRouter()
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const loadComics = async () => {
      try {
        const response = await getComics({ page: 1, limit: 1 })
        if (response.code === 200 && response.data) {
          if (response.data.comics.length === 0) {
            router.push('/comic/add')
          } else {
            router.push('/comic/')
          }
        }
      } catch (error) {
        console.error('Failed to load comics:', error)
        router.push('/comic/')
      } finally {
        setLoading(false)
      }
    }

    loadComics()
  }, [router])

  return null
}
