'use client'

import dynamic from 'next/dynamic'
import { observer } from 'mobx-react-lite'

const Pc = dynamic(() => import('./pc'), { ssr: false })

const ComicRead = observer(() => {
  return (
    <Pc />
  )
})

export default ComicRead
