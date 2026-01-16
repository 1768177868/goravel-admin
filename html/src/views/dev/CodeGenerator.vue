<template>
  <div class="code-generator">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('code_generator.title') }}</span>
          <el-button type="primary" :loading="generating" @click="handleGenerate">
            <el-icon><Document /></el-icon>
            {{ $t('code_generator.generate') }}
          </el-button>
        </div>
      </template>

      <el-form :model="form" :rules="rules" ref="formRef" label-width="120px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="$t('code_generator.module_name')" prop="module_name">
              <el-input v-model="form.module_name" :placeholder="$t('code_generator.module_name_placeholder')" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="$t('code_generator.table_name')" prop="table_name">
              <el-input v-model="form.table_name" :placeholder="$t('code_generator.table_name_placeholder')" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>

      <el-divider>{{ $t('code_generator.fields_config') }}</el-divider>

      <el-form-item :label="$t('code_generator.options')">
        <el-checkbox v-model="form.has_create">{{ $t('common.add') }}</el-checkbox>
        <el-checkbox v-model="form.has_edit">{{ $t('common.edit') }}</el-checkbox>
        <el-checkbox v-model="form.has_delete">{{ $t('common.delete') }}</el-checkbox>
      </el-form-item>

      <el-form-item :label="$t('code_generator.generated_files')">
        <el-checkbox-group v-model="form.files">
          <el-checkbox v-for="fileType in fileTypes" :key="fileType.value" :label="fileType.value">
            {{ fileType.label }}
          </el-checkbox>
        </el-checkbox-group>
      </el-form-item>

      <el-button type="primary" @click="handleAddField">
        <el-icon><Plus /></el-icon>
        {{ $t('code_generator.add_field') }}
      </el-button>

      <el-table :data="form.fields" border style="margin-top: 20px">
        <el-table-column type="index" :label="$t('table.index')" width="60" />
        <el-table-column prop="name" :label="$t('code_generator.field_name')" width="150">
          <template #default="{ row }">
            <el-input v-model="row.name" :placeholder="$t('code_generator.field_name_placeholder')" />
          </template>
        </el-table-column>
        <el-table-column prop="type" :label="$t('code_generator.field_type')" width="150">
          <template #default="{ row }">
            <el-select v-model="row.type" :placeholder="$t('common.select')">
              <el-option
                v-for="type in fieldTypes"
                :key="type.value"
                :label="type.label"
                :value="type.value"
              />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column prop="label" :label="$t('code_generator.field_label')" width="150">
          <template #default="{ row }">
            <el-input v-model="row.label" :placeholder="$t('code_generator.field_label_placeholder')" />
          </template>
        </el-table-column>
        <el-table-column prop="comment" :label="$t('code_generator.field_comment')" width="150">
          <template #default="{ row }">
            <el-input v-model="row.comment" :placeholder="$t('code_generator.field_comment_placeholder')" />
          </template>
        </el-table-column>
        <el-table-column prop="search_type" :label="$t('code_generator.search_type')" width="120">
          <template #default="{ row }">
            <el-select v-model="row.search_type" :placeholder="$t('common.select')">
              <el-option label="LIKE" value="like" />
              <el-option label="=" value="=" />
              <el-option label=">" value=">" />
              <el-option label=">=" value=">=" />
              <el-option label="<" value="<" />
              <el-option label="<=" value="<=" />
              <el-option label="!=" value="!=" />
              <el-option label="IN" value="in" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column prop="search_ui_type" :label="$t('code_generator.search_ui_type')" width="150">
          <template #default="{ row }">
            <el-select v-model="row.search_ui_type" :placeholder="$t('common.select')">
              <el-option :label="$t('code_generator.search_ui_types.input')" value="input" />
              <el-option :label="$t('code_generator.search_ui_types.select')" value="select" />
              <el-option :label="$t('code_generator.search_ui_types.date')" value="date" />
              <el-option :label="$t('code_generator.search_ui_types.datetime')" value="datetime" />
              <el-option :label="$t('code_generator.search_ui_types.daterange')" value="daterange" />
              <el-option :label="$t('code_generator.search_ui_types.datetimerange')" value="datetimerange" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column :label="$t('code_generator.relation')" width="150">
          <template #default="{ row }">
            <el-button size="small" @click="handleEditRelation(row)">
              {{ row.relation ? $t('code_generator.edit_relation') : $t('code_generator.add_relation') }}
            </el-button>
          </template>
        </el-table-column>
        <el-table-column :label="$t('code_generator.field_config')" width="100">
          <template #default="{ row }">
            <el-button 
              v-if="row.search_ui_type === 'select' || row.type === 'decimal'"
              size="small" 
              @click="handleEditFieldConfig(row)"
            >
              <el-icon><Setting /></el-icon>
            </el-button>
          </template>
        </el-table-column>
        <el-table-column :label="$t('code_generator.field_options')" width="300">
          <template #default="{ row }">
            <el-checkbox v-model="row.required">{{ $t('code_generator.required') }}</el-checkbox>
            <el-checkbox v-model="row.searchable">{{ $t('code_generator.searchable') }}</el-checkbox>
            <el-checkbox v-model="row.sortable">{{ $t('code_generator.sortable') }}</el-checkbox>
            <el-checkbox v-model="row.show_in_list">{{ $t('code_generator.show_in_list') }}</el-checkbox>
            <el-checkbox v-model="row.show_in_form">{{ $t('code_generator.show_in_form') }}</el-checkbox>
          </template>
        </el-table-column>
        <el-table-column :label="$t('table.operation')" width="100" fixed="right">
          <template #default="{ $index }">
            <el-button type="danger" size="small" @click="handleRemoveField($index)">
              <el-icon><Delete /></el-icon>
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-divider>{{ $t('code_generator.code_preview') }}</el-divider>

      <el-tabs v-model="activeTab" type="border-card">
        <el-tab-pane
          v-for="fileType in fileTypes"
          :key="fileType.value"
          :label="fileType.label"
          :name="fileType.value"
        >
          <div class="code-preview">
            <el-button
              type="primary"
              size="small"
              @click="handlePreview(fileType.value)"
              :loading="previewing === fileType.value"
            >
              {{ $t('code_generator.refresh_preview') }}
            </el-button>
            <pre v-if="previewCode[fileType.value]"><code>{{ previewCode[fileType.value] }}</code></pre>
            <el-empty v-else :description="$t('code_generator.click_preview')" />
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <el-dialog
      v-model="relationDialogVisible"
      :title="$t('code_generator.relation_config')"
      width="600px"
    >
      <el-form :model="relationForm" label-width="120px">
        <el-form-item :label="$t('code_generator.relation_table')">
          <el-input v-model="relationForm.table" :placeholder="$t('code_generator.relation_table_placeholder')" />
        </el-form-item>
        <el-form-item :label="$t('code_generator.relation_type')">
          <el-select v-model="relationForm.relation_type" :placeholder="$t('common.select')">
            <el-option label="belongsTo" value="belongsTo" />
            <el-option label="hasOne" value="hasOne" />
            <el-option label="hasMany" value="hasMany" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('code_generator.foreign_key')">
          <el-input v-model="relationForm.foreign_key" :placeholder="$t('code_generator.foreign_key_placeholder')" />
        </el-form-item>
        <el-form-item :label="$t('code_generator.display_field')">
          <el-input v-model="relationForm.display_field" :placeholder="$t('code_generator.display_field_placeholder')" />
        </el-form-item>
        <el-form-item :label="$t('code_generator.relation_alias')">
          <el-input v-model="relationForm.alias" :placeholder="$t('code_generator.relation_alias_placeholder')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="relationDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleSaveRelation">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
    <el-dialog
      v-model="fieldConfigDialogVisible"
      :title="$t('code_generator.field_config')"
      width="500px"
    >
      <el-form :model="fieldConfigForm" label-width="120px">
        <template v-if="currentField && currentField.type === 'decimal'">
          <el-form-item :label="$t('code_generator.precision')">
            <el-input-number v-model="fieldConfigForm.precision" :min="1" :max="65" />
          </el-form-item>
          <el-form-item :label="$t('code_generator.scale')">
            <el-input-number v-model="fieldConfigForm.scale" :min="0" :max="30" />
          </el-form-item>
        </template>
        <template v-else>
          <el-form-item :label="$t('code_generator.option_type')">
            <el-radio-group v-model="fieldConfigForm.option_type">
              <el-radio label="dictionary">{{ $t('code_generator.option_type_dictionary') }}</el-radio>
              <el-radio label="api">{{ $t('code_generator.option_type_api') }}</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item 
            v-if="fieldConfigForm.option_type === 'dictionary'"
            :label="$t('code_generator.dictionary_key')"
          >
            <el-select 
              v-model="fieldConfigForm.dictionary" 
              :placeholder="$t('code_generator.dictionary_key_placeholder')"
              filterable
              allow-create
              default-first-option
            >
              <el-option
                v-for="item in dictionaryTypes"
                :key="item"
                :label="item"
                :value="item"
              />
            </el-select>
          </el-form-item>
          <el-form-item 
            v-if="fieldConfigForm.option_type === 'api'"
            :label="$t('code_generator.api_url')"
          >
            <el-input v-model="fieldConfigForm.api_url" :placeholder="$t('code_generator.api_url_placeholder')" />
          </el-form-item>
        </template>
      </el-form>
      <template #footer>
        <el-button @click="fieldConfigDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleSaveFieldConfig">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Document, Delete, Setting } from '@element-plus/icons-vue'
