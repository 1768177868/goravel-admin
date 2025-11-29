import { ref, nextTick } from 'vue'

export function useFormHeight(formRef, expanded, hasAdvancedFields, defaultExpanded) {
  const formHeight = ref(0)
  const singleLineHeight = ref(0)
  const shouldShowExpandButton = ref(false)

  // 检测表单高度，判断是否需要显示展开按钮
  const checkFormHeight = () => {
    nextTick(() => {
      if (!formRef.value || !formRef.value.$el) return
      
      const formEl = formRef.value.$el
      const formItems = Array.from(formEl.querySelectorAll('.el-form-item:not(.action-item)'))
      
      if (formItems.length === 0) return
      
      // 获取第一个表单项的高度（作为单行高度参考）
      const firstItem = formItems[0]
      if (firstItem) {
        const firstItemRect = firstItem.getBoundingClientRect()
        const computedStyle = window.getComputedStyle(firstItem)
        const marginBottom = parseFloat(computedStyle.marginBottom) || 18
        // 单行高度 = 表单项高度 + 底部间距
        singleLineHeight.value = firstItemRect.height + marginBottom
      }
      
      // 获取表单字段容器的高度
      const fieldsWrapper = formEl.querySelector('.form-fields-wrapper')
      if (fieldsWrapper) {
        const wrapperRect = fieldsWrapper.getBoundingClientRect()
        formHeight.value = wrapperRect.height
        
        // 计算可以显示多少行（根据容器宽度和表单项宽度）
        // 获取容器宽度（减去 padding）
        const containerPadding = 40 // 左右 padding 各 20px
        const containerWidth = wrapperRect.width - containerPadding
        let currentRowWidth = 0
        let rowCount = 1
        const rowGap = 10 // 表单项之间的间距
        let firstRowItems = [] // 第一行能显示的字段
        
        formItems.forEach((item) => {
          const itemRect = item.getBoundingClientRect()
          const itemWidth = itemRect.width
          const itemMarginRight = parseFloat(window.getComputedStyle(item).marginRight) || 10
          
          // 如果当前行放不下这个表单项，换行
          if (currentRowWidth + itemWidth + itemMarginRight > containerWidth && currentRowWidth > 0) {
            rowCount++
            currentRowWidth = itemWidth + itemMarginRight
          } else {
            if (rowCount === 1) {
              firstRowItems.push(item)
            }
            currentRowWidth += itemWidth + itemMarginRight + rowGap
          }
        })
        
        // 如果超过一行，需要显示展开按钮
        if (rowCount > 1) {
          shouldShowExpandButton.value = true
          // 计算收起状态应该显示的高度（尽可能多地显示第一行的字段）
          // 如果第一行能显示所有基础字段，则高度就是单行高度
          // 否则需要计算第一行实际占用的高度
          if (firstRowItems.length > 0) {
            // 使用第一行最后一个元素的位置来计算高度
            const lastFirstRowItem = firstRowItems[firstRowItems.length - 1]
            const lastItemRect = lastFirstRowItem.getBoundingClientRect()
            const firstItemRect = firstRowItems[0].getBoundingClientRect()
            // 第一行的实际高度 = 最后一个元素底部 - 第一个元素顶部 + 底部间距
            const firstRowHeight = lastItemRect.bottom - firstItemRect.top + 18
            singleLineHeight.value = Math.max(singleLineHeight.value, firstRowHeight)
          }
          
          // 如果默认是收起状态，且表单高度超过单行，则默认收起
          if (!defaultExpanded && expanded.value === defaultExpanded) {
            expanded.value = false
          }
        } else {
          shouldShowExpandButton.value = false
          // 如果不需要展开按钮，则始终展开
          if (!hasAdvancedFields.value) {
            expanded.value = true
          }
        }
      }
    })
  }

  return {
    formHeight,
    singleLineHeight,
    shouldShowExpandButton,
    checkFormHeight
  }
}

