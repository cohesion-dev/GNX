'use client'

import dynamic from 'next/dynamic'
import { observer } from 'mobx-react-lite'
import Layout from '@/components/layout'

const Pc = dynamic(() => import('./pc'), { ssr: false })
const Mobile = dynamic(() => import('./mobile'), { ssr: false })

interface ComicDetailProps {
  params: {
    id: string
  }
}

const ComicDetail = observer(({ params }: ComicDetailProps) => {
  return <Layout pc={<Pc params={params} />} mobile={<Mobile params={params} />} />
})

export default ComicDetail
