import { ConfigProvider, App as AntdApp } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import enUS from 'antd/locale/en_US'
import { useI18n } from './I18nProvider'

// antd 组件内置文案（分页、空状态、日期等）随界面语言切换
export default function AntdLocaleWrapper({
  children,
}: {
  children: React.ReactNode
}) {
  const { locale } = useI18n()

  return (
    <ConfigProvider
      locale={locale === 'zh' ? zhCN : enUS}
      theme={{ token: { colorPrimary: '#8b5e3c' } }}
    >
      <AntdApp>{children}</AntdApp>
    </ConfigProvider>
  )
}
