import { useCallback, useState } from 'react'
import { App, Button, Modal, Slider, Space, Upload } from 'antd'
import Cropper from 'react-easy-crop'
import type { Area } from 'react-easy-crop'
import { useTranslation } from 'react-i18next'
import { uploadAttachment } from '@/api/attachment'
import { useUnhandledError } from '@/hooks/useUnhandledError'

interface CropUploadModalProps {
  open: boolean
  onClose: () => void
  onSuccess: () => void
}

async function getCroppedBlob(imageSrc: string, pixelCrop: Area): Promise<Blob> {
  const image = await createImage(imageSrc)
  const canvas = document.createElement('canvas')
  const ctx = canvas.getContext('2d')
  if (!ctx) throw new Error('Canvas not supported')

  canvas.width = pixelCrop.width
  canvas.height = pixelCrop.height
  ctx.drawImage(
    image,
    pixelCrop.x,
    pixelCrop.y,
    pixelCrop.width,
    pixelCrop.height,
    0,
    0,
    pixelCrop.width,
    pixelCrop.height,
  )

  return new Promise((resolve, reject) => {
    canvas.toBlob(
      (blob) => {
        if (blob) resolve(blob)
        else reject(new Error('Crop failed'))
      },
      'image/png',
      1,
    )
  })
}

function createImage(url: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const image = new Image()
    image.addEventListener('load', () => resolve(image))
    image.addEventListener('error', (error) => reject(error))
    image.setAttribute('crossOrigin', 'anonymous')
    image.src = url
  })
}

export default function CropUploadModal({ open, onClose, onSuccess }: CropUploadModalProps) {
  const { t } = useTranslation()
  const { message } = App.useApp()
  const showError = useUnhandledError()
  const [imageSrc, setImageSrc] = useState('')
  const [fileName, setFileName] = useState('')
  const [crop, setCrop] = useState({ x: 0, y: 0 })
  const [zoom, setZoom] = useState(1)
  const [croppedAreaPixels, setCroppedAreaPixels] = useState<Area | null>(null)
  const [uploading, setUploading] = useState(false)

  const reset = () => {
    setImageSrc('')
    setFileName('')
    setCrop({ x: 0, y: 0 })
    setZoom(1)
    setCroppedAreaPixels(null)
  }

  const handleClose = () => {
    reset()
    onClose()
  }

  const onCropComplete = useCallback((_area: Area, pixels: Area) => {
    setCroppedAreaPixels(pixels)
  }, [])

  const handleSelect = (file: File) => {
    if (!file.type.startsWith('image/')) {
      message.error(t('attachment.only_image_allowed'))
      return false
    }
    const maxSize = 10 * 1024 * 1024
    if (file.size > maxSize) {
      message.error(t('attachment.file_too_large'))
      return false
    }
    setFileName(file.name)
    const reader = new FileReader()
    reader.onload = () => {
      setImageSrc(String(reader.result || ''))
      setCrop({ x: 0, y: 0 })
      setZoom(1)
    }
    reader.readAsDataURL(file)
    return false
  }

  const handleConfirm = async () => {
    if (!imageSrc || !croppedAreaPixels) {
      message.warning(t('attachment.please_select_image'))
      return
    }
    setUploading(true)
    try {
      const blob = await getCroppedBlob(imageSrc, croppedAreaPixels)
      const file = new File([blob], fileName || 'cropped-image.png', { type: blob.type || 'image/png' })
      await uploadAttachment(file)
      message.success(t('attachment.upload_success'))
      reset()
      onSuccess()
      onClose()
    } catch (error) {
      showError(error, t('attachment.upload_failed'))
    } finally {
      setUploading(false)
    }
  }

  return (
    <Modal
      open={open}
      title={t('attachment.crop_upload')}
      width={800}
      onCancel={handleClose}
      destroyOnHidden
      footer={
        <Space>
          <Upload accept="image/*" showUploadList={false} beforeUpload={handleSelect}>
            <Button>{t('attachment.select_image')}</Button>
          </Upload>
          <Button onClick={handleClose}>{t('common.cancel')}</Button>
          <Button type="primary" loading={uploading} onClick={() => void handleConfirm()}>
            {t('common.confirm')}
          </Button>
        </Space>
      }
    >
      <div style={{ position: 'relative', height: 460, background: '#111', borderRadius: 8 }}>
        {imageSrc ? (
          <Cropper
            image={imageSrc}
            crop={crop}
            zoom={zoom}
            aspect={1}
            onCropChange={setCrop}
            onZoomChange={setZoom}
            onCropComplete={onCropComplete}
          />
        ) : (
          <div
            style={{
              height: '100%',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: '#999',
            }}
          >
            {t('attachment.please_select_image')}
          </div>
        )}
      </div>
      {imageSrc ? (
        <div style={{ marginTop: 16 }}>
          <Slider min={1} max={3} step={0.05} value={zoom} onChange={setZoom} />
        </div>
      ) : null}
    </Modal>
  )
}
