package icon

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	iconclient "github.com/icon-project/icon-bridge/cmd/iconbridge/chain/icon"
	icontypes "github.com/icon-project/icon-bridge/cmd/iconbridge/chain/icon/types"
	"github.com/icon-project/icon-bridge/common/log"
)

type IconTestnet struct {
	Config Config
	Client *iconclient.Client
}

func (it *IconTestnet) LoadConfig() {
	it.Config = LoadConfig(".")
	var l log.Logger
	it.Client = iconclient.NewClient(it.Config.URL, l)

}
func (it IconTestnet) GetLastBlock() (int64, error) {
	res, err := it.Client.GetLastBlock()
	return res.Height, err
}

func (it IconTestnet) GetBlockByHeight(height int64) (string, error) {
	h := icontypes.BlockHeightParam{Height: icontypes.NewHexInt(height)}
	block, _ := it.Client.GetBlockByHeight(&h)
	fmt.Println(block)
	return "", nil
}

// This function takes initMessage, scorePath and keytorePath, Deploys contract to testnet and returns score address
func (it IconTestnet) DeployContract(scorePath, keystorePath, initMessage string) (string, error) {
	var result *icontypes.TransactionResult
	var output string
	// before, _ := it.GetLastBlock()
	hash, _ := exec.Command(it.Config.Bin, "rpc", "sendtx", "deploy", scorePath,
		"--key_store", keystorePath, "--key_password", "gochain", "--step_limit", "5000000000",
		"--content_type", "application/java", "--param", initMessage,
		"--uri", it.Config.URL, "--nid", it.Config.NID).Output()
	json.Unmarshal(hash, &output)
	log.Info("Waitng for few blocks to complete")
	time.Sleep(3 * time.Second)
	out, err := exec.Command(it.Config.Bin, "rpc", "txresult", output, "--uri", it.Config.URL).Output()
	json.Unmarshal(out, &result)
	return string(result.SCOREAddress), err
}

func (it *IconTestnet) QueryContract(scoreAddress, methodName, params string) (string, error) {
	if params != "" {
		output, _ := exec.Command(it.Config.Bin, "rpc", "call", "--to", scoreAddress, "--method", methodName, "--param", params, "--uri", it.Config.URL).Output()
		return string(output), nil
	} else {
		output, _ := exec.Command(it.Config.Bin, "rpc", "call", "--to", scoreAddress, "--method", methodName, "--uri", it.Config.URL).Output()
		return string(output), nil
	}
}

func (it *IconTestnet) ExecuteContract(scoreAddress, keystorePath, methodName, params string) (string, error) {
	var hash string
	output, err := exec.Command(it.Config.Bin, "rpc", "sendtx", "call", "--to", scoreAddress, "--method", methodName, "--key_store", keystorePath,
		"--key_password", "gochain", "--step_limit", "5000000000", "--param", params, "--uri", it.Config.URL, "--nid", it.Config.NID).Output()
	json.Unmarshal(output, &hash)
	return hash, err
}

func (it *IconTestnet) GetTransactionResult(hash string) (*icontypes.TransactionResult, error) {
	var result *icontypes.TransactionResult
	out, err := exec.Command(it.Config.Bin, "rpc", "txresult", hash, "--uri", it.Config.URL).Output()
	json.Unmarshal(out, &result)
	return result, err
}
