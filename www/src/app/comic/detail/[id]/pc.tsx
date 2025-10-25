'use client'

import { observer } from 'mobx-react-lite'
import { useEffect, useState } from 'react'
import { useRouter, useParams } from 'next/navigation'
import { getComic, ComicDetail, ComicRole, ComicSection } from '@/apis/comic'

const ComicDetailPC = observer(() => {
  const router = useRouter()
  const params = useParams()
  const comicId = Number(params.id)

  const [comicDetail, setComicDetail] = useState<ComicDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (comicId) {
      loadComicDetail()
    }
  }, [comicId])

  const loadComicDetail = async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await getComic(comicId)
      if (response.code === 200) {
        setComicDetail(response.data)
      } else {
        setError(response.message || '获取漫画详情失败')
      }
    } catch (err) {
      setError('获取漫画详情失败，请稍后重试')
      console.error('Failed to load comic detail:', err)
    } finally {
      setLoading(false)
    }
  }

  const getStatusText = (status: ComicSection['status']) => {
    switch (status) {
      case 'completed':
        return '已完成'
      case 'pending':
        return '生成中'
      case 'failed':
        return '失败'
      default:
        return '未知'
    }
  }

  const getStatusColor = (status: ComicSection['status']) => {
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

  const handleSectionClick = (sectionId: number) => {
    router.push(`/comic/read/${sectionId}`)
  }

  const handleAddSection = () => {
    router.push(`/comic/detail/${comicId}/section/add`)
  }

  if (loading) {
    return (
      <div style={styles.container}>
        <div style={styles.loadingText}>加载中...</div>
      </div>
    )
  }

  if (error || !comicDetail) {
    return (
      <div style={styles.container}>
        <div style={styles.errorText}>{error || '漫画不存在'}</div>
        <button onClick={loadComicDetail} style={styles.retryButton}>
          重试
        </button>
      </div>
    )
  }

  return (
    <div style={styles.container}>
      {/* 返回按钮 */}
      <button onClick={() => router.push('/comic')} style={styles.backButton}>
        ← 返回列表
      </button>

      {/* 漫画概览模块 */}
      <div style={styles.overviewSection}>
        <div style={styles.coverWrapper}>
          {comicDetail.icon ? (
            <img
              src={comicDetail.icon}
              alt={comicDetail.title}
              style={styles.comicCover}
            />
          ) : (
            <div style={styles.coverPlaceholder}>
              <span style={styles.placeholderText}>无封面</span>
            </div>
          )}
        </div>
        <div style={styles.comicInfo}>
          <h1 style={styles.comicTitle}>{comicDetail.title}</h1>
          {comicDetail.brief && (
            <div style={styles.briefSection}>
              <h3 style={styles.sectionSubtitle}>简介</h3>
              <p style={styles.comicBrief}>{comicDetail.brief}</p>
            </div>
          )}
          {comicDetail.user_prompt && (
            <div style={styles.promptSection}>
              <h3 style={styles.sectionSubtitle}>创作提示</h3>
              <p style={styles.userPrompt}>{comicDetail.user_prompt}</p>
            </div>
          )}
          <div style={styles.metaInfo}>
            <span style={styles.metaItem}>
              创建时间：{new Date(comicDetail.created_at).toLocaleString()}
            </span>
            <span style={styles.metaItem}>
              更新时间：{new Date(comicDetail.updated_at).toLocaleString()}
            </span>
          </div>
        </div>
      </div>

      {/* 漫画角色列表模块 */}
      <div style={styles.section}>
        <h2 style={styles.sectionTitle}>角色列表</h2>
        {comicDetail.roles && comicDetail.roles.length > 0 ? (
          <div style={styles.rolesGrid}>
            {comicDetail.roles.map((role: ComicRole) => (
              <div key={role.id} style={styles.roleCard}>
                <div style={styles.roleAvatarWrapper}>
                  {role.image_url ? (
                    <img
                      src={role.image_url}
                      alt={role.name}
                      style={styles.roleAvatar}
                    />
                  ) : (
                    <div style={styles.avatarPlaceholder}>
                      <span style={styles.avatarInitial}>
                        {role.name.charAt(0)}
                      </span>
                    </div>
                  )}
                </div>
                <div style={styles.roleInfo}>
                  <h3 style={styles.roleName}>{role.name}</h3>
                  {role.brief && (
                    <p style={styles.roleBrief}>{role.brief}</p>
                  )}
                  {role.voice && (
                    <div style={styles.roleVoice}>
                      <span style={styles.voiceLabel}>配音：</span>
                      <span style={styles.voiceValue}>{role.voice}</span>
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div style={styles.emptyState}>
            <span style={styles.emptyText}>暂无角色</span>
          </div>
        )}
      </div>

      {/* 漫画章节列表模块 */}
      <div style={styles.section}>
        <div style={styles.sectionHeader}>
          <h2 style={styles.sectionTitle}>章节列表</h2>
          <button onClick={handleAddSection} style={styles.addButton}>
            + 添加章节
          </button>
        </div>
        {comicDetail.sections && comicDetail.sections.length > 0 ? (
          <div style={styles.sectionsList}>
            {comicDetail.sections
              .sort((a, b) => a.index - b.index)
              .map((section: ComicSection) => (
                <div
                  key={section.id}
                  style={styles.sectionItem}
                  onClick={() => section.status === 'completed' && handleSectionClick(section.id)}
                >
                  <div style={styles.sectionItemContent}>
                    <div style={styles.sectionIndex}>
                      第 {section.index} 章
                    </div>
                    <div style={styles.sectionDetail}>
                      <p style={styles.sectionDetailText}>{section.detail}</p>
                      <span style={styles.sectionMeta}>
                        创建时间：{new Date(section.created_at).toLocaleString()}
                      </span>
                    </div>
                  </div>
                  <div style={styles.sectionRight}>
                    <div
                      style={{
                        ...styles.sectionStatus,
                        color: getStatusColor(section.status)
                      }}
                    >
                      {getStatusText(section.status)}
                    </div>
                    {section.status === 'completed' && (
                      <span style={styles.readLink}>阅读 →</span>
                    )}
                  </div>
                </div>
              ))}
          </div>
        ) : (
          <div style={styles.emptyState}>
            <span style={styles.emptyText}>暂无章节</span>
            <button onClick={handleAddSection} style={styles.createButton}>
              创建第一个章节
            </button>
          </div>
        )}
      </div>
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
  backButton: {
    padding: '10px 20px',
    fontSize: '14px',
    backgroundColor: 'white',
    color: '#666',
    border: '1px solid #d9d9d9',
    borderRadius: '8px',
    cursor: 'pointer',
    marginBottom: '24px',
    transition: 'all 0.3s',
  },
  overviewSection: {
    display: 'flex',
    gap: '32px',
    backgroundColor: 'white',
    borderRadius: '12px',
    padding: '32px',
    marginBottom: '32px',
    boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
  },
  coverWrapper: {
    flexShrink: 0,
    width: '300px',
    height: '400px',
    borderRadius: '8px',
    overflow: 'hidden',
    backgroundColor: '#f0f0f0',
  },
  comicCover: {
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
  comicInfo: {
    flex: 1,
    display: 'flex',
    flexDirection: 'column',
  },
  comicTitle: {
    fontSize: '36px',
    fontWeight: 'bold',
    margin: '0 0 24px 0',
    color: '#333',
  },
  briefSection: {
    marginBottom: '24px',
  },
  sectionSubtitle: {
    fontSize: '18px',
    fontWeight: '600',
    margin: '0 0 12px 0',
    color: '#666',
  },
  comicBrief: {
    fontSize: '16px',
    lineHeight: '1.8',
    color: '#666',
    margin: 0,
  },
  promptSection: {
    marginBottom: '24px',
  },
  userPrompt: {
    fontSize: '14px',
    lineHeight: '1.6',
    color: '#999',
    margin: 0,
    padding: '12px',
    backgroundColor: '#f9f9f9',
    borderRadius: '6px',
    borderLeft: '3px solid #1890ff',
  },
  metaInfo: {
    display: 'flex',
    flexDirection: 'column',
    gap: '8px',
    marginTop: 'auto',
  },
  metaItem: {
    fontSize: '14px',
    color: '#999',
  },
  section: {
    backgroundColor: 'white',
    borderRadius: '12px',
    padding: '32px',
    marginBottom: '32px',
    boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
  },
  sectionHeader: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '24px',
  },
  sectionTitle: {
    fontSize: '24px',
    fontWeight: 'bold',
    margin: '0 0 24px 0',
    color: '#333',
  },
  addButton: {
    padding: '10px 20px',
    fontSize: '14px',
    backgroundColor: '#1890ff',
    color: 'white',
    border: 'none',
    borderRadius: '8px',
    cursor: 'pointer',
    fontWeight: '500',
    transition: 'background-color 0.3s',
  },
  rolesGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
    gap: '20px',
  },
  roleCard: {
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    padding: '24px',
    border: '1px solid #e8e8e8',
    borderRadius: '8px',
    transition: 'all 0.3s',
    backgroundColor: '#fafafa',
  },
  roleAvatarWrapper: {
    width: '120px',
    height: '120px',
    borderRadius: '50%',
    overflow: 'hidden',
    marginBottom: '16px',
    backgroundColor: '#f0f0f0',
  },
  roleAvatar: {
    width: '100%',
    height: '100%',
    objectFit: 'cover',
  },
  avatarPlaceholder: {
    width: '100%',
    height: '100%',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: '#1890ff',
  },
  avatarInitial: {
    fontSize: '48px',
    fontWeight: 'bold',
    color: 'white',
  },
  roleInfo: {
    textAlign: 'center',
    width: '100%',
  },
  roleName: {
    fontSize: '20px',
    fontWeight: '600',
    margin: '0 0 12px 0',
    color: '#333',
  },
  roleBrief: {
    fontSize: '14px',
    lineHeight: '1.6',
    color: '#666',
    margin: '0 0 12px 0',
  },
  roleVoice: {
    fontSize: '13px',
    color: '#999',
  },
  voiceLabel: {
    fontWeight: '500',
  },
  voiceValue: {
    marginLeft: '4px',
  },
  sectionsList: {
    display: 'flex',
    flexDirection: 'column',
    gap: '16px',
  },
  sectionItem: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: '20px',
    border: '1px solid #e8e8e8',
    borderRadius: '8px',
    transition: 'all 0.3s',
    cursor: 'pointer',
    backgroundColor: '#fafafa',
  },
  sectionItemContent: {
    display: 'flex',
    gap: '20px',
    flex: 1,
  },
  sectionIndex: {
    fontSize: '18px',
    fontWeight: '600',
    color: '#1890ff',
    minWidth: '80px',
  },
  sectionDetail: {
    flex: 1,
  },
  sectionDetailText: {
    fontSize: '15px',
    color: '#333',
    margin: '0 0 8px 0',
    lineHeight: '1.5',
  },
  sectionMeta: {
    fontSize: '12px',
    color: '#999',
  },
  sectionRight: {
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'flex-end',
    gap: '8px',
  },
  sectionStatus: {
    fontSize: '14px',
    fontWeight: '500',
  },
  readLink: {
    fontSize: '14px',
    color: '#1890ff',
    fontWeight: '500',
  },
  emptyState: {
    textAlign: 'center',
    padding: '60px 0',
  },
  emptyText: {
    fontSize: '16px',
    color: '#999',
    display: 'block',
    marginBottom: '16px',
  },
  createButton: {
    padding: '12px 24px',
    fontSize: '14px',
    backgroundColor: '#1890ff',
    color: 'white',
    border: 'none',
    borderRadius: '8px',
    cursor: 'pointer',
    fontWeight: '500',
  },
}

export default ComicDetailPC
