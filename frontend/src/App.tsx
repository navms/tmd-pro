import { useState, useEffect } from 'react'
import { Users, Activity, Settings, AlertCircle } from 'lucide-react'
import UsersPage from './pages/Users'
import ScannerPage from './pages/Scanner'
import SettingsPage from './pages/Settings'
import { GetInitError } from './wailsjs/go/app/App'

type Page = 'users' | 'scanner' | 'settings'

const navItems = [
  { id: 'users' as Page, icon: Users, label: '用户管理' },
  { id: 'scanner' as Page, icon: Activity, label: '扫描控制' },
  { id: 'settings' as Page, icon: Settings, label: '系统配置' },
]

export default function App() {
  const [page, setPage] = useState<Page>('users')
  const [initError, setInitError] = useState('')

  useEffect(() => {
    GetInitError().then((err) => {
      if (err) setInitError(err)
    }).catch(() => {})
  }, [])

  return (
    <div className="flex h-full" style={{ background: '#F5F5F7' }}>
      {/* Sidebar */}
      <aside
        className="flex flex-col"
        style={{
          width: 216,
          minWidth: 216,
          background: '#FFFFFF',
          borderRight: '1px solid rgba(0,0,0,0.07)',
          boxShadow: '1px 0 0 rgba(0,0,0,0.04)',
        }}
      >
        
        {/* App branding */}
        <div className="px-5 pb-4 pt-4 no-drag">
          <div
            style={{
              fontWeight: 700,
              fontSize: 15,
              color: '#1C1C1E',
              letterSpacing: '-0.02em',
            }}
          >
            TMD Pro
          </div>
          <div style={{ fontSize: 11, color: 'rgba(0,0,0,0.35)', marginTop: 2 }}>
            Screen Name 管理工具
          </div>
        </div>

        {/* Divider */}
        <div style={{ height: 1, background: 'rgba(0,0,0,0.06)', margin: '0 16px 8px' }} />

        {/* Navigation */}
        <nav className="flex-1 px-2.5 py-1 no-drag space-y-0.5">
          {navItems.map(({ id, icon: Icon, label }) => {
            const active = page === id
            return (
              <button
                key={id}
                onClick={() => setPage(id)}
                className="w-full flex items-center gap-2.5 px-3 py-2 rounded-lg text-left"
                style={{
                  background: active ? 'rgba(99,102,241,0.08)' : 'transparent',
                  color: active ? '#6366F1' : 'rgba(0,0,0,0.45)',
                  fontWeight: active ? 600 : 400,
                  fontSize: 13,
                  position: 'relative',
                  transition: 'all 0.12s',
                }}
                onMouseEnter={(e) => {
                  if (!active) {
                    (e.currentTarget as HTMLElement).style.background = 'rgba(0,0,0,0.04)'
                    ;(e.currentTarget as HTMLElement).style.color = 'rgba(0,0,0,0.7)'
                  }
                }}
                onMouseLeave={(e) => {
                  if (!active) {
                    ;(e.currentTarget as HTMLElement).style.background = 'transparent'
                    ;(e.currentTarget as HTMLElement).style.color = 'rgba(0,0,0,0.45)'
                  }
                }}
              >
                {active && (
                  <div
                    style={{
                      position: 'absolute',
                      left: 0,
                      top: '50%',
                      transform: 'translateY(-50%)',
                      width: 2.5,
                      height: 14,
                      background: '#6366F1',
                      borderRadius: '0 2px 2px 0',
                    }}
                  />
                )}
                <Icon
                  size={14}
                  strokeWidth={active ? 2.2 : 1.75}
                  style={{ flexShrink: 0, color: active ? '#6366F1' : 'rgba(0,0,0,0.4)' }}
                />
                {label}
              </button>
            )
          })}
        </nav>

        {/* Bottom info */}
        <div
          className="no-drag px-5 pb-5"
          style={{ fontSize: 11, color: 'rgba(0,0,0,0.22)' }}
        >
          v1.0.0
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 flex flex-col overflow-hidden" style={{ background: '#F5F5F7' }}>
        {initError && (
          <div
            className="flex items-center gap-2 px-4 py-2.5"
            style={{
              background: 'rgba(220,38,38,0.05)',
              borderBottom: '1px solid rgba(220,38,38,0.12)',
              fontSize: 12,
              color: '#DC2626',
              flexShrink: 0,
            }}
          >
            <AlertCircle size={13} />
            {initError}
          </div>
        )}

        {page === 'users' && <UsersPage />}
        {page === 'scanner' && <ScannerPage />}
        {page === 'settings' && <SettingsPage />}
      </main>
    </div>
  )
}
