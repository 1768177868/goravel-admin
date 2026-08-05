import type { TFunction } from 'i18next'

export type PaymentMethodType =
  | 'wechat'
  | 'alipay'
  | 'qq'
  | 'allinpay'
  | 'lakala'
  | 'paypal'
  | 'apple'
  | 'saobei'

export interface PaymentMethodSearchForm {
  name: string
  code: string
  type: string
  is_active: string | number
  description: string
  [key: string]: string | number
}

export interface PaymentMethodConfigField {
  key: string
  label: string
  type: 'input' | 'textarea'
  inputType?: 'password' | 'text'
  required?: boolean
  rows?: number
  placeholder?: string
}

export const paymentMethodInitialSearchForm: PaymentMethodSearchForm = {
  name: '',
  code: '',
  type: '',
  is_active: '',
  description: '',
}

export const PAYMENT_TYPE_CONFIG_FIELDS: Record<PaymentMethodType, PaymentMethodConfigField[]> = {
  wechat: [
    { key: 'app_id', label: 'AppID', type: 'input', required: true, placeholder: '请输入微信AppID' },
    {
      key: 'app_secret',
      label: 'AppSecret',
      type: 'input',
      inputType: 'password',
      required: true,
      placeholder: '请输入微信AppSecret',
    },
    { key: 'mch_id', label: '商户号(MchID)', type: 'input', required: true, placeholder: '请输入商户号' },
    {
      key: 'api_key',
      label: 'API密钥',
      type: 'input',
      inputType: 'password',
      required: true,
      placeholder: '请输入API密钥',
    },
    { key: 'cert_path', label: '证书路径', type: 'input', placeholder: '请输入证书路径（可选）' },
    { key: 'notify_url', label: '通知地址', type: 'input', placeholder: '请输入支付通知回调地址（可选）' },
  ],
  alipay: [
    { key: 'app_id', label: 'AppID', type: 'input', required: true, placeholder: '请输入支付宝AppID' },
    {
      key: 'private_key',
      label: '应用私钥',
      type: 'textarea',
      rows: 5,
      required: true,
      placeholder: '请输入应用私钥',
    },
    {
      key: 'public_key',
      label: '支付宝公钥',
      type: 'textarea',
      rows: 5,
      required: true,
      placeholder: '请输入支付宝公钥',
    },
    { key: 'gateway', label: '网关地址', type: 'input', placeholder: '请输入网关地址（可选）' },
    { key: 'notify_url', label: '通知地址', type: 'input', placeholder: '请输入支付通知回调地址（可选）' },
  ],
  qq: [
    { key: 'app_id', label: 'AppID', type: 'input', required: true, placeholder: '请输入QQ支付AppID' },
    {
      key: 'app_key',
      label: 'AppKey',
      type: 'input',
      inputType: 'password',
      required: true,
      placeholder: '请输入QQ支付AppKey',
    },
    { key: 'mch_id', label: '商户号', type: 'input', required: true, placeholder: '请输入商户号' },
    { key: 'notify_url', label: '通知地址', type: 'input', placeholder: '请输入支付通知回调地址（可选）' },
  ],
  allinpay: [
    { key: 'merchant_id', label: '商户号', type: 'input', required: true, placeholder: '请输入通联商户号' },
    { key: 'app_id', label: 'AppID', type: 'input', required: true, placeholder: '请输入通联AppID' },
    {
      key: 'app_key',
      label: 'AppKey',
      type: 'input',
      inputType: 'password',
      required: true,
      placeholder: '请输入通联AppKey',
    },
    { key: 'gateway', label: '网关地址', type: 'input', placeholder: '请输入网关地址（可选）' },
    { key: 'notify_url', label: '通知地址', type: 'input', placeholder: '请输入支付通知回调地址（可选）' },
  ],
  lakala: [
    { key: 'merchant_id', label: '商户号', type: 'input', required: true, placeholder: '请输入拉卡拉商户号' },
    { key: 'terminal_id', label: '终端号', type: 'input', required: true, placeholder: '请输入终端号' },
    {
      key: 'key',
      label: '密钥',
      type: 'input',
      inputType: 'password',
      required: true,
      placeholder: '请输入密钥',
    },
    { key: 'gateway', label: '网关地址', type: 'input', placeholder: '请输入网关地址（可选）' },
    { key: 'notify_url', label: '通知地址', type: 'input', placeholder: '请输入支付通知回调地址（可选）' },
  ],
  paypal: [
    { key: 'client_id', label: 'Client ID', type: 'input', required: true, placeholder: '请输入PayPal Client ID' },
    {
      key: 'client_secret',
      label: 'Client Secret',
      type: 'input',
      inputType: 'password',
      required: true,
      placeholder: '请输入PayPal Client Secret',
    },
    { key: 'mode', label: '模式', type: 'input', placeholder: '请输入模式（sandbox/live，可选）' },
    { key: 'notify_url', label: '通知地址', type: 'input', placeholder: '请输入支付通知回调地址（可选）' },
  ],
  apple: [
    { key: 'merchant_id', label: '商户ID', type: 'input', required: true, placeholder: '请输入Apple Pay商户ID' },
    { key: 'key_id', label: 'Key ID', type: 'input', required: true, placeholder: '请输入Key ID' },
    {
      key: 'private_key',
      label: '私钥',
      type: 'textarea',
      rows: 5,
      required: true,
      placeholder: '请输入私钥',
    },
    { key: 'certificate', label: '证书', type: 'textarea', rows: 5, placeholder: '请输入证书（可选）' },
  ],
  saobei: [
    { key: 'merchant_id', label: '商户号', type: 'input', required: true, placeholder: '请输入扫呗商户号' },
    { key: 'terminal_id', label: '终端号', type: 'input', required: true, placeholder: '请输入终端号' },
    {
      key: 'key',
      label: '密钥',
      type: 'input',
      inputType: 'password',
      required: true,
      placeholder: '请输入密钥',
    },
    { key: 'gateway', label: '网关地址', type: 'input', placeholder: '请输入网关地址（可选）' },
    { key: 'notify_url', label: '通知地址', type: 'input', placeholder: '请输入支付通知回调地址（可选）' },
  ],
}

