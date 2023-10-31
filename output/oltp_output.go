package output

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/childe/gohangout/codec"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"google.golang.org/grpc/metadata"
	"os"
	"time"

	"github.com/childe/gohangout/topology"
	"github.com/youmark/pkcs8"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/klog/v2"
)

const (
	defaultServiceAddress = "localhost:4317"
	defaultTimeout        = 5
	defaultCompression    = "none"
	userAgent             = "gohangout"
)

func init() {
	Register("Oltp", newOTLPOutput)
}

// OLTPOutput OpenTelemetry protocol output
type OLTPOutput struct {
	config      *OLTPOutputConfig
	oltpEncoder codec.OltpEncoder

	timeout          time.Duration
	grpcClientConn   *grpc.ClientConn
	logServiceClient plogotlp.GRPCClient
	callOptions      []grpc.CallOption
}

type OLTPOutputConfig struct {
	ServiceAddress string `json:"service_address"`
	ClientAuthConfig
	Timeout     int               `json:"timeout"`
	Compression string            `json:"compression"`
	Headers     map[string]string `json:"headers"`
}

type ClientAuthConfig struct {
	TLSCA               string `json:"tls_ca"`
	TLSCert             string `json:"tls_cert"`
	TLSKey              string `json:"tls_key"`
	TLSKeyPwd           string `json:"tls_key_pwd"`
	TLSMinVersion       string `json:"tls_min_version"`
	InsecureSkipVerify  bool   `json:"insecure_skip_verify"`
	ServerName          string `json:"tls_server_name"`
	RenegotiationMethod string `json:"tls_renegotiation_method"`
	Enable              *bool  `json:"tls_enable"`
}

func (c *OLTPOutputConfig) TLSConfig() (*tls.Config, error) {
	if c.Enable != nil && !*c.Enable {
		return nil, nil
	}

	empty := c.TLSCA == "" && c.TLSKey == "" && c.TLSCert == ""
	empty = empty && !c.InsecureSkipVerify && c.ServerName == ""
	empty = empty && (c.RenegotiationMethod == "" || c.RenegotiationMethod == "never")

	if empty {
		if c.Enable != nil && *c.Enable {
			return &tls.Config{}, nil
		}
		return nil, nil
	}

	var renegotiationMethod tls.RenegotiationSupport
	switch c.RenegotiationMethod {
	case "", "never":
		renegotiationMethod = tls.RenegotiateNever
	case "once":
		renegotiationMethod = tls.RenegotiateOnceAsClient
	case "freely":
		renegotiationMethod = tls.RenegotiateFreelyAsClient
	default:
		return nil, fmt.Errorf("unrecognized renegotation method %q, choose from: 'never', 'once', 'freely'", c.RenegotiationMethod)
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: c.InsecureSkipVerify,
		Renegotiation:      renegotiationMethod,
	}

	if c.TLSCA != "" {
		pool, err := makeCertPool([]string{c.TLSCA})
		if err != nil {
			return nil, err
		}
		tlsConfig.RootCAs = pool
	}

	if c.TLSCert != "" && c.TLSKey != "" {
		err := loadCertificate(tlsConfig, c.TLSCert, c.TLSKey, c.TLSKeyPwd)
		if err != nil {
			return nil, err
		}
	}

	if c.ServerName != "" {
		tlsConfig.ServerName = c.ServerName
	}

	return tlsConfig, nil
}

