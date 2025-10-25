'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { observer } from 'mobx-react-lite'

const Home = observer(() => {
  const router = useRouter()

  useEffect(() => {
    router.push('/comic')
  }, [router])

  return null
})

export default Home
