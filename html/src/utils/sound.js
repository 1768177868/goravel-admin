/**
 * 声音工具函数
 * 使用 Web Audio API 生成提示音
 */

/**
 * 播放通知提示音
 * 生成一个简短、悦耳的提示音
 */
export function playNotificationSound() {
  try {
    // 创建音频上下文
    const audioContext = new (window.AudioContext || window.webkitAudioContext)()
    
    // 创建一个振荡器节点（用于生成音调）
    const oscillator = audioContext.createOscillator()
    const gainNode = audioContext.createGain()
    
    // 连接节点
    oscillator.connect(gainNode)
    gainNode.connect(audioContext.destination)
    
    // 设置音调频率（800Hz，比较悦耳）
    oscillator.frequency.value = 800
    oscillator.type = 'sine' // 正弦波，声音更柔和
    
    // 设置音量包络（淡入淡出，避免突然的声音）
    const now = audioContext.currentTime
    gainNode.gain.setValueAtTime(0, now)
    gainNode.gain.linearRampToValueAtTime(0.3, now + 0.01) // 快速淡入
    gainNode.gain.exponentialRampToValueAtTime(0.01, now + 0.15) // 快速淡出
    
    // 播放音调
    oscillator.start(now)
    oscillator.stop(now + 0.15) // 持续时间 150ms
    
    // 清理资源
    oscillator.onended = () => {
      audioContext.close()
    }
  } catch (error) {
    // 如果浏览器不支持 Web Audio API，静默失败
    console.warn('Failed to play notification sound:', error)
  }
}

