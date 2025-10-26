import { useEffect, useRef } from 'react'

export function usePolling(
  callback: () => void | Promise<void>,
  shouldPoll: boolean,
  interval: number = 3000
) {
  const savedCallback = useRef(callback)
  const timeoutId = useRef<NodeJS.Timeout | null>(null)

  useEffect(() => {
    savedCallback.current = callback
  }, [callback])

  useEffect(() => {
    if (!shouldPoll) {
      if (timeoutId.current) {
        clearTimeout(timeoutId.current)
        timeoutId.current = null
      }
      return
    }

    const tick = async () => {
      await savedCallback.current()
      if (shouldPoll) {
        timeoutId.current = setTimeout(tick, interval)
      }
    }

    timeoutId.current = setTimeout(tick, interval)

    return () => {
      if (timeoutId.current) {
        clearTimeout(timeoutId.current)
      }
    }
  }, [shouldPoll, interval])
}
