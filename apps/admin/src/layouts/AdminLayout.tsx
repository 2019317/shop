import { useState } from 'react'
import { Outlet, useNavigate } from 'react-router-dom'
import { Layout, Menu, Button, Select, theme } from 'antd'
import {
  AppstoreOutlined,
  TagsOutlined,
  ShoppingCartOutlined,
  LogoutOutlined,
} from '@ant-design/icons'
import { clearToken } from '../api/client'
import { useI18n } from '../i18n/I18nProvider'
import { localeNames, locales } from '../i18n'

const { Header, Sider, Content } = Layout

export default function AdminLayout() {
  const [collapsed, setCollapsed] = useState(false)
  const navigate = useNavigate()
  const { dict, locale, setLocale } = useI18n()
  const {
    token: { colorBgContainer },
  } = theme.useToken()

  const handleLogout = () => {
    clearToken()
    navigate('/login', { replace: true })
  }

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider collapsible collapsed={collapsed} onCollapse={setCollapsed}>
        <div
          style={{
            height: 48,
            margin: 16,
            color: '#fff',
            fontSize: 16,
            fontWeight: 600,
            textAlign: 'center',
            lineHeight: '48px',
            whiteSpace: 'nowrap',
            overflow: 'hidden',
          }}
        >
          {collapsed ? 'S' : dict.brand}
        </div>
        <Menu
          theme="dark"
          mode="inline"
          defaultSelectedKeys={['products']}
          items={[
            { key: 'products', icon: <AppstoreOutlined />, label: dict.menu.products },
            { key: 'orders', icon: <ShoppingCartOutlined />, label: dict.menu.orders },
            { key: 'categories', icon: <TagsOutlined />, label: dict.menu.categories },
          ]}
          onClick={({ key }) => navigate(`/${key}`)}
        />
      </Sider>

      <Layout>
        <Header
          style={{
            padding: '0 24px',
            background: colorBgContainer,
            display: 'flex',
            justifyContent: 'flex-end',
            alignItems: 'center',
            gap: 12,
          }}
        >
          <Select
            value={locale}
            onChange={setLocale}
            style={{ width: 110 }}
            options={locales.map((l) => ({ value: l, label: localeNames[l] }))}
          />
          <Button icon={<LogoutOutlined />} onClick={handleLogout}>
            {dict.logout}
          </Button>
        </Header>
        <Content style={{ margin: 24 }}>
          <div
            style={{
              padding: 24,
              minHeight: 360,
              background: colorBgContainer,
              borderRadius: 8,
            }}
          >
            <Outlet />
          </div>
        </Content>
      </Layout>
    </Layout>
  )
}
