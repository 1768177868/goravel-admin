import type { ReactNode } from 'react'
import { Space, Typography } from 'antd'
import './PageContainer.scss'

interface PageContainerProps {
  title?: ReactNode
  extra?: ReactNode
  children: ReactNode
  className?: string
}

export default function PageContainer({ title, extra, children, className }: PageContainerProps) {
  return (
    <div className={`page-container${className ? ` ${className}` : ''}`}>
      {(title || extra) && (
        <div className="page-container__header">
          <div className="page-container__title-wrap">
            <Typography.Title level={4} className="page-container__title">
              {title}
            </Typography.Title>
          </div>
          {extra ? <Space className="page-container__extra" size={8}>{extra}</Space> : null}
        </div>
      )}
      <div className="page-container__body">{children}</div>
    </div>
  )
}
