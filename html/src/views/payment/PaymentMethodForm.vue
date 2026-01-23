<template>
  <el-dialog
    v-model="dialogVisible"
    :title="dialogTitle"
    width="800px"
    @close="handleDialogClose"
  >
    <div v-loading="loading">
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="120px"
      >
        <FormField
          v-for="f in formFields"
          :key="f.prop"
          :field="f"
          :model="formData"
        />
        
        <!-- 动态配置字段（单独处理，因为需要动态生成多个表单项） -->
        <template v-if="formData.type || formData.id">
          <el-divider content-position="left">{{ $t('payment_method.config') }}</el-divider>
          <template v-for="field in configFields" :key="field.key">
            <el-form-item 
              :label="field.label" 
              :prop="`config.${field.key}`"
              :rules="field.required ? [{ required: true, message: field.label + '不能为空', trigger: 'blur' }] : []"
            >
              <el-input
                v-if="field.type === 'input'"
                v-model="formData.config[field.key]"
                :type="field.inputType || 'text'"
                :disabled="loading"
                :placeholder="field.placeholder || `请输入${field.label}`"
                :show-password="field.inputType === 'password'"
              />
              <el-input
                v-else-if="field.type === 'textarea'"
                v-model="formData.config[field.key]"
                type="textarea"
                :rows="field.rows || 3"
                :disabled="loading"
                :placeholder="field.placeholder || `请输入${field.label}`"
              />
            </el-form-item>
          </template>
        </template>
        
        <!-- JSON 编辑模式（备用） -->
        <el-form-item v-else :label="$t('payment_method.config')" prop="config">
          <el-input
            v-model="configJson"
            type="textarea"
            :rows="10"
            :disabled="loading"
            :placeholder="$t('payment_method.config_placeholder')"
          />
        </el-form-item>
      </el-form>
    </div>
    <template #footer>
      <el-button @click="handleCancel">{{ $t('common.cancel') }}</el-button>
      <el-button type="primary" @click="handleSubmit" :loading="submitting">
        {{ $t('common.confirm') }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import FormField from '../../components/Form/FormField.vue'
import {
  getPaymentMethodDetail,
  createPaymentMethod,
  updatePaymentMethod
} from '../../api/paymentMethod'
import { mapFields, getField } from '../../utils/normalizeFormData'
import logger from '../../utils/logger'
import ErrorHandler from '../../utils/errorHandler'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  editId: {
    type: [Number, String],
    default: null
  }
})

const emit = defineEmits(['update:modelValue', 'success'])

const { t } = useI18n()
const formRef = ref(null)
const submitting = ref(false)
const loading = ref(false)

// 定义表单初始值的复用函数（返回新对象，避免引用问题）
const getFormInitialValue = () => ({
  id: null,
  name: '',
  code: '',
  type: '',
  config: {},
  is_active: true,
  sort: 0,
  description: ''
})

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const dialogTitle = computed(() => 
  formData.id ? t('payment_method.edit_payment_method') : t('payment_method.add_payment_method')
)

const formData = reactive(getFormInitialValue())

// 配置 JSON 字符串（用于备用）
const configJson = ref('')

