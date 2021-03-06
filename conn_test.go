package mint

import (
	"bytes"
	"crypto/x509"
	"encoding/hex"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

type pipeConn struct {
	r     *bytes.Buffer
	w     *bytes.Buffer
	rLock *sync.Mutex
	wLock *sync.Mutex
}

func pipe() (client *pipeConn, server *pipeConn) {
	client = new(pipeConn)
	server = new(pipeConn)

	c2s := bytes.NewBuffer(nil)
	server.r = c2s
	client.w = c2s

	c2sLock := new(sync.Mutex)
	server.rLock = c2sLock
	client.wLock = c2sLock

	s2c := bytes.NewBuffer(nil)
	client.r = s2c
	server.w = s2c

	s2cLock := new(sync.Mutex)
	client.rLock = s2cLock
	server.wLock = s2cLock
	return
}

func (p *pipeConn) Read(data []byte) (n int, err error) {
	p.rLock.Lock()
	n, err = p.r.Read(data)
	p.rLock.Unlock()

	// Suppress bytes.Buffer's EOF on an empty buffer
	if err == io.EOF {
		err = nil
	}
	return
}

func (p *pipeConn) Write(data []byte) (n int, err error) {
	p.wLock.Lock()
	defer p.wLock.Unlock()
	return p.w.Write(data)
}

func (p *pipeConn) Close() error {
	return nil
}

func (p *pipeConn) LocalAddr() net.Addr                { return nil }
func (p *pipeConn) RemoteAddr() net.Addr               { return nil }
func (p *pipeConn) SetDeadline(t time.Time) error      { return nil }
func (p *pipeConn) SetReadDeadline(t time.Time) error  { return nil }
func (p *pipeConn) SetWriteDeadline(t time.Time) error { return nil }

var (
	serverName = "example.com"

	// Certificate generated by Go
	// * CN: "example.com"
	// * SANs: "example.com", "www.example.com"
	// * Random 2048-bit public key
	// * Self-signed
	serverCertHex = "308202f7308201dfa00302010202012a300d06092a864886f70d01010b05003016311430" +
		"120603550403130b6578616d706c652e636f6d301e170d3135303130313030303030305a" +
		"170d3235303130313030303030305a3016311430120603550403130b6578616d706c652e" +
		"636f6d30820122300d06092a864886f70d01010105000382010f003082010a0282010100" +
		"a558ff3c12b8c4906b7f638878c71963ac95548c5d36975bc575de8775a141408c449c3e" +
		"7fe7eddf93329dd894ecb2705b7f79caa06f1477b7bd2d3ff32f43076dd32a7f9f97ed4d" +
		"4593db3f28adbea7794c14d8d206832652e93959e2b8d2b4781fadcf55c852641482f7fc" +
		"6b9e7e751442a0818c21c9cacc28e7594606ff692392510df57ce26d9c0d052f84e236b9" +
		"9e3f81daa98c554607432e3bb26a5fe3fa2b5fc5e5c1fcb1d76050328b92edc80238773d" +
		"16547ccc24c0784933d86b3f8d0ee33d90a1b47ecbfbaad12e77155f1b4e84b3e5c4d565" +
		"1717832fcbf82886eb6f925435b4ca9f87ec207b4338f03a846fbf0f68ea0e674bf50a21" +
		"d9165b690203010001a350304e300e0603551d0f0101ff04040302028430130603551d25" +
		"040c300a06082b0601050507030130270603551d110420301e820b6578616d706c652e63" +
		"6f6d820f7777772e6578616d706c652e636f6d300d06092a864886f70d01010b05000382" +
		"01010024ed08531171df6256ed5668b962ca3740e61c20d014b3aae8ac436b200ee89b50" +
		"dc5e5a74859fdf1d0f1136ad16d5fe018dac83d0e4dcfd1b5951e1502e23b7b740621b10" +
		"b3376f8b3bc5494c25917b4103d88670390e33c2b089e66326316e4bbd75fd6e5dced78f" +
		"79caf97d83b981950ed10449f61d826af4a6eb70e291fccdaa76145f7ba085d27698197f" +
		"60e944646640ea18d5439955d91a80d4dfb1e4c12f539da9423a33f479ee19a0fa9c5339" +
		"1e0d164633bea4346dc0c8081172d67ee7bca4bd5463cc147d8c062ebb31be6e9c39518c" +
		"37f5607a2d6f36114800f6c6f509893fa352a468b30ad874ae56db769f1786567e9c96c1" +
		"6b4a4b2a25dda3"
	serverCertDER, _ = hex.DecodeString(serverCertHex)
	serverCert, _    = x509.ParseCertificate(serverCertDER)

	// The corresponding private key
	serverKeyHex = "308204a40201000282010100a558ff3c12b8c4906b7f638878c71963ac95548c5d36975b" +
		"c575de8775a141408c449c3e7fe7eddf93329dd894ecb2705b7f79caa06f1477b7bd2d3f" +
		"f32f43076dd32a7f9f97ed4d4593db3f28adbea7794c14d8d206832652e93959e2b8d2b4" +
		"781fadcf55c852641482f7fc6b9e7e751442a0818c21c9cacc28e7594606ff692392510d" +
		"f57ce26d9c0d052f84e236b99e3f81daa98c554607432e3bb26a5fe3fa2b5fc5e5c1fcb1" +
		"d76050328b92edc80238773d16547ccc24c0784933d86b3f8d0ee33d90a1b47ecbfbaad1" +
		"2e77155f1b4e84b3e5c4d5651717832fcbf82886eb6f925435b4ca9f87ec207b4338f03a" +
		"846fbf0f68ea0e674bf50a21d9165b6902030100010282010074f08262ec22bcf21ef4d3" +
		"621b79445d981b6cd670be4141e85f3a68b72abac979eab44e078bf25222fab3640fbf6f" +
		"5bc37a5e9a8de8c1a301d1cb84e4ead20f18ff35995937cbded08c878d1da9f3a2e2488a" +
		"9de5bc3159135e5aef5547bdcd60ff969f825dd0d77322455cc2882f8b822eb4f1aa37e3" +
		"4d88228dac37b88f3d9b671ef6b05e2f47b562265e0d09fefb01c190c7fb4b3682231cd8" +
		"564c59b6cc788ff742fb040562110b1f849f1535164503b0a402399e2c6cf1c0847dd50a" +
		"a917b62fc3215e4eb43d7d07fa9731a51e01f0f7b694dd002b48c0bad04b9ff34e576393" +
		"c0a213a12dda4bf43a7dd4ee0563c5e0de2025eb76e049cd771c96330102818100c590bd" +
		"8f226cec50c818afb3ebe7ceeacabb107ac73ac159b1eca1a194ea550a0609c432a183e2" +
		"fee62dafdc0201426f90cb46f9b2fc7a9bcc2365b58177529cf78c209eb6a3afd1896466" +
		"63e8462729e8bf902dc1c42c7d46c1c0c99c632f0560418604b4260a1ed8d165375c674c" +
		"806c2a8e202d0b7c5a8b8717309106fb3102818100d640cae7b6adea649a8c11863a3ba8" +
		"098025a972d130aecaa4db08154fd0feb8af79bf7009c1ea2a790752464e923b53b41ff4" +
		"3ff84e6ddb94bfc5b157e6a21e1fefe11cc082e7e8b31d07eab5e13d7a84cdeeba24d283" +
		"699a8fa5138e753e88856a033ab2153c1a8200caac28377a1d09d6318ac2e946cef879a0" +
		"5acbd8e5b902818100bfe142ea189257b66190f05d3bba8951aa92a27fccadf90a076f7e" +
		"cff354e040fafa534ea565f57a81ce4fa5cb60b3c8ad8570aaa5b6e7d217232dee6a0e9c" +
		"f30cce510434f8a79347f0762d84735628330092a48e33dccdd381ec9f233f8574a03723" +
		"55c02dcdd885d6618ab23935a8e8e52fe27a3d548a90472533ab376f910281805253fd64" +
		"02875bbd22c1d5ee0d2c654a994a5f8d7622cdd7a27763e8c48ddb835e325b44930b478e" +
		"e088d6ad9b7d877c87878bd494f696323d3b5f9ce0d907cca99b049686c706941d577776" +
		"524365db5172cc5c0cd0339cfdbe5ac164095b691c52fb40afb3872fec6a9f767dd1ab83" +
		"c306e26c9eaf02fd7eef4595fe24af4902818100b5a2294d7567283f3f4bf54be7b98785" +
		"fc564f24ff2d67215ecdc7955cbf05260f48c9608a59a8ebfbedc62b4d110c1704ade704" +
		"cb27a591f69752d1d6ebe21291aec29b301efe47eced0187125f741ce52b3826beac3778" +
		"f3560448e91644fd52460f8c3afa1596c01e6cd2c37120d8122c09edf326988b48d98c27" +
		"f788eb83"
	serverKeyDER, _ = hex.DecodeString(serverKeyHex)
	serverKey, _    = x509.ParsePKCS1PrivateKey(serverKeyDER)

	psk = PreSharedKey{
		CipherSuite:  TLS_AES_128_GCM_SHA256,
		IsResumption: false,
		Identity:     []byte{0, 1, 2, 3},
		Key:          []byte{4, 5, 6, 7},
	}
	certificates = []*Certificate{
		{
			Chain:      []*x509.Certificate{serverCert},
			PrivateKey: serverKey,
		},
	}
	psks = &PSKMapCache{
		serverName: psk,
		"00010203": psk,
	}

	basicConfig = &Config{
		ServerName:   serverName,
		Certificates: certificates,
	}

	hrrConfig = &Config{
		ServerName:    serverName,
		Certificates:  certificates,
		RequireCookie: true,
	}

	alpnConfig = &Config{
		ServerName:   serverName,
		Certificates: certificates,
		NextProtos:   []string{"http/1.1", "h2"},
	}

	clientAuthConfig = &Config{
		ServerName:        serverName,
		RequireClientAuth: true,
		Certificates:      certificates,
	}

	pskConfig = &Config{
		ServerName:     serverName,
		CipherSuites:   []CipherSuite{TLS_AES_128_GCM_SHA256},
		PSKs:           psks,
		AllowEarlyData: true,
	}

	pskECDHEConfig = &Config{
		ServerName:   serverName,
		CipherSuites: []CipherSuite{TLS_AES_128_GCM_SHA256},
		Certificates: certificates,
		PSKs:         psks,
	}

	pskDHEConfig = &Config{
		ServerName:   serverName,
		CipherSuites: []CipherSuite{TLS_AES_128_GCM_SHA256},
		Certificates: certificates,
		PSKs:         psks,
		Groups:       []NamedGroup{FFDHE2048},
	}

	resumptionConfig = &Config{
		ServerName:         serverName,
		Certificates:       certificates,
		SendSessionTickets: true,
	}

	ffdhConfig = &Config{
		ServerName:   serverName,
		Certificates: certificates,
		CipherSuites: []CipherSuite{TLS_AES_128_GCM_SHA256},
		Groups:       []NamedGroup{FFDHE2048},
	}

	x25519Config = &Config{
		ServerName:   serverName,
		Certificates: certificates,
		CipherSuites: []CipherSuite{TLS_AES_128_GCM_SHA256},
		Groups:       []NamedGroup{X25519},
	}
)

func assertContextEquals(t *testing.T, c *cryptoContext, s *cryptoContext) {
	assertEquals(t, c.suite, s.suite)
	// XXX: Figure out a way to compare ciphers?
	assertEquals(t, c.params.hash, s.params.hash)
	assertEquals(t, c.params.keyLen, s.params.keyLen)
	assertEquals(t, c.params.ivLen, s.params.ivLen)
	assertByteEquals(t, c.zero, s.zero)

	assertByteEquals(t, c.h2, s.h2)
	assertByteEquals(t, c.h3, s.h3)
	assertByteEquals(t, c.h4, s.h4)
	assertByteEquals(t, c.h5, s.h5)
	assertByteEquals(t, c.h6, s.h6)

	assertByteEquals(t, c.pskSecret, s.pskSecret)
	assertByteEquals(t, c.earlySecret, s.earlySecret)

	if c.binderKey != nil && s.binderKey != nil {
		assertByteEquals(t, c.binderKey, s.binderKey)
	}

	if c.earlyTrafficSecret != nil && s.earlyTrafficSecret != nil {
		assertByteEquals(t, c.earlyTrafficSecret, s.earlyTrafficSecret)
		assertByteEquals(t, c.earlyExporterSecret, s.earlyExporterSecret)
		assertDeepEquals(t, c.clientEarlyTrafficKeys, s.clientEarlyTrafficKeys)
	}

	assertByteEquals(t, c.dhSecret, s.dhSecret)
	assertByteEquals(t, c.handshakeSecret, s.handshakeSecret)
	assertByteEquals(t, c.clientHandshakeTrafficSecret, s.clientHandshakeTrafficSecret)
	assertByteEquals(t, c.serverHandshakeTrafficSecret, s.serverHandshakeTrafficSecret)
	assertDeepEquals(t, c.clientHandshakeKeys, s.clientHandshakeKeys)
	assertDeepEquals(t, c.serverHandshakeKeys, s.serverHandshakeKeys)

	assertByteEquals(t, c.serverFinishedKey, s.serverFinishedKey)
	assertByteEquals(t, c.serverFinishedData, s.serverFinishedData)

	assertByteEquals(t, c.clientFinishedKey, s.clientFinishedKey)
	assertByteEquals(t, c.clientFinishedData, s.clientFinishedData)

	assertByteEquals(t, c.masterSecret, s.masterSecret)
	assertByteEquals(t, c.clientTrafficSecret, s.clientTrafficSecret)
	assertByteEquals(t, c.serverTrafficSecret, s.serverTrafficSecret)
	assertDeepEquals(t, c.clientTrafficKeys, s.clientTrafficKeys)
	assertDeepEquals(t, c.serverTrafficKeys, s.serverTrafficKeys)
	assertByteEquals(t, c.exporterSecret, s.exporterSecret)
	assertByteEquals(t, c.resumptionSecret, s.resumptionSecret)
}

func TestBasicFlows(t *testing.T) {
	for _, conf := range []*Config{basicConfig, hrrConfig, alpnConfig, ffdhConfig, x25519Config} {
		cConn, sConn := pipe()

		client := Client(cConn, conf)
		server := Server(sConn, conf)

		done := make(chan bool)
		go func(t *testing.T) {
			err := server.Handshake()
			assertNotError(t, err, "Server failed handshake")
			done <- true
		}(t)

		err := client.Handshake()
		assertNotError(t, err, "Client failed handshake")

		<-done

		assertDeepEquals(t, client.handshake.ConnectionParams(), server.handshake.ConnectionParams())
		assertContextEquals(t, client.handshake.cryptoContext(), server.handshake.cryptoContext())
	}
}

func TestClientAuth(t *testing.T) {
	cConn, sConn := pipe()

	client := Client(cConn, clientAuthConfig)
	server := Server(sConn, clientAuthConfig)

	done := make(chan bool)
	go func(t *testing.T) {
		err := server.Handshake()
		assertNotError(t, err, "Server failed handshake")
		done <- true
	}(t)

	err := client.Handshake()
	assertNotError(t, err, "Client failed handshake")

	<-done

	assertContextEquals(t, client.handshake.cryptoContext(), server.handshake.cryptoContext())
	assertDeepEquals(t, client.handshake.ConnectionParams(), server.handshake.ConnectionParams())
	assert(t, client.handshake.ConnectionParams().UsingClientAuth, "Session did not negotiate client auth")
}

func TestPSKFlows(t *testing.T) {
	for _, conf := range []*Config{pskConfig, pskECDHEConfig, pskDHEConfig} {
		cConn, sConn := pipe()

		client := Client(cConn, conf)
		server := Server(sConn, conf)

		done := make(chan bool)
		go func(t *testing.T) {
			err := server.Handshake()
			assertNotError(t, err, "Server failed handshake")
			done <- true
		}(t)

		err := client.Handshake()
		assertNotError(t, err, "Client failed handshake")

		<-done

		assertDeepEquals(t, client.handshake.ConnectionParams(), server.handshake.ConnectionParams())
		assert(t, client.handshake.ConnectionParams().UsingPSK, "Session did not use the provided PSK")

		assertContextEquals(t, client.handshake.cryptoContext(), server.handshake.cryptoContext())
	}
}

func TestResumption(t *testing.T) {
	// Phase 1: Verify that the session ticket gets sent and stored
	clientConfig := *resumptionConfig
	serverConfig := *resumptionConfig

	cConn1, sConn1 := pipe()
	client1 := Client(cConn1, &clientConfig)
	server1 := Server(sConn1, &serverConfig)

	done := make(chan bool)
	go func(t *testing.T) {
		err := server1.Handshake()
		assertNotError(t, err, "Server failed handshake")
		done <- true
	}(t)

	err := client1.Handshake()
	assertNotError(t, err, "Client failed handshake")

	client1.Read(nil)
	<-done

	assertDeepEquals(t, client1.handshake.ConnectionParams(), server1.handshake.ConnectionParams())
	assertContextEquals(t, client1.handshake.cryptoContext(), server1.handshake.cryptoContext())
	assertEquals(t, clientConfig.PSKs.Size(), 1)
	assertEquals(t, serverConfig.PSKs.Size(), 1)

	clientCache := clientConfig.PSKs.(*PSKMapCache)
	serverCache := serverConfig.PSKs.(*PSKMapCache)

	var serverPSK PreSharedKey
	for _, key := range *serverCache {
		serverPSK = key
	}
	var clientPSK PreSharedKey
	for _, key := range *clientCache {
		clientPSK = key
	}
	assertDeepEquals(t, clientPSK, serverPSK)

	// Phase 2: Verify that the session ticket gets used as a PSK
	cConn2, sConn2 := pipe()
	client2 := Client(cConn2, &clientConfig)
	server2 := Server(sConn2, &serverConfig)

	go func(t *testing.T) {
		err := server2.Handshake()
		assertNotError(t, err, "Server failed second handshake")
		done <- true
	}(t)

	err = client2.Handshake()
	assertNotError(t, err, "Client failed second handshake")

	client2.Read(nil)
	<-done

	assertDeepEquals(t, client2.handshake.ConnectionParams(), server2.handshake.ConnectionParams())
	assertContextEquals(t, client2.handshake.cryptoContext(), server2.handshake.cryptoContext())
}

func Test0xRTT(t *testing.T) {
	conf := pskConfig
	cConn, sConn := pipe()

	client := Client(cConn, conf)
	client.earlyData = []byte("hello 0xRTT world!")

	server := Server(sConn, conf)

	done := make(chan bool)
	go func(t *testing.T) {
		err := server.Handshake()
		assertNotError(t, err, "Server failed handshake")
		done <- true
	}(t)

	err := client.Handshake()
	assertNotError(t, err, "Client failed handshake")

	<-done

	assertContextEquals(t, client.handshake.cryptoContext(), server.handshake.cryptoContext())
	assertDeepEquals(t, client.handshake.ConnectionParams(), server.handshake.ConnectionParams())
	assert(t, client.handshake.ConnectionParams().UsingEarlyData, "Session did not negotiate early data")
	assertByteEquals(t, client.earlyData, server.readBuffer)
}

func Test0xRTTFailure(t *testing.T) {
	// Client thinks it has a PSK
	clientConfig := &Config{
		ServerName:   serverName,
		CipherSuites: []CipherSuite{TLS_AES_128_GCM_SHA256},
		PSKs:         psks,
	}

	// Server doesn't
	serverConfig := &Config{
		ServerName:   serverName,
		CipherSuites: []CipherSuite{TLS_AES_128_GCM_SHA256},
	}

	cConn, sConn := pipe()

	client := Client(cConn, clientConfig)
	client.earlyData = []byte("hello 0xRTT world!")

	server := Server(sConn, serverConfig)

	done := make(chan bool)
	go func(t *testing.T) {
		err := server.Handshake()
		assertNotError(t, err, "Server failed handshake")
		done <- true
	}(t)

	err := client.Handshake()
	assertNotError(t, err, "Client failed handshake")

	<-done
}

func TestKeyUpdate(t *testing.T) {
	cConn, sConn := pipe()

	conf := basicConfig
	client := Client(cConn, conf)
	server := Server(sConn, conf)

	zeroBuf := []byte{}
	c2s := make(chan bool)
	s2c := make(chan bool)
	go func(t *testing.T) {
		err := server.Handshake()
		assertNotError(t, err, "Server failed handshake")
		s2c <- true

		// Test server-initiated KeyUpdate
		<-c2s
		err = server.SendKeyUpdate(false)
		assertNotError(t, err, "Key update send failed")
		s2c <- true

		// Null read to trigger key update
		<-c2s
		server.Read(zeroBuf)
		s2c <- true

		// Null read to trigger key update and KeyUpdate response
		<-c2s
		server.Read(zeroBuf)
		s2c <- true
	}(t)

	err := client.Handshake()
	assertNotError(t, err, "Client failed handshake")
	<-s2c

	clientContext0 := *client.handshake.cryptoContext()
	serverContext0 := *server.handshake.cryptoContext()
	assertContextEquals(t, &clientContext0, &serverContext0)

	// Null read to trigger key update
	c2s <- true
	<-s2c
	client.Read(zeroBuf)

	clientContext1 := *client.handshake.cryptoContext()
	serverContext1 := *server.handshake.cryptoContext()
	assertContextEquals(t, &clientContext1, &serverContext1)
	assertNotByteEquals(t, clientContext0.serverTrafficKeys.key, clientContext1.serverTrafficKeys.key)
	assertNotByteEquals(t, clientContext0.serverTrafficKeys.iv, clientContext1.serverTrafficKeys.iv)
	assertByteEquals(t, clientContext0.clientTrafficKeys.key, clientContext1.clientTrafficKeys.key)
	assertByteEquals(t, clientContext0.clientTrafficKeys.iv, clientContext1.clientTrafficKeys.iv)

	// Test client-initiated KeyUpdate
	client.SendKeyUpdate(false)
	c2s <- true
	<-s2c

	clientContext2 := *client.handshake.cryptoContext()
	serverContext2 := *server.handshake.cryptoContext()
	assertContextEquals(t, &clientContext2, &serverContext2)
	assertByteEquals(t, clientContext1.serverTrafficKeys.key, clientContext2.serverTrafficKeys.key)
	assertByteEquals(t, clientContext1.serverTrafficKeys.iv, clientContext2.serverTrafficKeys.iv)
	assertNotByteEquals(t, clientContext1.clientTrafficKeys.key, clientContext2.clientTrafficKeys.key)
	assertNotByteEquals(t, clientContext1.clientTrafficKeys.iv, clientContext2.clientTrafficKeys.iv)

	// Test client-initiated with keyUpdateRequested
	client.SendKeyUpdate(true)
	c2s <- true
	<-s2c
	client.Read(zeroBuf)

	clientContext3 := *client.handshake.cryptoContext()
	serverContext3 := *server.handshake.cryptoContext()
	assertContextEquals(t, &clientContext3, &serverContext3)
	assertNotByteEquals(t, clientContext2.serverTrafficKeys.key, clientContext3.serverTrafficKeys.key)
	assertNotByteEquals(t, clientContext2.serverTrafficKeys.iv, clientContext3.serverTrafficKeys.iv)
	assertNotByteEquals(t, clientContext2.clientTrafficKeys.key, clientContext3.clientTrafficKeys.key)
	assertNotByteEquals(t, clientContext2.clientTrafficKeys.iv, clientContext3.clientTrafficKeys.iv)
}
