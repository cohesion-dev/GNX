import { makeAutoObservable } from 'mobx'

export class AudioPlayer {
  currentAudio: HTMLAudioElement | null = null
  private currentAudioUrl: string | null = null
  isPlaying: boolean = false
  duration: number = 0
  currentTime: number = 0
  onEnded: (() => void) | null = null
  onError: ((error: Error) => void) | null = null

  constructor() {
    makeAutoObservable(this)
  }

  async play(audioBlob: Blob): Promise<void> {
    this.stop()

    const audioUrl = URL.createObjectURL(audioBlob)
    const audio = new Audio(audioUrl)

    this.currentAudioUrl = audioUrl

    audio.onended = () => {
      URL.revokeObjectURL(audioUrl)
      if (this.currentAudioUrl === audioUrl) {
        this.currentAudioUrl = null
      }
      this.isPlaying = false
      this.currentAudio = null
      if (this.onEnded) {
        this.onEnded()
      }
    }

    audio.onerror = (e) => {
      URL.revokeObjectURL(audioUrl)
      if (this.currentAudioUrl === audioUrl) {
        this.currentAudioUrl = null
      }
      this.isPlaying = false
      if (this.onError) {
        this.onError(new Error('Audio playback failed'))
      }
    }

    audio.ontimeupdate = () => {
      this.currentTime = audio.currentTime
    }

    audio.onloadedmetadata = () => {
      this.duration = audio.duration
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
      const audio = this.currentAudio
      const audioUrl = this.currentAudioUrl

      audio.onended = null
      audio.onerror = null
      audio.ontimeupdate = null
      audio.onloadedmetadata = null

      if (audioUrl) {
        URL.revokeObjectURL(audioUrl)
        this.currentAudioUrl = null
      }

      audio.pause()
      audio.currentTime = 0
      audio.src = ''

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
