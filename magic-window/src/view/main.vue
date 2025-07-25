<script setup>
import { ref, onMounted, watch, nextTick } from 'vue';


import {
  addAssistantService,
  selectAllService,
  deleteAssistantService,
  editAssistantService
} from '../api/Assistant.js';

import {
  selectVoiceChatContentService,
  deleteVoiceChatContentService
} from '../api/voiceChat.js';

import {
  Phone,
  PhoneOff,
  Keyboard,
  KeyboardOff,
  TriangleAlert,
  Mic,
  Pencil,
  X,
  UserRoundCog,
  Trash
} from 'lucide-vue-next';

// 响应式数据
const numInput = ref('');
const assistants = ref([]);
const chatHistory = ref([]);
const llmResponse = ref('');
const currentAssistant = ref(null); // 当前选中的机器人

// 聊天容器引用
const chatContainer = ref(null);

// 监听聊天历史变化，自动滚动到底部
watch(chatHistory, () => {
  nextTick(() => {
    if (chatContainer.value) {
      chatContainer.value.scrollTop = chatContainer.value.scrollHeight;
    }
  });
}, { deep: true });

// 模态框状态管理
const modalState = ref({
  showModalBox: false,
  hasCall: false,
  hasInput: false,
  numberInputDisk: false,
  toDeleteModalBox: false,
  toEditModalBox: false
});

// 表单数据（已将userId替换为email）
const formData = ref({
  addNewRobot: {
    name: '',
    description: '',
    email: '' 
  },
  currentRobot: {
    id: '',
    name: '',
    description: '',
    email: ''
  },
  editRobotCopy: {
    id: '',
    name: '',
    description: '',
    email: ''
  }
});

// WebRTC状态管理
const rtcState = ref({
  ws: null,
  peerConnection: null,
  isConnected: false,
  mediaRecorder: null,
  audioChunks: []
});

// 数字键盘处理
const handleNum = (num) => {
  numInput.value += num;
};

const cleanInput = () => {
  numInput.value = '';
};

// 获取机器人列表
const fetchAssistants = async () => {
  try {
    const result = await selectAllService();
    console.log(result);
    assistants.value = Array.isArray(result) ? result : result.data || [];
  } catch (error) {
    console.error('拉取机器人列表失败:', error);
    alert('获取机器人列表失败，请检查网络或后端服务');
  }
};

// 历史聊天记录
const showChatHistory = async (assistant) => {
  currentAssistant.value = assistant;
  try {
    // 从后端获取聊天历史
    const historyData = await selectVoiceChatContentService({ robotId: parseInt(assistant.id) });
    console.log(historyData);
    const conversations = historyData.data || [];
    chatHistory.value = Array.isArray(conversations) ? conversations : [];
    console.log("长度：" + chatHistory.value.length);
    // 如果没有历史记录，添加欢迎消息
    if (chatHistory.value.length === 0) {
      const welcomeMsg = {
        robotId: assistant.id,
        from: 'AI',
        context: '',//你好，我是AI助手，有什么可以帮助你的吗？
        time: new Date().toLocaleString() // 仅前端使用
      };
      chatHistory.value.push(welcomeMsg);
    }
  } catch (error) {
    console.error('加载聊天记录失败:', error);
    // 错误处理时显示默认消息
    chatHistory.value = [{
      robotId: assistant.id,
      text: '你好，我是AI助手，有什么可以帮助你的吗？',
      time: new Date().toLocaleTimeString()
    }];
  }
};

// 新增机器人
const addBox = () => {
  modalState.value.showModalBox = true;
  const email = sessionStorage.getItem("user"); 
  formData.value.addNewRobot = { name: '', description: '', email: email };
};

const confirmAdd = async () => {
  if (!formData.value.addNewRobot.name) {
    formData.value.addNewRobot.name = '默认机器人';
  }
  try {
    await addAssistantService(formData.value.addNewRobot);
    await fetchAssistants();
    modalState.value.showModalBox = false;
  } catch (error) {
    console.error('添加失败:', error);
    alert(`添加失败: ${error.message || '未知错误'}`);
  }
};

// 取消操作
const cancelAction = () => {
  modalState.value.showModalBox = false;
  modalState.value.toDeleteModalBox = false;
  modalState.value.toEditModalBox = false;
  formData.value.addNewRobot = { name: '', description: '', email: '' };  // 清空email
  formData.value.editRobotCopy = { id: '', name: '', description: '', email: '' };  // 清空email
};