import { getFieldTypes, previewCode as previewCodeApi, generateCode, saveCode } from '../../api/codeGenerator'
import { getDictionaryTypes } from '../../api/dictionary'
import { isDev } from '../../utils/env'
import logger from '../../utils/logger'

const { t } = useI18n()

if (!isDev()) {
  ElMessage.error(t('code_generator.dev_only'))
}

const formRef = ref(null)
const generating = ref(false)
const previewing = ref('')
const activeTab = ref('model')
const fieldTypes = ref([])
const dictionaryTypes = ref([])
const previewCode = reactive({})
const relationDialogVisible = ref(false)
const fieldConfigDialogVisible = ref(false)
const currentField = ref(null)
const relationForm = reactive({
  table: '',
  relation_type: 'belongsTo',
  foreign_key: '',
  display_field: '',
  alias: ''
})
const fieldConfigForm = reactive({
  option_type: 'dictionary',
  dictionary: '',
  api_url: '',
  precision: 8,
  scale: 2
})

const form = reactive({
  module_name: '',
  table_name: '',
  has_create: true,
  has_edit: true,
  has_delete: true,
  files: ['model', 'controller', 'service', 'request_create', 'request_update', 'api', 'list_page', 'form_page'],
  fields: [
    {
      name: 'name',
      type: 'string',
      label: '名称',
      required: true,
      searchable: true,
      sortable: true,
      show_in_list: true,
      show_in_form: true,
      show_in_detail: true,
      is_primary_key: false,
      comment: '名称',
      search_type: 'like',
      search_ui_type: 'input',
      relation: null,
      dictionary: '',
      api_url: ''
    },
    {
      name: 'status',
      type: 'bool',
      label: '状态',
      required: false,
      searchable: true,
      sortable: true,
      show_in_list: true,
      show_in_form: true,
      show_in_detail: true,
      is_primary_key: false,
      comment: '状态',
      search_type: '=',
      search_ui_type: 'select',
      relation: null,
      dictionary: 'status',
      api_url: ''
    }
  ]
})

