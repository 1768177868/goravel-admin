
// ====================== 通用枚举 ======================
/**
 * @param {Function} t - vue-i18n 翻译函数
 * @returns {Array}
 */
export const getShowHideOptions = (t) => {
  return [
    { label: t('menu_management.is_hidden_show'), value: 0 },
    { label: t('menu_management.is_hidden_hide'), value: 1 }
  ]
}


export const getEnableDisableOptions = (t) => {
  return [
    { label: t('common.enabled'), value: 1 },
    { label: t('common.disabled'), value: 0 }
  ]
}

/**
 * @param {Function} t - vue-i18n 翻译函数
 * @returns {Array}
 */
export const getOpenTypeOptions = (t) => {
  return [
    { label: t('menu_management.open_type_iframe'), value: 1 },
    { label: t('menu_management.open_type_new_window'), value: 2 }
  ]
}


export const getMenuLinkTypeOptions = (t) => {
  return [
    { label: t('menu_management.link_type_internal'), value: 1 },
    { label: t('menu_management.link_type_external'), value: 2 }
  ]
}

export const getMethodOptions = (t) => {
  return [
    { label: 'GET', value: 'GET' },
    { label: 'POST', value: 'POST' },
    { label: 'PUT', value: 'PUT' },
    { label: 'DELETE', value: 'DELETE' }
  ]
}