<template>
  <div class="form-demo-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ t('form_demo.title') }}</span>
          <div class="header-actions">
            <el-button @click="loadFromApi">{{ t('form_demo.load_api_data') }}</el-button>
            <el-button @click="fillSampleData">{{ t('form_demo.fill_sample') }}</el-button>
            <el-button @click="resetForm">{{ t('common.reset') }}</el-button>
          </div>
        </div>
      </template>

      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="250px"
        autocomplete="off"
      >
        <FormField
          v-for="field in formFields"
          :key="field.prop"
          :field="field"
          :model="formData"
        >
          <template v-if="field.prop === 'domain_url'">
            <div class="domain-inline">
              <el-select
                v-model="formData.domain_protocol"
                :clearable="false"
                style="width: 92px"
              >
                <el-option label="https://" value="https" />
                <el-option label="http://" value="http" />
              </el-select>
              <el-input
                v-model="formData.domain"
                placeholder="example.com"
                clearable
              />
            </div>
          </template>
          <template v-if="field.prop === 'custom_note'">
            <el-input
              v-model="formData.custom_note"
              :placeholder="t('form_demo.custom_note_placeholder')"
              clearable
            />
            <div class="custom-tip">{{ t('form_demo.custom_slot_tip') }}</div>
          </template>
          <template v-if="field.prop === 'rich_content_wang'">
            <WangEditor
              v-model="formData.rich_content_wang"
              :width="800"
              :height="300"
              :exclude-toolbar-keys="['group-video', 'group-image']"
              :placeholder="t('form_demo.rich_content_wang_placeholder')"
            />
          </template>
          <template v-if="field.prop === 'rich_content_markdown'">
            <MarkdownEditor
              v-model="formData.rich_content_markdown"
              width="80%"
              :height="300"
              :toolbars="['bold', 'underline', 'italic', '-', 'title', 'quote', 'link', 'image', 'table', '-', 'revoke', 'next', '=', 'preview', 'fullscreen']"
              :placeholder="t('form_demo.rich_content_markdown_placeholder')"
            />
          </template>
        </FormField>
      </el-form>

      <div class="actions">
        <el-button type="primary" @click="submitForm">{{ t('common.submit') }}</el-button>
      </div>
    </el-card>

    <el-card class="preview-card">
      <template #header>
        <span>{{ t('form_demo.preview') }}</span>
      </template>
      <pre>{{ previewJson }}</pre>
    </el-card>
  </div>
</template>