// 编辑机器人
const editAssistant = (assistant) => {
  formData.value.editRobotCopy = { ...assistant };  
  modalState.value.toEditModalBox = true;
};

const confirmEdit = async () => {
  try {
    await editAssistantService(formData.value.editRobotCopy);
    const index = assistants.value.findIndex(item => item.id === formData.value.editRobotCopy.id);
    if (index !== -1) {
      assistants.value[index] = { ...formData.value.editRobotCopy };
    }
    // 更新当前选中的机器人
    if (currentAssistant.value?.id === formData.value.editRobotCopy.id) {
      currentAssistant.value = { ...formData.value.editRobotCopy };
    }
    modalState.value.toEditModalBox = false;
  } catch (error) {
    console.error('编辑失败:', error);
    alert(`编辑失败: ${error.message || '未知错误'}`);
  }
};

// 删除机器人
const prepareDelete = (assistant) => {
  formData.value.currentRobot = { ...assistant };  // 复制包含email的机器人数据
  modalState.value.toDeleteModalBox = true;
};

const confirmDelete = async () => {
  try {
    await deleteVoiceChatContentService({
      robotId: parseInt(formData.value.currentRobot.id)
    });

    console.log(formData.value.currentRobot);
    await deleteAssistantService(formData.value.currentRobot);

    await fetchAssistants();

    if (currentAssistant.value?.id === formData.value.currentRobot.id) {
      chatHistory.value = [];
      currentAssistant.value = null;
    }
    modalState.value.toDeleteModalBox = false;
  } catch (error) {
    console.error('删除失败:', error);
    alert(`删除失败: ${error.message || '未知错误'}`);
  }
};

// 通话控制
const endCall = () => {
  modalState.value.hasCall = false;
  modalState.value.hasInput = false;
  modalState.value.numberInputDisk = false;
  hangupCall();
  numInput.value = '';
  llmResponse.value = '';
};

const showInputPad = () => {
  if (modalState.value.hasCall) {
    modalState.value.hasInput = true;
    modalState.value.numberInputDisk = true;
  }
};

const hideInputPad = () => {
  modalState.value.hasInput = false;
  modalState.value.numberInputDisk = false;
  numInput.value = '';
};

// 重置对话
const resetConversation = async () => {
  if (currentAssistant.value) {
    try {
      await deleteVoiceChatContentService({
        robotId: parseInt(currentAssistant.value.id)
      })
    } catch (error) {
      console.error('删除后端聊天记录失败:', error);
    }
  }

  chatHistory.value = [];
  llmResponse.value = '';
};

const logMessage = (text, from) => {
  chatHistory.value.push({
    context: text,
    from: from
  });
};

// 开始通话
const startCall = async () => {
  if (!currentAssistant.value) {
    alert('请先选择一个机器人');
    return;
  }

  modalState.value.hasCall = true;
  modalState.value.numberInputDisk = false;

  rtcState.value.ws = new WebSocket('ws://localhost:8081/ws2');
  rtcState.value.isConnected = true;

  rtcState.value.ws.onopen = () => {
    console.log('WebSocket connected!');
    connectWebRTC();
    startRecording();
  };

  rtcState.value.ws.onmessage = async (event) => {
    try {
      let message;
      let dataString;

      if (event.data instanceof ArrayBuffer || event.data instanceof Blob) {
        dataString = await new Response(event.data).text();
      } else if (typeof event.data === 'string') {
        dataString = event.data;
      } else {
        throw new Error('Unsupported message format');
      }

      try {
        message = JSON.parse(dataString);
      } catch (e) {
        console.log("捕捉到了错误");
        message = { event: 'llmResponse', text: dataString };
      }

      switch (message.event) {
        case 'answer':
          await handleAnswer(message.sdp);
          break;
        case 'candidate':
          await handleCandidate(message.candidate);
          break;
        case 'trackStart':
          console.log('Remote media track started');
          break;
        case 'asrDelta':
          console.log('ASR realtime: ' + message.text);
          break;
        case 'asrFinal':
          console.log('ASR Final:' + message.text)
          logMessage(message.text, "User");
          break;
        case 'llmResponse':
          console.log('llmResponse:' + message.text)
          logMessage(message.text, "AI");
          llmResponse.value = message.text;
          break;
        case 'hangup':
          if (rtcState.value.peerConnection) {
            rtcState.value.peerConnection.close();
            rtcState.value.peerConnection = null;
          }
          console.log('Call has been hung up');
          logMessage("很荣幸为您服务，下次再见！！！", "AI")
          rtcState.value.isConnected = false;
        default:
          console.log('未知消息类型:', message.event);
      }
    } catch (error) {
      logMessage('处理消息失败: ' + error.message, "AI");
    }
  };

  rtcState.value.ws.onclose = () => {
    console.log('WebSocket closed!');
    rtcState.value.isConnected = false;
  };

  rtcState.value.ws.onerror = (error) => {
    console.log('WebSocket error: ' + error.message);
    alert('连接socket失败');
  };
};

