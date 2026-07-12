export const paymentMethodInitialSearchForm = {
  name: '',
  code: '',
  type: '',
  is_active: '',
  description: ''
}

export function createPaymentMethodSearchFields(t) {
  return [
    { prop: 'name', type: 'input', placeholder: t('payment_method.name_placeholder') },
    { prop: 'code', type: 'input', placeholder: t('payment_method.code_placeholder') },
    {
      prop: 'type',
      type: 'select',
      placeholder: t('payment_method.type_placeholder'),
      options: [
        { label: t('payment_method.type_wechat'), value: 'wechat' },
        { label: t('payment_method.type_alipay'), value: 'alipay' },
        { label: t('payment_method.type_qq'), value: 'qq' },
        { label: t('payment_method.type_allinpay'), value: 'allinpay' },
        { label: t('payment_method.type_lakala'), value: 'lakala' },
        { label: t('payment_method.type_paypal'), value: 'paypal' },
        { label: t('payment_method.type_apple'), value: 'apple' },
        { label: t('payment_method.type_saobei'), value: 'saobei' }
      ]
    },
    {
      prop: 'is_active',
      type: 'select',
      placeholder: t('payment_method.is_active_placeholder'),
      options: [
        { label: t('common.enabled'), value: '1' },
        { label: t('common.disabled'), value: '0' }
      ]
    },
    { prop: 'description', type: 'input', placeholder: t('payment_method.description_placeholder') }
  ]
}

export function createPaymentMethodTableColumns(t) {
  return [
    { field: 'id', title: t('table.id'), width: 80, sortable: true, key: 'id' },
    { field: 'name', title: t('payment_method.name'), minWidth: 150, sortable: true, key: 'name' },
    { field: 'code', title: t('payment_method.code'), width: 120, sortable: true, key: 'code' },
    { field: 'type', title: t('payment_method.type'), width: 120, sortable: true, key: 'type' },
    { field: 'is_active', title: t('table.status'), width: 100, slot: 'is_active', key: 'is_active' },
    { field: 'sort', title: t('table.sort'), width: 100, sortable: true, key: 'sort' },
    { field: 'description', title: t('table.description'), minWidth: 200, key: 'description' },
    { field: 'created_at', title: t('table.created_at'), width: 180, sortable: true, key: 'created_at' },
    { title: t('table.operation'), width: 150, fixed: 'right', slot: 'operation', sortable: false, key: 'operation' }
  ]
}
