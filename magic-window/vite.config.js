import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'


//跨域解决
export default defineConfig({
  plugins: [vue()],
    server:{
      proxy:{
        '/api':{
          target:'http://localhost:8080',
          changeOrigin:true,
          rewrite:(path)=>path.replace(/^\/api/,'')
        }
      }
  }
})
