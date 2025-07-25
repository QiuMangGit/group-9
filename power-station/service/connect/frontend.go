package connect

import (
	"encoding/json"
	"group-9/llm"
	"group-9/repository"
	"group-9/service/client"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/shenjinti/go711"
	"github.com/shenjinti/go722"
)

// FrontendServer 管理前端WebSocket连接
type FrontendServer struct {
	upgrader      websocket.Upgrader
	clients       map[*websocket.Conn]bool
	conn          *websocket.Conn
	RealTimeConn  *websocket.Conn
	llm           *llm.LLMHandler
	backendConn   *websocket.Conn // shane: 与后端的连接
	backendServer *BackendServer  // shane: 后端服务实例
	codec         string          // shane: codec for audio stream
	asrOption     *client.ASROption
	ttsOption     *client.TTSOption
	repo          *repository.Repository
}

func NewFrontendServer(llm *llm.LLMHandler, backendConn *websocket.Conn, backendServer *BackendServer, codec string, asrOption *client.ASROption, ttsOption *client.TTSOption, repo *repository.Repository) *FrontendServer {
	return &FrontendServer{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true }, // shane: 允许跨域
		}, // http的升级
		clients:       make(map[*websocket.Conn]bool),
		llm:           llm,
		backendConn:   backendConn,
		backendServer: backendServer,
		codec:         codec,
		asrOption:     asrOption,
		ttsOption:     ttsOption,
		repo:          repo,
	}
}

// Start shane: http 升级为 websocket
func (s *FrontendServer) Start(r *gin.Engine, port string) {
	// shane: 使用gin处理WebSocket连接
	r.GET("/ws", func(c *gin.Context) {
		s.handleWebSocket(c.Writer, c.Request) // shane: http请求升级为WebSocket连接
	})
	// shane: 新增路由 /ws2 处理实时语音
	r.GET("/ws2", func(c *gin.Context) {
		s.handleWebSocket2(c.Writer, c.Request)
	})

	// shane: 监听后端消息
	if s.backendConn != nil {
		go s.receiveBackendMessages()
	}

	go func() {
		if err := r.Run(":" + port); err != nil {
			log.Println("Connection failed:", err)
		}
	}() // shane: 开协程防止阻塞
	log.Printf("Connected to the front end! Serve on %s", port)
}

// shane: 处理前端WebSocket连接
func (s *FrontendServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade Connection Failed:", err)
		return
	}
	defer func() {
		conn.Close()
		delete(s.clients, conn)
		log.Println("WebSocket Connection closed")
	}()
	s.clients[conn] = true // shane: 设置已经连接（状态信息）
	s.conn = conn          // shane: 一定要记住保存连接，后面需要用到

	done := make(chan struct{})
	go s.ReceiveMessages(conn, done) // shane: 启动接收消息的协程
	// shane: 阻塞当前函数
	<-done
}

func (s *FrontendServer) handleWebSocket2(w gin.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade Connection Failed:", err)
		return
	}
	defer func() {
		conn.Close()
		delete(s.clients, conn)
		s.RealTimeConn = nil
		log.Println("RealTime WebSocket Connection closed")
	}()
	s.clients[conn] = true // shane: 设置已经连接（状态信息）
	s.RealTimeConn = conn  // shane: 一定要记住保存连接，后面需要用到

	done := make(chan struct{})
	go s.ReceiveRealTimeMessage(conn, done) // shane: 启动接收实时消息的协程
	<-done
}

// convertAudio shane: parse audio data and convert to g722 or pcmu
func (s *FrontendServer) convertAudio(audioData []byte) ([]byte, error) {
	if s.codec == "g722" {
		encoder := go722.NewG722Encoder(go722.Rate64000, 0)
		return encoder.Encode(audioData), nil
	} else if s.codec == "pcmu" {
		return go711.EncodePCMU(audioData)
	} else if s.codec == "pcm" {
		return audioData, nil
	} else if s.codec == "wav" {
		return audioData, nil
	}

	return audioData, nil
}

