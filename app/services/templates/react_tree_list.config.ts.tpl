import type { SearchField } from '@/components/SearchForm'
import { buildSearchParams } from '@/utils/buildSearchParams'
import { entityField, normalizeTreeList } from '@/utils/normalize'
import { flattenTree } from '@/utils/tree'

export const <<.ModuleNameCamel>>InitialSearchForm: Record<string, unknown> = {
<<range .SearchableFields>>
  <<.Name>>: <<if or (eq .SearchUIType "daterange") (eq .SearchUIType "datetimerange")>>[]<<else>>''<<end>>,
<<- end>>
}

export type <<.ModelName>>SearchForm = typeof <<.ModuleNameCamel>>InitialSearchForm

export interface <<.ModelName>>Row {
  id: number | string
  parent_id?: number | string
<<range .ListFields>>
<<- if and .ShowInList (ne .Name "id") (ne .Name "operation") (ne .Name "parent_id")>>
  <<.Name>>?: unknown
<<- end>>
<<- end>>
  children?: <<.ModelName>>Row[]
}

export type <<.ModelName>>TreeSelectNode = {
  title: string
  value: number | string
  children?: <<.ModelName>>TreeSelectNode[]
}

export function build<<.ModelName>>ListParams(
  form: <<.ModelName>>SearchForm,
  baseParams: Record<string, unknown>,
) {
  return { ...buildSearchParams(form, baseParams), page_size: 1000 }
}

export function create<<.ModelName>>SearchFields(t: (key: string) => string): SearchField[] {
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

export function map<<.ModelName>>Rows(list: unknown[]): <<.ModelName>>Row[] {
  return normalizeTreeList(list as Record<string, unknown>[]).map((item) => {
    const row = item as Record<string, unknown>
    const children = Array.isArray(row.children)
      ? map<<.ModelName>>Rows(row.children as unknown[])
      : undefined
    return {
      id: entityField(row, 'id', '')!,
      parent_id: entityField(row, 'parent_id', 0),
<<range .ListFields>>
<<- if and .ShowInList (ne .Name "id") (ne .Name "operation") (ne .Name "parent_id")>>
      <<.Name>>: entityField(row, '<<.Name>>', <<if eq .FormType "number">>0<<else if eq .FormType "switch">>0<<else>>''<<end>>),
<<- end>>
<<- end>>
      children: children?.length ? children : undefined,
    } as <<.ModelName>>Row
  })
}

/** Build nested rows from a flat API list (search mode). */
export function build<<.ModelName>>Tree(flat: <<.ModelName>>Row[]): <<.ModelName>>Row[] {
  const byParent = new Map<string, <<.ModelName>>Row[]>()
  flat.forEach((item) => {
    const parentKey = String(Number(item.parent_id ?? 0))
    const bucket = byParent.get(parentKey) ?? []
    bucket.push({ ...item, children: undefined })
    byParent.set(parentKey, bucket)
  })

  const attach = (parentId: number | string): <<.ModelName>>Row[] => {
    const nodes = byParent.get(String(Number(parentId))) ?? []
    return nodes.map((node) => {
      const children = attach(node.id)
      return children.length ? { ...node, children } : node
    })
  }

  return attach(0)
}

export function normalize<<.ModelName>>TreeList(list: unknown[]): <<.ModelName>>Row[] {
  const mapped = map<<.ModelName>>Rows(list)
  if (mapped.some((row) => Array.isArray(row.children) && row.children.length > 0)) {
    return mapped
  }
  return build<<.ModelName>>Tree(mapped)
}

export function to<<.ModelName>>TreeSelect(items: <<.ModelName>>Row[]): <<.ModelName>>TreeSelectNode[] {
  return items.map((item) => ({
    title: String(item.<<.TreeLabelField>> ?? item.id),
    value: item.id,
    children: item.children?.length ? to<<.ModelName>>TreeSelect(item.children) : undefined,
  }))
}

export function collect<<.ModelName>>Ids(items: <<.ModelName>>Row[]): Array<string | number> {
  return flattenTree(items).map((node) => (node as <<.ModelName>>Row).id)
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
