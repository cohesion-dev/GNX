'use client'

import { observer } from 'mobx-react-lite'
import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { getComics, Comic } from '@/apis/comic'

const ComicListPC = observer(() => {
  const router = useRouter()
  const [comics, setComics] = useState<Comic[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadComics()
  }, [])

  const loadComics = async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await getComics()
      if (response.code === 200) {
        setComics(response.data.comics)
      } else {
        setError(response.message || '获取漫画列表失败')
      }
    } catch (err) {
      setError('获取漫画列表失败，请稍后重试')
      console.error('Failed to load comics:', err)
    } finally {
      setLoading(false)
    }
  }

  const getStatusText = (status: Comic['status']) => {
    switch (status) {
      case 'completed':
        return '完成'
      case 'pending':
        return '生成中'
      case 'failed':
        return '失败'
      default:
        return '未知'
    }
  }

  const getStatusColor = (status: Comic['status']) => {
    switch (status) {
      case 'completed':
        return '#52c41a'
      case 'pending':
        return '#1890ff'
      case 'failed':
        return '#ff4d4f'
      default:
        return '#999'
    }
  }

  const handleComicClick = (id: number) => {
    router.push(`/comic/detail/${id}`)
  }

  if (loading) {
    return (
      <div style={styles.container}>
        <div style={styles.loadingText}>加载中...</div>
      </div>
    )
  }

  if (error) {
    return (
      <div style={styles.container}>
        <div style={styles.errorText}>{error}</div>
        <button onClick={loadComics} style={styles.retryButton}>
          重试
        </button>
      </div>
    )
  }

  return (
    <div style={styles.container}>
      <div style={styles.header}>
        <h1 style={styles.title}>漫画列表</h1>
        <button
          onClick={() => router.push('/comic/add')}
          style={styles.addButton}
        >
          创建漫画
        </button>
      </div>

      {comics.length === 0 ? (
        <div style={styles.emptyState}>
          <div style={styles.emptyText}>暂无漫画</div>
          <button
            onClick={() => router.push('/comic/add')}
            style={styles.createButton}
          >
            创建第一个漫画
          </button>
        </div>
      ) : (
        <div style={styles.comicGrid}>
          {comics.map((comic) => (
            <div
              key={comic.id}
              style={styles.comicCard}
              onClick={() => handleComicClick(comic.id)}
            >
              <div style={styles.coverContainer}>
                {comic.icon ? (
                  <img
                    src={comic.icon}
                    alt={comic.title}
                    style={styles.cover}
                  />
                ) : (
                  <div style={styles.coverPlaceholder}>
                    <span style={styles.placeholderText}>无封面</span>
                  </div>
                )}
                <div
                  style={{
                    ...styles.statusBadge,
                    backgroundColor: getStatusColor(comic.status)
                  }}
                >
                  {getStatusText(comic.status)}
                </div>
              </div>
              <div style={styles.comicInfo}>
                <h3 style={styles.comicTitle}>{comic.title}</h3>
                {comic.brief && (
                  <p style={styles.comicBrief}>{comic.brief}</p>
                )}
                <div style={styles.comicMeta}>
                  <span style={styles.metaText}>
                    {new Date(comic.created_at).toLocaleDateString()}
                  </span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
})

const styles: { [key: string]: React.CSSProperties } = {
  container: {
    padding: '40px',
    maxWidth: '1400px',
    margin: '0 auto',
    minHeight: '100vh',
    backgroundColor: '#f5f5f5',
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '32px',
  },
  title: {
    fontSize: '32px',
    fontWeight: 'bold',
    margin: 0,
    color: '#333',
  },
  addButton: {
    padding: '12px 24px',
    fontSize: '16px',
    backgroundColor: '#1890ff',
    color: 'white',
    border: 'none',
    borderRadius: '8px',
    cursor: 'pointer',
    fontWeight: '500',
    transition: 'background-color 0.3s',
  },
  loadingText: {
    textAlign: 'center',
    fontSize: '18px',
    color: '#666',
    padding: '60px 0',
  },
  errorText: {
    textAlign: 'center',
    fontSize: '18px',
    color: '#ff4d4f',
    marginBottom: '20px',
  },
  retryButton: {
    padding: '12px 24px',
    fontSize: '16px',
    backgroundColor: '#1890ff',
    color: 'white',
    border: 'none',
    borderRadius: '8px',
    cursor: 'pointer',
    display: 'block',
    margin: '0 auto',
  },
  emptyState: {
    textAlign: 'center',
    padding: '100px 0',
  },
  emptyText: {
    fontSize: '18px',
    color: '#999',
    marginBottom: '24px',
  },
  createButton: {
    padding: '14px 32px',
    fontSize: '16px',
    backgroundColor: '#1890ff',
    color: 'white',
    border: 'none',
    borderRadius: '8px',
    cursor: 'pointer',
    fontWeight: '500',
  },
  comicGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
    gap: '24px',
  },
  comicCard: {
    backgroundColor: 'white',
    borderRadius: '12px',
    overflow: 'hidden',
    cursor: 'pointer',
    transition: 'all 0.3s',
    boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
  },
  coverContainer: {
    position: 'relative',
    width: '100%',
    height: '320px',
    overflow: 'hidden',
    backgroundColor: '#f0f0f0',
  },
  cover: {
    width: '100%',
    height: '100%',
    objectFit: 'cover',
  },
  coverPlaceholder: {
    width: '100%',
    height: '100%',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: '#f0f0f0',
  },
  placeholderText: {
    color: '#999',
    fontSize: '16px',
  },
  statusBadge: {
    position: 'absolute',
    top: '12px',
    right: '12px',
    padding: '6px 12px',
    borderRadius: '20px',
    color: 'white',
    fontSize: '14px',
    fontWeight: '500',
    boxShadow: '0 2px 4px rgba(0,0,0,0.2)',
  },
  comicInfo: {
    padding: '16px',
  },
  comicTitle: {
    fontSize: '18px',
    fontWeight: '600',
    margin: '0 0 8px 0',
    color: '#333',
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap',
  },
  comicBrief: {
    fontSize: '14px',
    color: '#666',
    margin: '0 0 12px 0',
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    display: '-webkit-box',
    WebkitLineClamp: 2,
    WebkitBoxOrient: 'vertical',
    lineHeight: '1.5',
  },
  comicMeta: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  metaText: {
    fontSize: '12px',
    color: '#999',
  },
}

export default ComicListPC
