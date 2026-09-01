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

export function create<<.ModelName>>TableColumns(t) {
  return [
    { type: 'index', width: 60, title: t('table.seq'), key: 'index' },
<<range .ListFields>>
<<- if and .ShowInList (ne .Name "operation")>>
    <<- if eq .Name "status">>
    { field: 'status', title: t('table.status'), width: 120, slot: 'status', key: 'status' },
    <<- else if eq .Name "sort">>
    { field: 'sort', title: t('common.sort'), width: 100, key: 'sort' },
    <<- else if eq .Name "created_at">>
    { field: 'created_at', title: t('table.created_at'), width: 180, sortable: true, key: 'created_at' },
    <<- else if eq .Name "updated_at">>
    { field: 'updated_at', title: t('table.updated_at'), width: 180, sortable: true, key: 'updated_at' },
    <<- else if eq .FormType "image-upload">>
    { field: '<<.Name>>', title: t('<<.Name>>'), slot: '<<.Name>>', width: 120, key: '<<.Name>>' },
    <<- else if or (eq .FormType "editor") (eq .FormType "markdown")>>
    { field: '<<.Name>>', title: t('<<.Name>>'), slot: '<<.Name>>', width: 220, key: '<<.Name>>' },
    <<- else if .Relation>>
    { field: '<<.Name>>', title: t('<<.Name>>'), slot: '<<.Name>>', minWidth: 150, key: '<<.Name>>' },
    <<- else>>
    { field: '<<.Name>>', title: t('<<.Name>>'), minWidth: 150, key: '<<.Name>>' },
    <<- end>>
<<- end>>
<<- end>>
    { title: t('table.operation'), width: 150, fixed: 'right', slot: 'operation', key: 'operation' }
  ]
}

<<range .ListFields>>
<<- if and .ShowInList .Relation>>
export function get<<.Relation.JsonName>>DisplayName(value) {
  if (!value) return '-'
  return value.<<.Relation.DisplayField>> || value.<<.Relation.JsonName>> || '-'
}
<<- end>>
<<- end>>
