import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import {
  Form,
  Input,
  Button,
  Select,
  InputNumber,
  Card,
  Space,
  App,
  Upload,
  Row,
  Col,
  Divider,
  Typography,
  Tabs,
} from 'antd'
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons'
import {
  api,
  type ProductInput,
  type Category,
  type VariantInput,
  type ImageInput,
  type TranslationInput,
} from '../api/client'
import { uploadImage } from '../lib/upload'
import { useI18n } from '../i18n/I18nProvider'

const IMG_BASE = import.meta.env.VITE_IMG_BASE_URL || ''

// isValidJson 判断文本是否为合法的 JSON 对象（用于 Options 输入实时校验）
function isValidJson(text: string): boolean {
  try {
    const parsed = JSON.parse(text || '{}')
    return parsed !== null && typeof parsed === 'object' && !Array.isArray(parsed)
  } catch {
    return false
  }
}

export default function ProductEdit() {
  const { id } = useParams()
  const isEdit = Boolean(id)
  const [form] = Form.useForm()
  const [categories, setCategories] = useState<Category[]>([])
  const [variants, setVariants] = useState<VariantInput[]>([emptyVariant()])
  const [images, setImages] = useState<ImageInput[]>([])
  const [saving, setSaving] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [zhTrans, setZhTrans] = useState<TranslationInput>({})
  // 变体 Options 的原始文本（避免受控输入被 JSON.stringify 覆盖，导致无法正常编辑）
  const [optionsText, setOptionsText] = useState<Record<number, string>>({})
  const navigate = useNavigate()
  const { message } = App.useApp()
  const { dict } = useI18n()

  useEffect(() => {
    api.get<Category[]>('/admin/categories').then(setCategories).catch(() => {})
    if (isEdit) {
      api
        .get<ProductInput>(`/admin/products/${id}`)
        .then((data) => {
          form.setFieldsValue(data)
          const loadedVariants = data.variants?.length ? data.variants : [emptyVariant()]
          setVariants(loadedVariants)
          // 用加载到的 options 初始化文本态，避免受控输入被覆盖
          const textMap: Record<number, string> = {}
          loadedVariants.forEach((v, i) => {
            textMap[i] = JSON.stringify(v.options || {})
          })
          setOptionsText(textMap)
          setImages(data.images || [])
          if (data.translations?.zh) setZhTrans(data.translations.zh)
        })
        .catch((err) => message.error(err.message))
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id])

  const beforeUpload = async (file: File) => {
    setUploading(true)
    try {
      const asset = await uploadImage(file)
      setImages((prev) => [
        ...prev,
        { object_key: asset.objectKey, alt: file.name, sort_order: prev.length },
      ])
      message.success('Uploaded')
    } catch (err) {
      message.error(err instanceof Error ? err.message : 'upload failed')
    } finally {
      setUploading(false)
    }
    return false // 阻止 antd 默认上传行为
  }

  const submit = async (status: ProductInput['status']) => {
    const values = await form.validateFields()
    const hasZh = Object.values(zhTrans).some((v) => String(v ?? '').trim() !== '')
    const payload: ProductInput = {
      ...values,
      status,
      currency: values.currency || 'USD',
      tags: values.tags || [],
      attributes: values.attributes || {},
      variants: variants.map((v, i) => ({ ...v, sort_order: i })),
      images: images.map((img, i) => ({ ...img, sort_order: i })),
      translations: hasZh ? { zh: zhTrans } : undefined,
    }

    setSaving(true)
    try {
      if (isEdit) {
        await api.put(`/admin/products/${id}`, payload)
      } else {
        await api.post('/admin/products', payload)
      }
      message.success('Saved')
      navigate('/products')
    } catch (err) {
      message.error(err instanceof Error ? err.message : 'save failed')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div>
      <Typography.Title level={4}>
        {isEdit ? dict.product.edit : dict.product.create}
      </Typography.Title>

      <Form form={form} layout="vertical" initialValues={{ currency: 'USD', status: 'draft' }}>
        <Card title={dict.product.basicInfo}>
          <Tabs
            defaultActiveKey="en"
            items={[
              {
                key: 'en',
                label: dict.product.baseLang,
                children: (
                  <>
                    <Row gutter={16}>
                      <Col span={12}>
                        <Form.Item name="title" label={dict.product.title} rules={[{ required: true }]}>
                          <Input />
                        </Form.Item>
                      </Col>
                      <Col span={12}>
                        <Form.Item
                          name="slug"
                          label={dict.product.slug}
                          rules={[
                            { required: true },
                            { pattern: /^[a-z0-9-]+$/, message: 'lowercase, numbers and hyphens' },
                          ]}
                        >
                          <Input placeholder="a5-dot-grid-journal" />
                        </Form.Item>
                      </Col>
                    </Row>

                    <Form.Item name="subtitle" label={dict.product.subtitle}>
                      <Input />
                    </Form.Item>

                    <Form.Item name="description" label={dict.product.description}>
                      <Input.TextArea rows={5} />
                    </Form.Item>

                    <Row gutter={16}>
                      <Col span={8}>
                        <Form.Item name="category_id" label={dict.product.category}>
                          <Select
                            allowClear
                            placeholder={dict.product.categoryPlaceholder}
                            options={categories.map((c) => ({ value: c.id, label: c.name }))}
                          />
                        </Form.Item>
                      </Col>
                      <Col span={8}>
                        <Form.Item name="currency" label={dict.product.currency}>
                          <Select
                            options={[
                              { value: 'USD', label: 'USD' },
                              { value: 'EUR', label: 'EUR' },
                              { value: 'JPY', label: 'JPY' },
                              { value: 'GBP', label: 'GBP' },
                            ]}
                          />
                        </Form.Item>
                      </Col>
                      <Col span={8}>
                        <Form.Item name="status" label={dict.product.status}>
                          <Select
                            options={[
                              { value: 'draft', label: dict.product.draft },
                              { value: 'published', label: dict.product.published },
                              { value: 'archived', label: dict.product.archived },
                            ]}
                          />
                        </Form.Item>
                      </Col>
                    </Row>

                    <Form.Item
                      name="tags"
                      label={dict.product.tags}
                      tooltip="Used for filtering and recommendations"
                    >
                      <Select mode="tags" placeholder={dict.product.tagsPlaceholder} />
                    </Form.Item>
                  </>
                ),
              },
              {
                key: 'zh',
                label: dict.product.zhLang,
                children: (
                  <>
                    <Form.Item label={dict.product.title}>
                      <Input
                        value={zhTrans.title}
                        onChange={(e) => setZhTrans((s) => ({ ...s, title: e.target.value }))}
                        placeholder={dict.product.titlePlaceholder}
                      />
                    </Form.Item>
                    <Form.Item label={dict.product.subtitle}>
                      <Input
                        value={zhTrans.subtitle}
                        onChange={(e) => setZhTrans((s) => ({ ...s, subtitle: e.target.value }))}
                        placeholder={dict.product.subtitlePlaceholder}
                      />
                    </Form.Item>
                    <Form.Item label={dict.product.description}>
                      <Input.TextArea
                        rows={5}
                        value={zhTrans.description}
                        onChange={(e) => setZhTrans((s) => ({ ...s, description: e.target.value }))}
                        placeholder={dict.product.descriptionPlaceholder}
                      />
                    </Form.Item>
                    <Form.Item label={dict.product.seoTitle}>
                      <Input
                        value={zhTrans.seo_title}
                        onChange={(e) => setZhTrans((s) => ({ ...s, seo_title: e.target.value }))}
                        placeholder={dict.product.seoTitle}
                      />
                    </Form.Item>
                    <Form.Item label={dict.product.seoDescription}>
                      <Input.TextArea
                        rows={2}
                        value={zhTrans.seo_description}
                        onChange={(e) => setZhTrans((s) => ({ ...s, seo_description: e.target.value }))}
                        placeholder={dict.product.seoDescription}
                      />
                    </Form.Item>
                  </>
                ),
              },
            ]}
          />
        </Card>

        <Card title={dict.product.images} style={{ marginTop: 16 }}>
          <Upload
            listType="picture-card"
            beforeUpload={beforeUpload}
            showUploadList={false}
            accept="image/jpeg,image/png,image/webp,image/avif"
          >
            <div>
              <PlusOutlined />
              <div style={{ marginTop: 8 }}>{uploading ? 'Uploading...' : 'Upload'}</div>
            </div>
          </Upload>

          <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
            {images.map((img, idx) => (
              <Col key={img.object_key} span={6}>
                <Card
                  size="small"
                  cover={
                    <img
                      alt={img.alt}
                      src={
                        img.object_key.startsWith('http')
                          ? img.object_key
                          : `${IMG_BASE}/${img.object_key}`
                      }
                      style={{ height: 140, objectFit: 'cover' }}
                    />
                  }
                  actions={[
                    <DeleteOutlined
                      key="del"
                      onClick={() =>
                        setImages((prev) => prev.filter((_, i) => i !== idx))
                      }
                    />,
                  ]}
                >
                  <Input
                    size="small"
                    placeholder="Alt text"
                    value={img.alt}
                    onChange={(e) =>
                      setImages((prev) =>
                        prev.map((it, i) =>
                          i === idx ? { ...it, alt: e.target.value } : it,
                        ),
                      )
                    }
                  />
                </Card>
              </Col>
            ))}
          </Row>
        </Card>

        <Card
          title={dict.product.variants}
          style={{ marginTop: 16 }}
          extra={
            <Button
              icon={<PlusOutlined />}
              onClick={() => setVariants((v) => [...v, emptyVariant()])}
            >
              {dict.product.addVariant}
            </Button>
          }
        >
          {variants.map((v, idx) => (
            <div key={idx} style={{ marginBottom: 16 }}>
              <Divider orientation="left">Variant {idx + 1}</Divider>
              <Row gutter={16}>
                <Col span={6}>
                  <Form.Item label="SKU Code" tooltip="留空将自动生成（基于 slug + 序号）">
                    <Input
                      placeholder="Auto if empty"
                      value={v.sku_code}
                      onChange={(e) => updateVariant(idx, 'sku_code', e.target.value)}
                    />
                  </Form.Item>
                </Col>
                <Col span={6}>
                  <Form.Item label="Title">
                    <Input
                      value={v.title}
                      onChange={(e) => updateVariant(idx, 'title', e.target.value)}
                    />
                  </Form.Item>
                </Col>
                <Col span={6}>
                  <Form.Item label="Price (USD)">
                    <InputNumber
                      min={0}
                      step={0.01}
                      style={{ width: '100%' }}
                      value={v.price_cents / 100}
                      onChange={(val) =>
                        updateVariant(idx, 'price_cents', Math.round((val || 0) * 100))
                      }
                    />
                  </Form.Item>
                </Col>
                <Col span={6}>
                  <Form.Item label="Compare At (USD)">
                    <InputNumber
                      min={0}
                      step={0.01}
                      style={{ width: '100%' }}
                      value={v.compare_at_cents / 100}
                      onChange={(val) =>
                        updateVariant(idx, 'compare_at_cents', Math.round((val || 0) * 100))
                      }
                    />
                  </Form.Item>
                </Col>
              </Row>
              <Row gutter={16}>
                <Col span={6}>
                  <Form.Item label="Weight (g)">
                    <InputNumber
                      min={0}
                      style={{ width: '100%' }}
                      value={v.weight_g}
                      onChange={(val) => updateVariant(idx, 'weight_g', val || 0)}
                    />
                  </Form.Item>
                </Col>
                <Col span={6}>
                  <Form.Item label="Stock">
                    <InputNumber
                      min={0}
                      style={{ width: '100%' }}
                      value={v.stock}
                      onChange={(val) => updateVariant(idx, 'stock', val || 0)}
                    />
                  </Form.Item>
                </Col>
                <Col span={10}>
                  <Form.Item
                    label="Options (JSON)"
                    validateStatus={
                      optionsText[idx] !== undefined && !isValidJson(optionsText[idx])
                        ? 'error'
                        : undefined
                    }
                    help={
                      optionsText[idx] !== undefined && !isValidJson(optionsText[idx])
                        ? 'Invalid JSON'
                        : undefined
                    }
                  >
                    <Input
                      placeholder='{"cover":"Linen","size":"A5"}'
                      value={optionsText[idx] ?? JSON.stringify(v.options || {})}
                      onChange={(e) => {
                        const text = e.target.value
                        setOptionsText((s) => ({ ...s, [idx]: text }))
                        // 输入合法 JSON 时才同步到 v.options；非法时保留原文，允许继续编辑
                        try {
                          const parsed = JSON.parse(text || '{}')
                          if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
                            updateVariant(idx, 'options', parsed)
                          }
                        } catch {
                          /* 允许编辑过程中出现非法 JSON */
                        }
                      }}
                    />
                  </Form.Item>
                </Col>
                <Col span={2}>
                  <Button
                    danger
                    icon={<DeleteOutlined />}
                    onClick={() =>
                      setVariants((prev) => prev.filter((_, i) => i !== idx))
                    }
                  />
                </Col>
              </Row>
            </div>
          ))}
        </Card>
      </Form>

      <Space style={{ marginTop: 24 }}>
        <Button type="primary" loading={saving} onClick={() => submit('published')}>
          {dict.product.savePublish}
        </Button>
        <Button loading={saving} onClick={() => submit('draft')}>
          {dict.product.saveDraft}
        </Button>
        <Button onClick={() => navigate('/products')}>{dict.product.cancel}</Button>
      </Space>
    </div>
  )

  function updateVariant(idx: number, key: keyof VariantInput, value: unknown) {
    setVariants((prev) =>
      prev.map((v, i) => (i === idx ? { ...v, [key]: value } : v)),
    )
  }
}

function emptyVariant(): VariantInput {
  return {
    sku_code: '',
    title: '',
    options: {},
    price_cents: 0,
    compare_at_cents: 0,
    weight_g: 0,
    image_key: '',
    stock: 0,
    sort_order: 0,
  }
}
