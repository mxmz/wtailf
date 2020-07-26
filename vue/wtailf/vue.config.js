module.exports = {
  devServer: {
    proxy: {
      '^/events': {
        target: 'http://172.27.193.21:8081/',
        ws: true,
        changeOrigin: true
      }
    }
  }
}
