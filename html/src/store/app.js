import { defineStore } from 'pinia'
import Storage from '../utils/storage'

const detectBrowserTimezone = () => {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC'
  } catch {
    return 'UTC'
  }
}

export const useAppStore = defineStore('app', {
  state: () => ({
    sidebarCollapsed: Storage.getItem('sidebarCollapsed', 'false') === 'true',
    layoutSize: Storage.getItem('layoutSize', 'default') || 'default', // default, large, small
    isFullscreen: false,
    timezone: Storage.getItem('timezone', detectBrowserTimezone()) || detectBrowserTimezone(),
    theme: Storage.getItem('theme', 'light') || 'light' // light, dark
  }),

  actions: {
    toggleSidebar() {
      this.sidebarCollapsed = !this.sidebarCollapsed
      Storage.setItem('sidebarCollapsed', this.sidebarCollapsed.toString())
    },

    setSidebarCollapsed(collapsed) {
      this.sidebarCollapsed = collapsed
      Storage.setItem('sidebarCollapsed', collapsed.toString())
    },

    setLayoutSize(size) {
      this.layoutSize = size
      Storage.setItem('layoutSize', size)
      // 应用布局大小到 body
      document.body.className = document.body.className.replace(/layout-\w+/g, '')
      document.body.classList.add(`layout-${size}`)
    },

    toggleFullscreen() {
      if (!document.fullscreenElement) {
        document.documentElement.requestFullscreen().then(() => {
          this.isFullscreen = true
        }).catch(() => {
          console.error('无法进入全屏模式')
        })
      } else {
        document.exitFullscreen().then(() => {
          this.isFullscreen = false
        }).catch(() => {
          console.error('无法退出全屏模式')
        })
      }
    },

    setTimezone(timezone) {
      this.timezone = timezone || 'UTC'
      Storage.setItem('timezone', this.timezone)
    },

    setTheme(theme) {
      this.theme = theme || 'light'
      Storage.setItem('theme', this.theme)
      // 应用主题到 html 和 body 元素（body 用于 teleport 渲染的元素）
      document.documentElement.setAttribute('data-theme', this.theme)
      document.documentElement.classList.toggle('dark-mode', this.theme === 'dark')
      document.body.classList.toggle('dark-mode', this.theme === 'dark')
      
      // 动态注入下拉框样式
      this.injectDropdownStyles()
      // 应用选择器输入框样式
      this.applySelectInputStyles()
    },
    
    injectDropdownStyles() {
      const styleId = 'dark-mode-dropdown-styles'
      let styleElement = document.getElementById(styleId)
      
      if (!styleElement) {
        styleElement = document.createElement('style')
        styleElement.id = styleId
        document.head.appendChild(styleElement)
      }
      
      if (this.theme === 'dark') {
        styleElement.textContent = `
          .el-select-dropdown {
            background-color: #252526 !important;
            border-color: #3d3e40 !important;
            box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.3) !important;
          }
          .el-select-dropdown__wrap {
            background-color: #252526 !important;
          }
          .el-select-dropdown__list {
            background-color: #252526 !important;
          }
          .el-select-dropdown__item {
            color: #cfd3dc !important;
            background-color: transparent !important;
          }
          .el-select-dropdown__item:hover {
            background-color: #2d2d30 !important;
            color: #e5eaf3 !important;
          }
          .el-select-dropdown__item.is-selected {
            color: #409EFF !important;
            background-color: #2d2d30 !important;
            font-weight: 500;
          }
          .el-select-dropdown__item.is-disabled {
            color: #6d6d6d !important;
            cursor: not-allowed;
          }
          .el-select-dropdown__empty {
            color: #a3a6ad !important;
          }
          .el-select-dropdown__loading {
            color: #a3a6ad !important;
          }
        `
        
        // 使用MutationObserver监听下拉框的创建并直接应用样式
        this.setupDropdownObserver()
      } else {
        styleElement.textContent = ''
        // 清理observer
        if (this.dropdownObserver) {
          this.dropdownObserver.disconnect()
          this.dropdownObserver = null
        }
      }
    },
    
    setupDropdownObserver() {
      // 如果已经存在observer，先断开
      if (this.dropdownObserver) {
        this.dropdownObserver.disconnect()
      }
      
      const self = this
      
      // 创建MutationObserver来监听下拉框的创建
      this.dropdownObserver = new MutationObserver((mutations) => {
        mutations.forEach((mutation) => {
          mutation.addedNodes.forEach((node) => {
            if (node.nodeType === 1) { // Element node
              // 检查是否是下拉框元素
              if (node.classList && node.classList.contains('el-select-dropdown')) {
                self.applyDropdownStyles(node)
              }
              // 也检查子元素
              if (node.querySelectorAll) {
                const dropdowns = node.querySelectorAll('.el-select-dropdown')
                dropdowns.forEach(dropdown => self.applyDropdownStyles(dropdown))
              }
            }
          })
        })
      })
      
      // 开始观察body的变化
      this.dropdownObserver.observe(document.body, {
        childList: true,
        subtree: true
      })
      
      // 立即检查已存在的下拉框
      setTimeout(() => {
        document.querySelectorAll('.el-select-dropdown').forEach(dropdown => {
          this.applyDropdownStyles(dropdown)
        })
      }, 100)
    },
    
    applyDropdownStyles(element) {
      if (!element) return
      
      // 如果是白天模式，清除内联样式
      if (this.theme !== 'dark') {
        element.style.removeProperty('background-color')
        element.style.removeProperty('border-color')
        element.style.removeProperty('box-shadow')
        
        const wrap = element.querySelector('.el-select-dropdown__wrap')
        if (wrap) {
          wrap.style.removeProperty('background-color')
        }
        
        const list = element.querySelector('.el-select-dropdown__list')
        if (list) {
          list.style.removeProperty('background-color')
        }
        
        const items = element.querySelectorAll('.el-select-dropdown__item')
        items.forEach(item => {
          item.style.removeProperty('color')
          item.style.removeProperty('background-color')
        })
        return
      }
      
      // 直接设置样式，使用setProperty确保优先级
      element.style.setProperty('background-color', '#252526', 'important')
      element.style.setProperty('border-color', '#3d3e40', 'important')
      element.style.setProperty('box-shadow', '0 2px 12px 0 rgba(0, 0, 0, 0.3)', 'important')
      
      // 设置子元素样式
      const wrap = element.querySelector('.el-select-dropdown__wrap')
      if (wrap) {
        wrap.style.setProperty('background-color', '#252526', 'important')
      }
      
      const list = element.querySelector('.el-select-dropdown__list')
      if (list) {
        list.style.setProperty('background-color', '#252526', 'important')
      }
      
      const items = element.querySelectorAll('.el-select-dropdown__item')
      items.forEach(item => {
        item.style.setProperty('color', '#cfd3dc', 'important')
        item.style.setProperty('background-color', 'transparent', 'important')
      })
    },
    
    applySelectInputStyles() {
      // 为所有选择器输入框应用样式
      const styleId = 'dark-mode-select-input-styles'
      let styleElement = document.getElementById(styleId)
      
      if (!styleElement) {
        styleElement = document.createElement('style')
        styleElement.id = styleId
        document.head.appendChild(styleElement)
      }
      
      // 如果是白天模式，清除样式
      if (this.theme !== 'dark') {
        styleElement.textContent = ''
        return
      }
      
      styleElement.textContent = `
        .el-select__wrapper {
          background-color: #252526 !important;
          border-width: 1px !important;
        }
        .el-select .el-input__wrapper {
          background-color: #252526 !important;
          border-width: 1px !important;
        }
        .el-select .el-input__inner {
          background-color: #252526 !important;
          color: #e5eaf3 !important;
        }
        .el-select__selection {
          color: #e5eaf3 !important;
        }
        .el-select__placeholder {
          color: #6d6d6d !important;
        }
        .el-select__caret {
          color: #cfd3dc !important;
        }
        .el-input__wrapper::before,
        .el-input__wrapper::after {
          background-color: transparent !important;
        }
        .el-input__wrapper * {
          background-color: transparent !important;
        }
        .el-input__wrapper .el-input__inner {
          background-color: #252526 !important;
        }
        .el-input__wrapper.is-disabled::before,
        .el-input__wrapper.is-disabled::after,
        .el-input__wrapper:disabled::before,
        .el-input__wrapper:disabled::after {
          background-color: transparent !important;
        }
        .el-input__wrapper.is-disabled *,
        .el-input__wrapper:disabled * {
          background-color: transparent !important;
        }
        .el-input__wrapper.is-disabled .el-input__inner,
        .el-input__wrapper:disabled .el-input__inner,
        .el-input.is-disabled .el-input__wrapper .el-input__inner {
          background-color: #252526 !important;
        }
        .el-select__wrapper::before,
        .el-select__wrapper::after {
          background-color: transparent !important;
        }
        .el-select__wrapper * {
          background-color: transparent !important;
        }
        .el-select__wrapper .el-select__selection {
          background-color: transparent !important;
        }
        .el-select__wrapper .el-input__inner {
          background-color: #252526 !important;
        }
        .el-select.is-disabled .el-select__wrapper::before,
        .el-select.is-disabled .el-select__wrapper::after {
          background-color: transparent !important;
        }
        .el-select.is-disabled .el-select__wrapper * {
          background-color: transparent !important;
        }
        .el-select.is-disabled .el-select__wrapper .el-input__inner {
          background-color: #252526 !important;
        }
        .vxe-input::before,
        .vxe-input::after {
          background-color: transparent !important;
        }
        .vxe-input * {
          background-color: transparent !important;
        }
        .vxe-input .vxe-input--inner {
          background-color: #252526 !important;
        }
        .vxe-input.is--disabled::before,
        .vxe-input.is--disabled::after,
        .vxe-input:disabled::before,
        .vxe-input:disabled::after {
          background-color: transparent !important;
        }
        .vxe-input.is--disabled *,
        .vxe-input:disabled * {
          background-color: transparent !important;
        }
        .vxe-input.is--disabled .vxe-input--inner,
        .vxe-input:disabled .vxe-input--inner {
          background-color: #252526 !important;
        }
        .vxe-select::before,
        .vxe-select::after {
          background-color: transparent !important;
        }
        .vxe-select * {
          background-color: transparent !important;
        }
        .vxe-select .vxe-input--inner {
          background-color: #252526 !important;
        }
        .vxe-select.is--disabled::before,
        .vxe-select.is--disabled::after,
        .vxe-select:disabled::before,
        .vxe-select:disabled::after {
          background-color: transparent !important;
        }
        .vxe-select.is--disabled *,
        .vxe-select:disabled * {
          background-color: transparent !important;
        }
        .vxe-select.is--disabled .vxe-input--inner,
        .vxe-select:disabled .vxe-input--inner {
          background-color: #252526 !important;
        }
        .el-input__wrapper.is-disabled,
        .el-input__wrapper:disabled,
        .el-input.is-disabled .el-input__wrapper {
          background-color: #252526 !important;
          border-color: #3d3e40 !important;
          cursor: not-allowed !important;
        }
        .el-input__inner:disabled,
        .el-input.is-disabled .el-input__inner {
          background-color: #252526 !important;
          color: #6d6d6d !important;
          cursor: not-allowed !important;
        }
        .el-textarea__inner:disabled,
        .el-textarea.is-disabled .el-textarea__inner {
          background-color: #2d2d30 !important;
          color: #6d6d6d !important;
          cursor: not-allowed !important;
        }
        .el-select.is-disabled .el-input__wrapper,
        .el-select.is-disabled .el-select__wrapper {
          background-color: #252526 !important;
          border-color: #3d3e40 !important;
          cursor: not-allowed !important;
        }
        .el-select.is-disabled .el-input__inner {
          background-color: #252526 !important;
          color: #6d6d6d !important;
          cursor: not-allowed !important;
        }
        .el-input-number.is-disabled .el-input__wrapper {
          background-color: #252526 !important;
          border-color: #3d3e40 !important;
          cursor: not-allowed !important;
        }
        .el-input-number.is-disabled .el-input__inner {
          background-color: #252526 !important;
          color: #6d6d6d !important;
          cursor: not-allowed !important;
        }
        .vxe-input.is--disabled,
        .vxe-input:disabled {
          background-color: #252526 !important;
          border-color: #3d3e40 !important;
          color: #6d6d6d !important;
          cursor: not-allowed !important;
        }
        .vxe-input.is--disabled .vxe-input--inner,
        .vxe-input:disabled .vxe-input--inner {
          background-color: #252526 !important;
          color: #6d6d6d !important;
        }
        .vxe-select.is--disabled,
        .vxe-select:disabled {
          background-color: #252526 !important;
          border-color: #3d3e40 !important;
          cursor: not-allowed !important;
        }
        .vxe-select.is--disabled .vxe-input--inner,
        .vxe-select:disabled .vxe-input--inner {
          background-color: #252526 !important;
          color: #6d6d6d !important;
        }
        .vxe-number-input.is--disabled,
        .vxe-number-input:disabled {
          background-color: #252526 !important;
          border-color: #3d3e40 !important;
          cursor: not-allowed !important;
        }
        .vxe-number-input.is--disabled .vxe-number-input--input,
        .vxe-number-input:disabled .vxe-number-input--input {
          background-color: #252526 !important;
          color: #6d6d6d !important;
        }
        .el-tree-node__content:hover {
          background-color: #2d2d30 !important;
        }
        .el-tree-node__content:hover .el-tree-node__label {
          color: #e5eaf3 !important;
        }
        .el-table__row:hover {
          background-color: #2d2d30 !important;
        }
        .el-table__row:hover > td {
          background-color: #2d2d30 !important;
        }
        .el-table__body-wrapper .el-table__row:hover {
          background-color: #2d2d30 !important;
        }
        .el-table__body-wrapper .el-table__row:hover > td {
          background-color: #2d2d30 !important;
        }
        .vxe-pager .vxe-pager--btn-wrapper {
          background-color: #252526 !important;
          border-color: #4c4d4f !important;
          color: #cfd3dc !important;
        }
        .vxe-pager .vxe-pager--btn-wrapper:hover {
          color: #409EFF !important;
          background-color: #2d2d30 !important;
        }
        .vxe-pager .vxe-pager--btn-wrapper.is--active {
          background-color: #409EFF !important;
          color: #fff !important;
          border-color: #409EFF !important;
        }
        .vxe-pager--jump-prev,
        .vxe-pager--jump-next,
        .vxe-pager--prev-page,
        .vxe-pager--next-page,
        .vxe-pager--prev-btn,
        .vxe-pager--next-btn,
        .vxe-pager--num-btn,
        .vxe-pager .vxe-pager--jump-prev,
        .vxe-pager .vxe-pager--jump-next,
        .vxe-pager .vxe-pager--prev-page,
        .vxe-pager .vxe-pager--next-page,
        .vxe-pager .vxe-pager--prev-btn,
        .vxe-pager .vxe-pager--next-btn,
        .vxe-pager .vxe-pager--num-btn,
        [class*="vxe-pager--jump-prev"],
        [class*="vxe-pager--jump-next"],
        [class*="vxe-pager--prev-page"],
        [class*="vxe-pager--next-page"],
        [class*="vxe-pager--prev-btn"],
        [class*="vxe-pager--next-btn"],
        [class*="vxe-pager--num-btn"] {
          background-color: #252526 !important;
          border-color: #4c4d4f !important;
          color: #cfd3dc !important;
        }
        .vxe-pager--jump-prev:hover:not(.is--disabled),
        .vxe-pager--jump-next:hover:not(.is--disabled),
        .vxe-pager--prev-page:hover:not(.is--disabled),
        .vxe-pager--next-page:hover:not(.is--disabled),
        .vxe-pager--prev-btn:hover:not(.is--disabled),
        .vxe-pager--next-btn:hover:not(.is--disabled),
        .vxe-pager--num-btn:hover:not(.is--disabled),
        .vxe-pager .vxe-pager--jump-prev:hover:not(.is--disabled),
        .vxe-pager .vxe-pager--jump-next:hover:not(.is--disabled),
        .vxe-pager .vxe-pager--prev-page:hover:not(.is--disabled),
        .vxe-pager .vxe-pager--next-page:hover:not(.is--disabled),
        .vxe-pager .vxe-pager--prev-btn:hover:not(.is--disabled),
        .vxe-pager .vxe-pager--next-btn:hover:not(.is--disabled),
        .vxe-pager .vxe-pager--num-btn:hover:not(.is--disabled),
        [class*="vxe-pager--jump-prev"]:hover:not(.is--disabled),
        [class*="vxe-pager--jump-next"]:hover:not(.is--disabled),
        [class*="vxe-pager--prev-page"]:hover:not(.is--disabled),
        [class*="vxe-pager--next-page"]:hover:not(.is--disabled),
        [class*="vxe-pager--prev-btn"]:hover:not(.is--disabled),
        [class*="vxe-pager--next-btn"]:hover:not(.is--disabled),
        [class*="vxe-pager--num-btn"]:hover:not(.is--disabled) {
          color: #409EFF !important;
          background-color: #2d2d30 !important;
        }
        .vxe-pager--jump-prev.is--disabled,
        .vxe-pager--jump-next.is--disabled,
        .vxe-pager--prev-page.is--disabled,
        .vxe-pager--next-page.is--disabled,
        .vxe-pager--prev-btn.is--disabled,
        .vxe-pager--next-btn.is--disabled,
        .vxe-pager--num-btn.is--disabled,
        .vxe-pager .vxe-pager--jump-prev.is--disabled,
        .vxe-pager .vxe-pager--jump-next.is--disabled,
        .vxe-pager .vxe-pager--prev-page.is--disabled,
        .vxe-pager .vxe-pager--next-page.is--disabled,
        .vxe-pager .vxe-pager--prev-btn.is--disabled,
        .vxe-pager .vxe-pager--next-btn.is--disabled,
        .vxe-pager .vxe-pager--num-btn.is--disabled,
        [class*="vxe-pager--jump-prev"].is--disabled,
        [class*="vxe-pager--jump-next"].is--disabled,
        [class*="vxe-pager--prev-page"].is--disabled,
        [class*="vxe-pager--next-page"].is--disabled,
        [class*="vxe-pager--prev-btn"].is--disabled,
        [class*="vxe-pager--next-btn"].is--disabled,
        [class*="vxe-pager--num-btn"].is--disabled {
          color: #6d6d6d !important;
          background-color: #2d2d30 !important;
          cursor: not-allowed !important;
        }
        .vxe-pager--num-btn.is--active,
        .vxe-pager .vxe-pager--num-btn.is--active {
          background-color: #409EFF !important;
          color: #fff !important;
          border-color: #409EFF !important;
        }
        .vxe-pager .vxe-pager--sizes .vxe-select {
          background-color: #252526 !important;
          border-color: #4c4d4f !important;
        }
        .vxe-pager .vxe-pager--sizes .vxe-select .vxe-input--inner {
          background-color: #252526 !important;
          color: #cfd3dc !important;
        }
        .vxe-pager .vxe-pager--sizes .vxe-select .vxe-input,
        .vxe-pager .vxe-pager--sizes .vxe-select .vxe-input__wrapper {
          background-color: #252526 !important;
          border-color: #4c4d4f !important;
        }
        .vxe-pager .vxe-pager--sizes .vxe-select .vxe-input__inner {
          background-color: #252526 !important;
          color: #cfd3dc !important;
        }
        .vxe-select,
        .vxe-pager .vxe-select {
          background-color: #252526 !important;
          border-color: #4c4d4f !important;
        }
        .vxe-select .vxe-input,
        .vxe-select .vxe-input__wrapper,
        .vxe-pager .vxe-select .vxe-input,
        .vxe-pager .vxe-select .vxe-input__wrapper {
          background-color: #252526 !important;
          border-color: #4c4d4f !important;
        }
        .vxe-select .vxe-input__inner,
        .vxe-select .vxe-input--inner,
        .vxe-pager .vxe-select .vxe-input__inner,
        .vxe-pager .vxe-select .vxe-input--inner {
          background-color: #252526 !important;
          color: #cfd3dc !important;
        }
        .vxe-select .vxe-input__suffix,
        .vxe-select .vxe-input__icon,
        .vxe-select__suffix,
        .vxe-select__icon,
        .vxe-select__arrow,
        .vxe-pager .vxe-select .vxe-input__suffix,
        .vxe-pager .vxe-select .vxe-input__icon,
        .vxe-pager .vxe-select__suffix,
        .vxe-pager .vxe-select__icon,
        .vxe-pager .vxe-select__arrow {
          background-color: transparent !important;
          color: #cfd3dc !important;
        }
        .vxe-select .vxe-input__suffix-inner,
        .vxe-select .vxe-input__icon-wrapper,
        .vxe-pager .vxe-select .vxe-input__suffix-inner,
        .vxe-pager .vxe-select .vxe-input__icon-wrapper {
          background-color: transparent !important;
        }
        .vxe-input--prefix,
        .vxe-input--suffix,
        .vxe-pager .vxe-input--prefix,
        .vxe-pager .vxe-input--suffix,
        .vxe-select .vxe-input--prefix,
        .vxe-select .vxe-input--suffix,
        .vxe-pager .vxe-select .vxe-input--prefix,
        .vxe-pager .vxe-select .vxe-input--suffix {
          background-color: transparent !important;
          color: #cfd3dc !important;
        }
        .vxe-input--suffix-icon,
        .vxe-pager .vxe-input--suffix-icon,
        .vxe-select .vxe-input--suffix-icon,
        .vxe-pager .vxe-select .vxe-input--suffix-icon {
          background-color: transparent !important;
          color: #cfd3dc !important;
        }
        .vxe-input--suffix-icon .vxe-icon-caret-down,
        .vxe-pager .vxe-input--suffix-icon .vxe-icon-caret-down,
        .vxe-select .vxe-input--suffix-icon .vxe-icon-caret-down,
        .vxe-pager .vxe-select .vxe-input--suffix-icon .vxe-icon-caret-down {
          color: #cfd3dc !important;
        }
        .vxe-pager .vxe-pager--jump .vxe-input--inner {
          background-color: #252526 !important;
          color: #cfd3dc !important;
        }
        .vxe-number-input--input,
        .vxe-pager .vxe-number-input--input {
          background-color: #252526 !important;
          color: #cfd3dc !important;
          border-color: #4c4d4f !important;
        }
        .vxe-number-input,
        .vxe-pager .vxe-number-input {
          background-color: #252526 !important;
          border-color: #4c4d4f !important;
        }
        .pagination-jumper .el-input-number .el-input__wrapper {
          background-color: #252526 !important;
          border-color: #4c4d4f !important;
        }
        .pagination-jumper .el-input-number .el-input__inner {
          background-color: #252526 !important;
          color: #cfd3dc !important;
        }
        .pagination-jumper .el-button {
          background-color: #252526 !important;
          border-color: #4c4d4f !important;
          color: #cfd3dc !important;
        }
        .pagination-jumper .el-button:hover:not(:disabled) {
          background-color: #2d2d30 !important;
          border-color: #409EFF !important;
          color: #409EFF !important;
        }
        .pagination-jumper .el-button:active:not(:disabled) {
          background-color: #2d2d30 !important;
          border-color: #409EFF !important;
          color: #409EFF !important;
        }
        .vxe-pager .vxe-select--panel {
          background-color: #252526 !important;
          border-color: #3d3e40 !important;
        }
        .vxe-pager .vxe-select-option {
          background-color: #252526 !important;
          color: #cfd3dc !important;
        }
        .vxe-pager .vxe-select-option:hover {
          background-color: #2d2d30 !important;
        }
        .el-input-number .el-input__wrapper {
          background-color: #252526 !important;
          border-color: #4c4d4f !important;
          border-width: 1px !important;
        }
        .el-input-number .el-input__inner {
          background-color: #252526 !important;
          color: #cfd3dc !important;
        }
        .vxe-table--header-wrapper,
        .vxe-table__header-wrapper {
          background-color: #2d2d30 !important;
        }
        .vxe-table--header-inner-wrapper,
        .vxe-table__header-inner-wrapper {
          background-color: #2d2d30 !important;
        }
        .vxe-table--body-inner-wrapper,
        .vxe-table__body-inner-wrapper {
          background-color: #252526 !important;
        }
        .vxe-table--body-wrapper,
        .vxe-table__body-wrapper {
          background-color: #252526 !important;
        }
        .vxe-table--body,
        .vxe-table__body {
          background-color: #252526 !important;
        }
        .vxe-table {
          background-color: #252526 !important;
        }
        .vxe-body--row.row--stripe,
        .vxe-body--row:nth-child(even),
        .vxe-table .vxe-body--row.row--stripe,
        .vxe-table .vxe-body--row:nth-child(even) {
          background-color: #2d2d30 !important;
        }
        .vxe-body--row.row--stripe .vxe-body--column,
        .vxe-body--row:nth-child(even) .vxe-body--column,
        .vxe-table .vxe-body--row.row--stripe .vxe-body--column,
        .vxe-table .vxe-body--row:nth-child(even) .vxe-body--column {
          background-color: #2d2d30 !important;
        }
        .el-loading-mask {
          background-color: rgba(0, 0, 0, 0.85) !important;
        }
        .vxe-loading,
        .vxe-loading--mask {
          background-color: rgba(0, 0, 0, 0.85) !important;
        }
        .vxe-table .vxe-loading,
        .vxe-table .vxe-loading--mask {
          background-color: rgba(0, 0, 0, 0.85) !important;
        }
      `
    },

    toggleTheme() {
      const newTheme = this.theme === 'light' ? 'dark' : 'light'
      this.setTheme(newTheme)
    }
  }
})

