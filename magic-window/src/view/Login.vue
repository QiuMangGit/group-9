<script setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import{loginService} from '../api/User.js'


const email = ref('');
const password = ref('');
const router = useRouter();


const handleLogin = async () => {
  try {
    const result = await loginService({ email: email.value, password: password.value });
    if (result.code === 1) {
      sessionStorage.setItem('user', result.data.email);
      router.push('/main');
    } else {
      alert(result.message);
    }
  } catch (error) {
    alert('登录失败: ' + error.message);
  }
};
</script>


<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full space-y-8 bg-white rounded-xl shadow-lg overflow-hidden transform transition-all duration-300 hover:shadow-2xl">
      <div class="relative overflow-hidden">
        <div class="absolute top-0 right-0 -mt-16 -mr-16 w-32 h-32 bg-green-400/20 rounded-full blur-2xl"></div>
        <div class="relative p-8">
          <div class="text-center">
            <h2 class="mt-6 text-3xl font-bold text-gray-900 tracking-tight">
              用户登录
            </h2>
            <div class="w-12 h-1 bg-green-500 rounded-full mx-auto mt-4"></div>
          </div>

          <form class="mt-8 space-y-6" @submit.prevent="handleLogin">
            <div class="rounded-md -space-y-px">
              <div class="mb-4">
                <label for="username" class="block text-sm font-medium text-gray-700 mb-1">
                  邮箱
                </label>
                <div class="relative">
                  <div class="absolute inset-y-0 left-0 pl-3 flex pointer-events-none">
                    <i class="fa fa-user text-gray-400"></i>
                  </div>
                  <input 
                    id="username" 
                    name="username" 
                    type="text" 
                    v-model="email" 
                    required
                    class=" w-full h-full px-4 py-3 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-green-500 focus:border-green-500 transition-all duration-300"
                    placeholder="请输入邮箱"
                  >
                </div>
              </div>
              
              <div class="mb-4">
                <label for="password" class="block text-sm font-medium text-gray-700 mb-1">
                  密码
                </label>
                <div class="relative">
                  <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <i class="fa fa-lock text-gray-400"></i>
                  </div>
                  <input 
                    id="password" 
                    name="password" 
                    type="password" 
                    v-model="password" 
                    required
                    class=" w-full px-4 py-3 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-green-500 focus:border-green-500 transition-all duration-300"
                    placeholder="请输入密码"
                  >
                </div>
              </div>
            </div>

            <div>
              <button
                type="submit"
                class="group relative w-full flex justify-center py-3 px-4 border border-transparent text-sm font-medium rounded-lg text-white bg-gradient-to-r from-green-500 to-green-600 hover:from-green-600 hover:to-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500 transition-all duration-300 transform hover:-translate-y-1 hover:shadow-lg"
              >
                <span class="absolute left-0 inset-y-0 flex items-center pl-3">
                  <i class="fa fa-sign-in text-white/80 group-hover:text-white transition-colors duration-300"></i>
                </span>
                登录
              </button>
            </div>
            
            <div class="text-center text-sm text-gray-600 flex flex-nowrap items-center justify-center">
              还没有账号? 
              <router-link 
                to="/register" 
                class="font-medium text-green-600 hover:text-green-700 transition-colors duration-300 flex justify-center items-center gap-1"
              >
                <span>立即注册</span>
                <i class="fa fa-arrow-right text-xs"></i>
              </router-link>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