const rules = {
  module_name: [
    { required: true, message: t('code_generator.module_name_required'), trigger: 'blur' }
  ],
  table_name: [
    { required: true, message: t('code_generator.table_name_required'), trigger: 'blur' }
  ]
}

const fileTypes = [
  { value: 'model', label: t('code_generator.file_model') },
  { value: 'controller', label: t('code_generator.file_controller') },
  { value: 'service', label: t('code_generator.file_service') },
  { value: 'request_create', label: t('code_generator.file_request_create') },
  { value: 'request_update', label: t('code_generator.file_request_update') },
  { value: 'api', label: t('code_generator.file_api') },
  { value: 'list_page', label: t('code_generator.file_list_page') },
  { value: 'form_page', label: t('code_generator.file_form_page') }
]

const loadFieldTypes = async () => {
  try {
    const response = await getFieldTypes()
    fieldTypes.value = response.data.field_types || []
  } catch (error) {
    logger.error('Failed to load field types:', error)
    ElMessage.error(t('code_generator.load_field_types_failed'))
  }
}

const loadDictionaryTypes = async () => {
  try {
    const response = await getDictionaryTypes()
    dictionaryTypes.value = response.data.types || []
  } catch (error) {
    logger.error('Failed to load dictionary types:', error)
  }
}

const handleAddField = () => {
  form.fields.push({
    name: '',
    type: 'string',
    label: '',
    required: false,
    searchable: false,
    sortable: false,
    show_in_list: true,
    show_in_form: true,
    show_in_detail: true,
    is_primary_key: false,
    comment: '',
    search_type: 'like',
    search_ui_type: 'input',
    relation: null,
    dictionary: '',
    api_url: ''
  })
}

const handleRemoveField = (index) => {
  form.fields.splice(index, 1)
}

const handleEditRelation = (row) => {
  currentField.value = row
  if (row.relation) {
    relationForm.table = row.relation.table
    relationForm.relation_type = row.relation.relation_type
    relationForm.foreign_key = row.relation.foreign_key
    relationForm.display_field = row.relation.display_field
    relationForm.alias = row.relation.alias || ''
  } else {
    relationForm.table = ''
    relationForm.relation_type = 'belongsTo'
    relationForm.foreign_key = ''
    relationForm.display_field = ''
    relationForm.alias = ''
  }
  relationDialogVisible.value = true
}

