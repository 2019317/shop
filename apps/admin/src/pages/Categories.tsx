import { useEffect, useState } from 'react'
import { Table, Button, Modal, Form, Input, InputNumber, App, Tabs } from 'antd'
import { api, type Category, type CategoryTranslationInput } from '../api/client'
import { useI18n } from '../i18n/I18nProvider'

export default function Categories() {
  const [list, setList] = useState<Category[]>([])
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const [zhTrans, setZhTrans] = useState<CategoryTranslationInput>({})
  const [form] = Form.useForm()
  const { message } = App.useApp()
  const { dict } = useI18n()

  const load = async () => {
    setLoading(true)
    try {
      setList(await api.get<Category[]>('/admin/categories'))
    } catch (err) {
      message.error(err instanceof Error ? err.message : dict.error.loadFailed)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  const submit = async () => {
    const values = await form.validateFields()
    const hasZh = !!((zhTrans.name ?? '') || (zhTrans.description ?? ''))
    try {
      await api.post('/admin/categories', {
        ...values,
        status: values.status || 'active',
        translations: hasZh ? { zh: zhTrans } : undefined,
      })
      message.success(dict.categories.created)
      setOpen(false)
      setZhTrans({})
      form.resetFields()
      load()
    } catch (err) {
      message.error(err instanceof Error ? err.message : dict.error.saveFailed)
    }
  }

  return (
    <div>
      <Button type="primary" style={{ marginBottom: 16 }} onClick={() => setOpen(true)}>
        {dict.categories.new}
      </Button>

      <Table
        rowKey="id"
        loading={loading}
        dataSource={list}
        pagination={false}
        columns={[
          { title: dict.categories.name, dataIndex: 'name' },
          { title: dict.categories.slug, dataIndex: 'slug' },
          { title: dict.categories.description, dataIndex: 'description' },
          { title: dict.categories.sortOrder, dataIndex: 'sort_order' },
        ]}
      />

      <Modal
        title={dict.categories.new}
        open={open}
        onOk={submit}
        onCancel={() => setOpen(false)}
        destroyOnClose
      >
        <Form form={form} layout="vertical">
          <Tabs
            defaultActiveKey="en"
            items={[
              {
                key: 'en',
                label: dict.product.baseLang,
                children: (
                  <>
                    <Form.Item name="name" label={dict.categories.name} rules={[{ required: true, message: dict.error.required }]}>
                      <Input />
                    </Form.Item>
                    <Form.Item
                      name="slug"
                      label={dict.categories.slug}
                      rules={[
                        { required: true, message: dict.error.required },
                        { pattern: /^[a-z0-9-]+$/, message: dict.categories.slugPattern },
                      ]}
                    >
                      <Input placeholder="journals-notebooks" />
                    </Form.Item>
                    <Form.Item name="description" label={dict.categories.description}>
                      <Input.TextArea rows={3} />
                    </Form.Item>
                    <Form.Item name="sort_order" label={dict.categories.sortOrder}>
                      <InputNumber defaultValue={0} />
                    </Form.Item>
                  </>
                ),
              },
              {
                key: 'zh',
                label: dict.product.zhLang,
                children: (
                  <>
                    <Form.Item label={dict.categories.name}>
                      <Input
                        value={zhTrans.name}
                        onChange={(e) => setZhTrans((s) => ({ ...s, name: e.target.value }))}
                      />
                    </Form.Item>
                    <Form.Item label={dict.categories.description}>
                      <Input.TextArea
                        rows={3}
                        value={zhTrans.description}
                        onChange={(e) => setZhTrans((s) => ({ ...s, description: e.target.value }))}
                      />
                    </Form.Item>
                  </>
                ),
              },
            ]}
          />
        </Form>
      </Modal>
    </div>
  )
}
