package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/cuckoo"
	"github.com/ethereum/go-ethereum/core/types"
	"net"
	"os"
	"strconv"
	"sync"
)

type ReqObj struct {
	Id      int      `json:"id"` // struct标签， 如果指定，jsonrpc包会在序列化json时，将该聚合字段命名为指定的字符串
	Jsonrpc string   `json:"jsonrpc"`
	Method  string   `json:"method"`
	Params  []string `json:"params"`
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func read(reader *bufio.Reader) map[string]interface{} {
	rep := make([]byte, 0, 4096) // big buffer
	for {
		tmp, isPrefix, err := reader.ReadLine()
		checkError(err)
		rep = append(rep, tmp...)
		if isPrefix == false {
			break
		}
	}
	fmt.Println("received ", len(rep), " bytes: ", string(rep), "\n")
	var repObj map[string]interface{}
	err := json.Unmarshal(rep, &repObj)
	checkError(err)
	return repObj
}

func write(reqObj ReqObj, conn *net.TCPConn) {
	req, err := json.Marshal(reqObj)
	checkError(err)

	req = append(req, uint8('\n'))
	n, err := conn.Write(req)
	checkError(err)

	fmt.Println("write ", n, " bytes: ", string(req))
}

func main() {

	type Step struct {
		lock sync.Mutex
		step uint32
	}
	var step Step
	step.step = 0
	figureout := false

	var THREAD uint = 3
	cuckoo.CuckooInit(THREAD)
	var wg sync.WaitGroup
	for nthread := 0; nthread < int(THREAD); nthread++ {
		wg.Add(1)
		go func() {
			step.lock.Lock()
			step.step += 1
			var curstep uint32 = step.step
			step.lock.Unlock()

			//-------- connect to server -------------
			var server = "139.196.32.192:8009"
			tcpAddr, err := net.ResolveTCPAddr("tcp", server)
			checkError(err)

			conn, err := net.DialTCP("tcp", nil, tcpAddr)
			checkError(err)

			defer conn.Close()

			reader := bufio.NewReader(conn)

			//------- login -------------
			var reqLogin = ReqObj{
				Id:      73,
				Jsonrpc: "2.0",
				Method:  "eth_submitLogin",
				Params:  []string{"0xc3d7a1ef810983847510542edfd5bc5551a6321c"},
			}
			write(reqLogin, conn)
			_ = read(reader)

			//------ get work ------------
			var reqGetwork = ReqObj{
				Id:      73,
				Jsonrpc: "2.0",
				Method:  "eth_getWork",
				Params:  []string{""},
			}
			write(reqGetwork, conn)
			work := read(reader)
			workinfo, _ := work["result"].([]interface{})

			//--------------------- mining -----------------------
			var header [32]uint8
			var start uint32 = 0
			var target [32]uint8

			for i := 0; i < 32; i = i + 1 {
				header[i] = 0
			}

			wr := workinfo[0].(string)
			lenr := len(wr)
			for i, j, k := 2, lenr-2, 31; i < lenr; i, j, k = i+2, j-2, k-1 {
				v, _ := strconv.ParseUint(wr[j:j+2], 16, 8)
				header[k] = uint8(v)
			}

			wr = workinfo[1].(string)
			lenr = len(wr)
			for i, j := 2, lenr-2; i < lenr; i, j = i+2, j-2 {
				v, _ := strconv.ParseUint(wr[j:j+2], 16, 8)
				start = start*256 + uint32(v)
			}
			start += curstep

			wr = workinfo[2].(string)
			lenr = len(wr)
			targetHack := false
			if targetHack {
				for i := 0; i < 32; i++ {
					target[i] = 255
				}
			} else {
				for i, j, k := 2, lenr-2, 31; i < lenr; i, j, k = i+2, j-2, k-1 {
					v, _ := strconv.ParseUint(wr[j:j+2], 16, 8)
					target[k] = uint8(v)
				}
			}
			fmt.Println("target and workInfo: ", workinfo[0], header, workinfo[1], start, workinfo[2], target)

			//------------- solve process -------------------
			var result types.BlockSolution
			for i, _ := range result {
				result[i] = 0
			}
			var result_len uint32
			var result_hash [32]uint8
			var intval uint32 = uint32(THREAD)
			shareTarget := common.HexToHash(workinfo[2].(string))
			var targetMinerTest [32]uint8
			for i := 0; i < 32; i++ {
				targetMinerTest[i] = 255
			}
			for {

				r := cuckoo.CuckooSolve(&header[0], 32, uint32(start), &result[0], &result_len, &targetMinerTest[0], &result_hash[0])
				if byte(r) == 1 {
					figureout = true
				}
				if figureout == true {
					// fmt.Println("result: ", result)
					sha3hash := common.BytesToHash(cuckoo.Sha3Solution(&result))
					// fmt.Println("r, sha3hash = ", r, sha3hash)
					fmt.Println("nonce:", start, ", sha3hash:\n", sha3hash.Big(), "\n", shareTarget.Big())
					if sha3hash.Big().Cmp(shareTarget.Big()) <= 0 {
						break
					}
				}
				start += intval
			}
			// r := cuckoo.CuckooVerify(&header[0], 32, uint32(start), &result[0], &target[0], &result_hash[0])

			nonce := strconv.FormatUint(uint64(start), 16)
			for len(nonce) < 16 {
				nonce = "0" + nonce
			}

			headerst := ""
			for _, val := range header {
				s := strconv.FormatUint(uint64(val), 16)
				if len(s) < 2 {
					s = "0" + s
				}
				headerst += s
			}

			digest := ""
			for _, val := range target {
				s := strconv.FormatUint(uint64(val), 16)
				if len(s) < 2 {
					s = "0" + s
				}
				digest += s
			}
			nonce = "0x" + nonce
			headerst = "0x" + headerst
			digest = "0x" + digest
			solution, sol_err := result.MarshalText()
			if sol_err != nil {
				fmt.Println("sol err")
			}
			fmt.Println("solution: ", solution)
			if headerst != workinfo[0] {
				fmt.Println("error header", header, headerst, workinfo[0])
			}
			var reqSubmit = ReqObj{
				Id:      73,
				Jsonrpc: "2.0",
				Method:  "eth_submitWork",
				Params:  []string{nonce, workinfo[0].(string), digest, string(solution)},
				// Params: []string{nonce, headerst, digest, solution},
			}
			write(reqSubmit, conn)
			_ = read(reader)
			wg.Done()
		}()
	}
	wg.Wait()
	cuckoo.CuckooFinalize()
}