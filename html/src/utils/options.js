
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