import { makeAutoObservable } from 'mobx'
import { getImageUrl, getTTSAudio } from '@/apis'

interface AudioData {
  id: string
  blob: Blob | null
  loading: boolean
  error: string | null
}

export class PageManager {
  pageId: string = ''
  imageUrl: string | null = null
  imageLoading: boolean = false
  imageError: string | null = null
  audioMap: Map<string, AudioData> = new Map()
  
  private imagePollingTimer: NodeJS.Timeout | null = null
  private imagePollingAttempts: number = 0
  private readonly MAX_POLLING_ATTEMPTS = 30
  private readonly POLLING_INTERVAL = 2000

  constructor() {
    makeAutoObservable(this)
  }

  get isImageReady(): boolean {
    return this.imageUrl !== null
  }

  get areAllAudiosReady(): boolean {
    if (this.audioMap.size === 0) return false
    for (const audio of this.audioMap.values()) {
      if (audio.blob === null) return false
    }
    return true
  }

  initialize(pageId: string, audioIds: string[]): void {
    this.pageId = pageId
    this.imageUrl = null
    this.imageError = null
    this.audioMap.clear()
    
    audioIds.forEach(id => {
      this.audioMap.set(id, {
        id,
        blob: null,
        loading: false,
        error: null
      })
    })
  }

  async loadImage(): Promise<void> {
    this.imageLoading = true
    this.imageError = null
    this.imagePollingAttempts = 0
    
    try {
      const url = await this.pollImageUrl()
      this.imageUrl = url
    } catch (error) {
      this.imageError = error instanceof Error ? error.message : 'Failed to load image'
      throw error
    } finally {
      this.imageLoading = false
    }
  }

  private async pollImageUrl(): Promise<string> {
    return new Promise((resolve, reject) => {
      const poll = async () => {
        try {
          this.imagePollingAttempts++
          
          const response = await getImageUrl(this.pageId)
          
          if (response.code === 200 && response.data?.url) {
            if (this.imagePollingTimer) {
              clearTimeout(this.imagePollingTimer)
              this.imagePollingTimer = null
            }
            resolve(response.data.url)
            return
          }
          
          if (this.imagePollingAttempts >= this.MAX_POLLING_ATTEMPTS) {
            if (this.imagePollingTimer) {
              clearTimeout(this.imagePollingTimer)
              this.imagePollingTimer = null
            }
            reject(new Error('Image polling timeout'))
            return
          }
          
          this.imagePollingTimer = setTimeout(poll, this.POLLING_INTERVAL)
        } catch (error) {
          if (this.imagePollingAttempts >= this.MAX_POLLING_ATTEMPTS) {
            if (this.imagePollingTimer) {
              clearTimeout(this.imagePollingTimer)
              this.imagePollingTimer = null
            }
            reject(error)
            return
          }
          
          this.imagePollingTimer = setTimeout(poll, this.POLLING_INTERVAL)
        }
      }
      
      poll()
    })
  }

  async loadAudio(audioId: string, retryCount = 3): Promise<void> {
    const audioData = this.audioMap.get(audioId)
    if (!audioData) return
    
    audioData.loading = true
    audioData.error = null
    
    let lastError: Error | null = null
    
    for (let attempt = 0; attempt < retryCount; attempt++) {
      try {
        const blob = await getTTSAudio(audioId)
        audioData.blob = blob
        audioData.loading = false
        audioData.error = null
        return
      } catch (error) {
        lastError = error instanceof Error ? error : new Error('Failed to load audio')
        if (attempt < retryCount - 1) {
          await new Promise(resolve => setTimeout(resolve, 1000 * (attempt + 1)))
        }
      }
    }
    
    audioData.loading = false
    audioData.error = lastError?.message || 'Failed to load audio'
    throw lastError
  }

  async loadAllAudios(): Promise<void> {
    const promises = Array.from(this.audioMap.keys()).map(id => this.loadAudio(id))
    await Promise.all(promises)
  }

  getAudioBlob(audioId: string): Blob | null {
    return this.audioMap.get(audioId)?.blob || null
  }

  dispose(): void {
    if (this.imagePollingTimer) {
      clearTimeout(this.imagePollingTimer)
      this.imagePollingTimer = null
    }
    this.audioMap.clear()
  }
}
