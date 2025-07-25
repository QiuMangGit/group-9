主函数使用：
func createClient(ctx context.Context, option CreateClientOption, id string) *rustpbxgo.Client {}
createClient 创建一个 rustpbxgo.Client 实例，并注册所有事件回调。

在main中直接调用createClient创建实例，传入			ctx，	 option结构体，	通话唯一标识

option结构体
type CreateClientOption struct {
	Endpoint       string                           //rustpbxgo 网关 WebSocket 地址
	Logger         *logrus.Logger                   //全链路日志
	SigChan        chan bool                        //外部协程退出信号
	LLMHandler     *handler.LLMHandler              //提前注入的 LLM 客户端
	OpenaiKey      string                           //OpenAI 配置
	OpenaiEndpoint string                           
	OpenaiModel    string
	SystemPrompt   string                           //系统提示词
	BreakOnVad     bool                             //是否一检测到说话就打断 TTS
	CallOption     rustpbxgo.CallOption             //底层 rustpbxgo 呼叫参数
}

调用示例：
client := createClient(ctx, option, "")
	// Connect to server
	err = client.Connect(callType)
	if err != nil {
		logger.Fatalf("Failed to connect to server: %v", err)
	}
	defer client.Shutdown()
	// Start the call
	answer, err := client.Invite(ctx, callOption)
	if err != nil {
		logger.Fatalf("Failed to invite: %v", err)
	}
	logger.Infof("Answer SDP: %v", answer.Sdp)


llm需求：
option.LLMHandler = handler.NewLLMHandler(ctx, option.OpenaiKey, option.OpenaiEndpoint, option.SystemPrompt, option.Logger)

llm需要提供给llm处理器
func NewLLMHandler(ctx context.Context, apiKey, endpoint, systemPrompt string, logger *logrus.Logger) *LLMHandler {}
                    //上下文             //配置              //提示词             //日志                 //返回一个llm客户端，写在option结构体中

调用llm流式输出接口，逐段输出文本
response, err := option.LLMHandler.QueryStream(option.OpenaiModel, event.Text,
			func(segment string, playID string, autoHangup bool) error {
			})
llm提供的QueryStream流式输出方法，
func (h *LLMHandler) QueryStream(model, text string, ttsCallback func(segment string,   playID string,    autoHangup bool) error) (	string,    	error) {}
                                //模型    //传入问题                   //当前流式返回片段   //唯一标识符      //是否自动挂断         //返回完整文本  //错误


返回TTS结果在网页播放（？）

