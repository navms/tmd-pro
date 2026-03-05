import { useState, useEffect } from 'react'
import { Save, RotateCcw, FolderOpen, Check, X, Eye, EyeOff } from 'lucide-react'
import { GetConfig, SaveConfig, OpenConfigDir } from '../wailsjs/go/app/App'
import type { ConfigData } from '../wailsjs/go/app/App'
import { PageHeader } from './Users'

const defaultConfig: ConfigData = {
  scanInterval: 30,
  dataDir: '',
  httpProxy: '',
  httpsProxy: '',
  noProxy: '',
  dbHost: '',
  dbPort: 3306,
  dbUsername: '',
  dbPassword: '',
  dbDatabase: '',
  dbCharset: 'utf8mb4',
}

export default function SettingsPage() {
  const [config, setConfig] = useState<ConfigData>(defaultConfig)
  const [original, setOriginal] = useState<ConfigData>(defaultConfig)
  const [saving, setSaving] = useState(false)
  const [showPassword, setShowPassword] = useState(false)
  const [toast, setToast] = useState<{ msg: string; type: 'ok' | 'err' } | null>(null)

  const isDirty = JSON.stringify(config) !== JSON.stringify(original)

  const load = async () => {
    try {
      const data = await GetConfig()
      setConfig(data)
      setOriginal(data)
    } catch (e: any) {
      showToast(String(e), 'err')
    }
  }

  useEffect(() => { load() }, [])

  const showToast = (msg: string, type: 'ok' | 'err') => {
    setToast({ msg, type })
    setTimeout(() => setToast(null), 3500)
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      await SaveConfig(config)
      setOriginal(config)
      showToast('配置已保存，部分项重启后生效', 'ok')
    } catch (e: any) {
      showToast(String(e), 'err')
    } finally {
      setSaving(false)
    }
  }

  const handleReset = () => {
    setConfig(original)
    showToast('已重置为已保存的值', 'ok')
  }

  const handleOpenDir = async () => {
    try {
      await OpenConfigDir()
    } catch (e: any) {
      showToast(String(e), 'err')
    }
  }

  const set = <K extends keyof ConfigData>(key: K, value: ConfigData[K]) =>
    setConfig((prev) => ({ ...prev, [key]: value }))

  return (
    <div className="flex flex-col h-full">
      <PageHeader title="系统配置">
        <div className="flex items-center gap-2 no-drag">
          {isDirty && (
            <span style={{
              fontSize: 11, color: '#D97706', padding: '2px 8px',
              background: 'rgba(217,119,6,0.07)', borderRadius: 999,
              border: '1px solid rgba(217,119,6,0.15)', fontWeight: 500,
            }}>
              未保存
            </span>
          )}
          <ToolBtn icon={FolderOpen} label="配置目录" onClick={handleOpenDir} />
          <ToolBtn icon={RotateCcw} label="重置" onClick={handleReset} disabled={!isDirty} />
          <ToolBtn
            icon={Save}
            label={saving ? '保存中…' : '保存'}
            onClick={handleSave}
            disabled={saving || !isDirty}
            primary
          />
        </div>
      </PageHeader>

      {toast && (
        <div className="mx-5 mt-4 flex items-center gap-2 px-3 py-2 rounded-lg" style={{
          background: toast.type === 'ok' ? 'rgba(5,150,105,0.06)' : 'rgba(220,38,38,0.06)',
          border: `1px solid ${toast.type === 'ok' ? 'rgba(5,150,105,0.18)' : 'rgba(220,38,38,0.18)'}`,
          color: toast.type === 'ok' ? '#059669' : '#DC2626',
          fontSize: 12, flexShrink: 0,
        }}>
          {toast.type === 'ok' ? <Check size={12} /> : <X size={12} />}
          {toast.msg}
        </div>
      )}

      <div className="flex-1 overflow-y-auto px-5 py-4 space-y-3">

        {/* 扫描配置 */}
        <SettingSection title="扫描配置">
          <FieldRow label="扫描间隔" hint="每轮轮转之间的等待时间">
            <div className="flex items-center gap-2">
              <input
                type="number"
                value={config.scanInterval}
                min={1}
                onChange={(e) => set('scanInterval', parseInt(e.target.value) || 1)}
                style={{ width: 90, padding: '6px 10px', textAlign: 'right' }}
              />
              <span style={{ fontSize: 12, color: 'rgba(0,0,0,0.35)' }}>分钟</span>
            </div>
          </FieldRow>
          <FieldRow label="数据目录" hint="需重启生效">
            <input
              value={config.dataDir}
              onChange={(e) => set('dataDir', e.target.value)}
              placeholder="~/.tmd-pro/data"
              style={{ width: '100%', padding: '6px 10px', fontFamily: 'monospace', fontSize: 12 }}
            />
          </FieldRow>
        </SettingSection>

        {/* 代理配置 */}
        <SettingSection title="代理配置">
          <FieldRow label="HTTP 代理">
            <input
              value={config.httpProxy}
              onChange={(e) => set('httpProxy', e.target.value)}
              placeholder="http://127.0.0.1:7890"
              style={{ width: '100%', padding: '6px 10px' }}
            />
          </FieldRow>
          <FieldRow label="HTTPS 代理">
            <input
              value={config.httpsProxy}
              onChange={(e) => set('httpsProxy', e.target.value)}
              placeholder="http://127.0.0.1:7890"
              style={{ width: '100%', padding: '6px 10px' }}
            />
          </FieldRow>
          <FieldRow label="不走代理">
            <input
              value={config.noProxy}
              onChange={(e) => set('noProxy', e.target.value)}
              placeholder="localhost,127.0.0.1"
              style={{ width: '100%', padding: '6px 10px' }}
            />
          </FieldRow>
        </SettingSection>

        {/* 数据库配置 */}
        <SettingSection title="数据库配置" hint="需重启生效">
          {/* Host + Port 同行 */}
          <div className="flex gap-3">
            <div className="flex-1">
              <FieldLabel label="主机" required />
              <input
                value={config.dbHost}
                onChange={(e) => set('dbHost', e.target.value)}
                placeholder="127.0.0.1"
                style={{ width: '100%', padding: '6px 10px', marginTop: 4 }}
              />
            </div>
            <div style={{ width: 110 }}>
              <FieldLabel label="端口" required />
              <input
                type="number"
                value={config.dbPort}
                min={1}
                max={65535}
                onChange={(e) => set('dbPort', parseInt(e.target.value) || 3306)}
                style={{ width: '100%', padding: '6px 10px', marginTop: 4 }}
              />
            </div>
          </div>

          {/* Database + Charset 同行 */}
          <div className="flex gap-3">
            <div className="flex-1">
              <FieldLabel label="数据库名" required />
              <input
                value={config.dbDatabase}
                onChange={(e) => set('dbDatabase', e.target.value)}
                placeholder="tmd_pro"
                style={{ width: '100%', padding: '6px 10px', marginTop: 4 }}
              />
            </div>
            <div style={{ width: 140 }}>
              <FieldLabel label="字符集" />
              <input
                value={config.dbCharset}
                onChange={(e) => set('dbCharset', e.target.value)}
                placeholder="utf8mb4"
                style={{ width: '100%', padding: '6px 10px', marginTop: 4 }}
              />
            </div>
          </div>

          {/* Username */}
          <div>
            <FieldLabel label="用户名" />
            <input
              value={config.dbUsername}
              onChange={(e) => set('dbUsername', e.target.value)}
              placeholder="root"
              style={{ width: '100%', padding: '6px 10px', marginTop: 4 }}
            />
          </div>

          {/* Password */}
          <div>
            <FieldLabel label="密码" />
            <div className="relative flex items-center" style={{ marginTop: 4 }}>
              <input
                type={showPassword ? 'text' : 'password'}
                value={config.dbPassword}
                onChange={(e) => set('dbPassword', e.target.value)}
                placeholder="••••••••"
                style={{ width: '100%', padding: '6px 36px 6px 10px' }}
              />
              <button
                onClick={() => setShowPassword((v) => !v)}
                style={{
                  position: 'absolute',
                  right: 10,
                  background: 'transparent',
                  border: 'none',
                  color: 'rgba(0,0,0,0.3)',
                  cursor: 'pointer',
                  padding: 2,
                  display: 'flex',
                  alignItems: 'center',
                }}
              >
                {showPassword ? <EyeOff size={13} /> : <Eye size={13} />}
              </button>
            </div>
          </div>
        </SettingSection>

        <p style={{ fontSize: 11, color: 'rgba(0,0,0,0.25)', textAlign: 'center', paddingBottom: 8 }}>
          标注「需重启生效」的项保存后需重启应用才能完全生效
        </p>
      </div>
    </div>
  )
}

