module.exports = {
  devServer: {
    proxy: {
      '^/(events|sources)': {
        // target: 'http://172.27.193.21:8081/',
        target: 'http://127.0.0.1:8081/',
        ws: true,
        changeOrigin: true
      }
    }
  },
  transpileDependencies: [
    'vuetify'
  ]
}
