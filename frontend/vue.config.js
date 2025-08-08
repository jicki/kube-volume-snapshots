const { defineConfig } = require('@vue/cli-service')

module.exports = defineConfig({
  transpileDependencies: true,
  // 设置静态资源路径，与后端路由匹配
  publicPath: process.env.NODE_ENV === 'production' ? '/static/' : '/',
  devServer: {
    port: 8080,
    proxy: {
      '/api': {
        target: 'http://localhost:8081',
        changeOrigin: true,
        logLevel: 'debug'
      }
    }
  },
  chainWebpack: config => {
    config.plugin('define').tap(definitions => {
      Object.assign(definitions[0], {
        __VUE_OPTIONS_API__: 'true',
        __VUE_PROD_DEVTOOLS__: 'false',
        __VUE_PROD_HYDRATION_MISMATCH_DETAILS__: 'false'
      })
      return definitions
    })

    // 在生产构建中禁用 ESLint
    if (process.env.NODE_ENV === 'production') {
      config.module.rule('eslint').uses.clear()
    }
  }
})
