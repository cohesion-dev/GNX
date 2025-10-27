import { makeAutoObservable, runInAction } from 'mobx'
import { getSection, getComic, type SectionDetail, type Page, type PageDetail } from '@/apis'
import { AudioPlayer } from './AudioPlayer'
import { PageManager } from './PageManager'

export class ComicReadStore {
  comicId: string = ''
  sectionId: string = ''
  sectionData: SectionDetail | null = null
  comicTitle: string = ''
  
  isPlaying: boolean = false
  showOverlay: boolean = false
  currentPageIndex: number = 0
  currentAudioIndex: number = 0
  
  isLoadingImage: boolean = false
  isLoadingAudio: boolean = false
  
  error: string | null = null
  
  private audioPlayer: AudioPlayer
  private pageManager: PageManager
  private sections: Array<{ id: string; index: number }> = []

  constructor() {
    makeAutoObservable(this)
    this.audioPlayer = new AudioPlayer()
    this.pageManager = new PageManager()
    
    this.audioPlayer.onEnded = () => {
      this.nextAudio()
    }
    
    this.audioPlayer.onError = (error) => {
      console.error('Audio playback error:', error)
      this.nextAudio()
    }
  }

  get currentPage(): Page | null {
    if (!this.sectionData || this.currentPageIndex >= this.sectionData.pages.length) {
      return null
    }
    return this.sectionData.pages[this.currentPageIndex]
  }

  get currentPageDetail(): PageDetail | null {
    const page = this.currentPage
    if (!page || this.currentAudioIndex >= page.details.length) {
      return null
    }
    return page.details[this.currentAudioIndex]
  }

  get totalPages(): number {
    return this.sectionData?.pages.length || 0
  }

  get isLoading(): boolean {
    return this.isLoadingImage || this.isLoadingAudio
  }

  get currentChapter(): string {
    return this.sectionData ? `第${this.sectionData.index}章` : ''
  }

  get chapterTitle(): string {
    return this.sectionData?.title || ''
  }

  get currentPageNumber(): number {
    return this.currentPageIndex + 1
  }

  get currentImageUrl(): string | null {
    return this.pageManager.imageUrl
  }

  async initialize(comicId: string, sectionId: string): Promise<void> {
    this.comicId = comicId
    this.sectionId = sectionId
    
    try {
      await this.loadComicSections()
      await this.loadSectionData()
      await this.checkAndLoadResources()
      runInAction(() => {
        this.showOverlay = true
      })
    } catch (error) {
      runInAction(() => {
        this.error = error instanceof Error ? error.message : 'Initialization failed'
      })
    }
  }

  private async loadComicSections(): Promise<void> {
    try {
      const response = await getComic(this.comicId)
      if (response.code === 200 && response.data.sections) {
        runInAction(() => {
          this.sections = response.data.sections.map(s => ({
            id: s.id,
            index: s.index
          }))
          this.comicTitle = response.data.title
        })
      }
    } catch (error) {
      console.error('Failed to load comic sections:', error)
    }
  }

  private async loadSectionData(): Promise<void> {
    try {
      const response = await getSection(this.comicId, this.sectionId)
      if (response.code === 200) {
        runInAction(() => {
          this.sectionData = response.data
          this.currentPageIndex = 0
          this.currentAudioIndex = 0
        })
      } else {
        throw new Error('Failed to load section data')
      }
    } catch (error) {
      throw error
    }
  }

  private async checkAndLoadResources(): Promise<void> {
    const page = this.currentPage
    if (!page) return
    
    runInAction(() => {
      this.isLoadingImage = true
      this.isLoadingAudio = true
    })
    
    try {
      const audioIds = page.details.map(d => d.id)
      this.pageManager.initialize(page.id, audioIds)
      
      await this.pageManager.loadImage()
      
      runInAction(() => {
        this.isLoadingImage = false
      })
      
      await this.pageManager.loadAllAudios()
      
      runInAction(() => {
        this.isLoadingAudio = false
      })
    } catch (error) {
      runInAction(() => {
        this.isLoadingImage = false
        this.isLoadingAudio = false
        this.error = error instanceof Error ? error.message : 'Failed to load resources'
      })
      throw error
    }
  }

  async play(): Promise<void> {
    if (!this.currentPageDetail) return
    
    runInAction(() => {
      this.isPlaying = true
      this.showOverlay = false
    })
    
    try {
      await this.playCurrentAudio()
    } catch (error) {
      console.error('Failed to start playback:', error)
      runInAction(() => {
        this.isPlaying = false
        this.showOverlay = true
      })
    }
  }

  pause(): void {
    this.audioPlayer.pause()
    this.isPlaying = false
  }

  togglePlayPause(): void {
    if (this.isPlaying) {
      this.pause()
    } else {
      this.play()
    }
  }

  private async playCurrentAudio(): Promise<void> {
    const detail = this.currentPageDetail
    if (!detail) return
    
    const audioBlob = this.pageManager.getAudioBlob(detail.id)
    if (!audioBlob) {
      console.error('Audio blob not found for:', detail.id)
      this.nextAudio()
      return
    }
    
    try {
      await this.audioPlayer.play(audioBlob)
    } catch (error) {
      console.error('Failed to play audio:', error)
      this.nextAudio()
    }
  }

  nextAudio(): void {
    const page = this.currentPage
    if (!page) return
    
    if (this.currentAudioIndex < page.details.length - 1) {
      runInAction(() => {
        this.currentAudioIndex++
      })
      if (this.isPlaying) {
        this.playCurrentAudio()
      }
    } else {
      this.nextPage()
    }
  }

  async nextPage(): Promise<void> {
    if (!this.sectionData) return
    
    if (this.currentPageIndex < this.sectionData.pages.length - 1) {
      runInAction(() => {
        this.currentPageIndex++
        this.currentAudioIndex = 0
      })
      
      try {
        await this.checkAndLoadResources()
        if (this.isPlaying) {
          this.play()
        }
      } catch (error) {
        console.error('Failed to load next page:', error)
      }
    } else {
      await this.nextSection()
    }
  }

  async nextSection(): Promise<void> {
    if (!this.sectionData) return
    
    const currentIndex = this.sections.findIndex(s => s.id === this.sectionId)
    if (currentIndex === -1 || currentIndex >= this.sections.length - 1) {
      runInAction(() => {
        this.isPlaying = false
      })
      return
    }
    
    const nextSection = this.sections[currentIndex + 1]
    runInAction(() => {
      this.sectionId = nextSection.id
    })
    
    try {
      await this.loadSectionData()
      await this.checkAndLoadResources()
      if (this.isPlaying) {
        this.play()
      }
    } catch (error) {
      console.error('Failed to load next section:', error)
      runInAction(() => {
        this.isPlaying = false
      })
    }
  }

  handleScreenTap(): void {
    if (this.isPlaying) {
      this.pause()
      this.showOverlay = true
    }
  }

  async handlePlayButtonClick(): Promise<void> {
    if (!this.isPlaying) {
      if (this.audioPlayer.currentAudio) {
        try {
          await this.audioPlayer.resume()
          runInAction(() => {
            this.isPlaying = true
            this.showOverlay = false
          })
        } catch (error) {
          console.error('Failed to resume playback:', error)
        }
      } else {
        await this.play()
      }
    }
  }

  handleBackClick(): void {
  }

  dispose(): void {
    this.audioPlayer.dispose()
    this.pageManager.dispose()
  }
}
