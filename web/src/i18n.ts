import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import en from './locales/en'
import zh from './locales/zh'

// 默认中文；用户切换后持久化到 localStorage
const saved = localStorage.getItem('cadentra_lang')
const lang = saved === 'en' || saved === 'zh' ? saved : 'zh'

i18n.use(initReactI18next).init({
  resources: {
    zh: { translation: zh },
    en: { translation: en },
  },
  lng: lang,
  fallbackLng: 'zh',
  interpolation: { escapeValue: false },
})

export function setLang(l: 'zh' | 'en') {
  localStorage.setItem('cadentra_lang', l)
  i18n.changeLanguage(l)
}

export default i18n