// ReceiveMessages shane: 接收前端发送的消息,没有返回值
func (s *FrontendServer) ReceiveMessages(conn *websocket.Conn, done chan struct{}) {
	// TODO: 设计读超时

	// shane: 接收前端发送的消息
	for {
		if conn == nil {
			log.Println("Connection is nil, waiting for connection")
			break
		}
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Receive from frontend failed:", err)
			delete(s.clients, s.conn)
			break
		}
		// shane: 主动关闭连接
		if string(msg) == "close" {
			log.Println("Frontend requested to close connection")
			break
		}
		log.Printf("Receive from frontend: %s", string(msg))
		// s.handleMessage(msg)
		s.SendMessages(conn, msg) // shane: 接收到消息之后发送消息
	}

	close(done) // shane: 关闭done通道，通知主协程结束
}

func (s *FrontendServer) ReceiveRealTimeMessage(conn *websocket.Conn, done chan struct{}) {
	for {
		if conn == nil {
			log.Println("Connection is nil, waiting for connection")
			break
		}
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Receive Frontend Message failed:", err)
			break
		}
		// shane: 主动关闭连接
		if string(msg) == "close" {
			log.Println("Frontend requested to close connection")
			break
		}

		if msgType == websocket.BinaryMessage {
			// shane: handle audio stram
			log.Println("Received audio stream from frontend")
			// log.Println(msg) // shane: 打印音频数据
			// shane: 转换音频格式
			convertedAudio, err := s.convertAudio(msg)
			if err != nil {
				log.Println("Convert audio format failed:", err)
				continue
			}

			// shane: forward audio stream to rust backend
			if err := s.backendConn.WriteMessage(websocket.BinaryMessage, convertedAudio); err != nil {
				log.Println("Forward audio stream to rust backend err:", err)
				err := s.backendServer.reconnect("webrtc")
				if err != nil {
					return
				} else {
					// shane: reforward audio stream to rust backend
					if err := s.backendConn.WriteMessage(websocket.BinaryMessage, convertedAudio); err != nil {
						log.Println("Retrying to forward audio stream failed:", err)
					} else {
						log.Println("Successfully retried forwarding audio stream to rust backend")
					}
				}
			} else {
				log.Println("Forwarded audio stream to rust backend successfully")
			}
		} else {
			// shane: parse message
			var frontendEvent struct {
				Event     string          `json:"event"`
				Sdp       string          `json:"sdp"`
				Candidate json.RawMessage `json:"candidate"`
				Command   string          `json:"command"`
				Reason    string          `json:"reason"`
				Initiator string          `json:"initiator"`
			}
			if err := json.Unmarshal(msg, &frontendEvent); err != nil {
				log.Println("parse front end message failed:", err)
				continue
			}

			// shane: receive offer
			if frontendEvent.Event == "offer" && frontendEvent.Sdp != "" {
				log.Printf("receive front end offer message, sdp: %s", frontendEvent.Sdp)

				inviteCmd := client.InviteCommand{
					Command: "invite",
					Option: client.CallOption{
						Offer:  frontendEvent.Sdp,
						Caller: "frontend",
						Callee: "rust",
						ASR:    s.asrOption,
						TTS:    s.ttsOption,
					},
				}

				cmdBytes, err := json.Marshal(inviteCmd)
				if err != nil {
					log.Println("marshal invite command failed:", err)
					continue
				}
				if err := s.backendConn.WriteMessage(websocket.TextMessage, cmdBytes); err != nil {
					log.Printf("forward candidate command to rust backend err: %v, Command data: %s", err, string(cmdBytes))
					if s.backendConn == nil {
						log.Println("Backend connection is nil, trying to reconnect")
						err := s.backendServer.reconnect("webrtc")
						if err != nil {
							return
						} else {
							// shane: 重发invite
							log.Println("Reconnected to backend successfully, will retry sending invite command")
							s.backendConn = s.backendServer.Conn
							if err := s.backendConn.WriteMessage(websocket.TextMessage, cmdBytes); err != nil {
								log.Println("Retrying to forward invite command failed:", err)
							} else {
								log.Println("Successfully retried forwarding invite command to rust backend")
							}
						}
					} else {
						log.Println("Failed to forward invite command to rust backend, will retry later")
					}
				} else {
					log.Println("Forwarded invite command with ASR config to rust backend")
				}
			}

			// shane: handle candidate
			if frontendEvent.Event == "candidate" && frontendEvent.Candidate != nil {
				log.Printf("Received ICE candidate: %s", string(frontendEvent.Candidate)) // shane: 调试

				// shane: parse candidate
				var candidate struct {
					Candidate     string `json:"candidate"`
					SdpMid        string `json:"sdpMid"`
					SdpMLineIndex int    `json:"sdpMLineIndex"`
				}
				if err := json.Unmarshal(frontendEvent.Candidate, &candidate); err != nil {
					log.Println("parse candidate failed:", err)
					continue
				}

				candidateCmd := client.CandidateCommand{
					Command:    "candidate",
					Candidates: []string{candidate.Candidate},
				}

				// shane: marshal to json
				cmdBytes, err := json.Marshal(candidateCmd)
				if err != nil {
					log.Println("marshal candidate command failed:", err)
					continue
				}
				if err := s.backendConn.WriteMessage(websocket.TextMessage, cmdBytes); err != nil {
					log.Println("forward candidate command to rust backend err:", err)
					err := s.backendServer.reconnect("webrtc")
					if err != nil {
						return
					} else {
						// shane: 重发candidate
						s.backendConn = s.backendServer.Conn
						if err := s.backendConn.WriteMessage(websocket.TextMessage, cmdBytes); err != nil {
							log.Println("Retrying to forward candidate command failed:", err)
						} else {
							log.Println("Successfully retried forwarding candidate command to rust backend")
						}

					}
				}
			}

			// shane: handle hangup event
			if frontendEvent.Command == "hangup" {
				hangupCmd := client.HangupCommand{
					Command:   "hangup",
					Reason:    frontendEvent.Reason,
					Initiator: frontendEvent.Initiator,
				}

				cmdBytes, err := json.Marshal(hangupCmd)
				if err != nil {
					log.Println("marshal hangup command failed:", err)
					continue
				}
				if err := s.backendConn.WriteMessage(websocket.TextMessage, cmdBytes); err != nil {
					log.Println("forward hangup command to rust backend err:", err)
					err := s.backendServer.reconnect("webrtc")
					if err != nil {
						return
					} else {
						// shane: 重发hangup
						s.backendConn = s.backendServer.Conn
						if err := s.backendConn.WriteMessage(websocket.TextMessage, cmdBytes); err != nil {
							log.Println("Retrying to forward hangup command failed:", err)
						} else {
							log.Println("Successfully retried forwarding hangup command to rust backend")
						}
					} // shane: 重连后端
				} else {
					log.Println("Forwarded hangup command to rust backend")
				}
			}
		}
	}
	close(done)
}