// ─── Section components ───────────────────────────────────────────────────────

function SettingSection({
  title,
  hint,
  children,
}: {
  title: string
  hint?: string
  children: React.ReactNode
}) {
  return (
    <div className="rounded-xl overflow-hidden" style={{
      border: '1px solid rgba(0,0,0,0.07)',
      background: '#FFFFFF',
      boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
    }}>
      <div style={{
        padding: '9px 16px',
        borderBottom: '1px solid rgba(0,0,0,0.05)',
        display: 'flex',
        alignItems: 'center',
        gap: 8,
      }}>
        <span style={{
          fontSize: 11, fontWeight: 600,
          color: 'rgba(0,0,0,0.45)',
          textTransform: 'uppercase',
          letterSpacing: '0.06em',
        }}>
          {title}
        </span>
        {hint && (
          <span style={{
            fontSize: 10, color: 'rgba(0,0,0,0.28)',
            padding: '1px 6px',
            background: 'rgba(0,0,0,0.04)',
            borderRadius: 999,
            border: '1px solid rgba(0,0,0,0.07)',
          }}>
            {hint}
          </span>
        )}
      </div>
      <div style={{ padding: '14px 16px' }} className="space-y-3">
        {children}
      </div>
    </div>
  )
}

function FieldRow({
  label,
  hint,
  children,
}: {
  label: string
  hint?: string
  children: React.ReactNode
}) {
  return (
    <div className="flex items-center gap-4">
      <div style={{ width: 100, flexShrink: 0 }}>
        <div style={{ fontSize: 13, color: '#3C3C43', fontWeight: 500 }}>{label}</div>
        {hint && <div style={{ fontSize: 10, color: 'rgba(0,0,0,0.3)', marginTop: 1 }}>{hint}</div>}
      </div>
      <div className="flex-1">{children}</div>
    </div>
  )
}

