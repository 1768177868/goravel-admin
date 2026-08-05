import { Avatar, Button, Input, Modal } from 'antd'
import { UserOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import { useUserStore } from '@/stores/user'
import type { UseLayoutLockScreenReturn } from '@/hooks/useLayoutLockScreen'
import './LayoutLockScreen.scss'

interface LayoutLockScreenProps {
  lockScreen: UseLayoutLockScreenReturn
}

export default function LayoutLockScreen({ lockScreen }: LayoutLockScreenProps) {
  const { t } = useTranslation()
  const adminInfo = useUserStore((s) => s.adminInfo)

  const {
    isScreenLocked,
    lockDialogVisible,
    pendingLockPassword,
    setPendingLockPassword,
    lockDialogError,
    unlockPassword,
    unlockError,
    lockInputName,
    lockDialogInputName,
    cancelLockDialog,
    confirmLockScreen,
    handleUnlockInput,
    handleUnlockScreen,
    goToLogin,
  } = lockScreen

  const displayName = adminInfo?.nickname || adminInfo?.username

  return (
    <>
      <Modal
        title={t('header.lock_screen')}
        open={lockDialogVisible}
        width={420}
        maskClosable={false}
        keyboard={false}
        onCancel={cancelLockDialog}
        onOk={confirmLockScreen}
        okText={t('common.confirm')}
        cancelText={t('common.cancel')}
        destroyOnHidden
      >
        <Input.Password
          value={pendingLockPassword}
          name={lockDialogInputName}
          autoComplete="new-password"
          autoCorrect="off"
          spellCheck={false}
          placeholder={t('header.lock_password_placeholder')}
          onChange={(e) => setPendingLockPassword(e.target.value)}
          onPressEnter={confirmLockScreen}
        />
        {lockDialogError && <div className="lock-screen-error">{lockDialogError}</div>}
      </Modal>

      {isScreenLocked && (
        <div className="lock-screen-overlay">
          <div className="lock-screen-card">
            <div className="lock-screen-avatar-wrap">
              <Avatar size={68} src={adminInfo?.avatar} icon={<UserOutlined />} />
            </div>
            <div className="lock-screen-title">{t('header.lock_screen_title')}</div>
            <div className="lock-screen-user">{displayName}</div>
            <Input.Password
              className="lock-screen-input"
              value={unlockPassword}
              name={lockInputName}
              autoComplete="new-password"
              autoCorrect="off"
              spellCheck={false}
              placeholder={t('header.lock_password_placeholder')}
              onChange={(e) => handleUnlockInput(e.target.value)}
              onPressEnter={handleUnlockScreen}
            />
            {unlockError && <div className="lock-screen-error">{unlockError}</div>}
            <div className="lock-screen-actions">
              <Button type="primary" onClick={handleUnlockScreen}>
                {t('header.unlock')}
              </Button>
              <Button onClick={() => void goToLogin()}>{t('header.back_to_login')}</Button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}
