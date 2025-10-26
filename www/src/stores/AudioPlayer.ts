import { makeAutoObservable } from 'mobx'

export class AudioPlayer {
  currentAudio: HTMLAudioElement | null = null
  isPlaying: boolean = false
  duration: number = 0
  currentTime: number = 0
  onEnded: (() => void) | null = null
  onError: ((error: Error) => void) | null = null

  constructor() {
    makeAutoObservable(this)
  }

  async play(audioBlob: Blob): Promise<void> {
    // 先停止并清理之前的音频
    this.stop()

    const audioUrl = URL.createObjectURL(audioBlob)
    const audio = new Audio(audioUrl)

    // 使用箭头函数确保 this 指向正确，并在事件触发时检查是否是当前音频
    audio.onended = () => {
      // 只有当这个音频仍然是当前音频时才触发回调
      if (this.currentAudio === audio) {
        URL.revokeObjectURL(audioUrl)
        this.isPlaying = false
        if (this.onEnded) {
          this.onEnded()
        }
      } else {
        // 如果不是当前音频，只清理 URL
        URL.revokeObjectURL(audioUrl)
      }
    }

    audio.onerror = (e) => {
      // 只有当这个音频仍然是当前音频时才触发错误回调
      if (this.currentAudio === audio) {
        URL.revokeObjectURL(audioUrl)
        this.isPlaying = false
        if (this.onError) {
          this.onError(new Error('Audio playback failed'))
        }
      } else {
        URL.revokeObjectURL(audioUrl)
      }
    }

    audio.ontimeupdate = () => {
      // 只有当这个音频仍然是当前音频时才更新时间
      if (this.currentAudio === audio) {
        this.currentTime = audio.currentTime
      }
    }

    audio.onloadedmetadata = () => {
      // 只有当这个音频仍然是当前音频时才更新时长
      if (this.currentAudio === audio) {
        this.duration = audio.duration
      }
    }

    this.currentAudio = audio

    try {
      await audio.play()
      this.isPlaying = true
    } catch (error) {
      URL.revokeObjectURL(audioUrl)
      throw error
    }
  }

  pause(): void {
    if (this.currentAudio) {
      this.currentAudio.pause()
      this.isPlaying = false
    }
  }

  async resume(): Promise<void> {
    if (this.currentAudio) {
      try {
        await this.currentAudio.play()
        this.isPlaying = true
      } catch (error) {
        this.isPlaying = false
        throw error
      }
    }
  }

  stop(): void {
    if (this.currentAudio) {
      this.currentAudio.pause()

      // 显式移除所有事件监听器，防止在清理后还触发
      this.currentAudio.onended = null
      this.currentAudio.onerror = null
      this.currentAudio.ontimeupdate = null
      this.currentAudio.onloadedmetadata = null

      this.currentAudio.src = ''
      this.currentAudio = null
      this.isPlaying = false
      this.currentTime = 0
      this.duration = 0
    }
  }

  dispose(): void {
    this.stop()
    this.onEnded = null
    this.onError = null
  }
}
