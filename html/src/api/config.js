import request from '../utils/request'

// 根据分组获取配置
export function getConfigByGroup(group) {
  return request({
    url: `/configs/group/${group}`,
    method: 'get'
  })
}

// 保存配置（按分组批量保存）
export function saveConfig(group, configs) {
  return request({
    url: '/configs/save',
    method: 'post',
    data: {
      group,
      configs
    }
  })
}

// 测试邮件发送
export function testEmail(emailConfig) {
  return request({
    url: '/configs/test-email',
    method: 'post',
    data: emailConfig
  })
}

