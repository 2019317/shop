import { useEffect, useState } from 'react'
import { Table, Button, Input, Select, Space, Tag, Modal, Form, App, Descriptions } from 'antd'
import { api, type Paged } from '../api/client'

interface OrderRow {
  id: string
  order_no: string
  email: string
  status: string
  total_cents: number
  currency: string
  created_at: string
}

interface OrderDetail {
  id: string
  order_no: string
  email: string
  status: string
  currency: string
  subtotal_cents: number
  shipping_cents: number
  discount_cents: number
  total_cents: number
  customer_note: string
  shipping_address: Record<string, string>
  items: {
    sku_code: string
    title: string
    qty: number
    unit_price_cents: number
    total_cents: number
  }[]
  payment_status?: string
  carrier?: string
  tracking_no?: string
  tracking_url?: string
  shipment_status?: string
  coupon_code?: string
  created_at: string
}

const statusColor: Record<string, string> = {
  pending: 'orange',
  paid: 'blue',
  fulfilled: 'green',
  cancelled: 'default',
  refunded: 'red',
}

export default function Orders() {
  const [data, setData] = useState<Paged<OrderRow>>({
    list: [],
    total: 0,
    page: 1,
    page_size: 20,
  })
  const [status, setStatus] = useState('')
  const [keyword, setKeyword] = useState('')
  const [loading, setLoading] = useState(false)
  const [detail, setDetail] = useState<OrderDetail | null>(null)
  const [shipOpen, setShipOpen] = useState(false)
  const [shipForm] = Form.useForm()
  const { message } = App.useApp()

  const load = async (page = 1) => {
    setLoading(true)
    try {
      const sp = new URLSearchParams({ page: String(page), page_size: '20' })
      if (status) sp.set('status', status)
      if (keyword) sp.set('keyword', keyword)
      setData(await api.get<Paged<OrderRow>>(`/admin/orders?${sp.toString()}`))
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

  const openDetail = async (id: string) => {
    try {
      setDetail(await api.get<OrderDetail>(`/admin/orders/${id}`))
    } catch (err) {
      message.error(err instanceof Error ? err.message : 'load failed')
    }
  }

  const ship = async () => {
    if (!detail) return
    const values = await shipForm.validateFields()
    try {
      await api.post(`/admin/orders/${detail.id}/ship`, values)
      message.success('Shipped')
      setShipOpen(false)
      shipForm.resetFields()
      setDetail(null)
      load(data.page)
    } catch (err) {
      message.error(err instanceof Error ? err.message : 'ship failed')
    }
  }

  const cancel = async () => {
    if (!detail) return
    try {
      await api.post(`/admin/orders/${detail.id}/cancel`, { reason: 'cancelled by admin' })
      message.success('Cancelled')
      setDetail(null)
      load(data.page)
    } catch (err) {
      message.error(err instanceof Error ? err.message : 'cancel failed')
    }
  }

  const markDelivered = async () => {
    if (!detail) return
    try {
      await api.post(`/admin/orders/${detail.id}/delivered`)
      message.success('Marked as delivered')
      await openDetail(detail.id)
    } catch (err) {
      message.error(err instanceof Error ? err.message : 'update failed')
    }
  }

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Input.Search
          placeholder="Order no or email"
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
        <Select
          allowClear
          placeholder="Status"
          style={{ width: 160 }}
          value={status || undefined}
          onChange={(v) => {
            setStatus(v || '')
            load(1)
          }}
          options={[
            { value: 'pending', label: 'Pending' },
            { value: 'paid', label: 'Paid' },
            { value: 'fulfilled', label: 'Fulfilled' },
            { value: 'cancelled', label: 'Cancelled' },
          ]}
        />
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
          { title: 'Order No', dataIndex: 'order_no' },
          { title: 'Email', dataIndex: 'email' },
          {
            title: 'Total',
            dataIndex: 'total_cents',
            render: (v: number, r) => `$${(v / 100).toFixed(2)} ${r.currency}`,
          },
          {
            title: 'Status',
            dataIndex: 'status',
            render: (v: string) => <Tag color={statusColor[v]}>{v}</Tag>,
          },
          { title: 'Created', dataIndex: 'created_at' },
          {
            title: 'Actions',
            render: (_, row) => (
              <Button type="link" onClick={() => openDetail(row.id)}>
                View
              </Button>
            ),
          },
        ]}
      />

      <Modal
        title={detail ? `Order ${detail.order_no}` : 'Order'}
        open={!!detail}
        onCancel={() => setDetail(null)}
        width={720}
        footer={[
          <Button key="close" onClick={() => setDetail(null)}>
            Close
          </Button>,
          detail?.status === 'paid' && (
            <Button key="ship" type="primary" onClick={() => setShipOpen(true)}>
              Ship
            </Button>
          ),
          detail?.status === 'pending' && (
            <Button key="cancel" danger onClick={cancel}>
              Cancel
            </Button>
          ),
          detail?.status === 'fulfilled' && detail?.shipment_status === 'shipped' && (
            <Button key="delivered" onClick={markDelivered}>
              Mark Delivered
            </Button>
          ),
        ]}
      >
        {detail && (
          <>
            <Descriptions column={2} size="small" bordered>
              <Descriptions.Item label="Status">
                <Tag color={statusColor[detail.status]}>{detail.status}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Email">{detail.email}</Descriptions.Item>
              <Descriptions.Item label="Subtotal">
                ${(detail.subtotal_cents / 100).toFixed(2)}
              </Descriptions.Item>
              <Descriptions.Item label="Shipping">
                ${(detail.shipping_cents / 100).toFixed(2)}
              </Descriptions.Item>
              <Descriptions.Item label="Total">
                ${(detail.total_cents / 100).toFixed(2)}
              </Descriptions.Item>
              <Descriptions.Item label="Created">{detail.created_at}</Descriptions.Item>
              <Descriptions.Item label="Ship to" span={2}>
                {[
                  detail.shipping_address?.name,
                  detail.shipping_address?.address1,
                  detail.shipping_address?.address2,
                  detail.shipping_address?.city,
                  detail.shipping_address?.state,
                  detail.shipping_address?.postal_code,
                  detail.shipping_address?.country,
                ]
                  .filter(Boolean)
                  .join(', ')}
              </Descriptions.Item>
              {detail.customer_note && (
                <Descriptions.Item label="Note" span={2}>
                  {detail.customer_note}
                </Descriptions.Item>
              )}
            </Descriptions>

            <Table
              style={{ marginTop: 16 }}
              rowKey="sku_code"
              size="small"
              pagination={false}
              dataSource={detail.items}
              columns={[
                { title: 'SKU', dataIndex: 'sku_code' },
                { title: 'Title', dataIndex: 'title' },
                { title: 'Qty', dataIndex: 'qty' },
                {
                  title: 'Price',
                  dataIndex: 'unit_price_cents',
                  render: (v: number) => `$${(v / 100).toFixed(2)}`,
                },
                {
                  title: 'Total',
                  dataIndex: 'total_cents',
                  render: (v: number) => `$${(v / 100).toFixed(2)}`,
                },
              ]}
            />
          </>
        )}
      </Modal>

      <Modal
        title="Ship order"
        open={shipOpen}
        onOk={ship}
        onCancel={() => setShipOpen(false)}
        destroyOnClose
      >
        <Form form={shipForm} layout="vertical">
          <Form.Item name="carrier" label="Carrier">
            <Input placeholder="USPS / UPS / YunExpress" />
          </Form.Item>
          <Form.Item
            name="tracking_no"
            label="Tracking No"
            rules={[{ required: true }]}
          >
            <Input />
          </Form.Item>
          <Form.Item name="tracking_url" label="Tracking URL">
            <Input placeholder="https://..." />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
