import { useTranslation } from 'react-i18next'
import SimpleCrudPage from '@/components/SimpleCrudPage'
import StatusTag from '@/components/StatusTag'
import {
  <<if .HasCreate>>create<<.ModelName>>,<<end>>
  <<if .HasDelete>>delete<<.ModelName>>,<<end>>
  get<<.ModelName>>List,
  <<if .HasEdit>>update<<.ModelName>>,<<end>>
} from '@/api/<<.ModuleNameK>>'

interface <<.ModelName>>Row {
  id: number | string
  name?: string
  status?: number
  created_at?: string
<<range .FormFields>>
<<- if and .ShowInList (ne .Name "id") (ne .Name "name") (ne .Name "status") (ne .Name "created_at") (ne .Name "updated_at")>>
  <<.Name>>?: <<if eq .FormType "number">>number<<else if eq .FormType "switch">>number<<else>>string<<end>>
<<- end>>
<<- end>>
}

export default function <<.ModelName>>List() {
  const { t } = useTranslation()

  return (
    <SimpleCrudPage<<<.ModelName>>Row>
      title={t('menu.<<.ModuleName>>')}
      permissions={{
        <<if .HasCreate>>store: '<<.ModuleName>>.store',<<end>>
        <<if .HasEdit>>update: '<<.ModuleName>>.update',<<end>>
        <<if .HasDelete>>destroy: '<<.ModuleName>>.destroy',<<end>>
      }}
      fetchApi={get<<.ModelName>>List}
      <<if .HasCreate>>createApi={create<<.ModelName>>}<<end>>
      <<if .HasEdit>>updateApi={update<<.ModelName>>}<<end>>
      <<if .HasDelete>>deleteApi={delete<<.ModelName>>}<<end>>
      createTitle={t('<<.ModuleName>>.add', { defaultValue: t('common.add') })}
      editTitle={t('<<.ModuleName>>.edit', { defaultValue: t('common.edit') })}
      initialSearchForm={<<.ModuleNameCamel>>InitialSearchForm}
      searchFields={<<.ModuleNameCamel>>SearchFields(t)}
      formFields={<<.ModuleNameCamel>>FormFields(t)}
      columns={<<.ModuleNameCamel>>Columns(t)}
      transformRow={(row) => ({
        id: row.id as string | number,
        name: String(row.name ?? ''),
        status: Number(row.status ?? 1),
        created_at: String(row.created_at ?? ''),
        ...row,
      })}
    />
  )
}

export const <<.ModuleNameCamel>>InitialSearchForm = {
<<range .SearchableFields>>
  <<.Name>>: <<if or (eq .SearchUIType "daterange") (eq .SearchUIType "datetimerange")>>[] as string[]<<else if eq .SearchUIType "select">>''<<else>>''<<end>>,
<<- end>>
}

export function <<.ModuleNameCamel>>SearchFields(t: (key: string) => string) {
  return [
<<range .SearchableFields>>
    {
      name: '<<.Name>>',
      label: <<if eq .Name "created_at">>t('common.created_at')<<else if eq .Name "updated_at">>t('common.updated_at')<<else if eq .Name "status">>t('common.status')<<else>>t('<<.Name>>', { defaultValue: '<<.Label>>' })<<end>>,
      <<if eq .SearchUIType "select">>
      type: 'select' as const,
      options: [
        { label: t('common.enabled'), value: 1 },
        { label: t('common.disabled'), value: 0 },
      ],
      <<end>>
    },
<<- end>>
  ]
}

export function <<.ModuleNameCamel>>FormFields(t: (key: string) => string) {
  return [
<<range .FormFields>>
<<- if and .ShowInForm (ne .Name "id") (ne .Name "created_at") (ne .Name "updated_at") (ne .Name "deleted_at")>>
    {
      name: '<<.Name>>',
      label: t('<<.Name>>', { defaultValue: '<<.Label>>' }),
      <<if .Required>>required: true,<<end>>
      <<if eq .FormType "textarea">>type: 'textarea' as const,<<else if eq .FormType "number">>type: 'number' as const,<<else if eq .FormType "switch">>type: 'status' as const,<<else if eq .FormType "password">>type: 'password' as const,<<end>>
    },
<<- end>>
<<- end>>
  ]
}

export function <<.ModuleNameCamel>>Columns(t: (key: string) => string) {
  return [
    { title: t('table.id'), dataIndex: 'id', width: 80, sorter: true },
<<range .ListFields>>
<<- if and .ShowInList (ne .Name "id") (ne .Name "created_at") (ne .Name "updated_at") (ne .Name "operation")>>
    <<if and (eq .Name "status") (eq .FormType "switch")>>
    {
      title: t('common.status'),
      dataIndex: 'status',
      width: 100,
      render: (status: number) => <StatusTag status={status} />,
    },
    <<else>>
    { title: t('<<.Name>>', { defaultValue: '<<.Label>>' }), dataIndex: '<<.Name>>'<<if .Sortable>>, sorter: true<<end>> },
    <<end>>
<<- end>>
<<- end>>
    { title: t('table.created_at'), dataIndex: 'created_at', width: 180, sorter: true },
  ]
}
