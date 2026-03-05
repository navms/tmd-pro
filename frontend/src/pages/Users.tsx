import { useState, useEffect, useMemo } from 'react'
import { Search, Plus, Trash2, RefreshCw, AtSign, Check, X } from 'lucide-react'
import {
  GetAllScreenNames,
  AddScreenNames,
  DeleteScreenName,
} from '../wailsjs/go/app/App'
import type { ScreenNameItem } from '../wailsjs/go/app/App'

export default function UsersPage() {
  const [users, setUsers] = useState<ScreenNameItem[]>([])
  const [search, setSearch] = useState('')
  const [selected, setSelected] = useState<ScreenNameItem | null>(null)
  const [addInput, setAddInput] = useState('')
  const [loading, setLoading] = useState(false)
  const [addLoading, setAddLoading] = useState(false)
  const [toast, setToast] = useState<{ msg: string; type: 'ok' | 'err' } | null>(null)

  const filtered = useMemo(() => {
    if (!search.trim()) return users
    const q = search.toLowerCase()
    return users.filter((u) => u.name.toLowerCase().includes(q))
  }, [users, search])

  const load = async () => {
    setLoading(true)
    try {
      const data = await GetAllScreenNames()
      setUsers(data || [])
      setSelected(null)
    } catch (e: any) {
      showToast(String(e), 'err')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  const showToast = (msg: string, type: 'ok' | 'err') => {
    setToast({ msg, type })
    setTimeout(() => setToast(null), 3000)
  }

  const handleAdd = async () => {
    if (!addInput.trim()) return
    setAddLoading(true)
    try {
      const count = await AddScreenNames(addInput)
      if (count > 0) {
        setAddInput('')
        showToast(`成功添加 ${count} 个用户`, 'ok')
        await load()
      } else {
        showToast('未添加任何用户（可能已存在）', 'err')
      }
    } catch (e: any) {
      showToast(String(e), 'err')
    } finally {
      setAddLoading(false)
    }
  }

  const handleDelete = async () => {
    if (!selected) return
    try {
      await DeleteScreenName(selected.name)
      showToast(`已删除 @${selected.name}`, 'ok')
      setSelected(null)
      await load()
    } catch (e: any) {
      showToast(String(e), 'err')
    }
  }

  return (
    <div className="flex flex-col h-full">
      <PageHeader
        title="用户管理"
        subtitle={loading ? '加载中…' : `共 ${users.length} 个用户`}
      >
        <IconBtn icon={RefreshCw} onClick={load} loading={loading} title="刷新" />
      </PageHeader>

      {/* Toast */}
      {toast && (
        <div
          className="mx-5 mb-3 flex items-center gap-2 px-3 py-2 rounded-lg"
          style={{
            background: toast.type === 'ok' ? 'rgba(5,150,105,0.06)' : 'rgba(220,38,38,0.06)',
            border: `1px solid ${toast.type === 'ok' ? 'rgba(5,150,105,0.18)' : 'rgba(220,38,38,0.18)'}`,
            color: toast.type === 'ok' ? '#059669' : '#DC2626',
            fontSize: 12,
            flexShrink: 0,
          }}
        >
          {toast.type === 'ok' ? <Check size={12} /> : <X size={12} />}
          {toast.msg}
        </div>
      )}

      {/* Content */}
      <div className="flex flex-1 overflow-hidden px-5 pb-5 pt-4 gap-4">
        {/* Left: User list */}
        <div
          className="flex flex-col flex-1 overflow-hidden rounded-xl"
          style={{
            background: '#FFFFFF',
            border: '1px solid rgba(0,0,0,0.07)',
            boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
          }}
        >
          {/* Search bar */}
          <div
            className="flex items-center gap-2 px-3"
            style={{
              height: 44,
              borderBottom: '1px solid rgba(0,0,0,0.06)',
              flexShrink: 0,
            }}
          >
            <Search size={13} style={{ color: 'rgba(0,0,0,0.25)', flexShrink: 0 }} />
            <input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="搜索用户名..."
              className="flex-1 bg-transparent border-none outline-none"
              style={{ color: '#1C1C1E', fontSize: 13, padding: 0, boxShadow: 'none' }}
            />
            {search && (
              <button
                onClick={() => setSearch('')}
                className="p-0.5 rounded"
                style={{ color: 'rgba(0,0,0,0.3)', background: 'transparent' }}
              >
                <X size={11} />
              </button>
            )}
          </div>

          {/* List */}
          <div className="flex-1 overflow-y-auto">
            {filtered.length === 0 ? (
              <div
                className="flex flex-col items-center justify-center h-full gap-2"
                style={{ color: 'rgba(0,0,0,0.2)' }}
              >
                <AtSign size={28} strokeWidth={1} />
                <span style={{ fontSize: 12 }}>
                  {search ? '没有匹配的用户' : '暂无用户，请先添加'}
                </span>
              </div>
            ) : (
              filtered.map((u) => {
                const isSelected = selected?.id === u.id
                return (
                  <button
                    key={u.id}
                    onClick={() => setSelected(isSelected ? null : u)}
                    className="w-full flex items-center gap-2.5 px-4 text-left"
                    style={{
                      height: 40,
                      background: isSelected ? 'rgba(99,102,241,0.07)' : 'transparent',
                      color: isSelected ? '#6366F1' : '#3C3C43',
                      borderBottom: '1px solid rgba(0,0,0,0.04)',
                      fontSize: 13,
                      fontFamily: "'JetBrains Mono', monospace",
                      transition: 'all 0.1s',
                    }}
                    onMouseEnter={(e) => {
                      if (!isSelected)
                        (e.currentTarget as HTMLElement).style.background = 'rgba(0,0,0,0.03)'
                    }}
                    onMouseLeave={(e) => {
                      if (!isSelected)
                        (e.currentTarget as HTMLElement).style.background = 'transparent'
                    }}
                  >
                    <AtSign
                      size={11}
                      style={{ color: isSelected ? '#6366F1' : 'rgba(0,0,0,0.2)', flexShrink: 0 }}
                    />
                    <span className="truncate">{u.name}</span>
                  </button>
                )
              })
            )}
          </div>
        </div>

        {/* Right: Actions */}
        <div className="flex flex-col gap-3" style={{ width: 260, flexShrink: 0 }}>
          {/* Add users */}
          <div
            className="flex flex-col gap-3 p-4 rounded-xl"
            style={{
              background: '#FFFFFF',
              border: '1px solid rgba(0,0,0,0.07)',
              boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
            }}
          >
            <SectionLabel>添加用户</SectionLabel>
            <textarea
              value={addInput}
              onChange={(e) => setAddInput(e.target.value)}
              placeholder={'输入用户名\n支持逗号、换行分隔批量添加'}
              rows={5}
              style={{
                width: '100%',
                padding: '8px 10px',
                resize: 'none',
                fontFamily: "'JetBrains Mono', monospace",
                fontSize: 12,
                lineHeight: 1.6,
              }}
              onKeyDown={(e) => {
                if (e.key === 'Enter' && e.metaKey) handleAdd()
              }}
            />
            <ActionBtn
              icon={Plus}
              label="添加"
              onClick={handleAdd}
              loading={addLoading}
              disabled={!addInput.trim()}
              variant="primary"
            />
          </div>

          {/* Selected user actions */}
          <div
            className="flex flex-col gap-3 p-4 rounded-xl"
            style={{
              background: '#FFFFFF',
              border: '1px solid rgba(0,0,0,0.07)',
              boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
            }}
          >
            <SectionLabel>操作</SectionLabel>
            {selected ? (
              <>
                <div
                  className="flex items-center gap-2 px-3 py-2 rounded-lg"
                  style={{
                    background: 'rgba(99,102,241,0.06)',
                    border: '1px solid rgba(99,102,241,0.12)',
                    fontSize: 12,
                  }}
                >
                  <AtSign size={11} style={{ color: '#6366F1', flexShrink: 0 }} />
                  <span style={{ color: '#6366F1', fontFamily: 'monospace' }} className="truncate">
                    {selected.name}
                  </span>
                </div>
                <ActionBtn
                  icon={Trash2}
                  label="删除此用户"
                  onClick={handleDelete}
                  variant="danger"
                />
              </>
            ) : (
              <div style={{ fontSize: 12, color: 'rgba(0,0,0,0.25)', textAlign: 'center', padding: '8px 0' }}>
                点击列表中的用户以选择
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

// ─── Shared UI atoms ────────────────────────────────────────────────────────

export function PageHeader({
  title,
  subtitle,
  children,
}: {
  title: string
  subtitle?: string
  children?: React.ReactNode
}) {
  return (
    <div
      className="drag-region flex items-center justify-between px-5"
      style={{
        height: 73,
        borderBottom: '1px solid rgba(0,0,0,0.06)',
        flexShrink: 0,
        background: '#F5F5F7',
      }}
    >
      <div className="no-drag">
        <div style={{ fontWeight: 700, fontSize: 15, color: '#1C1C1E', letterSpacing: '-0.01em' }}>
          {title}
        </div>
        {subtitle && (
          <div style={{ fontSize: 11, color: 'rgba(0,0,0,0.35)', marginTop: 1 }}>
            {subtitle}
          </div>
        )}
      </div>
      {children && <div className="flex items-center gap-2 no-drag">{children}</div>}
    </div>
  )
}

export function SectionLabel({ children }: { children: React.ReactNode }) {
  return (
    <div style={{
      fontSize: 11,
      fontWeight: 600,
      color: 'rgba(0,0,0,0.35)',
      textTransform: 'uppercase',
      letterSpacing: '0.06em',
    }}>
      {children}
    </div>
  )
}

export function IconBtn({
  icon: Icon,
  onClick,
  loading,
  title,
}: {
  icon: React.ElementType
  onClick: () => void
  loading?: boolean
  title?: string
}) {
  return (
    <button
      onClick={onClick}
      title={title}
      disabled={loading}
      className="flex items-center justify-center rounded-lg"
      style={{
        width: 32,
        height: 32,
        background: 'rgba(0,0,0,0.04)',
        color: loading ? 'rgba(0,0,0,0.2)' : 'rgba(0,0,0,0.45)',
        border: '1px solid rgba(0,0,0,0.08)',
        cursor: loading ? 'default' : 'pointer',
        transition: 'all 0.12s',
      }}
      onMouseEnter={(e) => {
        if (!loading) (e.currentTarget as HTMLElement).style.background = 'rgba(0,0,0,0.07)'
      }}
      onMouseLeave={(e) => {
        if (!loading) (e.currentTarget as HTMLElement).style.background = 'rgba(0,0,0,0.04)'
      }}
    >
      <Icon size={13} className={loading ? 'animate-spin' : ''} />
    </button>
  )
}

export function ActionBtn({
  icon: Icon,
  label,
  onClick,
  loading,
  disabled,
  variant = 'default',
}: {
  icon: React.ElementType
  label: string
  onClick: () => void
  loading?: boolean
  disabled?: boolean
  variant?: 'default' | 'primary' | 'danger'
}) {
  const styles = {
    primary: {
      background: disabled ? 'rgba(99,102,241,0.3)' : '#6366F1',
      color: disabled ? 'rgba(255,255,255,0.5)' : '#fff',
      border: 'none',
    },
    danger: {
      background: 'rgba(220,38,38,0.06)',
      color: '#DC2626',
      border: '1px solid rgba(220,38,38,0.15)',
    },
    default: {
      background: 'rgba(0,0,0,0.04)',
      color: 'rgba(0,0,0,0.55)',
      border: '1px solid rgba(0,0,0,0.08)',
    },
  }
  const s = styles[variant]

  return (
    <button
      onClick={onClick}
      disabled={loading || disabled}
      className="flex items-center justify-center gap-2 w-full py-2"
      style={{
        ...s,
        opacity: disabled && variant !== 'primary' ? 0.4 : 1,
        cursor: disabled || loading ? 'default' : 'pointer',
        fontWeight: 500,
        borderRadius: 8,
        transition: 'all 0.12s',
        fontSize: 13,
      }}
    >
      <Icon size={13} />
      {loading ? '处理中…' : label}
    </button>
  )
}
