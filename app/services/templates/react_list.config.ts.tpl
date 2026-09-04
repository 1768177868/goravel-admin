import type { TFunction } from 'i18next'
import type { SearchField } from '@/components/SearchForm'
import { buildSearchParams } from '@/utils/buildSearchParams'
import { entityField } from '@/utils/normalize'

export const <<.ModuleNameCamel>>InitialSearchForm: Record<string, unknown> = {
<<range .SearchableFields>>
  <<.Name>>: <<if or (eq .SearchUIType "daterange") (eq .SearchUIType "datetimerange")>>[]<<else>>''<<end>>,
<<- end>>
}

export type <<.ModelName>>SearchForm = typeof <<.ModuleNameCamel>>InitialSearchForm

export interface <<.ModelName>>Row {
  id: number | string
<<range .ListFields>>
<<- if and .ShowInList (ne .Name "id") (ne .Name "operation")>>
  <<.Name>>?: unknown
<<- if .Relation>>
  <<.Relation.JsonName>>?: unknown
<<- end>>
<<- end>>
<<- end>>
}

export function build<<.ModelName>>ListParams(
  form: <<.ModelName>>SearchForm,
  baseParams: Record<string, unknown>,
) {
  return buildSearchParams(form, baseParams)
}

export function create<<.ModelName>>SearchFields(t: TFunction): SearchField[] {
  return [
<<range .SearchableFields>>
    {
      name: '<<.Name>>',
      label: <<if eq .Name "created_at">>t('common.created_at')<<else if eq .Name "updated_at">>t('common.updated_at')<<else if eq .Name "status">>t('common.status')<<else>>t('<<.Name>>', { defaultValue: '<<.Label>>' })<<end>>,
      <<if eq .SearchUIType "select">>
      type: 'select',
      options: [
        { label: t('common.enabled'), value: 1 },
        { label: t('common.disabled'), value: 0 },
      ],
      <<else if eq .SearchUIType "datetime">>
      type: 'datetime',
      <<end>>
    },
<<- end>>
  ]
}

export function transform<<.ModelName>>Row(row: Record<string, unknown>): <<.ModelName>>Row {
  return {
    id: entityField(row, 'id', '')!,
<<range .ListFields>>
<<- if and .ShowInList (ne .Name "id") (ne .Name "operation")>>
    <<.Name>>: entityField(row, '<<.Name>>', <<if eq .FormType "number">>0<<else if eq .FormType "switch">>0<<else>>''<<end>>),
<<- if .Relation>>
    <<.Relation.JsonName>>: entityField(row, '<<.Relation.JsonName>>', null),
<<- end>>
<<- end>>
<<- end>>
  }
}

<<range .ListFields>>
<<- if and .ShowInList .Relation>>
export function get<<.Relation.JsonName>>DisplayName(value: unknown) {
  if (!value || typeof value !== 'object') return '-'
  const record = value as Record<string, unknown>
  return String(record['<<.Relation.DisplayField>>'] ?? record['<<.Relation.JsonName>>'] ?? '-')
}
<<- end>>
<<- end>>
