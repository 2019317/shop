import { useEffect, useState } from 'react'
import {
  Table,
  Button,
  Input,
  Select,
  Space,
  Tag,
  Modal,
  Form,
  InputNumber,
  DatePicker,
  App,
  Switch,
} from 'antd'
import type { Dayjs } from 'dayjs'
import dayjs from 'dayjs'
import { api, type CouponRow, type CouponInput, type Paged } from '../api/client'

const statusColor: Record<string, string> = {
  active: 'green',
  disabled: 'default',
}

interface CouponFormValues {
  code: string
  type: 'percent' | 'fixed'
  value: number
  min_amount_cents: number
  max_uses: number
  range?: [Dayjs, Dayjs]
  status: string
}

export default function Coupons() {
  const [data, setData] = useState<Paged<CouponRow>>({
    list: [],
    total: 0,
    page: 1,
    page_size: 20,
  })
  const [keyword, setKeyword] = useState('')
  const [status, setStatus] = useState('')
  const [loading, setLoading] = useState(false)
  const [editing, setEditing] = useState<CouponRow | null>(null)
  const [open, setOpen] = useState(false)
  const [form] = Form.useForm<CouponFormValues>()
  const { message } = App.useApp()

  const load = async (page = 1) => {
    setLoading(true)
    try {
      const sp = new URLSearchParams({ page: String(page), page_size: '20' })
      if (keyword) sp.set('keyword', keyword)
      if (status) sp.set('status', status)
      setData(await api.get<Paged<CouponRow>>(`/admin/coupons?${sp.toString()}`))
    } catch (err) {
      message.error(err instanceof Error ? err.message : 'load failed')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const openCreate = () => {
    setEditing(null)
    form.resetFields()
    form.setFieldsValue({ type: 'percent', status: 'active', min_amount_cents: 0, max_uses: 0, value: 10 })
    setOpen(true)
  }

  const openEdit = (row: CouponRow) => {
    setEditing(row)
    form.setFieldsValue({
      code: row.code,
      type: row.type,
      value: row.value,
      min_amount_cents: row.min_amount_cents,
      max_uses: row.max_uses,
      status: row.status,
      range:
        row.starts_at && row.ends_at
          ? [dayjs(row.starts_at), dayjs(row.ends_at)]
          : undefined,
    })
    setOpen(true)
  }

  const submit = async () => {
    const values = await form.validateFields()
    const payload: CouponInput = {
      code: values.code,
      type: values.type,
      value: values.value,
      min_amount_cents: values.min_amount_cents || 0,
      max_uses: values.max_uses || 0,
      status: values.status || 'active',
      starts_at: values.range?.[0] ? values.range[0].toISOString() : '',
      ends_at: values.range?.[1] ? values.range[1].toISOString() : '',
    }
    try {
      if (editing) {
        await api.put(`/admin/coupons/${editing.id}`, payload)
        message.success('Updated')
      } else {
        await api.post('/admin/coupons', payload)
        message.success('Created')
      }
      setOpen(false)
      load(data.page)
    } catch (err) {
      message.error(err instanceof Error ? err.message : 'save failed')
    }
  }

  const toggleStatus = async (row: CouponRow, checked: boolean) => {
    try {
      await api.post(`/admin/coupons/${row.id}/status`, {
        status: checked ? 'active' : 'disabled',
      })
      message.success('Updated')
      load(data.page)
    } catch (err) {
      message.error(err instanceof Error ? err.message : 'update failed')
    }
  }

  const describeValue = (row: CouponRow) =>
    row.type === 'percent' ? `${row.value / 10} off` : `$${(row.value / 100).toFixed(2)} off`

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Input.Search
          placeholder="Search by code"
          allowClear
          enterButton
          style={{ width: 220 }}
          value={keyword}
          onChange={(e) => setKeyword(e.target.value)}
          onSearch={(v) => {
            setKeyword(v)
            load(1)
          }}
        />
        <Select
          allowClear
          placeholder="Status"
          style={{ width: 140 }}
          value={status || undefined}
          onChange={(v) => {
            setStatus(v || '')
            load(1)
          }}
          options={[
            { value: 'active', label: 'Active' },
            { value: 'disabled', label: 'Disabled' },
          ]}
        />
        <Button type="primary" onClick={openCreate}>
          New Coupon
        </Button>
      </Space>

      <Table
        rowKey="id"
        loading={loading}
        dataSource={data.list}
        pagination={{
          current: data.page,
          pageSize: data.page_size,
          total: data.total,
          onChange: (p) => load(p),
        }}
        columns={[
          { title: 'Code', dataIndex: 'code' },
          { title: 'Discount', render: (_, row) => describeValue(row) },
          {
            title: 'Min Amount',
            dataIndex: 'min_amount_cents',
            render: (v: number) => (v > 0 ? `$${(v / 100).toFixed(2)}` : '-'),
          },
          {
            title: 'Usage',
            render: (_, row) => `${row.used_count} / ${row.max_uses === 0 ? '∞' : row.max_uses}`,
          },
          {
            title: 'Valid',
            render: (_, row) =>
              row.starts_at || row.ends_at
                ? `${row.starts_at || '…'} ~ ${row.ends_at || '…'}`
                : 'Always',
          },
          {
            title: 'Status',
            dataIndex: 'status',
            render: (v: string, row) => (
              <Space>
                <Tag color={statusColor[v]}>{v}</Tag>
                <Switch
                  size="small"
                  checked={v === 'active'}
                  onChange={(checked) => toggleStatus(row, checked)}
                />
              </Space>
            ),
          },
          {
            title: 'Actions',
            render: (_, row) => (
              <Button type="link" onClick={() => openEdit(row)}>
                Edit
              </Button>
            ),
          },
        ]}
      />

      <Modal
        title={editing ? 'Edit Coupon' : 'New Coupon'}
        open={open}
        onOk={submit}
        onCancel={() => setOpen(false)}
        destroyOnClose
      >
        <Form form={form} layout="vertical">
          <Form.Item name="code" label="Code" rules={[{ required: true }]}>
            <Input placeholder="SUMMER10" disabled={!!editing} />
          </Form.Item>
          <Form.Item name="type" label="Type" rules={[{ required: true }]}>
            <Select
              options={[
                { value: 'percent', label: 'Percent (10 = 10% off)' },
                { value: 'fixed', label: 'Fixed (cents)' },
              ]}
            />
          </Form.Item>
          <Form.Item name="value" label="Value" rules={[{ required: true }]}>
            <InputNumber min={1} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="min_amount_cents" label="Min Order (cents)">
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="max_uses" label="Max Uses (0 = unlimited)">
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="range" label="Valid Period">
            <DatePicker.RangePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="status" label="Status">
            <Select
              options={[
                { value: 'active', label: 'Active' },
                { value: 'disabled', label: 'Disabled' },
              ]}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
