package icon

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"interchaintest/internal/blockdb"
	"interchaintest/internal/dockerutil"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	iconclient "github.com/icon-project/icon-bridge/cmd/iconbridge/chain/icon"
	icontypes "github.com/icon-project/icon-bridge/cmd/iconbridge/chain/icon/types"
	iconlog "github.com/icon-project/icon-bridge/common/log"
	"github.com/strangelove-ventures/interchaintest/v6/ibc"
	"github.com/strangelove-ventures/interchaintest/v6/testutil"
	"go.uber.org/zap"
)

type IconNode struct {
	VolumeName   string
	Index        int
	Chain        ibc.Chain
	NetworkID    string
	DockerClient *dockerclient.Client
	Client       iconclient.Client
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
	Config       Config
}

type IconNodes []*IconNode

func (in *IconNode) LoadConfig() {
	in.Config = LoadConfig("/home/dell/practice/ibc-bdd/ibctest/godogs/config.json")
}

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

func (tn *IconNode) CreateKey(ctx context.Context, password string) error {
	tn.lock.Lock()
	defer tn.lock.Unlock()

	_, _, err := tn.ExecBin(ctx,
		"ks", "gen",
		"--password", password,
	)
	return err
}

func (tn *IconNode) ExecBin(ctx context.Context, command ...string) ([]byte, []byte, error) {
	return tn.Exec(ctx, tn.BinCommand(command...), nil)
}

func (tn *IconNode) ExecRPC(ctx context.Context, command []string) ([]byte, []byte, error) {
	job := dockerutil.NewImage(tn.logger(), tn.DockerClient, tn.NetworkID, tn.TestName, tn.Image.Repository, tn.Image.Version)
	opts := dockerutil.ContainerOptions{
		Env:   nil,
		Binds: tn.Bind(),
	}
	res := job.Run(ctx, command, opts)
	if err := testutil.WaitForBlocks(ctx, 2, tn); err != nil {
		return nil, nil, err
	}
	return res.Stdout, res.Stderr, res.Err
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

func (in *IconNode) GetBlockByHeight(ctx context.Context, height int64) (string, error) {
	if in.Config.Environment == "local" {
		in.lock.Lock()
		defer in.lock.Unlock()
		uri := "http://" + in.hostRPCPort + "/api/v3"
		block, _, err := in.ExecBin(ctx,
			"rpc", "blockbyheight", fmt.Sprint(height),
			"--uri", uri,
		)
		fmt.Println(string(block))
		return string(block), err
	} else {
		in.LoadConfig()
		in.lock.Lock()
		defer in.lock.Unlock()
		block, _, err := in.ExecBin(ctx,
			"rpc", "blockbyheight", fmt.Sprint(height),
			"--uri", in.Config.URL,
		)
		fmt.Println(string(block))
		return string(block), err
	}

}

func (p *IconNode) CreateNodeContainer(ctx context.Context) error {
	imageRef := p.Image.Ref()
	containerConfig := &types.ContainerCreateConfig{
		Config: &container.Config{
			Image: imageRef,
			ExposedPorts: nat.PortSet{
				"8080/tcp": {},
				"9080/tcp": {},
			},
			Hostname: p.HostName(),

			Labels: map[string]string{dockerutil.CleanupLabel: p.TestName},
		},
		HostConfig: &container.HostConfig{
			Binds:           p.Bind(),
			PublishAllPorts: true,
			AutoRemove:      false,
			DNS:             []string{},
			PortBindings: nat.PortMap{
				"9080/tcp": {
					nat.PortBinding{
						HostIP:   "172.17.0.1",
						HostPort: "9080",
					},
				},
			},
		},
		NetworkingConfig: &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				p.NetworkID: {},
			},
		},
	}
	cc, err := p.DockerClient.ContainerCreate(ctx, containerConfig.Config, containerConfig.HostConfig, containerConfig.NetworkingConfig, nil, p.Name())
	if err != nil {
		panic(err)
	}
	if err != nil {
		return err
	}
	p.containerID = cc.ID
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
	p.hostRPCPort = dockerutil.GetHostPort(c, rpcPort)
	p.hostGRPCPort = dockerutil.GetHostPort(c, grpcPort)
	p.logger().Info("Icon chain node started", zap.String("container", p.Name()), zap.String("rpc_port", p.hostRPCPort))

	uri := "http://" + p.hostRPCPort + "/api/v3"
	var l iconlog.Logger
	p.Client = *iconclient.NewClient(uri, l)

	return nil
}

