import { useState, useEffect, useRef, useCallback } from 'react'
import { Play, Square, SkipForward, Trash2, Copy, Check } from 'lucide-react'
import { StartScan, StopScan, IsScanning, RunOnce, GetConfig } from '../wailsjs/go/app/App'
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime'
import { PageHeader } from './Users'

interface LogEntry {
  id: number
  message: string
  time: string
  kind: 'info' | 'success' | 'error' | 'warning' | 'system'
}

let logIdCounter = 0

function classifyLog(msg: string): LogEntry['kind'] {
  const lower = msg.toLowerCase()
  if (lower.includes('失败') || lower.includes('error') || lower.includes('err')) return 'error'
  if (lower.includes('成功') || lower.includes('完成') || lower.includes('ok')) return 'success'
  if (lower.includes('警告') || lower.includes('warn')) return 'warning'
  if (lower.startsWith('正在启动') || lower.startsWith('扫描已停止')) return 'system'
  return 'info'
}

function nowTime() {
  return new Date().toLocaleTimeString('zh-CN', { hour12: false })
}

const logColors: Record<LogEntry['kind'], string> = {
  success: '#059669',
  error: '#DC2626',
  warning: '#D97706',
  system: '#6366F1',
  info: '#6E6E73',
}

export default function ScannerPage() {
  const [scanning, setScanning] = useState(false)
  const [runningOnce, setRunningOnce] = useState(false)
  const [interval, setInterval_] = useState(0)
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [copied, setCopied] = useState(false)
  const logEndRef = useRef<HTMLDivElement>(null)
  const logsRef = useRef<LogEntry[]>([])

  const appendLog = useCallback((message: string) => {
    const entry: LogEntry = {
      id: ++logIdCounter,
      message,
      time: nowTime(),
      kind: classifyLog(message),
    }
    logsRef.current = [...logsRef.current.slice(-999), entry]
    setLogs([...logsRef.current])
  }, [])

  useEffect(() => {
    IsScanning().then(setScanning).catch(() => {})
    GetConfig().then((cfg) => setInterval_(cfg.scanInterval)).catch(() => {})

    EventsOn('log', appendLog)
    EventsOn('scan:status', (running: boolean) => {
      setScanning(running)
      if (!running) setRunningOnce(false)
    })

    return () => {
      EventsOff('log', 'scan:status')
    }
  }, [appendLog])

  useEffect(() => {
    logEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [logs])

  const handleStart = async () => {
    try {
      await StartScan()
    } catch (e: any) {
      appendLog(`启动失败: ${e}`)
    }
  }

  const handleStop = async () => {
    try {
      await StopScan()
    } catch (e: any) {
      appendLog(`停止失败: ${e}`)
    }
  }

  const handleRunOnce = async () => {
    setRunningOnce(true)
    try {
      await RunOnce()
    } catch (e: any) {
      appendLog(`执行失败: ${e}`)
      setRunningOnce(false)
    }
  }

  const clearLogs = () => {
    logsRef.current = []
    setLogs([])
  }

  const copyLogs = () => {
    const text = logs.map((l) => `[${l.time}] ${l.message}`).join('\n')
    navigator.clipboard.writeText(text).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    })
  }

  const isBusy = scanning || runningOnce

  return (
    <div className="flex flex-col h-full">
      <PageHeader
        title="扫描控制"
        subtitle={interval > 0 ? `轮转间隔 ${interval} 分钟` : undefined}
      />

      {/* Status + Controls */}
      <div
        className="px-5 py-4 flex-shrink-0"
        style={{ borderBottom: '1px solid rgba(0,0,0,0.06)', background: '#F5F5F7' }}
      >
        {/* Status indicator */}
        <div className="flex items-center gap-3 mb-4">
          <div className="flex items-center gap-2">
            <div
              style={{
                width: 7,
                height: 7,
                borderRadius: '50%',
                background: scanning ? '#059669' : 'rgba(0,0,0,0.18)',
                boxShadow: scanning ? '0 0 0 3px rgba(5,150,105,0.15)' : 'none',
                transition: 'all 0.3s',
                animation: scanning ? 'statusPulse 2s infinite' : 'none',
              }}
            />
            <span
              style={{
                fontSize: 13,
                fontWeight: 500,
                color: scanning ? '#059669' : 'rgba(0,0,0,0.4)',
                transition: 'color 0.3s',
              }}
            >
              {scanning ? '运行中' : runningOnce ? '单次执行中' : '已停止'}
            </span>
          </div>

          {(scanning || runningOnce) && (
            <div className="flex gap-1">
              {[0, 1, 2].map((i) => (
                <div
                  key={i}
                  style={{
                    width: 3,
                    height: 3,
                    borderRadius: '50%',
                    background: '#059669',
                    opacity: 0.5,
                    animation: `dotBounce 1.2s ${i * 0.2}s infinite`,
                  }}
                />
              ))}
            </div>
          )}
        </div>

        {/* Control buttons */}
        <div className="flex items-center gap-2">
          <CtrlBtn icon={Play} label="启动轮转" onClick={handleStart} disabled={isBusy} variant="primary" />
          <CtrlBtn icon={Square} label="停止扫描" onClick={handleStop} disabled={!isBusy} variant="danger" />
          <div style={{ width: 1, height: 20, background: 'rgba(0,0,0,0.1)', margin: '0 4px' }} />
          <CtrlBtn icon={SkipForward} label="单次执行" onClick={handleRunOnce} disabled={isBusy} variant="default" />
        </div>
      </div>

      {/* Log console */}
      <div className="flex flex-col flex-1 overflow-hidden px-5 pb-5 pt-4 gap-3">
        {/* Log header */}
        <div className="flex items-center justify-between flex-shrink-0">
          <div className="flex items-center gap-2">
            <span style={{
              fontSize: 11,
              fontWeight: 600,
              color: 'rgba(0,0,0,0.35)',
              textTransform: 'uppercase',
              letterSpacing: '0.06em',
            }}>
              执行日志
            </span>
            {logs.length > 0 && (
              <span style={{
                fontSize: 10,
                padding: '1px 6px',
                background: 'rgba(0,0,0,0.06)',
                color: 'rgba(0,0,0,0.35)',
                borderRadius: 999,
              }}>
                {logs.length}
              </span>
            )}
          </div>
          <div className="flex items-center gap-1.5">
            <LogActionBtn icon={copied ? Check : Copy} label={copied ? '已复制' : '复制'} onClick={copyLogs} success={copied} />
            <LogActionBtn icon={Trash2} label="清空" onClick={clearLogs} />
          </div>
        </div>

        {/* Terminal */}
        <div
          className="flex-1 overflow-y-auto rounded-xl"
          style={{
            background: '#FFFFFF',
            border: '1px solid rgba(0,0,0,0.07)',
            boxShadow: '0 1px 3px rgba(0,0,0,0.04)',
            padding: '12px 14px',
          }}
        >
          {logs.length === 0 ? (
            <div
              className="flex items-center justify-center h-full"
              style={{ fontSize: 12, color: 'rgba(0,0,0,0.2)', fontFamily: 'monospace' }}
            >
              等待日志输出…
            </div>
          ) : (
            logs.map((log) => (
              <div
                key={log.id}
                style={{
                  display: 'flex',
                  gap: 10,
                  marginBottom: 3,
                  fontFamily: "'JetBrains Mono', 'SF Mono', monospace",
                  fontSize: 12,
                  lineHeight: 1.65,
                }}
              >
                <span style={{ color: 'rgba(0,0,0,0.2)', flexShrink: 0, fontSize: 11 }}>{log.time}</span>
                <span style={{ color: logColors[log.kind], wordBreak: 'break-all' }}>
                  {log.message}
                </span>
              </div>
            ))
          )}
          <div ref={logEndRef} />
        </div>
      </div>

      <style>{`
        @keyframes statusPulse {
          0%, 100% { box-shadow: 0 0 0 3px rgba(5,150,105,0.15); }
          50% { box-shadow: 0 0 0 5px rgba(5,150,105,0.06); }
        }
        @keyframes dotBounce {
          0%, 100% { transform: translateY(0); opacity: 0.5; }
          50% { transform: translateY(-3px); opacity: 1; }
        }
      `}</style>
    </div>
  )
}