<script setup>
import { computed, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import FormField from '@/components/Form/FormField.vue'
import WangEditor from '@/components/WangEditor.vue'
import MarkdownEditor from '@/components/MarkdownEditor.vue'
import request from '@/utils/request'

const { t } = useI18n()

const formRef = ref(null)
const avatarListMinCount = 2

const getInitialValue = () => ({
  username: '',
  domain_url: '',
  domain_protocol: 'https',
  domain: '',
  password: '',
  intro: '',
  role: '',
  role_remote: null,
  department_id: null,
  department_remote_id: null,
  department_id_custom: null,
  department_remote_id_custom: null,
  gender: '',
  gender_remote: null,
  interests: [],
  interests_remote: [],
  hobbies: [],
  hobbies_remote: [],
  birthday: '',
  meeting_at: '',
  active_days: [],
  active_period: [],
  transfer_permissions: [],
  transfer_permissions_remote: [],
  score: 0,
  score_input_number: 0,
  score_input_number_suffix: 0,
  satisfaction: 3,
  volume: 40,
  enabled: true,
  icon_name: '',
  icon_name_remote: '',
  avatar: '',
  avatar_list: [],
  color: '#409EFF',
  region: [],
  region_remote: [],
  rich_content_wang: '',
  rich_content_markdown: '',
  custom_note: ''
})

const formData = reactive(getInitialValue())

const requiredRule = (message, trigger = 'blur') => [{ required: true, message, trigger }]
const requiredSelectRule = (message) => [{ required: true, message, trigger: 'change' }]
const nonEmptyArrayRule = (message, trigger = 'change') => [{
  validator: (_rule, value, callback) => {
    if (!Array.isArray(value) || value.length === 0) {
      callback(new Error(message))
      return
    }
    callback()
  },
  trigger
}]
const domainRegex = /^(?=.{1,253}$)(?!-)(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,63}$/

const formRules = computed(() => ({
  username: requiredRule(t('form.username_required')),
  domain_url: [
    {
      validator: (_rule, _value, callback) => {
        const domain = String(formData.domain || '').trim()
        if (!domain) {
          callback(new Error(t('form.please_enter') + '域名'))
          return
        }
        if (!domainRegex.test(domain)) {
          callback(new Error('域名格式不正确，例如：example.com'))
          return
        }
        callback()
      },
      trigger: 'blur'
    }
  ],
  password: requiredRule(t('form.password_required')),
  intro: requiredRule(t('form.please_enter') + t('common.description')),
  role: requiredSelectRule(t('form.please_select') + t('table.roles')),
  role_remote: requiredSelectRule(t('form.please_select') + t('form_demo.role_remote')),
  department_id: requiredSelectRule(t('form.please_select') + t('table.department')),
  department_remote_id: requiredSelectRule(t('form.please_select') + t('form_demo.department_remote')),
  department_id_custom: requiredSelectRule(t('form.please_select') + `${t('table.department')} (Custom)`),
  department_remote_id_custom: requiredSelectRule(t('form.please_select') + `${t('form_demo.department_remote')} (Custom)`),
  gender: [{ required: true, message: t('form_demo.gender_required'), trigger: 'change' }],
  gender_remote: requiredSelectRule(t('form.please_select') + t('form_demo.gender_remote')),
  interests: nonEmptyArrayRule(t('form.please_select') + t('form_demo.interests')),
  interests_remote: nonEmptyArrayRule(t('form.please_select') + t('form_demo.interests_remote')),
  hobbies: nonEmptyArrayRule(t('form.please_select') + t('form_demo.hobbies')),
  hobbies_remote: nonEmptyArrayRule(t('form.please_select') + t('form_demo.hobbies_remote')),
  birthday: requiredSelectRule(t('form.please_select') + t('form_demo.birthday')),
  meeting_at: requiredSelectRule(t('form.please_select') + t('form_demo.meeting_at')),
  active_days: nonEmptyArrayRule(t('form.please_select') + t('form_demo.active_days')),
  active_period: nonEmptyArrayRule(t('form.please_select') + t('form_demo.active_period')),
  transfer_permissions: nonEmptyArrayRule(t('form.please_select') + t('form_demo.transfer_permissions')),
  transfer_permissions_remote: nonEmptyArrayRule(t('form.please_select') + t('form_demo.transfer_permissions_remote')),
  score: [{
    type: 'number',
    required: true,
    min: 1,
    max: 100,
    message: t('form_demo.score') + ' 1-100',
    trigger: 'blur'
  }],
  score_input_number: [{
    type: 'number',
    required: true,
    min: 0,
    message: t('form.please_enter') + t('form_demo.score_input_number'),
    trigger: 'change'
  }],
  score_input_number_suffix: [{
    type: 'number',
    required: true,
    min: 0,
    message: t('form.please_enter') + t('form_demo.score_input_number_suffix'),
    trigger: 'change'
  }],
  satisfaction: [{
    type: 'number',
    required: true,
    min: 1,
    message: t('form.please_select') + t('form_demo.satisfaction'),
    trigger: 'change'
  }],
  volume: [{
    type: 'number',
    required: true,
    min: 0,
    max: 100,
    message: t('form_demo.volume') + ' 0-100',
    trigger: 'change'
  }],
  icon_name: requiredRule(t('form.please_enter') + t('form_demo.icon_name')),
  icon_name_remote: requiredRule(t('form.please_enter') + t('form_demo.icon_name_remote')),
  avatar: requiredRule(t('form.please_enter') + t('form_demo.avatar')),
  avatar_list: [{
    validator: (_rule, value, callback) => {
      const count = Array.isArray(value) ? value.filter(Boolean).length : 0
      if (count < avatarListMinCount) {
        callback(new Error(t('form_demo.avatar_list_min', { count: avatarListMinCount })))
        return
      }
      callback()
    },
    trigger: 'change'
  }],
  color: requiredSelectRule(t('form.please_select') + t('form_demo.color')),
  region: nonEmptyArrayRule(t('form.please_select') + t('form_demo.region')),
  region_remote: nonEmptyArrayRule(t('form.please_select') + t('form_demo.region_remote')),
  rich_content_wang: requiredRule(t('form.please_enter') + t('form_demo.rich_content_wang')),
  rich_content_markdown: requiredRule(t('form.please_enter') + t('form_demo.rich_content_markdown')),
  custom_note: requiredRule(t('form.please_enter') + t('form_demo.custom_slot'))
}))

const genderOptions = computed(() => [
  { label: t('form_demo.gender_male'), value: 'male' },
  { label: t('form_demo.gender_female'), value: 'female' }
])

const hobbyOptions = computed(() => [
  { label: t('form_demo.hobby_reading'), value: 'reading' },
  { label: t('form_demo.hobby_sports'), value: 'sports' },
  { label: t('form_demo.hobby_music'), value: 'music' }
])

const roleOptions = computed(() => [
  { label: t('form_demo.role_admin'), value: 'admin' },
  { label: t('form_demo.role_editor'), value: 'editor' },
  { label: t('form_demo.role_viewer'), value: 'viewer' }
])

const transferOptions = computed(() => [
  { label: t('form_demo.permission_user_view'), value: 'user.view' },
  { label: t('form_demo.permission_user_edit'), value: 'user.edit' },
  { label: t('form_demo.permission_role_view'), value: 'role.view' },
  { label: t('form_demo.permission_role_edit'), value: 'role.edit' },
  { label: t('form_demo.permission_menu_manage'), value: 'menu.manage' }
])

const departmentTree = computed(() => [
  {
    id: 1,
    name: t('form_demo.dept_headquarters'),
    children: [
      { id: 11, name: t('form_demo.dept_rd') },
      { id: 12, name: t('form_demo.dept_operations') }
    ]
  }
])

const fieldWidthSm = { width: '100%', maxWidth: '360px' }
const fieldWidthMd = { width: '100%', maxWidth: '420px' }
const fieldWidthLg = { width: '100%', maxWidth: '520px' }

const formFields = computed(() => [
  { prop: 'username', label: t('table.username'), type: 'input', autocomplete: 'off' },
  {
    prop: 'domain_url',
    label: '域名',
    type: 'custom',
    style: fieldWidthLg
  },
  { prop: 'password', label: t('common.password'), type: 'password', style: fieldWidthMd },
  { prop: 'intro', label: t('common.description'), type: 'textarea', rows: 3, style: fieldWidthLg },
  { prop: 'role', label: t('table.roles'), type: 'select', options: roleOptions.value, style: fieldWidthMd },
  {
    prop: 'role_remote',
    label: t('form_demo.role_remote'),
    type: 'select',
    apiUrl: '/options?type=form_demo&scene=default',
    clearable: true,
    style: fieldWidthMd
  },
  {
    prop: 'department_id',
    label: t('table.department'),
    type: 'tree-select',
    native: true,
    treeData: () => departmentTree.value,
    treeProps: { label: 'name', value: 'id', children: 'children' },
    clearable: true
  },
  {
    prop: 'department_remote_id',
    label: t('form_demo.department_remote'),
    type: 'tree-select',
    native: true,
    apiUrl: '/options?type=form_demo&scene=tree',
    treeProps: { label: 'name', value: 'id', children: 'children' },
    clearable: true
  },
  {
    prop: 'department_id_custom',
    label: `${t('table.department')} (Custom)`,
    type: 'tree-select',
    treeData: () => departmentTree.value,
    treeProps: { label: 'name', value: 'id', children: 'children' },
    clearable: true
  },
  {
    prop: 'department_remote_id_custom',
    label: `${t('form_demo.department_remote')} (Custom)`,
    type: 'tree-select',
    apiUrl: '/options?type=form_demo&scene=tree',
    treeProps: { label: 'name', value: 'id', children: 'children' },
    clearable: true
  },
  { prop: 'gender', label: t('form_demo.gender'), type: 'radio', options: genderOptions.value },
  {
    prop: 'gender_remote',
    label: t('form_demo.gender_remote'),
    type: 'radio',
    apiUrl: '/options?type=form_demo&scene=default'
  },
  { prop: 'interests', label: t('form_demo.interests'), type: 'checkbox', options: hobbyOptions.value },
  {
    prop: 'interests_remote',
    label: t('form_demo.interests_remote'),
    type: 'checkbox',
    apiUrl: '/options?type=form_demo&scene=default'
  },
  { prop: 'hobbies', label: t('form_demo.hobbies'), type: 'checkbox-group', options: hobbyOptions.value },
  {
    prop: 'hobbies_remote',
    label: t('form_demo.hobbies_remote'),
    type: 'checkbox-group',
    apiUrl: '/options?type=form_demo&scene=default'
  },
  { prop: 'birthday', label: t('form_demo.birthday'), type: 'date' },
  { prop: 'meeting_at', label: t('form_demo.meeting_at'), type: 'datetime' },
  { prop: 'active_days', label: t('form_demo.active_days'), type: 'daterange', style: fieldWidthLg },
  { prop: 'active_period', label: t('form_demo.active_period'), type: 'datetimerange', style: fieldWidthLg },
  {
    prop: 'transfer_permissions',
    label: t('form_demo.transfer_permissions'),
    type: 'transfer',
    options: transferOptions.value,
    titles: [t('form_demo.transfer_source'), t('form_demo.transfer_target')]
  },
  {
    prop: 'transfer_permissions_remote',
    label: t('form_demo.transfer_permissions_remote'),
    type: 'transfer',
    apiUrl: '/options?type=form_demo&scene=default',
    optionLabelKey: 'label',
    optionValueKey: 'value',
    titles: [t('form_demo.transfer_source'), t('form_demo.transfer_target')]
  },
  { prop: 'score', label: t('form_demo.score'), type: 'number', min: 0, max: 100, style: fieldWidthSm },
  {
    prop: 'score_input_number',
    label: t('form_demo.score_input_number'),
    type: 'input-number',
    min: 0,
    max: 999,
    prefix: '￥'
  },
  {
    prop: 'score_input_number_suffix',
    label: t('form_demo.score_input_number_suffix'),
    type: 'input-number',
    min: 0,
    max: 999,
    suffix: 'RMB'
  },
  { prop: 'satisfaction', label: t('form_demo.satisfaction'), type: 'rate' },
  { prop: 'volume', label: t('form_demo.volume'), type: 'slider', min: 0, max: 100, step: 5, style: fieldWidthMd },
  { prop: 'enabled', label: t('table.status'), type: 'switch' },
  { prop: 'icon_name', label: t('form_demo.icon_name'), type: 'icon' },
  {
    prop: 'icon_name_remote',
    label: t('form_demo.icon_name_remote'),
    type: 'icon',
    apiUrl: '/options?type=form_demo&scene=icon'
  },
  {
    prop: 'avatar',
    label: t('form_demo.avatar'),
    type: 'image-upload',
    uploadMode: 'both',
    cropWidth: 300,
    cropHeight: 300
  },
  {
    prop: 'avatar_list',
    label: t('form_demo.avatar_list'),
    type: 'image-upload-multiple',
    limit: 6,
    minCount: avatarListMinCount,
    maxSizeMB: 2
  },
  { prop: 'color', label: t('form_demo.color'), type: 'color', showAlpha: false },
  {
    prop: 'region',
    label: t('form_demo.region'),
    type: 'cascader',
    options: [
      {
        value: 'china',
        label: t('form_demo.country_china'),
        children: [
          {
            value: 'guangdong',
            label: t('form_demo.province_guangdong'),
            children: [{ value: 'guangzhou', label: t('form_demo.city_guangzhou') }]
          }
        ]
      }
    ]
  },
  {
    prop: 'region_remote',
    label: t('form_demo.region_remote'),
    type: 'cascader',
    apiUrl: '/options?type=form_demo&scene=cascader',
    cascaderProps: { value: 'id', label: 'name', children: 'children' }
  },
  {
    prop: 'rich_content_wang',
    label: t('form_demo.rich_content_wang'),
    type: 'custom'
  },
  {
    prop: 'rich_content_markdown',
    label: t('form_demo.rich_content_markdown'),
    type: 'custom'
  },
  {
    prop: 'custom_note',
    label: t('form_demo.custom_slot'),
    type: 'custom'
  }
])

const previewJson = computed(() => JSON.stringify(formData, null, 2))

const resetForm = () => {
  Object.assign(formData, getInitialValue())
  formRef.value?.resetFields()
}

const fillSampleData = () => {
  Object.assign(formData, {
    username: 'demo_admin',
    domain_protocol: 'https',
    domain: 'demo.example.com',
    password: '123456',
    intro: 'This is a sample form for FormField testing.',
    role: 'editor',
    role_remote: 'A',
    department_id: 11,
    department_remote_id: 11,
    department_id_custom: 11,
    department_remote_id_custom: 11,
    gender: 'male',
    gender_remote: 'A',
    interests: ['sports'],
    interests_remote: ['A'],
    hobbies: ['reading', 'music'],
    hobbies_remote: ['A'],
    birthday: '2026-03-31',
    meeting_at: '2026-03-31 09:30:00',
    active_days: ['2026-03-01', '2026-03-31'],
    active_period: ['2026-03-01 08:00:00', '2026-03-31 18:00:00'],
    transfer_permissions: ['user.view', 'role.view'],
    transfer_permissions_remote: ['A', 'C'],
    score: 88,
    score_input_number: 92,
    score_input_number_suffix: 120,
    satisfaction: 4,
    volume: 65,
    enabled: true,
    icon_name: 'User',
    icon_name_remote: 'Setting',
    color: '#67C23A',
    region: ['china', 'guangdong', 'guangzhou'],
    region_remote: [100, 110, 111],
    rich_content_wang: '<p>WangEditor demo content</p>',
    rich_content_markdown: '## Markdown Demo\n\n- item 1\n- item 2',
    custom_note: 'This field is rendered via FormField default slot.'
  })
}

const loadFromApi = async () => {
  try {
    const res = await request({
      url: '/form-demo/data',
      method: 'get'
    })
    if (res?.data) {
      Object.assign(formData, res.data)
      ElMessage.success(t('form_demo.loaded_from_api'))
    }
  } catch {
    // Global interceptor handles error
  }
}

const submitForm = async () => {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  console.log('[FormDemo] submit formData:', JSON.parse(JSON.stringify(formData)))
  ElMessage.success(t('common.success'))
}
</script>

<style scoped>
.form-demo-page {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.actions {
  margin-top: 16px;
}

.preview-card {
  margin-top: 16px;
}

pre {
  margin: 0;
  padding: 12px;
  border-radius: 4px;
  background: var(--el-fill-color-light);
  overflow: auto;
}

.custom-tip {
  margin-top: 6px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.domain-inline {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 8px;
}

.domain-inline .el-input {
  flex: 1;
}
</style>
