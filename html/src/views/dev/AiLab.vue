<template>
  <div class="ai-lab-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <div>
            <div class="page-title">{{ t('ai_lab.title') }}</div>
            <div class="page-subtitle">{{ t('ai_lab.subtitle') }}</div>
          </div>
        </div>
      </template>

      <el-alert
        v-if="!loadingStatus && !aiEnabled"
        type="warning"
        :title="t('ai_lab.not_configured')"
        show-icon
        :closable="false"
        class="mb-16"
      />

      <el-card v-if="status" shadow="never" class="mb-16 status-card">
        <template #header>
          <span>{{ t('ai_lab.models') }}</span>
        </template>
        <el-descriptions :column="4" border size="small">
          <el-descriptions-item
            v-for="item in modelItems"
            :key="item.label"
            :label="item.label"
          >
            {{ item.value }}
          </el-descriptions-item>
        </el-descriptions>
        <div class="rate-limit-hint">
          {{ t('ai_lab.rate_limit_hint', {
            perMinute: status.rate_limit_per_minute,
            perDay: status.rate_limit_per_day,
            maxUpload: status.max_upload_mb ?? 10,
          }) }}
        </div>
      </el-card>

      <el-tabs v-model="activeTab" type="border-card">
        <el-tab-pane :label="t('ai_lab.tab_text')" name="text">
          <div class="ai-lab-tab-body">
            <el-form label-position="top">
              <el-form-item :label="t('ai_lab.prompt')">
                <el-input
                  v-model="textPrompt"
                  type="textarea"
                  :rows="4"
                  :placeholder="t('ai_lab.prompt_placeholder')"
                />
              </el-form-item>
              <el-form-item :label="t('ai_lab.system_prompt')">
                <el-input
                  v-model="systemPrompt"
                  type="textarea"
                  :rows="2"
                  :placeholder="t('ai_lab.system_prompt_placeholder')"
                />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" :loading="textLoading" :disabled="!aiEnabled" @click="handleText">
                  {{ t('ai_lab.submit') }}
                </el-button>
              </el-form-item>
            </el-form>
            <div v-if="textResult" class="lab-result">
              <div class="lab-result__label">{{ t('ai_lab.result') }}</div>
              <el-input v-model="textResult" type="textarea" :rows="8" readonly />
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane :label="t('ai_lab.tab_vision')" name="vision">
          <div class="ai-lab-tab-body">
            <div class="ai-lab-workspace">
              <div class="ai-lab-media">
                <div
                  class="ai-lab-media-box"
                  :class="{ 'ai-lab-media-box--filled': visionPreviewUrl }"
                >
                  <el-upload
                    v-if="!visionPreviewUrl"
                    ref="visionUploadRef"
                    drag
                    class="ai-lab-upload"
                    :auto-upload="false"
                    :show-file-list="false"
                    :limit="1"
                    accept="image/*"
                    :on-change="onVisionFileChange"
                    :on-remove="onVisionFileRemove"
                  >
                    <el-icon class="ai-lab-upload__icon"><UploadFilled /></el-icon>
                    <div class="ai-lab-upload__text">{{ t('ai_lab.upload_image') }}</div>
                    <template #tip>
                      <div class="ai-lab-upload__tip">{{ t('ai_lab.upload_image_hint') }}</div>
                    </template>
                  </el-upload>
                  <div v-else class="ai-lab-media-preview">
                    <img :src="visionPreviewUrl" alt="preview" class="preview-image preview-image--upload" />
                    <el-button
                      class="ai-lab-media-preview__remove"
                      circle
                      size="small"
                      :icon="CircleCloseFilled"
                      @click="clearVisionUpload"
                    />
                  </div>
                </div>
              </div>
              <div class="ai-lab-form-panel">
                <el-form label-position="top">
                  <el-form-item :label="t('ai_lab.prompt')">
                    <el-input v-model="visionPrompt" type="textarea" :rows="5" />
                  </el-form-item>
                  <el-form-item>
                    <el-button type="primary" :loading="visionLoading" :disabled="!aiEnabled" @click="handleVision">
                      {{ t('ai_lab.submit') }}
                    </el-button>
                  </el-form-item>
                </el-form>
              </div>
            </div>
            <div v-if="visionResult" class="lab-result">
              <div class="lab-result__label">{{ t('ai_lab.result') }}</div>
              <el-input v-model="visionResult" type="textarea" :rows="8" readonly />
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane :label="t('ai_lab.tab_image')" name="image">
          <div class="ai-lab-tab-body">
            <el-form label-position="top">
              <el-form-item :label="t('ai_lab.prompt')">
                <el-input v-model="imagePrompt" type="textarea" :rows="3" />
              </el-form-item>
              <el-form-item :label="t('ai_lab.size')">
                <el-select v-model="imageSize" style="width: 200px">
                  <el-option
                    v-for="opt in sizeOptions"
                    :key="opt.value"
                    :label="opt.label"
                    :value="opt.value"
                  />
                </el-select>
              </el-form-item>
              <el-form-item>
                <el-button type="primary" :loading="imageLoading" :disabled="!aiEnabled" @click="handleImage">
                  {{ t('ai_lab.generate') }}
                </el-button>
              </el-form-item>
            </el-form>
            <div v-if="imagePreview" class="media-preview-block">
              <div class="media-preview-block__label">{{ t('ai_lab.preview_image') }}</div>
              <img :src="imagePreview" alt="generated" class="preview-image" />
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane :label="t('ai_lab.tab_audio')" name="audio">
          <div class="ai-lab-tab-body">
            <el-form label-position="top">
              <el-form-item :label="t('ai_lab.prompt')">
                <el-input v-model="audioPrompt" type="textarea" :rows="3" />
              </el-form-item>
              <el-form-item :label="t('ai_lab.voice')">
                <el-select v-model="audioVoice" style="width: 200px">
                  <el-option
                    v-for="opt in voiceOptions"
                    :key="opt.value"
                    :label="opt.label"
                    :value="opt.value"
                  />
                </el-select>
              </el-form-item>
              <el-form-item>
                <el-button type="primary" :loading="audioLoading" :disabled="!aiEnabled" @click="handleAudio">
                  {{ t('ai_lab.generate') }}
                </el-button>
              </el-form-item>
            </el-form>
            <div v-if="audioPreview" class="media-preview-block">
              <div class="media-preview-block__label">{{ t('ai_lab.preview_audio') }}</div>
              <audio controls :src="audioPreview" class="preview-audio" />
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane :label="t('ai_lab.tab_transcription')" name="transcription">
          <div class="ai-lab-tab-body">
            <div class="ai-lab-workspace">
              <div class="ai-lab-media">
                <div
                  class="ai-lab-media-box"
                  :class="{ 'ai-lab-media-box--filled': transcriptionFileName }"
                >
                  <el-upload
                    v-if="!transcriptionFileName"
                    ref="transcriptionUploadRef"
                    drag
                    class="ai-lab-upload"
                    :auto-upload="false"
                    :show-file-list="false"
                    :limit="1"
                    accept="audio/*"
                    :on-change="onTranscriptionFileChange"
                    :on-remove="onTranscriptionFileRemove"
                  >
                    <el-icon class="ai-lab-upload__icon"><UploadFilled /></el-icon>
                    <div class="ai-lab-upload__text">{{ t('ai_lab.upload_audio') }}</div>
                    <template #tip>
                      <div class="ai-lab-upload__tip">{{ t('ai_lab.upload_audio_hint') }}</div>
                    </template>
                  </el-upload>
                  <div v-else class="ai-lab-file-chip">
                    <span class="ai-lab-file-chip__name" :title="transcriptionFileName">
                      {{ transcriptionFileName }}
                    </span>
                    <el-button link type="danger" @click="clearTranscriptionUpload">
                      {{ t('common.delete') }}
                    </el-button>
                  </div>
                </div>
              </div>
              <div class="ai-lab-form-panel">
                <el-form label-position="top">
                  <el-form-item :label="t('ai_lab.language')">
                    <el-input
                      v-model="transcriptionLanguage"
                      :placeholder="t('ai_lab.language_placeholder')"
                    />
                  </el-form-item>
                  <el-form-item>
                    <el-button
                      type="primary"
                      :loading="transcriptionLoading"
                      :disabled="!aiEnabled"
                      @click="handleTranscription"
                    >
                      {{ t('ai_lab.transcribe') }}
                    </el-button>
                  </el-form-item>
                </el-form>
              </div>
            </div>
            <div v-if="transcriptionResult" class="lab-result">
              <div class="lab-result__label">{{ t('ai_lab.result') }}</div>
              <el-input v-model="transcriptionResult" type="textarea" :rows="8" readonly />
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup>
import { CircleCloseFilled, UploadFilled } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import { useAiLab } from './ai-lab/useAiLab'

