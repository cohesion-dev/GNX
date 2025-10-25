'use client'

import { ReactNode, useEffect, useState } from 'react'

interface ResponsiveLayoutProps {
  pc: ReactNode
  mobile: ReactNode
}

export default function ResponsiveLayout({ pc, mobile }: ResponsiveLayoutProps) {
  const [isMobile, setIsMobile] = useState(false)

  useEffect(() => {
    const mediaQuery = window.matchMedia('(max-width: 767px)')
    
    const handleChange = (e: MediaQueryListEvent | MediaQueryList) => {
      setIsMobile(e.matches)
    }

    handleChange(mediaQuery)
    
    mediaQuery.addEventListener('change', handleChange)

    return () => {
      mediaQuery.removeEventListener('change', handleChange)
    }
  }, [])

  return <>{isMobile ? mobile : pc}</>
}
