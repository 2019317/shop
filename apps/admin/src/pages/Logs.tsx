import { useCallback, useEffect, useRef, useState } from 'react'
import {
  Table,
  Button,
  Input,
  Select,
  Space,
  Tag,
  App,
  Popconfirm,
  Switch,
  Tooltip,
} from 'antd'
import { ReloadOutlined, DeleteOutlined } from '@ant-design/icons'
import { api, type Paged } from '../api/client'

interface LogEntry {
  id: number
  time: string
  level: 'info' | 'warn' | 'error'
  method: string
  path: string
  status: number
  duration_ms: number
  ip: string
  message: string
  error?: string
}

const levelColor: Record<string, string> = {
  info: 'blue',
  warn: 'orange',
  error: 'red',
}

export default function Logs() {
  const [data, setData] = useState<Paged<LogEntry>>({
    list: [],
    total: 0,
    page: 1,
    page_size: 50,
  })
  const [level, setLevel] = useState('')
  const [keyword, setKeyword] = useState('')
  const [loading, setLoading] = useState(false)
  const [auto, setAuto] = useState(true)
  const { message } = App.useApp()
  const timerRef = useRef<number | null>(null)
  const filtersRef = useRef({ level, keyword })

  filtersRef.current = { level, keyword }

  const load = useCallback(
    async (page = 1) => {
      setLoading(true)
      try {
        const sp = new URLSearchParams({ page: String(page), page_size: '50' })
        if (filtersRef.current.level) sp.set('level', filtersRef.current.level)
        if (filtersRef.current.keyword) sp.set('keyword', filtersRef.current.keyword)
        const res = await api.get<Paged<LogEntry>>(`/admin/logs?${sp.toString()}`)
        setData(res)
      } catch (err) {
        message.error(err instanceof Error ? err.message : 'load failed')
      } finally {
        setLoading(false)
      }
    },
    [message],
  )

  useEffect(() => {
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // 自动刷新（每 5 秒），仅在开关打开时生效；不改变当前页码之外的筛选
  useEffect(() => {
    if (!auto) {
      if (timerRef.current) window.clearInterval(timerRef.current)
      return
    }
    timerRef.current = window.setInterval(() => {
      load(1)
    }, 5000)
    return () => {
      if (timerRef.current) window.clearInterval(timerRef.current)
    }
  }, [auto, load])

  const clear = async () => {
    try {
      await api.del('/admin/logs')
      message.success('Cleared')
      load(1)
    } catch (err) {
      message.error(err instanceof Error ? err.message : 'clear failed')
    }
  }

  const statusColor = (status: number) => {
    if (status >= 500) return 'red'
    if (status >= 400) return 'orange'
    return 'green'
  }

  return (
    <div>
      <Space style={{ marginBottom: 16 }} wrap>
        <Select
          allowClear
          placeholder="Level"
          style={{ width: 140 }}
          value={level || undefined}
          onChange={(v) => {
            setLevel(v || '')
            load(1)
          }}
          options={[
            { value: 'info', label: 'Info' },
            { value: 'warn', label: 'Warn' },
            { value: 'error', label: 'Error' },
          ]}
        />
        <Input.Search
          placeholder="Search path / error"
          allowClear
          enterButton
          style={{ width: 260 }}
          value={keyword}
          onChange={(e) => setKeyword(e.target.value)}
          onSearch={(v) => {
            setKeyword(v)
            load(1)
          }}
        />
        <Button icon={<ReloadOutlined />} onClick={() => load(data.page)}>
          Refresh
        </Button>
        <span style={{ color: 'var(--color-muted, #888)', fontSize: 13 }}>
          <Switch size="small" checked={auto} onChange={setAuto} /> Auto (5s)
        </span>
        <Popconfirm title="Clear all logs?" onConfirm={clear}>
          <Button danger icon={<DeleteOutlined />}>
            Clear
          </Button>
        </Popconfirm>
      </Space>

      <Table
        rowKey="id"
        size="small"
        loading={loading}
        dataSource={data.list}
        pagination={{
          current: data.page,
          pageSize: data.page_size,
          total: data.total,
          onChange: (p) => load(p),
        }}
        columns={[
          {
            title: 'Time',
            dataIndex: 'time',
            width: 180,
            render: (v: string) => (v ? new Date(v).toLocaleString() : '-'),
          },
          {
            title: 'Level',
            dataIndex: 'level',
            width: 90,
            render: (v: string) => <Tag color={levelColor[v]}>{v}</Tag>,
          },
          {
            title: 'Method',
            dataIndex: 'method',
            width: 80,
          },
          {
            title: 'Path',
            dataIndex: 'path',
            ellipsis: true,
          },
          {
            title: 'Status',
            dataIndex: 'status',
            width: 90,
            render: (v: number) => <Tag color={statusColor(v)}>{v}</Tag>,
          },
          {
            title: 'Duration',
            dataIndex: 'duration_ms',
            width: 100,
            render: (v: number) => `${v} ms`,
          },
          {
            title: 'IP',
            dataIndex: 'ip',
            width: 130,
          },
          {
            title: 'Error',
            dataIndex: 'error',
            ellipsis: true,
            render: (v: string) =>
              v ? (
                <Tooltip title={v}>
                  <span style={{ color: '#cf1322' }}>{v}</span>
                </Tooltip>
              ) : (
                '-'
              ),
          },
        ]}
      />
    </div>
  )
}
