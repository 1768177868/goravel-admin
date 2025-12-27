<template>
  <el-dialog
    v-model="dialogVisible"
    :title="dialogTitle"
    width="900px"
    @close="handleDialogClose"
  >
    <div v-loading="loading">
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="120px"
      >
        <el-form-item :label="$t('order.user_id')" prop="user_id">
          <el-input-number
            v-model="formData.user_id"
            :min="1"
            :disabled="loading"
            style="width: 100%"
          />
        </el-form-item>

        <el-form-item :label="$t('order.products')" prop="products">
          <div style="width: 100%">
            <el-button
              type="primary"
              size="small"
              :disabled="loading"
              @click="handleAddProduct"
            >
              <el-icon><Plus /></el-icon>
              {{ $t('order.add_product') }}
            </el-button>
            
            <el-table
              :data="formData.products"
              border
              style="width: 100%; margin-top: 10px"
              :max-height="400"
            >
              <el-table-column :label="$t('order.product_id')" width="130" align="center">
                <template #default="{ row, $index }">
                  <el-input-number
                    v-model="row.product_id"
                    :min="1"
                    :disabled="loading"
                    :controls="false"
                    style="width: 100%"
                    @change="calculateAmount"
                  />
                </template>
              </el-table-column>
              <el-table-column :label="$t('order.product_name')" min-width="200">
                <template #default="{ row, $index }">
                  <el-input
                    v-model="row.product_name"
                    :disabled="loading"
                    :placeholder="$t('order.product_name_placeholder')"
                    clearable
                  />
                </template>
              </el-table-column>
              <el-table-column :label="$t('order.price')" width="140" align="center">
                <template #default="{ row, $index }">
                  <el-input-number
                    v-model="row.price"
                    :min="0"
                    :precision="2"
                    :disabled="loading"
                    :controls="false"
                    style="width: 100%"
                    @change="calculateAmount"
                  />
                </template>
              </el-table-column>
              <el-table-column :label="$t('order.quantity')" width="120" align="center">
                <template #default="{ row, $index }">
                  <el-input-number
                    v-model="row.quantity"
                    :min="1"
                    :disabled="loading"
                    :controls="false"
                    style="width: 100%"
                    @change="calculateAmount"
                  />
                </template>
              </el-table-column>
              <el-table-column :label="$t('order.subtotal')" width="140" align="right">
                <template #default="{ row }">
                  <span style="font-weight: bold; color: #409eff;">
                    {{ formatAmount((row.price || 0) * (row.quantity || 0)) }}
                  </span>
                </template>
              </el-table-column>
              <el-table-column :label="$t('table.operation')" width="100" align="center" fixed="right">
                <template #default="{ $index }">
                  <el-button
                    type="danger"
                    size="small"
                    :disabled="loading"
                    @click="handleRemoveProduct($index)"
                  >
                    {{ $t('common.delete') }}
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-form-item>

        <el-form-item :label="$t('order.amount')">
          <el-input
            :model-value="formatAmount(totalAmount)"
            disabled
            style="width: 100%"
          />
        </el-form-item>

        <el-form-item :label="$t('order.remark')">
          <el-input
            v-model="formData.remark"
            type="textarea"
            :rows="3"
            :disabled="loading"
            :placeholder="$t('order.remark_placeholder')"
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
import { Plus } from '@element-plus/icons-vue'
import { createOrder } from '../../api/order'
import logger from '../../utils/logger'
import ErrorHandler from '../../utils/errorHandler'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue', 'success'])

const { t } = useI18n()
const formRef = ref(null)
const submitting = ref(false)
const loading = ref(false)

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const dialogTitle = computed(() => t('order.add_order'))

const formData = reactive({
  user_id: null,
  products: [],
  remark: ''
})

// 计算总金额
const totalAmount = computed(() => {
  return formData.products.reduce((sum, product) => {
    const price = product.price || 0
    const quantity = product.quantity || 0
    return sum + (price * quantity)
  }, 0)
})

// 格式化金额
const formatAmount = (amount) => {
  return `¥${Number(amount).toFixed(2)}`
}

// 计算总金额（手动触发）
const calculateAmount = () => {
  // 总金额是计算属性，会自动更新
}

// 添加商品（确保每次添加都是全新的空对象）
const handleAddProduct = () => {
  formData.products.push({
    product_id: null,
    product_name: '',
    price: 0,
    quantity: 1
  })
  // 触发金额重新计算
  calculateAmount()
}

// 删除商品
const handleRemoveProduct = (index) => {
  formData.products.splice(index, 1)
  calculateAmount()
}

const formRules = computed(() => ({
  user_id: [
    { required: true, message: t('order.user_id_required'), trigger: 'blur' },
    { type: 'number', min: 1, message: t('order.user_id_min'), trigger: 'blur' }
  ],
  products: [
    {
      required: true,
      validator: (rule, value, callback) => {
        if (!value || value.length === 0) {
          callback(new Error(t('order.products_required')))
        } else {
          // 检查每个商品是否完整
          for (let i = 0; i < value.length; i++) {
            const product = value[i]
            if (!product.product_id || product.product_id <= 0) {
              callback(new Error(t('order.product_id_required', { index: i + 1 })))
              return
            }
            if (!product.product_name || product.product_name.trim() === '') {
              callback(new Error(t('order.product_name_required', { index: i + 1 })))
              return
            }
            if (!product.price || product.price <= 0) {
              callback(new Error(t('order.price_required', { index: i + 1 })))
              return
            }
            if (!product.quantity || product.quantity <= 0) {
              callback(new Error(t('order.quantity_required', { index: i + 1 })))
              return
            }
          }
          callback()
        }
      },
      trigger: 'change'
    }
  ]
}))

// 监听 dialogVisible 变化
watch(dialogVisible, (visible) => {
  if (visible) {
    resetForm()
  }
})

const resetForm = () => {
  loading.value = false
  // 完全重置表单数据
  formData.user_id = null
  formData.remark = ''
  // 完全清空商品数组（使用 length = 0 更彻底）
  formData.products.length = 0
  // 默认添加一个空的商品行（用于新订单）
  formData.products.push({
    product_id: null,
    product_name: '',
    price: 0,
    quantity: 1
  })
  // 重置表单验证状态
  formRef.value?.resetFields()
  // 清除表单验证错误
  formRef.value?.clearValidate()
  // 重新计算金额
  calculateAmount()
}

const handleDialogClose = () => {
  // 关闭对话框时重置表单（如果还没有重置的话）
  if (formData.products.length > 0) {
    resetForm()
  }
  emit('update:modelValue', false)
}

const handleCancel = () => {
  handleDialogClose()
}

const handleSubmit = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
  } catch (error) {
    return
  }

  // 验证总金额
  if (totalAmount.value <= 0) {
    ElMessage.warning(t('order.amount_required'))
    return
  }

  submitting.value = true
  try {
    const requestData = {
      user_id: formData.user_id,
      amount: totalAmount.value,
      products: formData.products.map(product => ({
        product_id: product.product_id,
        product_name: product.product_name,
        price: product.price,
        quantity: product.quantity
      })),
      remark: formData.remark || ''
    }

    await createOrder(requestData)
    ElMessage.success(t('order.create_success'))
    // 立即重置表单（清空商品列表等）
    resetForm()
    emit('success')
    handleDialogClose()
  } catch (error) {
    logger.error('Create order error:', error)
    ErrorHandler.handle(error, { silent: true })
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
</style>

