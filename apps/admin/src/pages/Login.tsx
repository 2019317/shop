import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Card, Form, Input, Button, App, Typography, Select } from 'antd'
import { api, setToken, type LoginResult } from '../api/client'
import { useI18n } from '../i18n/I18nProvider'
import { localeNames, locales } from '../i18n'

export default function Login() {
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const { message } = App.useApp()
  const { dict, locale, setLocale } = useI18n()

  const onFinish = async (values: { email: string; password: string }) => {
    setLoading(true)
    try {
      const data = await api.post<LoginResult>('/admin/auth/login', values)
      setToken(data.token)
      message.success(dict.login.success)
      navigate('/products', { replace: true })
    } catch (err) {
      message.error(err instanceof Error ? err.message : 'login failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div
      style={{
        height: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      <Card style={{ width: 380 }}>
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
          }}
        >
          <Typography.Title level={4} style={{ margin: 0 }}>
            {dict.login.title}
          </Typography.Title>
          <Select
            value={locale}
            onChange={setLocale}
            size="small"
            style={{ width: 100 }}
            options={locales.map((l) => ({ value: l, label: localeNames[l] }))}
          />
        </div>

        <Form
          layout="vertical"
          onFinish={onFinish}
          autoComplete="off"
          style={{ marginTop: 16 }}
        >
          <Form.Item
            name="email"
            label={dict.login.email}
            rules={[
              { required: true, message: dict.error.required },
              { type: 'email', message: 'Invalid email' },
            ]}
          >
            <Input placeholder={dict.login.emailPlaceholder} />
          </Form.Item>
          <Form.Item
            name="password"
            label={dict.login.password}
            rules={[{ required: true, message: dict.error.required }]}
          >
            <Input.Password placeholder="********" />
          </Form.Item>
          <Button type="primary" htmlType="submit" block loading={loading}>
            {dict.login.submit}
          </Button>
        </Form>
      </Card>
    </div>
  )
}