func newOTLPOutput(config map[interface{}]interface{}) topology.Output {
	var oltpConfig OLTPOutputConfig
	configBytes, _ := json.Marshal(config)
	if configBytes != nil {
		err := json.Unmarshal(configBytes, &oltpConfig)
		if err != nil {
			klog.Fatalf("wrong format oltp config! %v", err)
		}
		klog.Info(string(configBytes))
	}
	if oltpConfig.ServiceAddress == "" {
		oltpConfig.ServiceAddress = defaultServiceAddress
	}
	if oltpConfig.Timeout <= 0 {
		oltpConfig.Timeout = defaultTimeout
	}
	if oltpConfig.Compression == "" {
		oltpConfig.Compression = defaultCompression
	}

	var grpcTLSDialOption grpc.DialOption
	if tlsConfig, err := oltpConfig.TLSConfig(); err != nil {
		klog.Fatalf("parse tls config err: %v", err)
	} else if tlsConfig != nil {
		grpcTLSDialOption = grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
	} else {
		grpcTLSDialOption = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	grpcClientConn, err := grpc.Dial(oltpConfig.ServiceAddress, grpcTLSDialOption, grpc.WithUserAgent(userAgent))
	if err != nil {
		klog.Fatal("dial tcp server err :%v", err)
	}

	// TODO: add metric and trace client
	logServiceClient := plogotlp.NewGRPCClient(grpcClientConn)

	var callOptions []grpc.CallOption
	if oltpConfig.Compression != "" && oltpConfig.Compression != "none" {
		callOptions = append(callOptions, grpc.UseCompressor(oltpConfig.Compression))
	}

	o := &OLTPOutput{
		config:           &oltpConfig,
		oltpEncoder:      codec.OltpEncoder{},
		timeout:          time.Duration(oltpConfig.Timeout) * time.Second,
		grpcClientConn:   grpcClientConn,
		logServiceClient: logServiceClient,
		callOptions:      callOptions,
	}

	return o
}

func (o *OLTPOutput) Emit(event map[string]interface{}) {
	request, _ := o.oltpEncoder.Encode(event)
	ctx, cancel := context.WithTimeout(context.Background(), o.timeout)
	if len(o.config.Headers) > 0 {
		ctx = metadata.NewOutgoingContext(ctx, metadata.New(o.config.Headers))
	}
	defer cancel()
	_, err := o.logServiceClient.Export(ctx, request, o.callOptions...)
	if err != nil {
		klog.Errorf("oltp log export err: %v", err)
	}
}

func (o *OLTPOutput) Shutdown() {
	if o.grpcClientConn != nil {
		err := o.grpcClientConn.Close()
		if err != nil {
			klog.Errorf("close grpc client err: %v", err)
		}
		o.grpcClientConn = nil
	}
}

func makeCertPool(certFiles []string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	for _, certFile := range certFiles {
		cert, err := os.ReadFile(certFile)
		if err != nil {
			return nil, fmt.Errorf("could not read certificate %q: %w", certFile, err)
		}
		if !pool.AppendCertsFromPEM(cert) {
			return nil, fmt.Errorf("could not parse any PEM certificates %q: %w", certFile, err)
		}
	}
	return pool, nil
}

func loadCertificate(config *tls.Config, certFile, keyFile, privateKeyPassphrase string) error {
	certBytes, err := os.ReadFile(certFile)
	if err != nil {
		return fmt.Errorf("could not load certificate %q: %w", certFile, err)
	}

	keyBytes, err := os.ReadFile(keyFile)
	if err != nil {
		return fmt.Errorf("could not load private key %q: %w", keyFile, err)
	}

	keyPEMBlock, _ := pem.Decode(keyBytes)
	if keyPEMBlock == nil {
		return fmt.Errorf("failed to decode private key: no PEM data found")
	}

	var cert tls.Certificate
	if keyPEMBlock.Type == "ENCRYPTED PRIVATE KEY" {
		if privateKeyPassphrase == "" {
			return fmt.Errorf("missing password for PKCS#8 encrypted private key")
		}
		var decryptedKey *rsa.PrivateKey
		decryptedKey, err = pkcs8.ParsePKCS8PrivateKeyRSA(keyPEMBlock.Bytes, []byte(privateKeyPassphrase))
		if err != nil {
			return fmt.Errorf("failed to parse encrypted PKCS#8 private key: %w", err)
		}
		cert, err = tls.X509KeyPair(certBytes, pem.EncodeToMemory(&pem.Block{Type: keyPEMBlock.Type, Bytes: x509.MarshalPKCS1PrivateKey(decryptedKey)}))
		if err != nil {
			return fmt.Errorf("failed to load cert/key pair: %w", err)
		}
	} else if keyPEMBlock.Headers["Proc-Type"] == "4,ENCRYPTED" {
		// The key is an encrypted private key with the DEK-Info header.
		// This is currently unsupported because of the deprecation of x509.IsEncryptedPEMBlock and x509.DecryptPEMBlock.
		return fmt.Errorf("password-protected keys in pkcs#1 format are not supported")
	} else {
		cert, err = tls.X509KeyPair(certBytes, keyBytes)
		if err != nil {
			return fmt.Errorf("failed to load cert/key pair: %w", err)
		}
	}
	config.Certificates = []tls.Certificate{cert}
	return nil
}