// 建立WebRTC连接
const connectWebRTC = async () => {
  try {
    rtcState.value.peerConnection = new RTCPeerConnection({
      iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
    });

    const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
    stream.getTracks().forEach(track => {
      rtcState.value.peerConnection.addTrack(track, stream);
    });

    rtcState.value.peerConnection.ontrack = (event) => {
      const audioElement = document.getElementById('audio');
      if (audioElement) audioElement.srcObject = event.streams[0];
    };

    rtcState.value.peerConnection.onicecandidate = (event) => {
      if (event.candidate && rtcState.value.ws) {
        rtcState.value.ws.send(JSON.stringify({
          event: 'candidate',
          candidate: event.candidate
        }));
      }
    };

    const offer = await rtcState.value.peerConnection.createOffer();
    await rtcState.value.peerConnection.setLocalDescription(offer);

    if (rtcState.value.ws) {
      rtcState.value.ws.send(JSON.stringify({
        event: 'offer',
        sdp: offer.sdp,
        robotId: currentAssistant.value.id
      }));
    }

    rtcState.value.isConnected = true;
  } catch (error) {
    console.log('WebRTC连接失败: ' + error.message);
  }
};

// 处理WebRTC应答
const handleAnswer = async (sdp) => {
  try {
    await rtcState.value.peerConnection.setRemoteDescription(new RTCSessionDescription({
      type: 'answer',
      sdp: sdp
    }));
    logMessage('你好，我是AI助手，有什么可以帮助你的吗？', "AI");
  } catch (error) {
    console.log('处理应答失败: ' + error.message);
  }
};

// 处理ICE候选者
const handleCandidate = async (candidate) => {
  try {
    await rtcState.value.peerConnection.addIceCandidate(candidate);
  } catch (error) {
    logMessage('处理ICE候选者失败: ' + error.message, "AI");
  }
};

// 录音控制
const startRecording = async () => {
  if (!rtcState.value.peerConnection) {
    console.log('WebRTC connection not established yet.');
    return;
  }

  const localStreams = rtcState.value.peerConnection.getLocalStreams();
  if (localStreams.length === 0) {
    console.log('No local stream available.');
    return;
  }

  const stream = localStreams[0];
  try {
    rtcState.value.mediaRecorder = new MediaRecorder(stream);
    rtcState.value.audioChunks = [];

    rtcState.value.mediaRecorder.ondataavailable = (event) => {
      if (event.data.size > 0) {
        rtcState.value.audioChunks.push(event.data);
        const reader = new FileReader();
        reader.onload = () => {
          if (rtcState.value.ws) {
            rtcState.value.ws.send(new Uint8Array(reader.result));
          }
        };
        reader.readAsArrayBuffer(event.data);
      }
    };

    rtcState.value.mediaRecorder.onstop = () => {
      logMessage('很高兴为您服务，下次再见！', "AI");
    };

    rtcState.value.mediaRecorder.start();
  } catch (error) {
    logMessage('录音启动失败: ' + error.message, "AI");
  }
};

function hangupCall() {
  if (rtcState.value.isConnected) {
    const hangupCommand = {
      command: "hangup",
      reason: "user_requested",
      initiator: "caller"
    };
    rtcState.value.ws.send(JSON.stringify(hangupCommand));
    console.log('Hangup command sent');
    document.getElementById('hangupButton').disabled = true;
    rtcState.value.isConnected = false;
  } else {
    console.log('WebRTC connection is not established.');
    logMessage("请检查webRtc连接", "AI")
  }
}

// 组件挂载时初始化
onMounted(() => {
  fetchAssistants();
});
</script>

