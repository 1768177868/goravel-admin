import type { ReactNode } from 'react'
import { Space, Typography } from 'antd'

interface PageContainerProps {
  title?: ReactNode
  extra?: ReactNode
  children: ReactNode
  className?: string
}

export default function PageContainer({ title, extra, children, className }: PageContainerProps) {
  return (
    <div className={className}>
      {(title || extra) && (
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            marginBottom: 16,
          }}
        >
          <Typography.Title level={4} style={{ margin: 0 }}>
            {title}
          </Typography.Title>
          {extra ? <Space>{extra}</Space> : null}
        </div>
      )}
      {children}
    </div>
  )
}
