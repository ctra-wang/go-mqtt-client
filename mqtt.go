package gm_client

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"crypto/tls"
	"crypto/x509"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// 这里是mqtt的配置

// NewMqttClient 创建mqtt链接
// document：https://www.emqx.com/zh/blog/how-to-use-mqtt-in-golang
func NewMqttClient(mc *MqttClient) mqtt.Client {
	// mqtt 初始化
	opts := mqtt.NewClientOptions()
	// 无证书 tcp
	if mc.Port == 1883 {
		opts.AddBroker(fmt.Sprintf("tcp://%s:%d", mc.Broker, mc.Port))
	} else {
		if mc.Ca != "" {
			// 自定义端口
			opts.AddBroker(fmt.Sprintf("ssl://%s:%d", mc.Broker, mc.Port))
			// 设置 TLS/SSL
			tlsConfig := NewSingleTlsConfig(mc.Ca)
			opts.SetTLSConfig(tlsConfig)
		} else {
			// 自定义端口
			opts.AddBroker(fmt.Sprintf("mqtt://%s:%d", mc.Broker, mc.Port))
		}
	}

	// 设置 client_id
	opts.SetClientID(mc.ClientID)
	// 设置 账户、密码验证
	opts.SetUsername(mc.User)
	opts.SetPassword(mc.Pass)
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
		//defer client.Disconnect(250)
	}

	return client
}

// mqtt 钩子函数（以下三个）
// messagePubHandler
var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Println("------------当前主题", msg.Topic())
	UnCompress(msg)
}

// connectHandler （钩子函数）
var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	// 首次连接
	fmt.Println("MQTT broker Connected")
}

// connectLostHandler （钩子函数）
var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

// --------------------------------------------------------------------- MQTT 认证：单向/双向 ---------------------------------------------------------------------

// NewSingleTlsConfig 创建单向 tls 认证
func NewSingleTlsConfig(ca string) *tls.Config {
	certpool := x509.NewCertPool()
	temp := []byte(ca)
	certpool.AppendCertsFromPEM(temp)

	return &tls.Config{
		RootCAs: certpool,
	}
}

// NewDoubleTlsConfig 创建双向 tls 认证 （暂不使用）
func NewDoubleTlsConfig(ca string, crtPem string, keyPem string) *tls.Config {
	certpool := x509.NewCertPool()
	temp := []byte(ca)
	certpool.AppendCertsFromPEM(temp)
	// Import client certificate/key pair
	// type1: with local file
	//clientKeyPair, err := tls.LoadX509KeyPair("client-crt.pem", "client-key.pem")
	//if err != nil {
	//	panic(err)
	//}
	// type2: with local string
	crtPemTemp := []byte(crtPem)
	keyPemTemp := []byte(keyPem)
	clientKeyPair, err := tls.X509KeyPair(crtPemTemp, keyPemTemp)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		RootCAs:            certpool,
		ClientAuth:         tls.NoClientCert,
		ClientCAs:          nil,
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{clientKeyPair},
	}
}

// --------------------------------------------------------------------- mqtt对2种压缩处理：gzip/zlib ---------------------------------------------------------------------

// UnCompress 解压缩payload报文
func UnCompress(msg mqtt.Message) ([]byte, error) {
	var decompressedData []byte
	// 判断消息体是否为压缩数据
	if isGzipData(msg.Payload()) {
		// 解压缩 gzip 格式的数据
		temp, err := gunzipData(msg.Payload())
		if err != nil {
			fmt.Println("MQTT broker failed to decompress data:", err)
			return nil, err
		}
		decompressedData = temp
		// 处理解压缩后的数据（这里只作为日志输出）
		fmt.Println("MQTT broker received gzip data:", string(decompressedData))

	} else if isZlibData(msg.Payload()) {
		// 解压缩 zlib 格式的数据
		temp, err := gunzlibData(msg.Payload())
		if err != nil {
			fmt.Println("MQTT broker failed to decompress data:", err)
			return nil, err
		}
		decompressedData = temp
		// 处理解压缩后的数据 （这里只作为日志输出）
		fmt.Println("MQTT broker received zlib data:", string(decompressedData))
	} else {
		decompressedData = msg.Payload()
		// 处理未压缩的数据（这里只作为日志输出）
		fmt.Println("MQTT broker received data:", string(msg.Payload()))
	}
	return decompressedData, nil
}

// 判断是否为 gzip 格式的数据
func isGzipData(data []byte) bool {
	return len(data) > 2 && data[0] == 0x1f && data[1] == 0x8b
}

// 判断是否为 zlib 格式的数据
func isZlibData(data []byte) bool {
	return len(data) > 2 && data[0] == 0x78 && (data[1] == 0x01 || data[1] == 0x9c || data[1] == 0xda)
}

// 解压缩gzip格式的数据
func gunzipData(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(data)
	gz, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	decompressedData := bytes.Buffer{}
	_, err = decompressedData.ReadFrom(gz)
	if err != nil {
		return nil, err
	}

	return decompressedData.Bytes(), nil
}

// 解压缩 zlib 格式的数据
func gunzlibData(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(data)
	z, err := zlib.NewReader(buf)
	if err != nil {
		return nil, err
	}
	defer z.Close()

	decompressedData := bytes.Buffer{}
	_, err = decompressedData.ReadFrom(z)
	if err != nil {
		return nil, err
	}

	return decompressedData.Bytes(), nil
}