const handleSaveRelation = () => {
  if (!relationForm.table || !relationForm.foreign_key || !relationForm.display_field) {
    ElMessage.error(t('code_generator.relation_required'))
    return
  }
  currentField.value.relation = {
    table: relationForm.table,
    relation_type: relationForm.relation_type,
    foreign_key: relationForm.foreign_key,
    display_field: relationForm.display_field,
    alias: relationForm.alias
  }
  relationDialogVisible.value = false
  ElMessage.success(t('code_generator.relation_saved'))
}

const handleEditFieldConfig = (row) => {
  currentField.value = row
  if (row.type === 'decimal') {
    fieldConfigForm.precision = row.precision || 8
    fieldConfigForm.scale = row.scale || 2
  } else if (row.api_url) {
    fieldConfigForm.option_type = 'api'
    fieldConfigForm.api_url = row.api_url
    fieldConfigForm.dictionary = ''
  } else {
    fieldConfigForm.option_type = 'dictionary'
    fieldConfigForm.dictionary = row.dictionary || ''
    fieldConfigForm.api_url = ''
  }
  loadDictionaryTypes()
  fieldConfigDialogVisible.value = true
}

const handleSaveFieldConfig = () => {
  if (currentField.value) {
    if (currentField.value.type === 'decimal') {
      currentField.value.precision = fieldConfigForm.precision
      currentField.value.scale = fieldConfigForm.scale
    } else if (fieldConfigForm.option_type === 'dictionary') {
      currentField.value.dictionary = fieldConfigForm.dictionary
      currentField.value.api_url = ''
    } else {
      currentField.value.dictionary = ''
      currentField.value.api_url = fieldConfigForm.api_url
    }
  }
  fieldConfigDialogVisible.value = false
  ElMessage.success(t('code_generator.field_config_saved'))
}

const handlePreview = async (fileType) => {
  previewing.value = fileType
  try {
    const response = await previewCodeApi({
      module_name: form.module_name,
      table_name: form.table_name,
      fields: form.fields,
      file_type: fileType,
      options: {
        has_create: form.has_create,
        has_edit: form.has_edit,
        has_delete: form.has_delete
      }
    })
    previewCode[fileType] = response.data.code || ''
  } catch (error) {
    logger.error('Failed to preview code:', error)
    ElMessage.error(t('code_generator.preview_failed'))
  } finally {
    previewing.value = ''
  }
}

const handleGenerate = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
  } catch (error) {
    return
  }

  try {
    await ElMessageBox.confirm(
      t('code_generator.generate_confirm'),
      t('common.confirm'),
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      }
    )

    generating.value = true
    const response = await saveCode({
      module_name: form.module_name,
      table_name: form.table_name,
      fields: form.fields,
      files: form.files,
      force: false,
      options: {
        has_create: form.has_create,
        has_edit: form.has_edit,
        has_delete: form.has_delete
      }
    })

    const savedFiles = response.data.saved_files || []
    ElMessage.success(t('code_generator.save_success', { count: savedFiles.length }))
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') {
      if (error.response && error.response.status === 409 && error.response.data && error.response.data.error_code === 'files_exist') {
        const existingFiles = error.response.data.files || []
        try {
          await ElMessageBox.confirm(
            t('code_generator.files_exist_confirm', { 
              count: existingFiles.length,
              files: existingFiles.map(f => `<br/>${f}`).join('') 
            }),
            t('common.warning'),
            {
              confirmButtonText: t('code_generator.overwrite'),
              cancelButtonText: t('common.cancel'),
              type: 'warning',
              dangerouslyUseHTMLString: true
            }
          )

          const response = await saveCode({
            module_name: form.module_name,
            table_name: form.table_name,
            fields: form.fields,
            files: form.files,
            force: true,
            options: {
              has_create: form.has_create,
              has_edit: form.has_edit,
              has_delete: form.has_delete
            }
          })

          const savedFiles = response.data.saved_files || []
          ElMessage.success(t('code_generator.save_success', { count: savedFiles.length }))
        } catch (innerError) {
          if (innerError !== 'cancel' && innerError !== 'close') {
            logger.error('Failed to overwrite code:', innerError)
            ElMessage.error(t('code_generator.generate_failed'))
          }
        }
      } else {
        logger.error('Failed to generate code:', error)
        ElMessage.error(t('code_generator.generate_failed'))
      }
    }
  } finally {
    generating.value = false
  }
}

onMounted(() => {
  loadFieldTypes()
})
</script>

<style scoped>
.code-generator {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.code-preview {
  position: relative;
}

.code-preview pre {
  background: #f5f5f5;
  padding: 15px;
  border-radius: 4px;
  max-height: 500px;
  overflow: auto;
  font-family: 'Courier New', Courier, monospace;
  font-size: 13px;
  line-height: 1.5;
}

.code-preview code {
  color: #333;
}
</style>