// 支付类型配置字段定义
const paymentTypeConfigFields = {
  wechat: [
    { key: 'app_id', label: 'AppID', type: 'input', required: true, placeholder: '请输入微信AppID' },
    { key: 'app_secret', label: 'AppSecret', type: 'input', inputType: 'password', required: true, placeholder: '请输入微信AppSecret' },
    { key: 'mch_id', label: '商户号(MchID)', type: 'input', required: true, placeholder: '请输入商户号' },
    { key: 'api_key', label: 'API密钥', type: 'input', inputType: 'password', required: true, placeholder: '请输入API密钥' },
    { key: 'cert_path', label: '证书路径', type: 'input', required: false, placeholder: '请输入证书路径（可选）' },
    { key: 'notify_url', label: '通知地址', type: 'input', required: false, placeholder: '请输入支付通知回调地址（可选）' }
  ],
  alipay: [
    { key: 'app_id', label: 'AppID', type: 'input', required: true, placeholder: '请输入支付宝AppID' },
    { key: 'private_key', label: '应用私钥', type: 'textarea', rows: 5, required: true, placeholder: '请输入应用私钥' },
    { key: 'public_key', label: '支付宝公钥', type: 'textarea', rows: 5, required: true, placeholder: '请输入支付宝公钥' },
    { key: 'gateway', label: '网关地址', type: 'input', required: false, placeholder: '请输入网关地址（可选）' },
    { key: 'notify_url', label: '通知地址', type: 'input', required: false, placeholder: '请输入支付通知回调地址（可选）' }
  ],
  qq: [
    { key: 'app_id', label: 'AppID', type: 'input', required: true, placeholder: '请输入QQ支付AppID' },
    { key: 'app_key', label: 'AppKey', type: 'input', inputType: 'password', required: true, placeholder: '请输入QQ支付AppKey' },
    { key: 'mch_id', label: '商户号', type: 'input', required: true, placeholder: '请输入商户号' },
    { key: 'notify_url', label: '通知地址', type: 'input', required: false, placeholder: '请输入支付通知回调地址（可选）' }
  ],
  allinpay: [
    { key: 'merchant_id', label: '商户号', type: 'input', required: true, placeholder: '请输入通联商户号' },
    { key: 'app_id', label: 'AppID', type: 'input', required: true, placeholder: '请输入通联AppID' },
    { key: 'app_key', label: 'AppKey', type: 'input', inputType: 'password', required: true, placeholder: '请输入通联AppKey' },
    { key: 'gateway', label: '网关地址', type: 'input', required: false, placeholder: '请输入网关地址（可选）' },
    { key: 'notify_url', label: '通知地址', type: 'input', required: false, placeholder: '请输入支付通知回调地址（可选）' }
  ],
  lakala: [
    { key: 'merchant_id', label: '商户号', type: 'input', required: true, placeholder: '请输入拉卡拉商户号' },
    { key: 'terminal_id', label: '终端号', type: 'input', required: true, placeholder: '请输入终端号' },
    { key: 'key', label: '密钥', type: 'input', inputType: 'password', required: true, placeholder: '请输入密钥' },
    { key: 'gateway', label: '网关地址', type: 'input', required: false, placeholder: '请输入网关地址（可选）' },
    { key: 'notify_url', label: '通知地址', type: 'input', required: false, placeholder: '请输入支付通知回调地址（可选）' }
  ],
  paypal: [
    { key: 'client_id', label: 'Client ID', type: 'input', required: true, placeholder: '请输入PayPal Client ID' },
    { key: 'client_secret', label: 'Client Secret', type: 'input', inputType: 'password', required: true, placeholder: '请输入PayPal Client Secret' },
    { key: 'mode', label: '模式', type: 'input', required: false, placeholder: '请输入模式（sandbox/live，可选）' },
    { key: 'notify_url', label: '通知地址', type: 'input', required: false, placeholder: '请输入支付通知回调地址（可选）' }
  ],
  apple: [
    { key: 'merchant_id', label: '商户ID', type: 'input', required: true, placeholder: '请输入Apple Pay商户ID' },
    { key: 'key_id', label: 'Key ID', type: 'input', required: true, placeholder: '请输入Key ID' },
    { key: 'private_key', label: '私钥', type: 'textarea', rows: 5, required: true, placeholder: '请输入私钥' },
    { key: 'certificate', label: '证书', type: 'textarea', rows: 5, required: false, placeholder: '请输入证书（可选）' }
  ],
  saobei: [
    { key: 'merchant_id', label: '商户号', type: 'input', required: true, placeholder: '请输入扫呗商户号' },
    { key: 'terminal_id', label: '终端号', type: 'input', required: true, placeholder: '请输入终端号' },
    { key: 'key', label: '密钥', type: 'input', inputType: 'password', required: true, placeholder: '请输入密钥' },
    { key: 'gateway', label: '网关地址', type: 'input', required: false, placeholder: '请输入网关地址（可选）' },
    { key: 'notify_url', label: '通知地址', type: 'input', required: false, placeholder: '请输入支付通知回调地址（可选）' }
  ]
}

// 根据支付类型获取配置字段
const configFields = computed(() => {
  const type = formData.type
  if (!type) return []
  return paymentTypeConfigFields[type] || []
})

const formFields = computed(() => {
  void formData.id // 依赖，使 id 变化时字段的 visible 能更新
  
  const fields = [
    {
      prop: 'name',
      label: t('payment_method.name'),
      type: 'input',
      disabled: loading.value
    },
    {
      prop: 'code',
      label: t('payment_method.code'),
      type: 'input',
      disabled: loading.value,
      visible: () => !formData.id
    },
    {
      prop: 'type',
      label: t('payment_method.type'),
      type: 'select',
      disabled: loading.value,
      visible: () => !formData.id,
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
      label: t('table.status'),
      type: 'switch',
      disabled: loading.value
    },
    {
      prop: 'sort',
      label: t('table.sort'),
      type: 'number',
      min: 0,
      disabled: loading.value
    },
    {
      prop: 'description',
      label: t('table.description'),
      type: 'textarea',
      rows: 3,
      disabled: loading.value,
      placeholder: t('payment_method.description_placeholder')
    }
  ]
  
  return fields
})


