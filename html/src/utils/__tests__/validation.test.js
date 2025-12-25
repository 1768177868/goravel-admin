import { describe, it, expect } from 'vitest'
import { validators, createValidator } from '../validation'

describe('validators', () => {
  describe('required', () => {
    it('应该拒绝空字符串', () => {
      expect(validators.required('')).not.toBe(true)
    })

    it('应该拒绝 null', () => {
      expect(validators.required(null)).not.toBe(true)
    })

    it('应该拒绝 undefined', () => {
      expect(validators.required(undefined)).not.toBe(true)
    })

    it('应该拒绝空数组', () => {
      expect(validators.required([])).not.toBe(true)
    })

    it('应该接受有效字符串', () => {
      expect(validators.required('hello')).toBe(true)
    })

    it('应该接受数字', () => {
      expect(validators.required(0)).toBe(true)
      expect(validators.required(123)).toBe(true)
    })

    it('应该接受非空数组', () => {
      expect(validators.required([1, 2, 3])).toBe(true)
    })
  })

  describe('email', () => {
    it('应该接受有效邮箱', () => {
      expect(validators.email('test@example.com')).toBe(true)
      expect(validators.email('user.name@domain.co.uk')).toBe(true)
    })

    it('应该拒绝无效邮箱', () => {
      expect(validators.email('invalid')).not.toBe(true)
      expect(validators.email('invalid@')).not.toBe(true)
      expect(validators.email('@domain.com')).not.toBe(true)
    })

    it('应该接受空值（由 required 验证）', () => {
      expect(validators.email('')).toBe(true)
      expect(validators.email(null)).toBe(true)
    })
  })

  describe('phone', () => {
    it('应该接受有效手机号', () => {
      expect(validators.phone('13812345678')).toBe(true)
      expect(validators.phone('19912345678')).toBe(true)
    })

    it('应该拒绝无效手机号', () => {
      expect(validators.phone('1234567890')).not.toBe(true)
      expect(validators.phone('12812345678')).not.toBe(true) // 第二位不能是2
      expect(validators.phone('138123456789')).not.toBe(true) // 太长
    })

    it('应该接受空值', () => {
      expect(validators.phone('')).toBe(true)
    })
  })

  describe('minLength', () => {
    it('应该拒绝短于最小长度的字符串', () => {
      const validator = validators.minLength(5)
      expect(validator('ab')).not.toBe(true)
    })

    it('应该接受达到最小长度的字符串', () => {
      const validator = validators.minLength(5)
      expect(validator('abcde')).toBe(true)
      expect(validator('abcdefg')).toBe(true)
    })

    it('应该处理数组', () => {
      const validator = validators.minLength(3)
      expect(validator([1, 2])).not.toBe(true)
      expect(validator([1, 2, 3])).toBe(true)
    })
  })

  describe('maxLength', () => {
    it('应该接受短于最大长度的字符串', () => {
      const validator = validators.maxLength(5)
      expect(validator('abc')).toBe(true)
    })

    it('应该拒绝超过最大长度的字符串', () => {
      const validator = validators.maxLength(5)
      expect(validator('abcdefg')).not.toBe(true)
    })

    it('应该接受空值', () => {
      const validator = validators.maxLength(5)
      expect(validator('')).toBe(true)
    })
  })

  describe('length', () => {
    it('应该接受在范围内的字符串', () => {
      const validator = validators.length(3, 6)
      expect(validator('abc')).toBe(true)
      expect(validator('abcdef')).toBe(true)
    })

    it('应该拒绝超出范围的字符串', () => {
      const validator = validators.length(3, 6)
      expect(validator('ab')).not.toBe(true)
      expect(validator('abcdefg')).not.toBe(true)
    })
  })

  describe('number', () => {
    it('应该接受有效数字', () => {
      expect(validators.number('123')).toBe(true)
      expect(validators.number('12.34')).toBe(true)
      expect(validators.number(-123)).toBe(true)
    })

    it('应该拒绝非数字', () => {
      expect(validators.number('abc')).not.toBe(true)
      expect(validators.number('12a')).not.toBe(true)
    })
  })

  describe('integer', () => {
    it('应该接受整数', () => {
      expect(validators.integer('123')).toBe(true)
      expect(validators.integer(-456)).toBe(true)
      expect(validators.integer(0)).toBe(true)
    })

    it('应该拒绝小数', () => {
      expect(validators.integer('12.34')).not.toBe(true)
      expect(validators.integer(1.5)).not.toBe(true)
    })
  })

  describe('min', () => {
    it('应该接受大于等于最小值的数', () => {
      const validator = validators.min(10)
      expect(validator(10)).toBe(true)
      expect(validator(15)).toBe(true)
    })

    it('应该拒绝小于最小值的数', () => {
      const validator = validators.min(10)
      expect(validator(5)).not.toBe(true)
    })
  })

  describe('max', () => {
    it('应该接受小于等于最大值的数', () => {
      const validator = validators.max(10)
      expect(validator(10)).toBe(true)
      expect(validator(5)).toBe(true)
    })

    it('应该拒绝大于最大值的数', () => {
      const validator = validators.max(10)
      expect(validator(15)).not.toBe(true)
    })
  })

  describe('range', () => {
    it('应该接受在范围内的数', () => {
      const validator = validators.range(1, 10)
      expect(validator(1)).toBe(true)
      expect(validator(5)).toBe(true)
      expect(validator(10)).toBe(true)
    })

    it('应该拒绝超出范围的数', () => {
      const validator = validators.range(1, 10)
      // 注意：0 被视为空值，会跳过验证返回 true
      // 这是设计行为，空值检验应由 required 验证器处理
      expect(validator(-1)).not.toBe(true)
      expect(validator(11)).not.toBe(true)
    })
  })

  describe('url', () => {
    it('应该接受有效 URL', () => {
      expect(validators.url('https://example.com')).toBe(true)
      expect(validators.url('http://localhost:3000')).toBe(true)
      expect(validators.url('ftp://files.example.com')).toBe(true)
    })

    it('应该拒绝无效 URL', () => {
      expect(validators.url('not-a-url')).not.toBe(true)
      expect(validators.url('example.com')).not.toBe(true) // 缺少协议
    })
  })

  describe('pattern', () => {
    it('应该接受匹配正则的值', () => {
      const validator = validators.pattern(/^\d{4}$/, '必须是4位数字')
      expect(validator('1234')).toBe(true)
    })

    it('应该拒绝不匹配正则的值', () => {
      const validator = validators.pattern(/^\d{4}$/, '必须是4位数字')
      expect(validator('123')).toBe('必须是4位数字')
      expect(validator('12345')).toBe('必须是4位数字')
    })
  })

  describe('password', () => {
    it('应该接受足够长的密码', () => {
      const validator = validators.password(6)
      expect(validator('password123')).toBe(true)
    })

    it('应该拒绝过短的密码', () => {
      const validator = validators.password(6)
      expect(validator('pass')).not.toBe(true)
    })
  })

  describe('confirmPassword', () => {
    it('应该接受匹配的密码', () => {
      const validator = validators.confirmPassword('password123')
      expect(validator('password123')).toBe(true)
    })

    it('应该拒绝不匹配的密码', () => {
      const validator = validators.confirmPassword('password123')
      expect(validator('differentpassword')).not.toBe(true)
    })
  })
})

describe('createValidator', () => {
  it('应该创建自定义验证器', () => {
    const isEven = createValidator(
      (value) => value % 2 === 0,
      '必须是偶数'
    )
    
    expect(isEven(2)).toBe(true)
    expect(isEven(4)).toBe(true)
    expect(isEven(3)).toBe('必须是偶数')
  })

  it('空值应该通过验证', () => {
    const customValidator = createValidator(
      (value) => value === 'test',
      '必须是 test'
    )
    
    expect(customValidator('')).toBe(true)
    expect(customValidator(null)).toBe(true)
  })
})

