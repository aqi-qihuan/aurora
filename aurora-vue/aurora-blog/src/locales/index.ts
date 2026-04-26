import { createI18n } from 'vue-i18n'
import cookies from 'js-cookie'

function loadLocaleMessages(): {
  [key: string]: { [key: string]: { [key: string]: string } }
} {
  const messages: {
    [key: string]: { [key: string]: { [key: string]: string } }
  } = {}
  
  // Vite 使用 import.meta.glob 替代 require.context
  const modules = import.meta.glob('./languages/*.json', { eager: true })
  
  Object.keys(modules).forEach((path) => {
    const matched = path.match(/([A-Za-z0-9-_]+)\.json$/i)
    if (matched && matched.length > 1) {
      const locale = matched[1]
      messages[locale] = (modules[path] as any).default
    }
  })
  return messages
}

export const i18n = createI18n({
  locale: cookies.get('locale') ? String(cookies.get('locale')) : 'en',
  fallbackLocale: cookies.get('locale') ? String(cookies.get('locale')) : 'en',
  messages: loadLocaleMessages()
})