function CtrlBtn({
  icon: Icon,
  label,
  onClick,
  disabled,
  variant = 'default',
}: {
  icon: React.ElementType
  label: string
  onClick: () => void
  disabled?: boolean
  variant?: 'primary' | 'danger' | 'default'
}) {
  const styles = {
    primary: {
      background: disabled ? 'rgba(99,102,241,0.25)' : '#6366F1',
      color: disabled ? 'rgba(99,102,241,0.5)' : '#fff',
      border: 'none',
    },
    danger: {
      background: disabled ? 'rgba(220,38,38,0.04)' : 'rgba(220,38,38,0.07)',
      color: disabled ? 'rgba(220,38,38,0.3)' : '#DC2626',
      border: `1px solid ${disabled ? 'rgba(220,38,38,0.07)' : 'rgba(220,38,38,0.15)'}`,
    },
    default: {
      background: disabled ? 'rgba(0,0,0,0.03)' : 'rgba(0,0,0,0.05)',
      color: disabled ? 'rgba(0,0,0,0.2)' : 'rgba(0,0,0,0.55)',
      border: '1px solid rgba(0,0,0,0.08)',
    },
  }
  const s = styles[variant]

  return (
    <button
      onClick={onClick}
      disabled={disabled}
      style={{
        ...s,
        display: 'flex',
        alignItems: 'center',
        gap: 6,
        padding: '7px 14px',
        borderRadius: 8,
        fontSize: 13,
        fontWeight: 500,
        cursor: disabled ? 'default' : 'pointer',
        transition: 'all 0.12s',
      }}
    >
      <Icon size={13} />
      {label}
    </button>
  )
}

function LogActionBtn({
  icon: Icon,
  label,
  onClick,
  success,
}: {
  icon: React.ElementType
  label: string
  onClick: () => void
  success?: boolean
}) {
  return (
    <button
      onClick={onClick}
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 5,
        padding: '4px 10px',
        borderRadius: 6,
        fontSize: 11,
        fontWeight: 500,
        background: success ? 'rgba(5,150,105,0.06)' : 'rgba(0,0,0,0.04)',
        color: success ? '#059669' : 'rgba(0,0,0,0.4)',
        border: `1px solid ${success ? 'rgba(5,150,105,0.15)' : 'rgba(0,0,0,0.08)'}`,
        cursor: 'pointer',
        transition: 'all 0.12s',
      }}
    >
      <Icon size={11} />
      {label}
    </button>
  )
}
