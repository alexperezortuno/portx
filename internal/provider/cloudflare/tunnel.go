package cloudflare

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

	"github.com/cloudflare/cloudflared/client"
	"github.com/cloudflare/cloudflared/config"
	"github.com/cloudflare/cloudflared/connection"
	"github.com/cloudflare/cloudflared/edgediscovery"
	"github.com/cloudflare/cloudflared/edgediscovery/allregions"
	"github.com/cloudflare/cloudflared/features"
	"github.com/cloudflare/cloudflared/ingress"
	"github.com/cloudflare/cloudflared/ingress/origins"
	"github.com/cloudflare/cloudflared/logger"
	"github.com/cloudflare/cloudflared/orchestration"
	"github.com/cloudflare/cloudflared/signal"
	"github.com/cloudflare/cloudflared/supervisor"
	"github.com/cloudflare/cloudflared/tlsconfig"
	"github.com/cloudflare/cloudflared/tunnelrpc/pogs"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func requestQuickTunnel() (*connection.TunnelProperties, error) {
	timeout := 30 * time.Second
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSHandshakeTimeout:   timeout,
			ResponseHeaderTimeout: timeout,
		},
		Timeout: timeout,
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.trycloudflare.com/tunnel", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build tunnel request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "cloudflared/portx")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to request tunnel")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read tunnel response")
	}

	var result CreateTunnelResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal tunnel response")
	}

	if !result.Success {
		for _, e := range result.Errors {
			return nil, fmt.Errorf("cloudflare tunnel error: %s", e.Message)
		}
	}

	tunnelID, err := uuid.Parse(result.Result.ID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse tunnel ID")
	}

	return &connection.TunnelProperties{
		Credentials: connection.Credentials{
			AccountTag:   result.Result.AccountTag,
			TunnelSecret: result.Result.Secret,
			TunnelID:     tunnelID,
		},
		QuickTunnelUrl: result.Result.Hostname,
	}, nil
}

type CreateTunnelResponse struct {
	Success bool             `json:"success"`
	Result  TunnelCreds      `json:"result"`
	Errors  []TunnelAPIError `json:"errors"`
}

type TunnelCreds struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Hostname   string `json:"hostname"`
	AccountTag string `json:"account_tag"`
	Secret     []byte `json:"secret"`
}

type TunnelAPIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func startQuickTunnel(ctx context.Context, port int, logLevel string, errCh chan<- error) (string, error) {
	logTransport := logger.Create(logger.CreateConfig(
		logLevel,
		false,
		false,
		"",
		"",
	))

	observer := connection.NewObserver(logTransport, logTransport)

	featureSelector, err := features.NewFeatureSelector(ctx, "", nil, false, logTransport)
	if err != nil {
		return "", errors.Wrap(err, "can't create feature selector")
	}

	clientConfig, err := client.NewConfig("portx", runtime.GOOS+"_"+runtime.GOARCH, featureSelector)
	if err != nil {
		return "", errors.Wrap(err, "can't create client config")
	}

	ing, err := ingress.ParseIngress(&config.Configuration{
		Ingress: []config.UnvalidatedIngressRule{
			{
				Service: fmt.Sprintf("http://localhost:%d", port),
			},
		},
	})
	if err != nil {
		return "", errors.Wrap(err, "can't parse ingress")
	}

	orchestrator, err := orchestration.NewOrchestrator(
		ctx,
		&orchestration.Config{
			Ingress:             &ing,
			WarpRouting:         ingress.NewWarpRoutingConfig(&config.WarpRoutingConfig{}),
			OriginDialerService: ingress.NewOriginDialer(ingress.OriginConfig{}, logTransport),
			ConfigurationFlags:  map[string]string{},
		},
		[]pogs.Tag{},
		[]ingress.Rule{},
		logTransport,
	)
	if err != nil {
		return "", errors.Wrap(err, "can't create orchestrator")
	}

	connectedSignal := signal.New(make(chan struct{}))
	reconnectCh := make(chan supervisor.ReconnectSignal, 4)

	protocolSelector, err := connection.NewProtocolSelector(
		connection.HTTP2.String(),
		"random value",
		false,
		false,
		edgediscovery.ProtocolPercentage,
		connection.ResolveTTL,
		logTransport,
	)
	if err != nil {
		return "", errors.Wrap(err, "unable to create protocol selector")
	}

	edgeTLSConfigs := make(map[connection.Protocol]*tls.Config, len(connection.ProtocolList))
	for _, p := range connection.ProtocolList {
		tlsSettings := p.TLSSettings()
		if tlsSettings == nil {
			return "", fmt.Errorf("%s has unknown TLS settings", p)
		}
		edgeTLSConfig, err := tlsconfig.CreateTunnelConfig(
			cli.NewContext(cli.NewApp(), &flag.FlagSet{}, nil),
			tlsSettings.ServerName,
		)
		if err != nil {
			return "", errors.Wrap(err, "unable to create TLS config to connect with edge")
		}
		if len(tlsSettings.NextProtos) > 0 {
			edgeTLSConfig.NextProtos = tlsSettings.NextProtos
		}
		edgeTLSConfigs[p] = edgeTLSConfig
	}

	tunnel, err := requestQuickTunnel()
	if err != nil {
		return "", err
	}

	tunnelConfig := &supervisor.TunnelConfig{
		ClientConfig:                        clientConfig,
		GracePeriod:                         30,
		EdgeAddrs:                           []string{},
		Region:                              "",
		EdgeIPVersion:                       allregions.Auto,
		EdgeBindAddr:                        nil,
		HAConnections:                       2,
		IsAutoupdated:                       false,
		LBPool:                              "",
		Tags:                                []pogs.Tag{},
		Log:                                 logTransport,
		LogTransport:                        logTransport,
		Observer:                            observer,
		ReportedVersion:                     "portx",
		Retries:                             5,
		RunFromTerminal:                     true,
		NamedTunnel:                         tunnel,
		ProtocolSelector:                    protocolSelector,
		EdgeTLSConfigs:                      edgeTLSConfigs,
		MaxEdgeAddrRetries:                  8,
		RPCTimeout:                          5 * time.Second,
		WriteStreamTimeout:                  time.Second * 0,
		DisableQUICPathMTUDiscovery:         false,
		QUICConnectionLevelFlowControlLimit: 30 * (1 << 20),
		QUICStreamLevelFlowControlLimit:     6 * (1 << 20),
		ICMPRouterServer:                    nil,
		OriginDNSService:                    origins.NewDNSResolverService(ingress.NewDialer(ingress.WarpRoutingConfig{}), logTransport, &noopMetrics{}),
	}

	shutdown := make(chan struct{})

	go func() {
		if err := supervisor.StartTunnelDaemon(ctx, tunnelConfig, orchestrator, connectedSignal, reconnectCh, shutdown); err != nil {
			select {
			case errCh <- fmt.Errorf("tunnel daemon error: %w", err):
			default:
			}
		}
	}()

	return "https://" + tunnel.QuickTunnelUrl, nil
}

type noopMetrics struct{}

func (noopMetrics) IncrementDNSUDPRequests() {}
func (noopMetrics) IncrementDNSTCPRequests() {}