function FieldLabel({ label, required }: { label: string; required?: boolean }) {
  return (
    <div style={{ fontSize: 12, color: '#6E6E73', fontWeight: 500 }}>
      {label}
      {required && <span style={{ color: '#DC2626', marginLeft: 2 }}>*</span>}
    </div>
  )
}

function ToolBtn({
  icon: Icon,
  label,
  onClick,
  disabled,
  primary,
}: {
  icon: React.ElementType
  label: string
  onClick: () => void
  disabled?: boolean
  primary?: boolean
}) {
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      style={{
        display: 'flex', alignItems: 'center', gap: 5,
        padding: '5px 12px', borderRadius: 7, fontSize: 12, fontWeight: 500,
        background: primary
          ? disabled ? 'rgba(99,102,241,0.25)' : '#6366F1'
          : 'rgba(0,0,0,0.04)',
        color: primary
          ? disabled ? 'rgba(99,102,241,0.5)' : '#fff'
          : disabled ? 'rgba(0,0,0,0.2)' : 'rgba(0,0,0,0.55)',
        border: primary ? 'none' : '1px solid rgba(0,0,0,0.08)',
        cursor: disabled ? 'default' : 'pointer',
        transition: 'all 0.12s',
      }}
    >
      <Icon size={12} />
      {label}
    </button>
  )
}
