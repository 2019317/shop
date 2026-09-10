import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Table, Button, Input, Space, Tag, App, Popconfirm } from 'antd'
import { api, type ProductRow, type Paged } from '../api/client'
import { useI18n } from '../i18n/I18nProvider'

const statusColor: Record<string, string> = {
  published: 'green',
  draft: 'orange',
  archived: 'default',
}

export default function Products() {
  const [data, setData] = useState<Paged<ProductRow>>({
    list: [],
    total: 0,
    page: 1,
    page_size: 20,
  })
  const [keyword, setKeyword] = useState('')
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const { message } = App.useApp()
  const { dict } = useI18n()

  const load = async (page = 1, kw = keyword) => {
    setLoading(true)
    try {
      const res = await api.get<Paged<ProductRow>>(
        `/admin/products?page=${page}&page_size=20${kw ? `&keyword=${encodeURIComponent(kw)}` : ''}`,
      )
      setData(res)
    } catch (err) {
      message.error(err instanceof Error ? err.message : dict.error.loadFailed)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const remove = async (id: string) => {
    try {
      await api.del(`/admin/products/${id}`)
      message.success(dict.products.archived)
      load(data.page)
    } catch (err) {
      message.error(err instanceof Error ? err.message : dict.error.saveFailed)
    }
  }

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Input.Search
          placeholder={dict.products.search}
          allowClear
          enterButton
          style={{ width: 300 }}
          value={keyword}
          onChange={(e) => setKeyword(e.target.value)}
          onSearch={(v) => load(1, v)}
        />
        <Button type="primary" onClick={() => navigate('/products/new')}>
          {dict.products.new}
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
          { title: dict.products.title, dataIndex: 'title' },
          { title: dict.products.slug, dataIndex: 'slug' },
          {
            title: dict.products.price,
            dataIndex: 'price_cents',
            render: (v: number, r) => `$${(v / 100).toFixed(2)} ${r.currency}`,
          },
          {
            title: dict.products.status,
            dataIndex: 'status',
            render: (v: string) => (
              <Tag color={statusColor[v]}>
                {dict.status[v as keyof typeof dict.status] ?? v}
              </Tag>
            ),
          },
          { title: dict.products.updated, dataIndex: 'updated_at' },
          {
            title: dict.products.actions,
            render: (_, row) => (
              <Space>
                <Button type="link" onClick={() => navigate(`/products/${row.id}`)}>
                  {dict.products.editAction}
                </Button>
                <Popconfirm
                  title={dict.products.archiveConfirm}
                  okText={dict.products.archive}
                  onConfirm={() => remove(row.id)}
                >
                  <Button type="link" danger>
                    {dict.products.archive}
                  </Button>
                </Popconfirm>
              </Space>
            ),
          },
        ]}
      />
    </div>
  )
}