// SendMessages shane: 接收到消息之后发送消息
func (s *FrontendServer) SendMessages(conn *websocket.Conn, msg []byte) {
	if conn == nil {
		log.Println("Connection is nil, waiting for connection")
		return
	}

	// shane: 发送LLM处理后的消息到前端
	response, _, err := s.llm.Query("qwen-turbo", string(msg))
	if err != nil {
		log.Println("LLM query failed:", err)
		return
	}
	// shane: 发送回复到前端
	if err := conn.WriteMessage(websocket.TextMessage, []byte(response)); err != nil {
		log.Println("sendMessage failed:", err)
	}
}

// receiveBackendMessages shane: 接收并打印后端发送的消息
// receiveBackendMessages shane: 接收并打印后端发送的消息
func (s *FrontendServer) receiveBackendMessages() {
	callType := "webrtc"
	for {
		if s.backendConn == nil {
			// shane: reconnect the backend connection
			err := s.backendServer.reconnect(callType)
			if err != nil {
				time.Sleep(1 * time.Second)
				continue
			}
			s.backendConn = s.backendServer.Conn
		}
		// shane: read message from backend
		messageType, msg, err := s.backendConn.ReadMessage()
		if err != nil {
			// shane: backend connection is closed, set it to nil, and continue the loop
			s.backendConn = nil
			continue
		}
		// shane: type down message fron rust backend
		log.Printf("Received from rust backend (type %d): %s", messageType, string(msg))

		// shane: parse the message
		var event struct {
			Event string `json:"event"`
			Text  string `json:"text"`
		}
		if err := json.Unmarshal(msg, &event); err == nil {
			// shane: handle asrFinal and send ASR result to LLM handler
			if event.Event == "asrFinal" && event.Text != "" {
				log.Printf("received ASR response: %s", event.Text) //用户说的话，存数据库

				// 关键修改：先转发asrFinal消息给前端（用户消息优先发送）
				s.forwardRustMessageToFrontend(msg)

				// 将event.Text存储到数据库中
				var voicdeContent = repository.VoiceConversation{
					Context: event.Text,
					From:    "User",
					RobotId: repository.RobotId,
				}
				if err := s.repo.AddVoiceConversation(&voicdeContent); err != nil {
					log.Printf("Failed to add conversation: %v", err)
				}

				// shane: use LLM to handle ASR result
				if s.llm != nil {
					log.Printf("handle ASR result via LLM...")
					response, _, err := s.llm.Query("qwen-turbo", event.Text) //ai回复，存数据库

					if err != nil {
						log.Println("LLM handle ASR result failed:", err)
					} else {
						log.Printf("LLM response: %s", response)

						var voicdeContent = repository.VoiceConversation{
							Context: response,
							From:    "AI",
							RobotId: repository.RobotId,
						}

						if err := s.repo.AddVoiceConversation(&voicdeContent); err != nil {
							log.Printf("Failed to add conversation: %v", err)
						}

						if s.RealTimeConn != nil {
							// 关键修改：将AI回复包装为llmResponse事件
							llmMsg := map[string]string{
								"event": "llmResponse",
								"text":  response,
							}
							llmBytes, err := json.Marshal(llmMsg)
							if err != nil {
								log.Println("Failed to marshal LLM response:", err)
								continue
							}

							if err := s.RealTimeConn.WriteMessage(websocket.TextMessage, llmBytes); err != nil {
								log.Printf("Failed to send LLM response to frontend: %v", err)
								if websocket.IsCloseError(err, websocket.CloseGoingAway) {
									s.RealTimeConn = nil
								}
							}
						} else {
							log.Println("RealTime conn is nil, cannot send LLM response")
						}

						// shane: send TTS command to Rust backend
						ttsCmd := client.TtsCommand{
							Command: "tts",
							Text:    response,
							Speaker: s.ttsOption.Speaker,
							Option:  s.ttsOption,
						}

						cmdBytes, err := json.Marshal(ttsCmd)
						if err != nil {
							log.Println("generate TTS Command failed:", err)
						} else {
							if err := s.backendConn.WriteMessage(websocket.TextMessage, cmdBytes); err != nil {
								log.Println("send TTS command to Rust backend failed:", err)
							} else {
								log.Println("TTS command sent to Rust backend successfully")
							}
						}
					}
				}

			} else if event.Event == "asrDelta" {
				// shane: handle ASR delta event
				log.Printf("ASR realtime recognize: %s", event.Text)
			} else if event.Event == "speaking" {
				log.Printf("detecting speaking")
			} else if event.Event == "silence" {
				log.Printf("detecting silence")
			} else if event.Event == "trackStart" {
				log.Printf("track started")
			} else if event.Event == "trackEnd" {
				log.Printf("track ended")
			}

		}

		// shane: forward the message to the frontend (除了asrFinal外的其他消息)
		if event.Event != "asrFinal" { // 避免重复发送asrFinal
			s.forwardRustMessageToFrontend(msg)
		}
	}
}

// forwardRustMessageToFrontend shane: 转发后端消息给前端
func (s *FrontendServer) forwardRustMessageToFrontend(msg []byte) {
	if s.RealTimeConn != nil {
		if err := s.RealTimeConn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Println("Failed to forward backend message to frontend:", err)
		} else {
			log.Println("Successfully forwarded backend message to frontend")
		}
	} else {
		log.Println("Frontend connection is nil, cannot forward message")
	}
}