export function createPaymentMethodTypeOptions(t: TFunction) {
  return [
    { label: t('payment_method.type_wechat'), value: 'wechat' },
    { label: t('payment_method.type_alipay'), value: 'alipay' },
    { label: t('payment_method.type_qq'), value: 'qq' },
    { label: t('payment_method.type_allinpay'), value: 'allinpay' },
    { label: t('payment_method.type_lakala'), value: 'lakala' },
    { label: t('payment_method.type_paypal'), value: 'paypal' },
    { label: t('payment_method.type_apple'), value: 'apple' },
    { label: t('payment_method.type_saobei'), value: 'saobei' },
  ]
}

export function getPaymentMethodTypeLabel(t: TFunction, type?: string): string {
  const map: Record<string, string> = {
    wechat: t('payment_method.type_wechat'),
    alipay: t('payment_method.type_alipay'),
    qq: t('payment_method.type_qq'),
    allinpay: t('payment_method.type_allinpay'),
    lakala: t('payment_method.type_lakala'),
    paypal: t('payment_method.type_paypal'),
    apple: t('payment_method.type_apple'),
    saobei: t('payment_method.type_saobei'),
  }
  return map[type || ''] || String(type ?? '-')
}

export function getConfigFieldsForType(type?: string): PaymentMethodConfigField[] {
  if (!type) return []
  return PAYMENT_TYPE_CONFIG_FIELDS[type as PaymentMethodType] || []
}

export function createEmptyConfig(type?: string): Record<string, string> {
  const fields = getConfigFieldsForType(type)
  return Object.fromEntries(fields.map((field) => [field.key, '']))
}
