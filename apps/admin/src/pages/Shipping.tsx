import { useEffect, useState } from 'react'
import {
  Table,
  Button,
  Space,
  Tag,
  Modal,
  Form,
  Input,
  InputNumber,
  Select,
  App,
  Popconfirm,
} from 'antd'
import {
  api,
  type ShippingRuleRow,
  type ShippingRuleInput,
  type Paged,
} from '../api/client'

interface RuleFormValues {
  name: string
  country_codes: string[]
  min_amount_dollars: number
  max_weight_g: number
  price_dollars: number
  free_threshold_dollars: number
  sort_order: number
  status: string
}

export default function Shipping() {
  const [data, setData] = useState<Paged<ShippingRuleRow>>({
    list: [],
    total: 0,
    page: 1,
    page_size: 50,
  })
  const [loading, setLoading] = useState(false)
  const [editing, setEditing] = useState<ShippingRuleRow | null>(null)
  const [open, setOpen] = useState(false)
  const [form] = Form.useForm<RuleFormValues>()
  const { message } = App.useApp()

  const load = async (page = 1) => {
    setLoading(true)
    try {
      const sp = new URLSearchParams({ page: String(page), page_size: '50' })
      setData(await api.get<Paged<ShippingRuleRow>>(`/admin/shipping/rules?${sp.toString()}`))
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
    form.setFieldsValue({
      status: 'active',
      country_codes: [],
      min_amount_dollars: 0,
      max_weight_g: 0,
      price_dollars: 0,
      free_threshold_dollars: 0,
      sort_order: 0,
    })
    setOpen(true)
  }

  const openEdit = (row: ShippingRuleRow) => {
    setEditing(row)
    form.setFieldsValue({
      name: row.name,
      country_codes: row.country_codes || [],
      min_amount_dollars: row.min_amount_cents / 100,
      max_weight_g: row.max_weight_g,
      price_dollars: row.price_cents / 100,
      free_threshold_dollars: row.free_threshold_cents / 100,
      sort_order: row.sort_order,
      status: row.status,
    })
    setOpen(true)
  }

  const submit = async () => {
    const v = await form.validateFields()
    const payload: ShippingRuleInput = {
      name: v.name,
      country_codes: (v.country_codes || []).map((c) => c.toUpperCase()),
      min_amount_cents: Math.round((v.min_amount_dollars || 0) * 100),
      max_weight_g: v.max_weight_g || 0,
      price_cents: Math.round((v.price_dollars || 0) * 100),
      free_threshold_cents: Math.round((v.free_threshold_dollars || 0) * 100),
      sort_order: v.sort_order || 0,
      status: v.status || 'active',
    }
    try {
      if (editing) {
        await api.put(`/admin/shipping/rules/${editing.id}`, payload)
        message.success('Updated')
      } else {
        await api.post('/admin/shipping/rules', payload)
        message.success('Created')
      }
      setOpen(false)
      load(data.page)
    } catch (err) {
      message.error(err instanceof Error ? err.message : 'save failed')
    }
  }

  const remove = async (row: ShippingRuleRow) => {
    try {
      await api.del(`/admin/shipping/rules/${row.id}`)
      message.success('Deleted')
      load(data.page)
    } catch (err) {
      message.error(err instanceof Error ? err.message : 'delete failed')
    }
  }

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button type="primary" onClick={openCreate}>
          New Rule
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
          { title: 'Name', dataIndex: 'name' },
          {
            title: 'Countries',
            dataIndex: 'country_codes',
            render: (codes: string[]) =>
              codes && codes.length ? codes.map((c) => <Tag key={c}>{c}</Tag>) : <Tag>All</Tag>,
          },
          {
            title: 'Price',
            dataIndex: 'price_cents',
            render: (v: number) => `$${(v / 100).toFixed(2)}`,
          },
          {
            title: 'Free Over',
            dataIndex: 'free_threshold_cents',
            render: (v: number) => (v > 0 ? `$${(v / 100).toFixed(2)}` : '-'),
          },
          {
            title: 'Min Order',
            dataIndex: 'min_amount_cents',
            render: (v: number) => (v > 0 ? `$${(v / 100).toFixed(2)}` : '-'),
          },
          { title: 'Sort', dataIndex: 'sort_order' },
          {
            title: 'Status',
            dataIndex: 'status',
            render: (v: string) => <Tag color={v === 'active' ? 'green' : 'default'}>{v}</Tag>,
          },
          {
            title: 'Actions',
            render: (_, row) => (
              <Space>
                <Button type="link" onClick={() => openEdit(row)}>
                  Edit
                </Button>
                <Popconfirm title="Deactivate this rule?" onConfirm={() => remove(row)}>
                  <Button type="link" danger>
                    Delete
                  </Button>
                </Popconfirm>
              </Space>
            ),
          },
        ]}
      />

      <Modal
        title={editing ? 'Edit Shipping Rule' : 'New Shipping Rule'}
        open={open}
        onOk={submit}
        onCancel={() => setOpen(false)}
        destroyOnClose
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="Name" rules={[{ required: true }]}>
            <Input placeholder="Standard International" />
          </Form.Item>
          <Form.Item
            name="country_codes"
            label="Countries (leave empty for all others)"
            tooltip="ISO country codes, e.g. US, CA, GB"
          >
            <Select
              mode="tags"
              placeholder="US, CA, GB"
              tokenSeparators={[',']}
              options={[
                { value: 'US', label: 'US' },
                { value: 'CA', label: 'CA' },
                { value: 'GB', label: 'GB' },
                { value: 'AU', label: 'AU' },
                { value: 'DE', label: 'DE' },
                { value: 'FR', label: 'FR' },
                { value: 'JP', label: 'JP' },
              ]}
            />
          </Form.Item>
          <Form.Item name="price_dollars" label="Shipping Price (USD)">
            <InputNumber min={0} step={0.1} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="free_threshold_dollars" label="Free Over (USD, 0 = never)">
            <InputNumber min={0} step={1} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="min_amount_dollars" label="Min Order (USD)">
            <InputNumber min={0} step={1} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="max_weight_g" label="Max Weight (g, 0 = unlimited)">
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="sort_order" label="Sort Order">
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="status" label="Status">
            <Select
              options={[
                { value: 'active', label: 'Active' },
                { value: 'inactive', label: 'Inactive' },
              ]}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