const (
	valKey   = "validator"
	rpcPort  = "9080/tcp"
	grpcPort = "7100/tcp"
)

func (tn *IconNode) Height(ctx context.Context) (uint64, error) {
	res, err := tn.Client.GetLastBlock()
	return uint64(res.Height), err
}

var flag = true

func (tn *IconNode) FindTxs(ctx context.Context, height uint64) ([]blockdb.Tx, error) {
	if flag {
		time.Sleep(3 * time.Second)
		flag = false
	}

	time.Sleep(2 * time.Second)
	blockHeight := icontypes.BlockHeightParam{Height: icontypes.NewHexInt(int64(height))}
	res, _ := tn.Client.GetBlockByHeight(&blockHeight)

	txs := make([]blockdb.Tx, 0, len(res.NormalTransactions)+2)
	var newTx blockdb.Tx
	for _, tx := range res.NormalTransactions {
		newTx.Data = []byte(fmt.Sprintf(`{"data":"%s"}`, tx.Data))
	}

	// ToDo Add events from block if any to newTx.Events.
	// Event is an alternative representation of tendermint/abci/types.Event
	return txs, nil
}

func (in *IconNode) GetBalance(ctx context.Context, address string) (int64, error) {
	addr := icontypes.AddressParam{Address: icontypes.Address(address)}
	bal, _ := in.Client.GetBalance(&addr)
	return bal.Int64(), nil
}

func (in *IconNode) GetLastBlock(ctx context.Context, height int64) error {
	in.lock.Lock()
	defer in.lock.Unlock()
	uri := "http://" + in.hostRPCPort + "/api/v3"
	out, _, err := in.ExecBin(ctx,
		"rpc", "lastblock",
		"--uri", uri,
	)
	fmt.Println(string(out))
	return err
}

func (in *IconNode) DeployContract(ctx context.Context, scorePath, keystorePath, initMessage string) (string, error) {

	// Write Contract file to Docker volume
	_, score := filepath.Split(scorePath)
	err := in.CopyFile(ctx, scorePath, score)
	if err != nil {
		return "", fmt.Errorf("error copying keystore to Docker volume: %w", err)
	}

	// Deploy the contract
	hash, err := in.ExecTx(ctx, initMessage, path.Join(in.HomeDir(), score), keystorePath)
	if err != nil {
		return "", err
	}

	time.Sleep(3 * time.Second)
	// Get Score Address
	trResult, _ := in.TransactionResult(ctx, hash)
	return string(trResult.SCOREAddress), err

}

func (in *IconNode) TransactionResult(ctx context.Context, hash string) (icontypes.TransactionResult, error) {
	var result icontypes.TransactionResult
	uri := "http://" + in.hostRPCPort + "/api/v3"
	out, _, _ := in.ExecBin(ctx, "rpc", "txresult", hash, "--uri", uri)
	json.Unmarshal(out, &result)
	return result, nil
}

// ExecTx executes a transaction, waits for 2 blocks if successful, then returns the tx hash.
func (in *IconNode) ExecTx(ctx context.Context, initMessage string, filePath string, keystorePath string, command ...string) (string, error) {
	var output string
	in.lock.Lock()
	defer in.lock.Unlock()
	stdout, _, err := in.Exec(ctx, in.TxCommand(ctx, initMessage, filePath, keystorePath, command...), nil)
	if err != nil {
		return "", err
	}
	json.Unmarshal(stdout, &output)
	return output, nil
}