const { t } = useI18n()

const {
  status,
  loadingStatus,
  activeTab,
  aiEnabled,
  modelItems,
  sizeOptions,
  voiceOptions,
  textPrompt,
  systemPrompt,
  textResult,
  textLoading,
  visionPrompt,
  visionPreviewUrl,
  visionUploadRef,
  visionResult,
  visionLoading,
  imagePrompt,
  imageSize,
  imagePreview,
  imageLoading,
  audioPrompt,
  audioVoice,
  audioPreview,
  audioLoading,
  transcriptionFileName,
  transcriptionUploadRef,
  transcriptionLanguage,
  transcriptionResult,
  transcriptionLoading,
  handleText,
  handleVision,
  handleImage,
  handleAudio,
  handleTranscription,
  onVisionFileChange,
  onVisionFileRemove,
  clearVisionUpload,
  onTranscriptionFileChange,
  onTranscriptionFileRemove,
  clearTranscriptionUpload,
} = useAiLab()
</script>

<style scoped>
.ai-lab-page {
  padding: 0;
}

.card-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
}

.page-title {
  font-size: 16px;
  font-weight: 600;
}

.page-subtitle {
  margin-top: 4px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.mb-16 {
  margin-bottom: 16px;
}

.status-card :deep(.el-card__body) {
  padding-top: 12px;
}

.rate-limit-hint {
  margin-top: 12px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.ai-lab-tab-body {
  padding: 4px 0;
}

.ai-lab-workspace {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
  align-items: flex-start;
}

.ai-lab-media {
  flex: 0 0 260px;
  max-width: 100%;
}

.ai-lab-form-panel {
  flex: 1 1 280px;
  min-width: 0;
}

.ai-lab-media-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 168px;
  padding: 16px;
  background: var(--el-fill-color-light);
  border: 1px dashed var(--el-border-color);
  border-radius: 10px;
  transition: border-color 0.2s, background 0.2s;
}

.ai-lab-media-box--filled {
  background: var(--el-bg-color);
  border-style: solid;
}

.ai-lab-upload {
  width: 100%;
}

.ai-lab-upload :deep(.el-upload) {
  width: 100%;
}

.ai-lab-upload :deep(.el-upload-dragger) {
  padding: 20px 16px;
  background: transparent;
  border: none;
}

.ai-lab-upload__icon {
  margin-bottom: 8px;
  font-size: 40px;
  color: var(--el-color-primary);
}

.ai-lab-upload__text {
  font-size: 14px;
  color: var(--el-text-color-regular);
}

.ai-lab-upload__tip {
  margin-top: 4px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.ai-lab-media-preview {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
}

.ai-lab-media-preview__remove {
  position: absolute;
  top: -8px;
  right: -8px;
}

.preview-image {
  display: block;
  max-width: 100%;
  max-height: 360px;
  margin: 0 auto;
  border-radius: 8px;
  object-fit: contain;
}

.preview-image--upload {
  max-height: 140px;
}

.ai-lab-file-chip {
  display: flex;
  gap: 8px;
  align-items: center;
  width: 100%;
  padding: 10px 12px;
  background: var(--el-fill-color-light);
  border-radius: 8px;
}

.ai-lab-file-chip__name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  font-size: 13px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.lab-result,
.media-preview-block {
  margin-top: 20px;
  padding: 16px;
  background: var(--el-fill-color-light);
  border-radius: 10px;
}

.lab-result__label,
.media-preview-block__label {
  margin-bottom: 10px;
  font-size: 14px;
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.preview-audio {
  width: 100%;
}

@media (max-width: 768px) {
  .ai-lab-media {
    flex-basis: 100%;
  }
}
</style>
