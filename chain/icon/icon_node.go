package icon

import (
	"context"
	"fmt"
	"path"
	"strings"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	dockerclient "github.com/docker/docker/client"
	"github.com/strangelove-ventures/ibctest/v6/ibc"
	"github.com/strangelove-ventures/ibctest/v6/internal/blockdb"
	"github.com/strangelove-ventures/ibctest/v6/internal/dockerutil"
	"go.uber.org/zap"
)

type IconNode struct {
	VolumeName   string
	Index        int
	Chain        ibc.Chain
	NetworkID    string
	DockerClient *dockerclient.Client
	TestName     string
	Image        ibc.DockerImage
	log          *zap.Logger
	containerID  string
	// Ports set during StartContainer.
	hostRPCPort  string
	hostGRPCPort string
	Validator    bool
	lock         sync.Mutex
	Address      string
}

type IconNodes []*IconNode

// Name of the test node container
func (in *IconNode) Name() string {
	var nodeType string
	if in.Validator {
		nodeType = "val"
	} else {
		nodeType = "fn"
	}
	return fmt.Sprintf("%s-%s-%d-%s", in.Chain.Config().ChainID, nodeType, in.Index, dockerutil.SanitizeContainerName(in.TestName))
}

func (tn *IconNode) CreateKey(ctx context.Context, name string) error {
	tn.lock.Lock()
	defer tn.lock.Unlock()

	_, _, err := tn.ExecBin(ctx,
		"ks", "gen",
		"--password", name,
	)
	return err
}

func (tn *IconNode) ExecBin(ctx context.Context, command ...string) ([]byte, []byte, error) {
	return tn.Exec(ctx, tn.BinCommand(command...), nil)
}

func (tn *IconNode) Exec(ctx context.Context, cmd []string, env []string) ([]byte, []byte, error) {
	job := dockerutil.NewImage(tn.logger(), tn.DockerClient, tn.NetworkID, tn.TestName, tn.Image.Repository, tn.Image.Version)
	opts := dockerutil.ContainerOptions{
		Env:   env,
		Binds: tn.Bind(),
	}
	res := job.Run(ctx, cmd, opts)
	output := string(res.Stdout)
	substr := strings.Split(output, " ")
	tn.Address = substr[0]

	return res.Stdout, res.Stderr, res.Err
}

func (tn *IconNode) BinCommand(command ...string) []string {
	command = append([]string{tn.Chain.Config().Bin}, command...)
	return command
}

func (tn *IconNode) logger() *zap.Logger {
	return tn.log.With(
		zap.String("chain_id", tn.Chain.Config().ChainID),
		zap.String("test", tn.TestName),
	)
}

func (tn *IconNode) Bind() []string {
	return []string{fmt.Sprintf("%s:%s", tn.VolumeName, tn.HomeDir())}
}

func (tn *IconNode) HomeDir() string {
	return path.Join("/var/icon-chain", tn.Chain.Config().Name)
}

func (tn *IconNode) GetBlockByHeight(ctx context.Context, uri string) error {
	// cmd := []string{
	// 	"goloop", "rpc", "blockbyheight", "105", "--uri", uri + "/api/v3",
	// }
	// StdOut, StdErr, Err := tn.Exec(ctx, cmd, nil)
	// fmt.Println(StdOut, StdErr)
	// return Err

	tn.lock.Lock()
	defer tn.lock.Unlock()
	out, _, err := tn.ExecBin(ctx,
		"rpc", "lastblock",
		"--uri", uri+"/api/v3",
	)
	fmt.Println(out)
	return err
}

func (p *IconNode) CreateNodeContainer(ctx context.Context) error {
	cmd := []string{"goloop", "-h"}
	fmt.Printf("{%s} -> '%s'\n", p.Name(), strings.Join(cmd, " "))

	cc, err := p.DockerClient.ContainerCreate(
		ctx,
		&container.Config{
			Image: p.Image.Ref(),

			Entrypoint: []string{},
			Cmd:        cmd,

			Hostname: p.HostName(),
			User:     p.Image.UidGid,

			Labels: map[string]string{dockerutil.CleanupLabel: p.TestName},
		},
		&container.HostConfig{
			Binds:           p.Bind(),
			PublishAllPorts: true,
			AutoRemove:      false,
			DNS:             []string{},
			// PortBindings: nat.PortMap{
			// 	"9080/tcp": {
			// 		nat.PortBinding{
			// 			HostIP:   "127.0.0.1",
			// 			HostPort: "9080",
			// 		},
			// 	},
			// },
		},
		&network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				p.NetworkID: {},
			},
		},
		nil,
		p.Name(),
	)
	if err != nil {
		return err
	}
	p.containerID = cc.ID
	fmt.Println(p.HostName())
	return nil

}

func (p *IconNode) HostName() string {
	return dockerutil.CondenseHostName(p.Name())
}

func (p *IconNode) StartContainer(ctx context.Context) error {
	if err := dockerutil.StartContainer(ctx, p.DockerClient, p.containerID); err != nil {
		return err
	}

	c, err := p.DockerClient.ContainerInspect(ctx, p.containerID)
	if err != nil {
		return err
	}
	a := c.NetworkSettings.Ports
	fmt.Println(a)

	p.hostRPCPort = dockerutil.GetHostPort(c, rpcPort)
	p.hostGRPCPort = dockerutil.GetHostPort(c, grpcPort)
	p.logger().Info("Icon chain node started", zap.String("container", p.Name()), zap.String("rpc_port", p.hostRPCPort))

	return nil
}

const (
	valKey   = "validator"
	rpcPort  = "9080/tcp"
	grpcPort = "7100/tcp"
)

func (tn *IconNode) Height(ctx context.Context) (uint64, error) {
	return 0, nil
}

func (tn *IconNode) FindTxs(ctx context.Context, height uint64) ([]blockdb.Tx, error) {
	return nil, nil
}
