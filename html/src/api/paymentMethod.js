import request from '../utils/request'
import { createCRUDApi } from '../utils/apiFactory'

// 创建基础 CRUD API
const basePaymentMethodApi = createCRUDApi('payment-methods')

// 导出所有方法
export const {
  list: getPaymentMethodList,
  detail: getPaymentMethodDetail,
  create: createPaymentMethod,
  update: updatePaymentMethod,
  delete: deletePaymentMethod
} = basePaymentMethodApi

