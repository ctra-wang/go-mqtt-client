package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

func onMessageReceived(client MQTT.Client, message MQTT.Message) {
	fmt.Printf("Received message on topic: %s\n", message.Topic())
	fmt.Printf("Message: %s\n", message.Payload())
}

func main() {
	// 创建 MQTT 连接参数
	opts := MQTT.NewClientOptions().AddBroker("tcp://your-address:your-port")
	opts.SetClientID("go_client2")
	opts.SetUsername("your-username") // 设置用户名
	opts.SetPassword("your-password") // 设置密码

	// 创建 MQTT 客户端实例
	client := MQTT.NewClient(opts)

	// 连接到 MQTT 代理
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// 订阅主题
	if token := client.Subscribe("test/topic", 0, onMessageReceived); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	fmt.Println("Connected to MQTT broker")

	// 等待信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	// 断开连接
	client.Disconnect(250)
	fmt.Println("Disconnected from MQTT broker")
}
