package repl

import (
	"context"
	"fmt"
	"strings"

	"github.com/artilugio0/efin-suite/internal/grpc/proto"
	"github.com/artilugio0/replit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func proxyConnectCmd(ctx context.Context, appState *appState) (*replit.Result, error) {
	if appState.proxyClient != nil {
		return nil, fmt.Errorf("repl is already connected to proxy")
	}

	conn, client, err := proxyConnect(ctx, "efin-repl")
	if err != nil {
		return nil, err
	}

	appState.proxyClient = client
	appState.proxyConn = conn
	return &replit.Result{
		Output: "connection established",
	}, nil
}

func proxyDisconnectCmd(appState *appState) (*replit.Result, error) {
	if appState.proxyClient == nil {
		return nil, fmt.Errorf("repl is not connected to proxy")
	}

	err := appState.proxyConn.Close()
	appState.proxyConn = nil
	appState.proxyClient = nil

	if err != nil {
		return nil, err
	}

	return &replit.Result{
		Output: "connection closed",
	}, nil
}

func proxyGetConfigCmd(ctx context.Context, appState *appState) (*replit.Result, error) {
	if appState.proxyClient == nil {
		return nil, fmt.Errorf("repl is not connected to proxy")
	}

	config, err := appState.proxyClient.GetConfig(ctx, &proto.Null{})
	if err != nil {
		return nil, err
	}

	output := "DB file: " + config.DbFile + "\n"
	output += fmt.Sprintf("Print logs: %t\n", config.PrintLogs)
	output += "Save dir: " + config.SaveDir + "\n"
	output += "Domain regex: " + config.ScopeDomainRe + "\n"
	output += "Excluded extensions: " + strings.Join(config.ScopeExcludedExtensions, ", ")
	return &replit.Result{
		Output: output,
	}, nil
}

func proxySetConfigCmd(ctx context.Context, appState *appState, args []string) (*replit.Result, error) {
	if appState.proxyClient == nil {
		return nil, fmt.Errorf("repl is not connected to proxy")
	}

	config, err := appState.proxyClient.GetConfig(ctx, &proto.Null{})
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(args[0]) {
	case "printlogs", "print-logs":
		config.PrintLogs = strings.ToLower(args[1]) == "true"
	case "dbfile", "db-file", "db":
		config.DbFile = args[1]
	case "save-dir", "savedir":
		config.SaveDir = args[1]
	case "scope", "domain", "domain-regex":
		config.ScopeDomainRe = args[1]
	case "excluded-ext", "eext":
		config.ScopeExcludedExtensions = strings.Split(strings.ReplaceAll(args[1], " ", ""), ",")
	default:
		return nil, fmt.Errorf("invalid config %s", args[1])
	}

	_, err = appState.proxyClient.SetConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("repl is not connected to proxy")
	}

	return &replit.Result{
		Output: "config updated",
	}, nil
}

func proxyUnsetConfigCmd(ctx context.Context, appState *appState, cfg string) (*replit.Result, error) {
	if appState.proxyClient == nil {
		return nil, fmt.Errorf("repl is not connected to proxy")
	}

	config, err := appState.proxyClient.GetConfig(ctx, &proto.Null{})
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(cfg) {
	case "printlogs", "print-logs":
		config.PrintLogs = false
	case "dbfile", "db-file", "db":
		config.DbFile = ""
	case "save-dir", "savedir":
		config.SaveDir = ""
	case "scope", "domain", "domain-regex":
		config.ScopeDomainRe = ""
	case "excluded-ext", "eext":
		config.ScopeExcludedExtensions = []string{}
	default:
		return nil, fmt.Errorf("invalid config %s", cfg)
	}

	_, err = appState.proxyClient.SetConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("repl is not connected to proxy")
	}

	return &replit.Result{
		Output: "config updated",
	}, nil
}

func proxyConnect(ctx context.Context, clientName string) (*grpc.ClientConn, proto.ProxyServiceClient, error) {
	// Connect to gRPC server
	const maxMsgSize = 1024 * 1024 * 1024 // 10MB
	conn, err := grpc.DialContext(
		ctx,
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxMsgSize),
			grpc.MaxCallSendMsgSize(maxMsgSize),
		),
		grpc.WithBlock(), // Forces connection to be established during Dialp
	)

	if err != nil {
		return nil, nil, err
	}

	client := proto.NewProxyServiceClient(conn)
	return conn, client, err
}
