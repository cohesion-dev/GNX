'use client'

import dynamic from 'next/dynamic'
import { observer } from 'mobx-react-lite'
import Layout from '@/components/layout'

const Pc = dynamic(() => import('./pc'), { ssr: false })
const Mobile = dynamic(() => import('./mobile'), { ssr: false })

const ComicRead = observer(() => {
  return <Layout pc={<Pc />} mobile={<Mobile />} />
})

export default ComicRead