<template>
  <div class="w-full h-screen bg-green-200 flex flex-wrap justify-center">
    <!-- 顶部标题区域 -->
    <div class="w-full h-[10%] border-b border-stone-300 bg-stone-50">
      <div class="m-5 font-bold cursor-default">智能语音机器人</div>
      <div class="mb-2 ml-5 font-normal text-xs text-slate-500 cursor-default">
        提供一站式的智能语音解决方案,基于大模型全面升级NLP能力,显著降低运营成本与门槛。
      </div>
    </div>

    <!-- 主内容区域 -->
    <div class="w-full h-[90%] bg-stone-50 flex items-center flex-nowrap">
      <!-- 左侧：机器人列表 -->
      <div class="w-[33%] h-full border-r border-stone-300 bg-stone-50 flex flex-wrap justify-center">
        <!-- 新增助手按钮 -->
        <div class="w-full h-[10%] flex items-center justify-center">
          <button @click="addBox"
            class="w-[80%] h-[80%] bg-blue-500 rounded text-stone-50 font-semibold text-sm flex items-center justify-center hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-400 shadow-lg transition-all">
            +新增助手
          </button>
        </div>

        <!-- 机器人列表展示 -->
        <div class="p-2 w-full h-[90%] overflow-auto flex items-center flex-col">
          <div v-for="assistant in assistants" :key="assistant.id"
            class="w-[80%] h-20 bg-blue-200 mb-2 min-w-[80%] min-h-20 parent relative rounded shadow-md flex flex-col cursor-pointer hover:bg-blue-300 transition-colors"
            @click="showChatHistory(assistant)">
            <p class="text-black text-base mt-3 ml-3 font-medium">{{ assistant.name }}</p>
            <p class="mt-4 ml-3 text-stone-500 text-xs line-clamp-1">{{ assistant.description }}</p>
            <UserRoundCog class="child absolute cursor-pointer right-2 top-2 rounded w-9 h-4 font-semibold"
              color="#696969" />
            <Pencil @click.stop="editAssistant(assistant)"
              class="child absolute cursor-pointer right-2 top-8 rounded w-9 h-4 font-semibold" color="#696969" />
            <Trash @click.stop="prepareDelete(assistant)"
              class="child absolute cursor-pointer right-3 top-14 text-xs w-8 h-4" color="#FF6666" />
          </div>
        </div>
      </div>

      <!-- 右侧：聊天区域 -->
      <div class="w-[67%] h-full bg-stone-50 flex flex-wrap justify-center">
        <!-- 顶部标题和操作 -->
        <div class="w-full h-[10%] bg-stone-50 border-b border-stone-300 flex items-center flex-nowrap">
          <div class="ml-5 cursor-default font-medium">
            {{ currentAssistant?.name || '请选择一个机器人' }}
          </div>
          <button type="button" @click="resetConversation"
            class="absolute right-5 text-stone-600 hover:text-stone-400 text-sm flex items-center gap-1"
            :disabled="!currentAssistant" :class="{ 'opacity-50 cursor-not-allowed': !currentAssistant }">
            <span>↺</span>
            <span>重置对话</span>
          </button>
        </div>

        <!-- 聊天内容区域 -->
        <div ref="chatContainer"
          class="w-full h-[75%] bg-stone-200 border-b border-stone-300 overflow-auto flex items-center flex-col pt-8 pb-16 scroll-smooth">
          <!-- 空状态提示 -->
          <div v-if="chatHistory.length === 0 && !currentAssistant"
            class="flex flex-col items-center justify-center h-full">
            <div class="text-stone-400 text-5xl mb-4">🤖</div>
            <p class="text-stone-500 text-lg">请从左侧选择一个机器人开始对话</p>
          </div>

          <!-- 聊天消息 -->
          <div v-else v-for="(message, index) in chatHistory" :key="index" class="w-full message-item mb-4 px-6">
            <div v-if="message.from === 'AI' && message.context != ''" class="flex items-start">
              <!-- 机器人头像 -->
              <div
                class="w-12 h-12 rounded-full bg-green-500 flex items-center justify-center ml-[5%] text-white cursor-default font-bold shadow-sm">
                AI
              </div>
              <div class="ml-3 max-w-[60%]">
                <div class="bg-white rounded-lg p-3 shadow-sm break-words">
                  <p class="text-gray-800 text-sm">{{ message.context }}</p>
                </div>
                <div class="text-stone-400 text-xs mt-1 ml-1">
                  {{ message.time || '刚刚' }}
                </div>
              </div>
            </div>
            <!-- 用户消息 -->
            <div v-else-if="message.from === 'User' && message.context != ''" class="flex items-start justify-end">
              <div class="mr-3 max-w-[80%]">
                <div class="bg-blue-500 text-white rounded-lg p-3 shadow-sm">
                  <p class="text-sm">{{ message.context }}</p>
                </div>
                <div class="text-stone-400 text-xs mt-1 ml-1">
                  {{ message.time || '刚刚' }}
                </div>
              </div>
              <!-- 用户头像 -->
              <div
                class="w-12 h-12 rounded-full bg-blue-500 flex items-center justify-center text-white font-bold shadow-sm">
                U
              </div>
            </div>


          </div>
        </div>

        <!-- 底部操作栏 -->
        <div class="w-full h-[15%] bg-stone-50 border-b border-stone-300 flex justify-center relative">
          <div class="flex justify-center items-center h-full">
            <!-- 通话按钮 -->
            <Transition enter-active-class="transition-all duration-300 ease-out" enter-from-class="scale-0 opacity-0"
              enter-to-class="scale-100 opacity-100">
              <div v-if="!modalState.hasCall && currentAssistant" @click="startCall"
                class="rounded-full active:bg-green-600 hover:bg-green-600 cursor-pointer flex items-center justify-center bg-green-500 w-14 h-14 m-3 transition-all">
                <Phone class="m-2" color="#fff" :size="35" :stroke-width="2" />
              </div>
            </Transition>

            <!-- 挂断按钮 -->
            <Transition enter-active-class="transition-all duration-300 ease-out" enter-from-class="scale-0 opacity-0"
              enter-to-class="scale-100 opacity-100">
              <div v-if="modalState.hasCall" @click="endCall" id="hangupButton"
                class="bg-red-500 active:bg-red-600 hover:bg-red-600 rounded-full cursor-pointer flex items-center justify-center w-14 h-14 m-3 transition-all">
                <PhoneOff color="#fff" :size="35" :stroke-width="2" />
              </div>
            </Transition>

            <!-- 显示键盘按钮 -->
            <Transition enter-active-class="transition-all duration-300 ease-out" enter-from-class="scale-0 opacity-0"
              enter-to-class="scale-100 opacity-100">
              <div v-if="!modalState.hasInput && modalState.hasCall" @click="showInputPad"
                class="bg-slate-300 active:bg-slate-400 hover:bg-slate-400 cursor-pointer rounded-full flex items-center justify-center w-14 h-14 m-3 transition-all">
                <Keyboard color="#333333" :size="40" :stroke-width="2" />
              </div>
            </Transition>

            <!-- 隐藏键盘按钮 -->
            <Transition enter-active-class="transition-all duration-300 ease-out" enter-from-class="scale-0 opacity-0"
              enter-to-class="scale-100 opacity-100">
              <div v-if="modalState.hasInput" @click="hideInputPad"
                class="border-solid border-indigo-600 border-2 active:bg-stone-100 hover:bg-stone-100 bg-stone-50 rounded-full flex items-center justify-center w-14 h-14 m-3 transition-all">
                <KeyboardOff color="#4338ca" :size="40" :stroke-width="2" />
              </div>
            </Transition>

            <!-- 麦克风按钮 -->
            <Transition enter-active-class="transition-all duration-300 ease-out" enter-from-class="scale-0 opacity-0"
              enter-to-class="scale-100 opacity-100">
              <div v-if="modalState.hasCall"
                class="bg-blue-500 rounded-full cursor-pointer flex items-center justify-center w-14 h-14 m-3 transition-all">
                <Mic color="#fff" :size="25" :stroke-width="2" />
              </div>
            </Transition>
          </div>

          <!-- 底部提示 -->
          <div class="absolute bottom-3 text-xs text-stone-400 text-center cursor-default w-full">
            服务生成的所有内容均人工智能生成,其生成内容的准确性和完整性无法保证,不代表我们的态度和观点
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- 新增助手模态框 -->
  <Transition name="modal">
    <div v-if="modalState.showModalBox" class="fixed z-50 inset-0 bg-gray-400/70 flex items-center justify-center p-4">
      <div class="w-full max-w-md bg-white rounded-lg shadow-xl overflow-hidden">
        <div class="p-5 border-b border-gray-200">
          <div class="text-xl font-serif font-semibold">新增助手</div>
        </div>
        <form class="p-5" @submit.prevent="confirmAdd" method="post">
          <div class="mb-4">
            <label class="block text-xs text-gray-600 mb-1">名称</label>
            <input type="text" name="name" placeholder="默认机器人" v-model="formData.addNewRobot.name"
              class="w-full border rounded border-gray-300 p-2 focus:outline-none focus:ring-2 focus:ring-blue-500">
          </div>
          <div class="mb-6">
            <label class="block text-xs text-gray-600 mb-1">描述</label>
            <textarea v-model="formData.addNewRobot.description" name="description" cols="30" rows="3"
              class="w-full border rounded border-gray-300 p-2 focus:outline-none focus:ring-2 focus:ring-blue-500"></textarea>
          </div>
          <div class="flex justify-end gap-3">
            <button type="button"
              class="px-4 py-2 border border-gray-300 rounded-full hover:bg-gray-100 transition-colors"
              @click="cancelAction">
              取消
            </button>
            <button class="px-4 py-2 bg-blue-600 text-white rounded-full hover:bg-blue-700 transition-colors"
              type="submit">
              确定
            </button>
          </div>
        </form>
      </div>
    </div>
  </Transition>

  <!-- 删除确认模态框 -->
  <Transition name="modal">
    <div v-if="modalState.toDeleteModalBox"
      class="fixed z-50 inset-0 bg-gray-400/70 flex items-center justify-center p-4">
      <div class="w-full max-w-md bg-white rounded-lg shadow-xl overflow-hidden">
        <div class="p-5 border-b border-gray-200">
          <div class="text-xl font-serif font-semibold flex items-center">
            <TriangleAlert class="inline-block mr-2" :stroke-width="3" color="#FF9933" />
            <span>确定删除对话？</span>
          </div>
        </div>
        <div class="p-5">
          <p class="text-sm text-gray-500 mb-5">删除后，聊天记录将不可恢复。</p>
          <div class="flex justify-end gap-3">
            <button type="button"
              class="px-4 py-2 border border-gray-300 rounded-full hover:bg-gray-100 transition-colors"
              @click.stop="cancelAction">
              取消
            </button>
            <button class="px-4 py-2 bg-red-600 text-white rounded-full hover:bg-red-700 transition-colors"
              @click="confirmDelete">
              删除
            </button>
          </div>
        </div>
      </div>
    </div>
  </Transition>

  <!-- 编辑助手模态框 -->
  <Transition name="modal">
    <div v-if="modalState.toEditModalBox"
      class="fixed z-50 inset-0 bg-gray-400/70 flex items-center justify-center p-4">
      <div class="w-full max-w-md bg-white rounded-lg shadow-xl overflow-hidden">
        <div class="p-5 border-b border-gray-200">
          <div class="text-xl font-serif font-semibold">编辑助手</div>
        </div>
        <form class="p-5" @submit.prevent="confirmEdit" method="post">
          <div class="mb-4">
            <label class="block text-xs text-gray-600 mb-1">名称</label>
            <input type="text" name="name" placeholder="默认机器人" v-model="formData.editRobotCopy.name"
              class="w-full border rounded border-gray-300 p-2 focus:outline-none focus:ring-2 focus:ring-blue-500">
          </div>
          <div class="mb-6">
            <label class="block text-xs text-gray-600 mb-1">描述</label>
            <textarea v-model="formData.editRobotCopy.description" name="description" cols="30" rows="3"
              class="w-full border rounded border-gray-300 p-2 focus:outline-none focus:ring-2 focus:ring-blue-500"></textarea>
          </div>
          <div class="flex justify-end gap-3">
            <button type="button"
              class="px-4 py-2 border border-gray-300 rounded-full hover:bg-gray-100 transition-colors"
              @click="cancelAction">
              取消
            </button>
            <button class="px-4 py-2 bg-blue-600 text-white rounded-full hover:bg-blue-700 transition-colors"
              type="submit">
              确定
            </button>
          </div>
        </form>
      </div>
    </div>
  </Transition>

  <!-- 数字键盘 -->
  <div v-if="modalState.numberInputDisk"
    class="w-[14%] h-80 shadow-2xl rounded-md fixed bottom-36 right-96 mr-14 bg-stone-50">
    <div class="flex items-center h-16">
      <input type="text" name="numInput" class="w-[75%] ml-4 h-8 focus:outline-none" placeholder="请输入"
        v-model="numInput">
      <X @click="cleanInput" color="#d1d5db" class="cursor-pointer ml-2" />
    </div>
    <div class="flex justify-around mb-2">
      <div @click="handleNum('1')"
        class="w-14 h-14 rounded-full bg-white border-2 border-neutral-300 flex items-center justify-center text-black font-bold shadow-sm active:bg-slate-100 cursor-pointer transition-colors">
        1
      </div>
      <div @click="handleNum('2')"
        class="w-14 h-14 rounded-full bg-white border-2 border-neutral-300 flex items-center justify-center text-black font-bold shadow-sm active:bg-slate-100 cursor-pointer transition-colors">
        2
      </div>
      <div @click="handleNum('3')"
        class="w-14 h-14 rounded-full bg-white border-2 border-neutral-300 flex items-center justify-center text-black font-bold shadow-sm active:bg-slate-100 cursor-pointer transition-colors">
        3
      </div>
    </div>
    <div class="flex justify-around mb-2">
      <div @click="handleNum('4')"
        class="w-14 h-14 rounded-full bg-white border-2 border-neutral-300 flex items-center justify-center text-black font-bold shadow-sm active:bg-slate-100 cursor-pointer transition-colors">
        4
      </div>
      <div @click="handleNum('5')"
        class="w-14 h-14 rounded-full bg-white border-2 border-neutral-300 flex items-center justify-center text-black font-bold shadow-sm active:bg-slate-100 cursor-pointer transition-colors">
        5
      </div>
      <div @click="handleNum('6')"
        class="w-14 h-14 rounded-full bg-white border-2 border-neutral-300 flex items-center justify-center text-black font-bold shadow-sm active:bg-slate-100 cursor-pointer transition-colors">
        6
      </div>
    </div>
    <div class="flex justify-around mb-2">
      <div @click="handleNum('7')"
        class="w-14 h-14 rounded-full bg-white border-2 border-neutral-300 flex items-center justify-center text-black font-bold shadow-sm active:bg-slate-100 cursor-pointer transition-colors">
        7
      </div>
      <div @click="handleNum('8')"
        class="w-14 h-14 rounded-full bg-white border-2 border-neutral-300 flex items-center justify-center text-black font-bold shadow-sm active:bg-slate-100 cursor-pointer transition-colors">
        8
      </div>
      <div @click="handleNum('9')"
        class="w-14 h-14 rounded-full bg-white border-2 border-neutral-300 flex items-center justify-center text-black font-bold shadow-sm active:bg-slate-100 cursor-pointer transition-colors">
        9
      </div>
    </div>
    <div class="flex justify-around mb-2">
      <div @click="handleNum('*')"
        class="w-14 h-14 rounded-full bg-white border-2 border-neutral-300 flex items-center justify-center text-black font-bold shadow-sm active:bg-slate-100 cursor-pointer transition-colors">
        <span class="mt-2 text-2xl">*</span>
      </div>
      <div @click="handleNum('0')"
        class="w-14 h-14 rounded-full bg-white border-2 border-neutral-300 flex items-center justify-center text-black font-bold shadow-sm active:bg-slate-100 cursor-pointer transition-colors">
        0
      </div>
      <div @click="handleNum('#')"
        class="w-14 h-14 rounded-full bg-white border-2 border-neutral-300 flex items-center justify-center text-black font-bold shadow-sm active:bg-slate-100 cursor-pointer transition-colors">
        #
      </div>
    </div>
  </div>
  <audio id="audio" autoplay></audio>
</template>

<style>
/* 模态框动画优化 */
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .bg-gray-400\/70,
.modal-leave-to .bg-gray-400\/70 {
  opacity: 0;
}

.modal-enter-from>div,
.modal-leave-to>div {
  transform: scale(0.95);
  opacity: 0;
}

.modal-enter-active,
.modal-leave-active {
  transition: all 0.2s ease-out;
}

/* 模态框容器样式 */
.modal-container {
  width: 100%;
  max-width: 500px;
  /* 最大宽度限制 */
  box-sizing: border-box;
  /* 确保padding不会增加总宽度 */
}

/* 响应式调整 */
@media (max-width: 768px) {
  .modal-container {
    max-width: 90%;
  }
}

#audio {
  display: none;
}
</style>