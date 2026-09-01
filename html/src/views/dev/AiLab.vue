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
          <el-form label-width="120px">
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
            <el-form-item v-if="textResult" :label="t('ai_lab.result')">
              <el-input v-model="textResult" type="textarea" :rows="8" readonly />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane :label="t('ai_lab.tab_vision')" name="vision">
          <el-row :gutter="16">
            <el-col :xs="24" :md="10">
              <el-upload
                drag
                :auto-upload="false"
                :limit="1"
                accept="image/*"
                :on-change="onVisionFileChange"
                :on-remove="onVisionFileRemove"
              >
                <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
                <div class="el-upload__text">{{ t('ai_lab.upload_image') }}</div>
                <template #tip>
                  <div class="el-upload__tip">{{ t('ai_lab.upload_image_hint') }}</div>
                </template>
              </el-upload>
            </el-col>
            <el-col :xs="24" :md="14">
              <el-form label-width="80px">
                <el-form-item :label="t('ai_lab.prompt')">
                  <el-input v-model="visionPrompt" type="textarea" :rows="4" />
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" :loading="visionLoading" :disabled="!aiEnabled" @click="handleVision">
                    {{ t('ai_lab.submit') }}
                  </el-button>
                </el-form-item>
              </el-form>
            </el-col>
          </el-row>
          <el-form v-if="visionResult" label-width="80px" class="mt-16">
            <el-form-item :label="t('ai_lab.result')">
              <el-input v-model="visionResult" type="textarea" :rows="8" readonly />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane :label="t('ai_lab.tab_image')" name="image">
          <el-form label-width="80px">
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
          <div v-if="imagePreview" class="media-preview">
            <div class="media-preview__label">{{ t('ai_lab.preview_image') }}</div>
            <img :src="imagePreview" alt="generated" class="preview-image" />
          </div>
        </el-tab-pane>

        <el-tab-pane :label="t('ai_lab.tab_audio')" name="audio">
          <el-form label-width="80px">
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
          <div v-if="audioPreview" class="media-preview">
            <div class="media-preview__label">{{ t('ai_lab.preview_audio') }}</div>
            <audio controls :src="audioPreview" class="preview-audio" />
          </div>
        </el-tab-pane>

        <el-tab-pane :label="t('ai_lab.tab_transcription')" name="transcription">
          <el-row :gutter="16">
            <el-col :xs="24" :md="10">
              <el-upload
                drag
                :auto-upload="false"
                :limit="1"
                accept="audio/*"
                :on-change="onTranscriptionFileChange"
                :on-remove="onTranscriptionFileRemove"
              >
                <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
                <div class="el-upload__text">{{ t('ai_lab.upload_audio') }}</div>
                <template #tip>
                  <div class="el-upload__tip">{{ t('ai_lab.upload_audio_hint') }}</div>
                </template>
              </el-upload>
            </el-col>
            <el-col :xs="24" :md="14">
              <el-form label-width="100px">
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
            </el-col>
          </el-row>
          <el-form v-if="transcriptionResult" label-width="80px" class="mt-16">
            <el-form-item :label="t('ai_lab.result')">
              <el-input v-model="transcriptionResult" type="textarea" :rows="8" readonly />
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup>
import { UploadFilled } from '@element-plus/icons-vue'
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
  onTranscriptionFileChange,
  onTranscriptionFileRemove,
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

.mt-16 {
  margin-top: 16px;
}

.status-card :deep(.el-card__body) {
  padding-top: 12px;
}

.rate-limit-hint {
  margin-top: 12px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.media-preview {
  margin-top: 16px;
}

.media-preview__label {
  margin-bottom: 8px;
  color: var(--el-text-color-regular);
}

.preview-image {
  max-width: 100%;
  max-height: 480px;
  border-radius: 8px;
}

.preview-audio {
  width: 100%;
}
</style>