const formRules = computed(() => {
  const rules = {
    name: [{ required: true, message: t('payment_method.name_required'), trigger: 'blur' }],
    code: formData.id ? [] : [{ required: true, message: t('payment_method.code_required'), trigger: 'blur' }],
    type: formData.id ? [] : [{ required: true, message: t('payment_method.type_required'), trigger: 'blur' }]
  }
  
  // 动态添加配置字段验证
  if (configFields.value.length > 0) {
    configFields.value.forEach(field => {
      if (field.required) {
        rules[`config.${field.key}`] = [{ required: true, message: `${field.label}不能为空`, trigger: 'blur' }]
      }
    })
  } else {
    // 如果没有配置字段，使用JSON模式验证
    rules.config = [{ required: true, message: t('payment_method.config_required'), trigger: 'blur' }]
  }
  
  return rules
})

// 监听支付类型变化，重置配置字段
watch(() => formData.type, (newType) => {
  if (newType && !formData.id) {
    // 新建时，切换类型重置配置
    formData.config = {}
    // 初始化配置字段
    const fields = paymentTypeConfigFields[newType] || []
    fields.forEach(field => {
      formData.config[field.key] = ''
    })
  }
})

// 监听 editId 变化，加载详情
watch(() => props.editId, async (newId) => {
  if (newId && dialogVisible.value) {
    await loadDetail(newId)
  } else if (!newId && dialogVisible.value) {
    resetForm()
  }
}, { immediate: true })

// 监听 dialogVisible 变化
watch(dialogVisible, (visible) => {
  if (visible) {
    if (props.editId) {
      loadDetail(props.editId)
    } else {
      resetForm()
    }
  }
})

const loadDetail = async (id) => {
  loading.value = true
  try {
    const response = await getPaymentMethodDetail(id)
    const data = response.data?.data || response.data || {}
    
    // 使用工具函数映射字段，自动处理 snake_case 和 PascalCase
    Object.assign(formData, mapFields(data, getFormInitialValue()))
    
    // 加载配置数据
    const configData = getField(data, 'config', {})
    formData.config = {}
    
    // 初始化配置字段并填充已有值
    const fields = paymentTypeConfigFields[paymentType] || []
    fields.forEach(field => {
      // 如果后端返回了该字段的值，使用返回的值，否则使用空字符串
      formData.config[field.key] = configData[field.key] !== undefined && configData[field.key] !== null 
        ? String(configData[field.key]) 
        : ''
    })
    
    // 如果有配置数据但字段未定义，使用 JSON 模式
    if (Object.keys(configData).length > 0 && fields.length === 0) {
      configJson.value = JSON.stringify(configData, null, 2)
    } else {
      configJson.value = ''
    }
  } catch (error) {
    logger.error('Load payment method detail error:', error)
    ErrorHandler.handle(error, { silent: true })
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  loading.value = false
  Object.assign(formData, getFormInitialValue())
  configJson.value = ''
  formRef.value?.resetFields()
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      // 收集配置数据
      let config = {}
      
      if (configFields.value.length > 0) {
        // 使用动态表单字段
        configFields.value.forEach(field => {
          const value = formData.config[field.key]
          if (value !== undefined && value !== null && value !== '') {
            config[field.key] = value
          }
        })
      } else if (configJson.value.trim()) {
        // 备用：解析 JSON
        try {
          config = JSON.parse(configJson.value.trim())
        } catch (error) {
          ElMessage.error(t('payment_method.config_invalid'))
          return
        }
      }
      
      // 验证配置是否为空
      if (Object.keys(config).length === 0) {
        ElMessage.error(t('payment_method.config_required'))
        return
      }
      
      submitting.value = true
      try {
        // 只处理需要转换的字段
        const data = {
          ...formData,
          name: formData.name.trim(), // 去除首尾空格
          sort: formData.sort !== null && formData.sort !== undefined ? Number(formData.sort) : 0, // 确保是数字
          config: config, // 使用处理后的配置
          description: formData.description || '' // 空字符串处理
        }
        
        if (formData.id) {
          await updatePaymentMethod(formData.id, data)
          ElMessage.success(t('payment_method.update_success'))
        } else {
          // 创建时额外处理 code 和 type
          data.code = formData.code.trim()
          data.type = formData.type
          await createPaymentMethod(data)
          ElMessage.success(t('payment_method.create_success'))
        }
        dialogVisible.value = false
        emit('success')
      } catch (error) {
        logger.error('Submit error:', error)
        if (!error.__handled) {
          if (error.response && error.response.data && error.response.data.message) {
            ElMessage.error(error.response.data.message)
          } else if (error.message) {
            ElMessage.error(error.message)
          }
        }
      } finally {
        submitting.value = false
      }
    }
  })
}

const handleCancel = () => {
  dialogVisible.value = false
}

const handleDialogClose = () => {
  formRef.value?.resetFields()
}
</script>

