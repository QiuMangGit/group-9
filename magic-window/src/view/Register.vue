<script setup>
import { ref, reactive, watch, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { registerService, sendVerificationCode } from '../api/User.js'

const router = useRouter();

const form = reactive({
  username: '',
  password: '',
  confirmPassword: '',
  email: '',
  verificationCode: ''
});

const postForm = reactive({
  username: '',
  password: '',
  email: '',
  verificationCode: ''
});

const formErrors = reactive({
  username: '',
  password: '',
  confirmPassword: '',
  email: '',
  verificationCode: ''
});

const countdown = ref(0);
const isFormValid = ref(false);

// 监听表单变化进行验证
watch(form, () => {
  validateForm();
});

// 表单验证函数
const validateForm = () => {
  formErrors.username = form.username ? '' : '用户名不能为空';
  formErrors.password = form.password.length >= 6 ? '' : '密码长度至少为6位';
  formErrors.confirmPassword = form.password === form.confirmPassword ? '' : '两次输入的密码不一致';
  formErrors.email = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email) ? '' : '请输入有效的邮箱地址';
  formErrors.verificationCode = form.verificationCode.length === 6 ? '' : '验证码为6位数字';

  isFormValid.value = Object.values(formErrors).every(error => !error);
};

// 发送验证码
const sendCode = async () => {
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email)) {
    formErrors.email = '请输入有效的邮箱地址';
    return;
  }

  if (countdown.value > 0) return;

  try {
    const result = await sendVerificationCode({ emailTo: form.email,subject:"注册验证码" });
    if (result.code === 1) {
      alert('验证码已发送，请注意查收');
      startCountdown();
    } else {
      alert(result.message);
    }
  } catch (error) {
    alert('发送验证码失败: ' + error.message);
  }
};


const startCountdown = () => {
  countdown.value = 60;
  const timer = setInterval(() => {
    countdown.value--;
    if (countdown.value <= 0) {
      clearInterval(timer);
    }
  }, 1000);
};


const handleRegister = async () => {
  validateForm();
  
  if (!isFormValid.value) {
    return;
  }

  postForm.username = form.username;
  postForm.password = form.password;
  postForm.email = form.email;
  postForm.verificationCode = form.verificationCode;

  try {
    const result = await registerService(postForm);
    console.log(result);
    if (result.code === 1) {
      alert('注册成功，请登录');
      router.push('/login');
    } else {
      alert(result.message);
      if (result.message === "用户已存在,请直接登录") {
        router.push('/login');
      }
      console.log('result.message');
    }
  } catch (error) {
    alert('注册失败: ' + error.message);
  }
};
</script>

<template>
  <div class="max-w-md mx-auto mt-20 p-8 bg-white rounded-2xl shadow-xl relative overflow-hidden">
    <div class="absolute -bottom-12 -left-12 w-32 h-32 bg-teal-400/10 rounded-full"></div>

    <h2 class="text-center text-2xl font-semibold text-gray-800 mb-8 relative">
      用户注册
      <span class="absolute bottom-[-8px] left-1/2 transform -translate-x-1/2 w-12 h-1 bg-teal-500 rounded"></span>
    </h2>

    <form @submit.prevent="handleRegister" class="relative z-10">
      <div class="mb-4">
        <label class="block text-sm font-medium text-gray-700 mb-1">用户名</label>
        <input type="text" v-model="form.username" required placeholder="请输入用户名"
          class="w-full px-4 py-3  focus:outline-none rounded-lg border border-teal-300 focus:border-teal-500 focus:ring-2 focus:ring-teal-200 transition-all bg-gray-50">
        <p class="text-red-500 text-xs mt-1">{{ formErrors.username }}</p>
      </div>

      <div class="mb-4">
        <label class="block text-sm font-medium text-gray-700 mb-1">密码</label>
        <input type="password" v-model="form.password" required minlength="6" placeholder="请输入密码"
          class="w-full px-4 py-3  focus:outline-none rounded-lg border border-teal-300 focus:border-teal-500 focus:ring-2 focus:ring-teal-200 transition-all bg-gray-50">
        <p class="text-red-500 text-xs mt-1">{{ formErrors.password }}</p>
      </div>

      <div class="mb-4">
        <label class="block text-sm font-medium text-gray-700 mb-1">确认密码</label>
        <input type="password" v-model="form.confirmPassword" required minlength="6" placeholder="确认密码"
          class="w-full px-4 py-3  focus:outline-none rounded-lg border border-teal-300 focus:border-teal-500 focus:ring-2 focus:ring-teal-200 transition-all bg-gray-50">
        <p class="text-red-500 text-xs mt-1">{{ formErrors.confirmPassword }}</p>
      </div>

      <div class="mb-4">
        <label class="block text-sm font-medium text-gray-700 mb-1">邮箱</label>
        <div class="flex">
          <input type="email" v-model="form.email" required placeholder="请输入邮箱"
            class="flex-1 px-4 py-3  focus:outline-none rounded-l-lg border border-teal-300 focus:border-teal-500 focus:ring-2 focus:ring-teal-200 transition-all bg-gray-50">
          <button type="button" @click="sendCode"
            class="px-4 py-3 bg-teal-500 text-white font-medium rounded-r-lg shadow hover:bg-teal-600 transition-colors"
            :disabled="countdown > 0">
            {{ countdown > 0 ? `${countdown}s后重试` : '获取验证码' }}
          </button>
        </div>
        <p class="text-red-500 text-xs mt-1">{{ formErrors.email }}</p>
      </div>

      <div class="mb-6">
        <label class="block text-sm font-medium text-gray-700 mb-1">验证码</label>
        <input type="text" v-model="form.verificationCode" required maxlength="6" placeholder="请输入验证码"
          class="w-full px-4 py-3 rounded-lg border border-teal-300 focus:border-teal-500 focus:ring-2 focus:ring-teal-200 transition-all bg-gray-50">
        <p class="text-red-500 text-xs mt-1">{{ formErrors.verificationCode }}</p>
      </div>

      <button type="submit"
        class="w-full py-3 px-4 bg-gradient-to-r from-teal-500 to-emerald-500 text-white font-medium rounded-lg shadow-md hover:shadow-lg hover:-translate-y-0.5 transition-all duration-300"
        :disabled="!isFormValid">
        注册
      </button>

      <div class="mt-6 text-center text-gray-600 text-sm">
        已有账号? <router-link to="/login"
          class="text-teal-600 font-medium hover:text-teal-700 hover:underline transition-colors">立即登录</router-link>
      </div>
    </form>
  </div>
</template>