// TxCommand is a helper to retrieve a full command for broadcasting a tx
// with the chain node binary.
func (in *IconNode) TxCommand(ctx context.Context, initMessage, filePath, keystorePath string, command ...string) []string {
	// Write keystore file to Docker volume
	_, key := filepath.Split(keystorePath)
	err := in.CopyFile(ctx, keystorePath, key)
	if err != nil {
		return []string{"error copying keystore to Docker volume"}
	}
	keystore := path.Join(in.HomeDir(), key)

	command = append([]string{"rpc", "sendtx", "deploy", filePath}, command...)
	return in.NodeCommand(append(command,
		"--key_store", keystore,
		"--key_password", "gochain",
		"--step_limit", "5000000000",
		"--content_type", "application/java",
		"--param", initMessage,
	)...)
}

// NodeCommand is a helper to retrieve a full command for a chain node binary.
// when interactions with the RPC endpoint are necessary.
// For example, if chain node binary is `gaiad`, and desired command is `gaiad keys show key1`,
// pass ("keys", "show", "key1") for command to return the full command.
// Will include additional flags for node URL, home directory, and chain ID.
func (in *IconNode) NodeCommand(command ...string) []string {
	command = in.BinCommand(command...)
	return append(command,
		"--uri", fmt.Sprintf("http://%s/api/v3", in.hostRPCPort),
		"--nid", "0xc5addf",
	)
}

// CopyFile adds a file from the host filesystem to the docker filesystem
// relPath describes the location of the file in the docker volume relative to
// the home directory
func (tn *IconNode) CopyFile(ctx context.Context, srcPath, dstPath string) error {
	content, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}
	return tn.WriteFile(ctx, content, dstPath)
}

// WriteFile accepts file contents in a byte slice and writes the contents to
// the docker filesystem. relPath describes the location of the file in the
// docker volume relative to the home directory
func (tn *IconNode) WriteFile(ctx context.Context, content []byte, relPath string) error {
	fw := dockerutil.NewFileWriter(tn.logger(), tn.DockerClient, tn.TestName)
	return fw.WriteFile(ctx, tn.VolumeName, relPath, content)
}

func (in *IconNode) QueryContract(ctx context.Context, scoreAddress, methodName, params string) (string, error) {
	uri := "http://" + in.hostRPCPort + "/api/v3"
	if params != "" {
		out, _, _ := in.ExecBin(ctx, "rpc", "call", "--to", scoreAddress, "--method", methodName, "--param", params, "--uri", uri)
		return string(out), nil
	} else {
		out, _, _ := in.ExecBin(ctx, "rpc", "call", "--to", scoreAddress, "--method", methodName, "--uri", uri)
		return string(out), nil
	}

}

func (in *IconNode) ExecuteContract(ctx context.Context, scoreAddress, methodName, keyStorePath, params string) (string, error) {
	hash, err := in.ExecCallTx(ctx, scoreAddress, methodName, keyStorePath, params)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func (in *IconNode) ExecCallTx(ctx context.Context, scoreAddress, methodName, keystorePath, params string) (string, error) {
	var output string
	in.lock.Lock()
	defer in.lock.Unlock()
	stdout, _, err := in.Exec(ctx, in.ExecCallTxCommand(ctx, scoreAddress, methodName, keystorePath, params), nil)
	if err != nil {
		return "", err
	}
	json.Unmarshal(stdout, &output)
	return output, nil
}

func (in *IconNode) ExecCallTxCommand(ctx context.Context, scoreAddress, methodName, keystorePath, params string) []string {
	// Write keystore file to Docker volume
	_, key := filepath.Split(keystorePath)
	err := in.CopyFile(ctx, keystorePath, key)
	if err != nil {
		return []string{"error copying keystore to Docker volume"}
	}
	keystore := path.Join(in.HomeDir(), key)
	command := []string{"rpc", "sendtx", "call"}
	return in.NodeCommand(append(command,
		"--to", scoreAddress,
		"--method", methodName,
		"--key_store", keystore,
		"--key_password", "gochain",
		"--step_limit", "5000000000",
		"--param", params,
	)...)
}
