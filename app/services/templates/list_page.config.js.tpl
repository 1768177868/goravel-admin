import { buildRangeSearchParams } from '@/utils/listPageHelpers'

export const <<.ModuleNameCamel>>InitialSearchForm = {
<<range .SearchableFields>>
  <<.Name>>: <<if or (eq .SearchUIType "daterange") (eq .SearchUIType "datetimerange")>>[]<<else>>''<<end>>,
<<- end>>
}

const rangeSearchFields = [
<<range .SearchableFields>>
<<- if or (eq .SearchUIType "daterange") (eq .SearchUIType "datetimerange")>>
  '<<.Name>>',
<<- end>>
<<- end>>
]

export function build<<.ModelName>>ListParams(form, baseParams) {
  return buildRangeSearchParams(form, baseParams, rangeSearchFields)
}

export function create<<.ModelName>>SearchFields(t) {
  return [
<<range .SearchableFields>>
    {
      prop: '<<.Name>>',
      label: <<if eq .Name "created_at">>t('common.created_at')<<else if eq .Name "updated_at">>t('common.updated_at')<<else if eq .Name "status">>t('table.status')<<else>>t('<<.Name>>')<<end>>,
      type: '<<.SearchUIType>>',
      clearable: true,
      width: '<<if eq .SearchUIType "datetimerange">>380px<<else if eq .SearchUIType "daterange">>320px<<else>>200px<<end>>',
      advanced: false<<if or (eq .SearchUIType "select") (eq .SearchUIType "radio")>>,
<<if .ApiUrl>>
      apiUrl: '<<.ApiUrl>>'<<else if .Dictionary>>,
      apiUrl: '/options?type=dictionary&dictionary_type=<<.Dictionary>>'<<else if eq .Name "status">>,
      apiUrl: '/options?type=dictionary&dictionary_type=status'<<end>><<end>>
    },
<<- end>>
  ]
}

export function create<<.ModelName>>TableColumns(t, options = {}) {
  const { enableBatchActions = <<if .EnableBatchActions>>true<<else>>false<<end>> } = options
  const baseColumns = [
    { field: 'id', title: t('table.id'), width: 80, sortable: true, key: 'id' },
<<range .ListFields>>
<<- if and .ShowInList (ne .Name "created_at") (ne .Name "updated_at") (ne .Name "operation")>>
    <<- if and (eq .Name "status") (eq .FormType "switch")>>
    { field: 'status', title: t('table.status'), width: 100, sortable: false, slot: 'status', key: 'status' },
    <<- else if eq .FormType "image-upload">>
    { field: '<<.Name>>', title: t('<<.Name>>'), slot: '<<.Name>>', sortable: false, width: 120, key: '<<.Name>>' },
    <<- else if or (eq .FormType "editor") (eq .FormType "markdown")>>
    { field: '<<.Name>>', title: t('<<.Name>>'), slot: '<<.Name>>', sortable: false, width: 220, key: '<<.Name>>' },
    <<- else if .Relation>>
    { field: '<<.Name>>', title: t('<<.Name>>'), slot: '<<.Name>>', sortable: false, key: '<<.Name>>' },
    <<- else>>
    { field: '<<.Name>>', title: t('<<.Name>>'), sortable: <<.Sortable>>, key: '<<.Name>>' },
    <<- end>>
<<- end>>
<<- end>>
    { field: 'updated_at', title: t('table.updated_at'), width: 180, sortable: true, key: 'updated_at' },
    { field: 'created_at', title: t('table.created_at'), width: 180, sortable: true, key: 'created_at' },
    { field: 'operation', title: t('table.operation'), width: 220, fixed: 'right', slot: 'operation', key: 'operation' }
  ]

  if (!enableBatchActions) {
    return baseColumns
  }

  return [{ type: 'checkbox', width: 52, fixed: 'left', key: 'checkbox' }, ...baseColumns]
}

<<range .ListFields>>
<<- if and .ShowInList .Relation>>
export function get<<.Relation.JsonName>>DisplayName(value) {
  if (!value) return '-'
  return value.<<.Relation.DisplayField>> || value.<<.Relation.JsonName>> || '-'
}
<<- end>>
<<- end>>
